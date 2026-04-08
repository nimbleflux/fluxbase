package migrations

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// TransitionService handles migration from imperative to declarative schema
type TransitionService struct {
	declarative *DeclarativeService
	validator   *Validator
	pool        *pgxpool.Pool
}

// NewTransitionService creates a new transition service
func NewTransitionService(declarative *DeclarativeService, pool *pgxpool.Pool) *TransitionService {
	return &TransitionService{
		declarative: declarative,
		validator:   NewValidator(declarative, pool),
		pool:        pool,
	}
}

// TransitionOptions holds options for the transition process
type TransitionOptions struct {
	SchemaDir         string
	KeepOldMigrations bool
	MigrationsDir     string
}

// TransitionResult holds the result of a transition
type TransitionResult struct {
	Success              bool
	SchemaFile           string // Kept for backward compatibility, now points to schema dir
	SchemaFingerprint    string
	LastMigrationVersion int64
	TransitionedAt       time.Time
	Error                error
}

// Transition performs the migration from imperative to declarative schema
func (s *TransitionService) Transition(ctx context.Context, opts TransitionOptions) (*TransitionResult, error) {
	result := &TransitionResult{
		TransitionedAt: time.Now(),
	}

	// Step 1: Validate current state
	log.Info().Msg("Validating current migration state...")
	if err := s.validator.ValidateForTransition(ctx); err != nil {
		result.Error = err
		return result, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Get current migration state
	state, err := s.validator.DetectMigrationState(ctx)
	if err != nil {
		result.Error = err
		return result, fmt.Errorf("failed to detect migration state: %w", err)
	}
	result.LastMigrationVersion = state.LastAppliedVersion

	log.Info().
		Int64("last_version", state.LastAppliedVersion).
		Msg("Current migration state detected")

	// Step 3: Export current schema
	log.Info().Msg("Exporting current schema...")
	schemaDir := opts.SchemaDir
	if schemaDir == "" {
		schemaDir = s.declarative.config.SchemaDir
	}

	if err := s.declarative.Dump(ctx, schemaDir); err != nil {
		result.Error = err
		return result, fmt.Errorf("failed to export schema: %w", err)
	}
	result.SchemaFile = schemaDir

	// Step 4: Calculate fingerprint
	fingerprint, err := s.declarative.CalculateFingerprint()
	if err != nil {
		result.Error = err
		return result, fmt.Errorf("failed to calculate fingerprint: %w", err)
	}
	result.SchemaFingerprint = fingerprint

	// Step 5: Record declarative state
	log.Info().Msg("Recording declarative state...")
	if err := s.recordDeclarativeState(ctx, fingerprint, "transitioned", state.LastAppliedVersion); err != nil {
		result.Error = err
		return result, fmt.Errorf("failed to record declarative state: %w", err)
	}

	// Step 6: Record transition log
	if err := s.recordTransitionLog(ctx, state.LastAppliedVersion, fingerprint); err != nil {
		log.Warn().Err(err).Msg("Failed to record transition log (non-fatal)")
	}

	// Step 7: Optionally remove old migrations directory
	if !opts.KeepOldMigrations && opts.MigrationsDir != "" {
		log.Info().Str("dir", opts.MigrationsDir).Msg("Removing old migrations directory...")
		if err := os.RemoveAll(opts.MigrationsDir); err != nil {
			log.Warn().Err(err).Msg("Failed to remove old migrations directory")
		}
	}

	result.Success = true
	log.Info().
		Str("schema_dir", schemaDir).
		Str("fingerprint", fingerprint[:min(16, len(fingerprint))]+"...").
		Msg("Transition completed successfully")

	return result, nil
}

// recordDeclarativeState records the transition in migrations.declarative_state
func (s *TransitionService) recordDeclarativeState(ctx context.Context, fingerprint, source string, lastVersion int64) error {
	query := `
		INSERT INTO migrations.declarative_state (schema_fingerprint, applied_by, source)
		VALUES ($1, 'transition', $2)
	`
	_, err := s.pool.Exec(ctx, query, fingerprint, source)
	if err != nil {
		return fmt.Errorf("failed to insert declarative state: %w", err)
	}
	return nil
}

// recordTransitionLog records the transition in migrations.transition_log
func (s *TransitionService) recordTransitionLog(ctx context.Context, lastVersion int64, fingerprint string) error {
	query := `
		INSERT INTO migrations.transition_log (from_system, to_system, last_migration_version, schema_fingerprint)
		VALUES ($1, $2, $3, $4)
	`
	_, err := s.pool.Exec(ctx, query, "imperative", "declarative", lastVersion, fingerprint)
	if err != nil {
		return fmt.Errorf("failed to insert transition log: %w", err)
	}
	return nil
}

// IsTransitioned returns true if the database has been transitioned to declarative
func (s *TransitionService) IsTransitioned(ctx context.Context) (bool, error) {
	state, err := s.validator.DetectMigrationState(ctx)
	if err != nil {
		return false, err
	}
	return state.HasDeclarativeState, nil
}

// GetTransitionStatus returns detailed status about the transition
func (s *TransitionService) GetTransitionStatus(ctx context.Context) (*TransitionStatus, error) {
	state, err := s.validator.DetectMigrationState(ctx)
	if err != nil {
		return nil, err
	}

	status := &TransitionStatus{
		HasImperativeMigrations: state.HasImperativeMigrations,
		HasDeclarativeState:     state.HasDeclarativeState,
		LastMigrationVersion:    state.LastAppliedVersion,
	}

	// Get transition details if available
	if state.HasDeclarativeState {
		var transitionedAt time.Time
		var source string
		query := `
			SELECT applied_at, source
			FROM migrations.declarative_state
			ORDER BY applied_at DESC
			LIMIT 1
		`
		err := s.pool.QueryRow(ctx, query).Scan(&transitionedAt, &source)
		if err == nil {
			status.TransitionedAt = &transitionedAt
			status.Source = source
		}
	}

	return status, nil
}

// TransitionStatus holds detailed status about the migration transition
type TransitionStatus struct {
	HasImperativeMigrations bool       `json:"has_imperative_migrations"`
	HasDeclarativeState     bool       `json:"has_declarative_state"`
	LastMigrationVersion    int64      `json:"last_migration_version"`
	TransitionedAt          *time.Time `json:"transitioned_at,omitempty"`
	Source                  string     `json:"source,omitempty"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
