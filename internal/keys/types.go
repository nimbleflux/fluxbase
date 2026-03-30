package keys

import (
	"time"

	"github.com/google/uuid"
)

const (
	KeyTypeAnon          = "anon"
	KeyTypePublishable   = "publishable"
	KeyTypeTenantService = "tenant_service"
	KeyTypeGlobalService = "global_service"
)

const (
	KeyPrefixAnon          = "fb_anon_"
	KeyPrefixPublishable   = "fb_pk_"
	KeyPrefixTenantService = "fb_tsk_"
	KeyPrefixGlobalService = "fb_gsk_"
)

type ServiceKey struct {
	ID                 uuid.UUID  `json:"id"`
	KeyType            string     `json:"key_type"`
	TenantID           *uuid.UUID `json:"tenant_id,omitempty"`
	Name               string     `json:"name"`
	Description        *string    `json:"description,omitempty"`
	KeyHash            string     `json:"-"`
	KeyPrefix          string     `json:"key_prefix"`
	UserID             *uuid.UUID `json:"user_id,omitempty"`
	Scopes             []string   `json:"scopes"`
	AllowedNamespaces  []string   `json:"allowed_namespaces,omitempty"`
	RateLimitPerMinute int        `json:"rate_limit_per_minute"`
	IsActive           bool       `json:"is_active"`
	IsConfigManaged    bool       `json:"is_config_managed"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty"`
	RevokedBy          *uuid.UUID `json:"revoked_by,omitempty"`
	RevocationReason   *string    `json:"revocation_reason,omitempty"`
	DeprecatedAt       *time.Time `json:"deprecated_at,omitempty"`
	GracePeriodEndsAt  *time.Time `json:"grace_period_ends_at,omitempty"`
	ReplacedBy         *uuid.UUID `json:"replaced_by,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	CreatedBy          *uuid.UUID `json:"created_by,omitempty"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
}

type KeyUsage struct {
	ID             uuid.UUID `json:"id"`
	KeyID          uuid.UUID `json:"key_id"`
	Endpoint       string    `json:"endpoint"`
	Method         string    `json:"method"`
	StatusCode     *int      `json:"status_code,omitempty"`
	ResponseTimeMs *int      `json:"response_time_ms,omitempty"`
	IPAddress      *string   `json:"ip_address,omitempty"`
	UserAgent      *string   `json:"user_agent,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
