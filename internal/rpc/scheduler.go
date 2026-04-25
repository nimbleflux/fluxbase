package rpc

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/scheduler"
)

type Scheduler struct {
	inner    *scheduler.CronScheduler
	storage  *Storage
	executor *Executor
}

func NewScheduler(storage *Storage, executor *Executor) *Scheduler {
	return &Scheduler{
		inner:    scheduler.NewCronScheduler(10),
		storage:  storage,
		executor: executor,
	}
}

func (s *Scheduler) Start() error {
	loader := func(ctx context.Context) ([]scheduler.Schedulable, error) {
		procedures, err := s.storage.ListScheduledProcedures(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]scheduler.Schedulable, 0, len(procedures))
		for _, proc := range procedures {
			schedule := ""
			if proc.Schedule != nil {
				schedule = *proc.Schedule
			}
			items = append(items, scheduler.Schedulable{
				Key:      proc.Namespace + "/" + proc.Name,
				Schedule: schedule,
				Enabled:  proc.Enabled,
			})
		}
		return items, nil
	}
	return s.inner.Start(loader, "RPC")
}

func (s *Scheduler) Stop() {
	s.inner.Stop("RPC")
}

func (s *Scheduler) ScheduleProcedure(proc *Procedure) error {
	if proc.Schedule == nil || *proc.Schedule == "" {
		return nil
	}

	procKey := proc.Namespace + "/" + proc.Name
	procName := proc.Name
	procNamespace := proc.Namespace

	_, err := s.inner.AddFunc(procKey, *proc.Schedule, func() {
		s.executeScheduledProcedure(procName, procNamespace)
	})
	if err != nil {
		return err
	}

	log.Info().
		Str("procedure", proc.Name).
		Str("namespace", proc.Namespace).
		Str("schedule", *proc.Schedule).
		Msg("Procedure scheduled successfully")

	return nil
}

func (s *Scheduler) UnscheduleProcedure(namespace, name string) {
	procKey := namespace + "/" + name
	if s.inner.IsScheduled(procKey) {
		s.inner.Remove(procKey)
		log.Info().Str("procedure", name).Str("namespace", namespace).Msg("Procedure unscheduled")
	}
}

func (s *Scheduler) RescheduleProcedure(proc *Procedure) error {
	s.UnscheduleProcedure(proc.Namespace, proc.Name)
	if proc.Enabled && proc.Schedule != nil && *proc.Schedule != "" {
		return s.ScheduleProcedure(proc)
	}
	return nil
}

func (s *Scheduler) executeScheduledProcedure(procName, procNamespace string) {
	if !s.inner.Guard.Acquire(procName) {
		return
	}
	defer s.inner.Guard.Release()

	proc, err := s.storage.GetProcedureByName(s.inner.Context(), procNamespace, procName)
	if err != nil || proc == nil {
		log.Error().Err(err).Str("procedure", procName).Msg("Failed to fetch procedure for scheduled execution")
		return
	}

	if !proc.Enabled {
		log.Debug().Str("procedure", procName).Msg("Skipping scheduled execution - procedure disabled")
		return
	}

	log.Info().
		Str("procedure", proc.Name).
		Str("namespace", proc.Namespace).
		Str("trigger", "cron").
		Msg("Executing scheduled procedure")

	execCtx := &ExecuteContext{
		Procedure: proc,
		Params: map[string]interface{}{
			"_trigger":      "cron",
			"_scheduled_at": time.Now().UTC().Format(time.RFC3339),
		},
		UserID:               "",
		UserRole:             "service_role",
		IsAsync:              false,
		DisableExecutionLogs: proc.DisableExecutionLogs,
	}

	result, err := s.executor.Execute(s.inner.Context(), execCtx)
	if err != nil {
		log.Error().Err(err).Str("procedure", proc.Name).Msg("Scheduled execution failed")
		return
	}

	log.Info().
		Str("procedure", proc.Name).
		Str("execution_id", result.ExecutionID).
		Str("status", string(result.Status)).
		Msg("Scheduled execution completed")
}

func (s *Scheduler) GetScheduledProcedures() []scheduler.ScheduledEntryInfo {
	return s.inner.GetScheduledEntries()
}

func (s *Scheduler) IsScheduled(namespace, name string) bool {
	procKey := namespace + "/" + name
	return s.inner.IsScheduled(procKey)
}

func (s *Scheduler) GetScheduleInfo(namespace, name string) (string, bool) {
	procKey := namespace + "/" + name
	return s.inner.GetScheduleInfo(procKey)
}
