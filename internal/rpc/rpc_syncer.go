package rpc

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	syncframework "github.com/nimbleflux/fluxbase/internal/sync"
)

type procedureSyncItem struct {
	loaded *LoadedProcedure
}

func (p procedureSyncItem) GetName() string {
	return p.loaded.Name
}

type rpcSyncer struct {
	handler  *Handler
	syncCtx  context.Context
	tenantID string
	source   string

	existing map[string]*Procedure
}

func newRPCSyncer(h *Handler, syncCtx context.Context, tenantID, source string) *rpcSyncer {
	return &rpcSyncer{
		handler:  h,
		syncCtx:  syncCtx,
		tenantID: tenantID,
		source:   source,
	}
}

func (s *rpcSyncer) ListExisting(ctx context.Context, opts syncframework.Options) (map[string]string, error) {
	procs, err := s.handler.storage.ListProceduresForSync(s.syncCtx, opts.Namespace, s.tenantID)
	if err != nil {
		return nil, err
	}

	s.existing = make(map[string]*Procedure, len(procs))
	result := make(map[string]string, len(procs))
	for _, p := range procs {
		s.existing[p.Name] = p
		result[p.Name] = p.ID
	}
	return result, nil
}

func (s *rpcSyncer) IsChanged(ctx context.Context, existingID string, item procedureSyncItem, opts syncframework.Options) (bool, error) {
	existing, ok := s.existing[item.loaded.Name]
	if !ok {
		return true, nil
	}
	proc := item.loaded.ToProcedure()
	proc.Namespace = opts.Namespace
	proc.Source = s.source
	return s.handler.needsUpdate(existing, proc), nil
}

func (s *rpcSyncer) Preprocess(ctx context.Context, item procedureSyncItem) error {
	proc := item.loaded.ToProcedure()
	validationResult := s.handler.validator.ValidateSQL(proc.SQLQuery, proc.AllowedTables, proc.AllowedSchemas)
	if !validationResult.Valid {
		return fmt.Errorf("SQL validation failed: %v", validationResult.Errors)
	}
	return nil
}

func (s *rpcSyncer) Create(ctx context.Context, item procedureSyncItem, opts syncframework.Options) error {
	proc := item.loaded.ToProcedure()
	proc.Namespace = opts.Namespace
	proc.Source = s.source

	if err := s.handler.storage.CreateProcedure(ctx, proc); err != nil {
		return err
	}

	if s.handler.scheduler != nil && proc.Schedule != nil && *proc.Schedule != "" {
		if err := s.handler.scheduler.ScheduleProcedure(proc); err != nil {
			log.Warn().Err(err).Str("procedure", proc.Name).Msg("Failed to schedule new procedure")
		}
	}
	return nil
}

func (s *rpcSyncer) Update(ctx context.Context, item procedureSyncItem, existingID string, opts syncframework.Options) error {
	proc := item.loaded.ToProcedure()
	proc.Namespace = opts.Namespace
	proc.Source = s.source
	proc.ID = existingID

	if err := s.handler.storage.UpdateProcedureForSync(s.syncCtx, s.tenantID, proc); err != nil {
		return err
	}

	if s.handler.scheduler != nil {
		if err := s.handler.scheduler.RescheduleProcedure(proc); err != nil {
			log.Warn().Err(err).Str("procedure", proc.Name).Msg("Failed to reschedule procedure")
		}
	}
	return nil
}

func (s *rpcSyncer) Delete(ctx context.Context, name string, existingID string, opts syncframework.Options) (bool, error) {
	existing, ok := s.existing[name]
	if !ok || existing.Source == "api" {
		return false, nil
	}

	if s.handler.scheduler != nil {
		s.handler.scheduler.UnscheduleProcedure(existing.Namespace, name)
	}

	if err := s.handler.storage.DeleteProcedureForSync(s.syncCtx, s.tenantID, existingID); err != nil {
		return false, err
	}
	return true, nil
}

func (s *rpcSyncer) PostSync(ctx context.Context, result *syncframework.Result, opts syncframework.Options) error {
	return nil
}
