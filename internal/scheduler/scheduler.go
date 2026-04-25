package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

var StandardParser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

type ScheduledEntryInfo struct {
	Key     string    `json:"key"`
	EntryID int       `json:"entry_id"`
	NextRun time.Time `json:"next_run"`
	PrevRun time.Time `json:"prev_run"`
}

type ConcurrencyGuard struct {
	mu            sync.Mutex
	activeCount   int
	MaxConcurrent int
}

func (g *ConcurrencyGuard) Acquire(name string) bool {
	g.mu.Lock()
	if g.activeCount >= g.MaxConcurrent {
		active := g.activeCount
		max := g.MaxConcurrent
		g.mu.Unlock()
		log.Warn().
			Str("name", name).
			Int("active", active).
			Int("max", max).
			Msg("Skipping scheduled execution - concurrent limit reached")
		return false
	}
	g.activeCount++
	g.mu.Unlock()
	return true
}

func (g *ConcurrencyGuard) Release() {
	g.mu.Lock()
	g.activeCount--
	g.mu.Unlock()
}

type CronScheduler struct {
	cron    *cron.Cron
	entries map[string]cron.EntryID
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	Guard   *ConcurrencyGuard
}

type LoadFunc func(ctx context.Context) ([]Schedulable, error)

type Schedulable struct {
	Key      string
	Schedule string
	Enabled  bool
}

func NewCronScheduler(maxConcurrent int) *CronScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &CronScheduler{
		cron:    cron.New(cron.WithParser(StandardParser)),
		entries: make(map[string]cron.EntryID),
		ctx:     ctx,
		cancel:  cancel,
		Guard:   &ConcurrencyGuard{MaxConcurrent: maxConcurrent},
	}
}

func (s *CronScheduler) Context() context.Context {
	return s.ctx
}

func (s *CronScheduler) Start(loader LoadFunc, label string) error {
	log.Info().Str("scheduler", label).Msg("Starting scheduler")
	s.cron.Start()

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error().Interface("panic", rec).Str("scheduler", label).Msg("Panic in scheduler async loader - recovered")
			}
		}()

		maxRetries := 5
		retryDelay := 100 * time.Millisecond

		for attempt := 1; attempt <= maxRetries; attempt++ {
			ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
			items, err := loader(ctx)
			cancel()

			if err != nil {
				if attempt < maxRetries {
					log.Debug().Err(err).Int("attempt", attempt).Int("max_retries", maxRetries).Dur("retry_delay", retryDelay).Str("scheduler", label).Msg("Failed to load scheduled items, retrying")
					time.Sleep(retryDelay)
					retryDelay *= 2
					continue
				}
				log.Error().Err(err).Str("scheduler", label).Msg("Failed to load scheduled items after all retries")
				return
			}

			for _, item := range items {
				if item.Enabled && item.Schedule != "" {
					if _, addErr := s.AddFunc(item.Key, item.Schedule, nil); addErr != nil {
						log.Error().Err(addErr).Str("key", item.Key).Str("schedule", item.Schedule).Str("scheduler", label).Msg("Failed to schedule item")
					}
				}
			}

			log.Info().Int("scheduled", len(s.entries)).Str("scheduler", label).Msg("Scheduler started successfully")
			return
		}
	}()

	return nil
}

func (s *CronScheduler) Stop(label string) {
	log.Info().Str("scheduler", label).Msg("Stopping scheduler")
	s.cancel()

	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
		log.Info().Str("scheduler", label).Msg("All scheduled executions completed")
	case <-time.After(30 * time.Second):
		log.Warn().Str("scheduler", label).Msg("Scheduler shutdown timeout")
	}
}

func (s *CronScheduler) AddFunc(key, schedule string, fn func()) (cron.EntryID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingID, exists := s.entries[key]; exists {
		s.cron.Remove(existingID)
		delete(s.entries, key)
	}

	entryID, err := s.cron.AddFunc(schedule, fn)
	if err != nil {
		return 0, err
	}

	s.entries[key] = entryID
	return entryID, nil
}

func (s *CronScheduler) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entries[key]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, key)
	}
}

func (s *CronScheduler) IsScheduled(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.entries[key]
	return exists
}

func (s *CronScheduler) GetScheduleInfo(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entryID, exists := s.entries[key]
	if !exists {
		return "", false
	}

	entry := s.cron.Entry(entryID)
	if entry.Next.IsZero() {
		return "Not scheduled", true
	}
	return entry.Next.Format(time.RFC3339), true
}

func (s *CronScheduler) GetScheduledEntries() []ScheduledEntryInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]ScheduledEntryInfo, 0, len(s.entries))
	for key, entryID := range s.entries {
		entry := s.cron.Entry(entryID)
		entries = append(entries, ScheduledEntryInfo{
			Key:     key,
			EntryID: int(entryID),
			NextRun: entry.Next,
			PrevRun: entry.Prev,
		})
	}
	return entries
}

func (s *CronScheduler) EntryCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
