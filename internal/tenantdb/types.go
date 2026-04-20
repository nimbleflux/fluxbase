package tenantdb

import "time"

type TenantStatus string

const (
	TenantStatusCreating TenantStatus = "creating"
	TenantStatusActive   TenantStatus = "active"
	TenantStatusDeleting TenantStatus = "deleting"
	TenantStatusError    TenantStatus = "error"
)

type Tenant struct {
	ID        string         `json:"id"`
	Slug      string         `json:"slug"`
	Name      string         `json:"name"`
	DBName    *string        `json:"db_name"`
	IsDefault bool           `json:"is_default"`
	Status    TenantStatus   `json:"status"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at,omitempty"`
}

func (t *Tenant) UsesMainDatabase() bool {
	return t.DBName == nil || *t.DBName == ""
}

type Config struct {
	Enabled        bool             `mapstructure:"enabled"`
	DatabasePrefix string           `mapstructure:"database_prefix"`
	MaxTenants     int              `mapstructure:"max_tenants"`
	Pool           PoolConfig       `mapstructure:"pool"`
	Migrations     MigrationsConfig `mapstructure:"migrations"`
}

type PoolConfig struct {
	MaxTotalConnections int32         `mapstructure:"max_total_connections"`
	EvictionAge         time.Duration `mapstructure:"eviction_age"`
}

type MigrationsConfig struct {
	CheckInterval time.Duration `mapstructure:"check_interval"`
	OnCreate      bool          `mapstructure:"on_create"`
	OnAccess      bool          `mapstructure:"on_access"`
	Background    bool          `mapstructure:"background"`
}

func DefaultConfig() Config {
	return Config{
		Enabled:        true,
		DatabasePrefix: "tenant_",
		MaxTenants:     100,
		Pool: PoolConfig{
			MaxTotalConnections: 100,
			EvictionAge:         30 * time.Minute,
		},
		Migrations: MigrationsConfig{
			CheckInterval: 5 * time.Minute,
			OnCreate:      true, // Run system migrations after bootstrap on tenant creation
			OnAccess:      false,
			Background:    false,
		},
	}
}

type CreateTenantRequest struct {
	Slug     string         `json:"slug"`
	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata,omitempty"`
	DBMode   string         `json:"db_mode,omitempty"` // "auto" (default) creates new DB, "existing" uses DBName
	DBName   *string        `json:"db_name,omitempty"` // Required when DBMode is "existing"
}

type UpdateTenantRequest struct {
	Name     *string        `json:"name,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type TenantAdminAssignment struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	UserID     string    `json:"user_id"`
	AssignedBy *string   `json:"assigned_by,omitempty"`
	AssignedAt time.Time `json:"assigned_at"`
}
