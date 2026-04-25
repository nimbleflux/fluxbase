package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/scheduler"
)

type ScheduleConfig struct {
	CronExpression string                 `json:"cron_expression"`
	Params         map[string]interface{} `json:"params,omitempty"`
}

type Scheduler struct {
	inner   *scheduler.CronScheduler
	storage *Storage
}

func NewScheduler(db *database.Connection) *Scheduler {
	return &Scheduler{
		inner:   scheduler.NewCronScheduler(20),
		storage: NewStorage(db),
	}
}

func (s *Scheduler) Start() error {
	return s.inner.Start(func(ctx context.Context) ([]scheduler.Schedulable, error) {
		functions, err := s.storage.ListAllScheduledJobFunctions(ctx)
		if err != nil {
			return nil, err
		}

		for _, fn := range functions {
			if err := s.ScheduleJob(fn); err != nil {
				log.Error().
					Err(err).
					Str("job", fn.Name).
					Str("namespace", fn.Namespace).
					Str("schedule", *fn.Schedule).
					Msg("Failed to schedule job")
			}
		}

		return nil, nil
	}, "jobs")
}

func (s *Scheduler) Stop() {
	s.inner.Stop("jobs")
}

func (s *Scheduler) ScheduleJob(fn *JobFunctionSummary) error {
	if fn.Schedule == nil || *fn.Schedule == "" {
		return nil
	}

	scheduleConfig := s.parseScheduleConfig(*fn.Schedule)

	if err := scheduler.ValidateCronSchedule(scheduleConfig.CronExpression); err != nil {
		log.Error().
			Err(err).
			Str("job", fn.Name).
			Str("schedule", scheduleConfig.CronExpression).
			Msg("Cron schedule validation failed")
		return err
	}

	jobKey := fn.Namespace + "/" + fn.Name
	jobName := fn.Name
	jobNamespace := fn.Namespace
	jobTenantID := fn.TenantID
	scheduleParams := scheduleConfig.Params

	entryID, err := s.inner.AddFunc(jobKey, scheduleConfig.CronExpression, func() {
		s.enqueueScheduledJob(jobName, jobNamespace, jobTenantID, scheduleParams)
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("job", fn.Name).
			Str("schedule", scheduleConfig.CronExpression).
			Msg("Failed to add cron schedule")
		return err
	}

	log.Info().
		Str("job", fn.Name).
		Str("namespace", jobNamespace).
		Str("schedule", scheduleConfig.CronExpression).
		Interface("params", scheduleParams).
		Uint("entry_id", uint(entryID)).
		Msg("Job scheduled successfully")

	return nil
}

func (s *Scheduler) ScheduleJobFunction(fn *JobFunction) error {
	summary := &JobFunctionSummary{
		ID:                     fn.ID,
		Name:                   fn.Name,
		Namespace:              fn.Namespace,
		Enabled:                fn.Enabled,
		Schedule:               fn.Schedule,
		TimeoutSeconds:         fn.TimeoutSeconds,
		MemoryLimitMB:          fn.MemoryLimitMB,
		MaxRetries:             fn.MaxRetries,
		ProgressTimeoutSeconds: fn.ProgressTimeoutSeconds,
		AllowNet:               fn.AllowNet,
		AllowEnv:               fn.AllowEnv,
		AllowRead:              fn.AllowRead,
		AllowWrite:             fn.AllowWrite,
		RequireRoles:           fn.RequireRoles,
		Source:                 fn.Source,
	}
	return s.ScheduleJob(summary)
}

func (s *Scheduler) UnscheduleJob(namespace, jobName string) {
	jobKey := namespace + "/" + jobName
	if s.inner.IsScheduled(jobKey) {
		s.inner.Remove(jobKey)
		log.Info().Str("job", jobName).Str("namespace", namespace).Msg("Job unscheduled")
	}
}

func (s *Scheduler) RescheduleJob(fn *JobFunctionSummary) error {
	s.UnscheduleJob(fn.Namespace, fn.Name)
	if fn.Enabled && fn.Schedule != nil && *fn.Schedule != "" {
		return s.ScheduleJob(fn)
	}
	return nil
}

func (s *Scheduler) parseScheduleConfig(schedule string) ScheduleConfig {
	config := ScheduleConfig{
		CronExpression: schedule,
		Params:         make(map[string]interface{}),
	}

	for i := len(schedule) - 1; i >= 0; i-- {
		if schedule[i] == '|' {
			config.CronExpression = schedule[:i]
			paramsJSON := schedule[i+1:]
			if err := json.Unmarshal([]byte(paramsJSON), &config.Params); err != nil {
				log.Warn().
					Err(err).
					Str("params", paramsJSON).
					Msg("Failed to parse schedule params, using cron expression only")
				config.CronExpression = schedule
				config.Params = make(map[string]interface{})
			}
			break
		}
	}

	return config
}

func (s *Scheduler) enqueueScheduledJob(jobName, jobNamespace, tenantID string, params map[string]interface{}) {
	if !s.inner.Guard.Acquire(jobName) {
		return
	}
	defer s.inner.Guard.Release()

	ctx := s.inner.Context()

	fn, err := s.storage.GetJobFunction(ctx, jobNamespace, jobName)
	if err != nil {
		if tenantID != "" {
			tenantCtx := database.ContextWithTenant(ctx, tenantID)
			fn, err = s.storage.GetJobFunction(tenantCtx, jobNamespace, jobName)
			if err != nil {
				log.Error().
					Err(err).
					Str("job", jobName).
					Str("namespace", jobNamespace).
					Msg("Failed to fetch job function for scheduled execution")
				return
			}
		} else {
			log.Error().
				Err(err).
				Str("job", jobName).
				Str("namespace", jobNamespace).
				Msg("Failed to fetch job function for scheduled execution")
			return
		}
	}

	if !fn.Enabled {
		log.Debug().
			Str("job", jobName).
			Msg("Skipping scheduled job - function is disabled")
		return
	}

	log.Info().
		Str("job", fn.Name).
		Str("namespace", fn.Namespace).
		Str("trigger", "cron").
		Interface("params", params).
		Msg("Enqueuing scheduled job")

	payload := map[string]interface{}{
		"_trigger":      "cron",
		"_scheduled_at": time.Now().UTC().Format(time.RFC3339),
	}

	for k, v := range params {
		payload[k] = v
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Error().
			Err(err).
			Str("job", fn.Name).
			Msg("Failed to marshal scheduled job payload")
		return
	}
	payloadStr := string(payloadJSON)

	job := &Job{
		ID:                     uuid.New(),
		Namespace:              fn.Namespace,
		JobFunctionID:          &fn.ID,
		JobName:                fn.Name,
		Status:                 JobStatusPending,
		Payload:                &payloadStr,
		Priority:               0,
		MaxRetries:             fn.MaxRetries,
		ProgressTimeoutSeconds: &fn.ProgressTimeoutSeconds,
		MaxDurationSeconds:     &fn.TimeoutSeconds,
		TenantID:               tenantID,
	}

	enqueueCtx := ctx
	if tenantID != "" {
		enqueueCtx = database.ContextWithTenant(ctx, tenantID)
	}

	if err := s.storage.EnqueueJob(enqueueCtx, job); err != nil {
		log.Error().
			Err(err).
			Str("job", fn.Name).
			Msg("Failed to enqueue scheduled job")
		return
	}

	log.Info().
		Str("job", fn.Name).
		Str("job_id", job.ID.String()).
		Str("namespace", fn.Namespace).
		Msg("Scheduled job enqueued successfully")
}

func (s *Scheduler) GetScheduledJobs() []scheduler.ScheduledEntryInfo {
	return s.inner.GetScheduledEntries()
}

func (s *Scheduler) GetScheduleInfo(namespace, jobName string) (string, bool) {
	return s.inner.GetScheduleInfo(namespace + "/" + jobName)
}

func (s *Scheduler) IsScheduled(namespace, jobName string) bool {
	return s.inner.IsScheduled(namespace + "/" + jobName)
}
