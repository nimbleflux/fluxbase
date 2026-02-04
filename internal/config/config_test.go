package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: ServerConfig{
				Address:      ":8080",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				BodyLimit:    1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "empty address",
			config: ServerConfig{
				Address:      "",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				BodyLimit:    1024 * 1024,
			},
			wantErr: true,
			errMsg:  "server address cannot be empty",
		},
		{
			name: "zero read timeout",
			config: ServerConfig{
				Address:      ":8080",
				ReadTimeout:  0,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				BodyLimit:    1024 * 1024,
			},
			wantErr: true,
			errMsg:  "read_timeout must be positive",
		},
		{
			name: "negative write timeout",
			config: ServerConfig{
				Address:      ":8080",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: -1 * time.Second,
				IdleTimeout:  60 * time.Second,
				BodyLimit:    1024 * 1024,
			},
			wantErr: true,
			errMsg:  "write_timeout must be positive",
		},
		{
			name: "zero idle timeout",
			config: ServerConfig{
				Address:      ":8080",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  0,
				BodyLimit:    1024 * 1024,
			},
			wantErr: true,
			errMsg:  "idle_timeout must be positive",
		},
		{
			name: "zero body limit",
			config: ServerConfig{
				Address:      ":8080",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				BodyLimit:    0,
			},
			wantErr: true,
			errMsg:  "body_limit must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConfig_Validate(t *testing.T) {
	validConfig := func() DatabaseConfig {
		return DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "password",
			Database:        "fluxbase",
			SSLMode:         "disable",
			MaxConnections:  50,
			MinConnections:  10,
			MaxConnLifetime: time.Hour,
			MaxConnIdleTime: 30 * time.Minute,
			HealthCheck:     time.Minute,
		}
	}

	tests := []struct {
		name    string
		modify  func(*DatabaseConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			modify:  func(c *DatabaseConfig) {},
			wantErr: false,
		},
		{
			name:    "empty host",
			modify:  func(c *DatabaseConfig) { c.Host = "" },
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name:    "invalid port - zero",
			modify:  func(c *DatabaseConfig) { c.Port = 0 },
			wantErr: true,
			errMsg:  "database port must be between 1 and 65535",
		},
		{
			name:    "invalid port - too high",
			modify:  func(c *DatabaseConfig) { c.Port = 70000 },
			wantErr: true,
			errMsg:  "database port must be between 1 and 65535",
		},
		{
			name:    "empty user",
			modify:  func(c *DatabaseConfig) { c.User = "" },
			wantErr: true,
			errMsg:  "database user is required",
		},
		{
			name:    "empty database name",
			modify:  func(c *DatabaseConfig) { c.Database = "" },
			wantErr: true,
			errMsg:  "database name is required",
		},
		{
			name:    "invalid ssl mode",
			modify:  func(c *DatabaseConfig) { c.SSLMode = "invalid" },
			wantErr: true,
			errMsg:  "invalid ssl_mode",
		},
		{
			name:    "valid ssl mode - require",
			modify:  func(c *DatabaseConfig) { c.SSLMode = "require" },
			wantErr: false,
		},
		{
			name:    "valid ssl mode - verify-full",
			modify:  func(c *DatabaseConfig) { c.SSLMode = "verify-full" },
			wantErr: false,
		},
		{
			name:    "zero max connections",
			modify:  func(c *DatabaseConfig) { c.MaxConnections = 0 },
			wantErr: true,
			errMsg:  "max_connections must be positive",
		},
		{
			name:    "negative min connections",
			modify:  func(c *DatabaseConfig) { c.MinConnections = -1 },
			wantErr: true,
			errMsg:  "min_connections cannot be negative",
		},
		{
			name: "max less than min",
			modify: func(c *DatabaseConfig) {
				c.MaxConnections = 5
				c.MinConnections = 10
			},
			wantErr: true,
			errMsg:  "max_connections",
		},
		{
			name:    "admin user defaults to user",
			modify:  func(c *DatabaseConfig) { c.AdminUser = "" },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.modify(&config)
			err := config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConfig_ConnectionStrings(t *testing.T) {
	config := DatabaseConfig{
		Host:          "localhost",
		Port:          5432,
		User:          "app_user",
		Password:      "app_pass",
		AdminUser:     "admin_user",
		AdminPassword: "admin_pass",
		Database:      "testdb",
		SSLMode:       "disable",
	}

	t.Run("RuntimeConnectionString", func(t *testing.T) {
		connStr := config.RuntimeConnectionString()
		assert.Contains(t, connStr, "app_user")
		assert.Contains(t, connStr, "app_pass")
		assert.Contains(t, connStr, "localhost:5432")
		assert.Contains(t, connStr, "testdb")
	})

	t.Run("AdminConnectionString", func(t *testing.T) {
		connStr := config.AdminConnectionString()
		assert.Contains(t, connStr, "admin_user")
		assert.Contains(t, connStr, "admin_pass")
		assert.Contains(t, connStr, "localhost:5432")
	})

	t.Run("AdminConnectionString falls back to User when AdminUser empty", func(t *testing.T) {
		config.AdminUser = ""
		config.AdminPassword = ""
		connStr := config.AdminConnectionString()
		assert.Contains(t, connStr, "app_user")
		assert.Contains(t, connStr, "app_pass")
	})

	t.Run("ConnectionString is deprecated alias for RuntimeConnectionString", func(t *testing.T) {
		config.AdminUser = "admin"
		assert.Equal(t, config.RuntimeConnectionString(), config.ConnectionString())
	})
}

func TestAuthConfig_Validate(t *testing.T) {
	validConfig := func() AuthConfig {
		return AuthConfig{
			JWTSecret:           "this-is-a-very-secure-secret-key-for-testing-purposes",
			JWTExpiry:           15 * time.Minute,
			RefreshExpiry:       7 * 24 * time.Hour,
			MagicLinkExpiry:     15 * time.Minute,
			PasswordResetExpiry: time.Hour,
			PasswordMinLen:      8,
			BcryptCost:          10,
		}
	}

	tests := []struct {
		name    string
		modify  func(*AuthConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			modify:  func(c *AuthConfig) {},
			wantErr: false,
		},
		{
			name:    "empty jwt secret",
			modify:  func(c *AuthConfig) { c.JWTSecret = "" },
			wantErr: true,
			errMsg:  "jwt_secret is required",
		},
		{
			name:    "insecure default jwt secret",
			modify:  func(c *AuthConfig) { c.JWTSecret = "your-secret-key-change-in-production" },
			wantErr: true,
			errMsg:  "please set a secure JWT secret",
		},
		{
			name:    "zero jwt expiry",
			modify:  func(c *AuthConfig) { c.JWTExpiry = 0 },
			wantErr: true,
			errMsg:  "jwt_expiry must be positive",
		},
		{
			name:    "zero refresh expiry",
			modify:  func(c *AuthConfig) { c.RefreshExpiry = 0 },
			wantErr: true,
			errMsg:  "refresh_expiry must be positive",
		},
		{
			name:    "zero magic link expiry",
			modify:  func(c *AuthConfig) { c.MagicLinkExpiry = 0 },
			wantErr: true,
			errMsg:  "magic_link_expiry must be positive",
		},
		{
			name:    "zero password reset expiry",
			modify:  func(c *AuthConfig) { c.PasswordResetExpiry = 0 },
			wantErr: true,
			errMsg:  "password_reset_expiry must be positive",
		},
		{
			name:    "zero password min length",
			modify:  func(c *AuthConfig) { c.PasswordMinLen = 0 },
			wantErr: true,
			errMsg:  "password_min_length must be at least 1",
		},
		{
			name:    "bcrypt cost too low",
			modify:  func(c *AuthConfig) { c.BcryptCost = 3 },
			wantErr: true,
			errMsg:  "bcrypt_cost must be between 4 and 31",
		},
		{
			name:    "bcrypt cost too high",
			modify:  func(c *AuthConfig) { c.BcryptCost = 32 },
			wantErr: true,
			errMsg:  "bcrypt_cost must be between 4 and 31",
		},
		{
			name:    "bcrypt cost valid minimum",
			modify:  func(c *AuthConfig) { c.BcryptCost = 4 },
			wantErr: false,
		},
		{
			name:    "bcrypt cost valid maximum",
			modify:  func(c *AuthConfig) { c.BcryptCost = 31 },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.modify(&config)
			err := config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStorageConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  StorageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid local storage",
			config: StorageConfig{
				Provider:      "local",
				LocalPath:     "./storage",
				MaxUploadSize: 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "valid s3 storage",
			config: StorageConfig{
				Provider:      "s3",
				S3Endpoint:    "s3.amazonaws.com",
				S3AccessKey:   "access-key",
				S3SecretKey:   "secret-key",
				S3Bucket:      "my-bucket",
				MaxUploadSize: 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "invalid provider",
			config: StorageConfig{
				Provider:      "azure",
				MaxUploadSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "storage provider must be 'local' or 's3'",
		},
		{
			name: "local without path",
			config: StorageConfig{
				Provider:      "local",
				LocalPath:     "",
				MaxUploadSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "local_path is required",
		},
		{
			name: "s3 without endpoint",
			config: StorageConfig{
				Provider:      "s3",
				S3Endpoint:    "",
				S3AccessKey:   "key",
				S3SecretKey:   "secret",
				S3Bucket:      "bucket",
				MaxUploadSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "s3_endpoint is required",
		},
		{
			name: "s3 without access key",
			config: StorageConfig{
				Provider:      "s3",
				S3Endpoint:    "endpoint",
				S3AccessKey:   "",
				S3SecretKey:   "secret",
				S3Bucket:      "bucket",
				MaxUploadSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "s3_access_key is required",
		},
		{
			name: "s3 without secret key",
			config: StorageConfig{
				Provider:      "s3",
				S3Endpoint:    "endpoint",
				S3AccessKey:   "key",
				S3SecretKey:   "",
				S3Bucket:      "bucket",
				MaxUploadSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "s3_secret_key is required",
		},
		{
			name: "s3 without bucket",
			config: StorageConfig{
				Provider:      "s3",
				S3Endpoint:    "endpoint",
				S3AccessKey:   "key",
				S3SecretKey:   "secret",
				S3Bucket:      "",
				MaxUploadSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "s3_bucket is required",
		},
		{
			name: "zero upload size",
			config: StorageConfig{
				Provider:      "local",
				LocalPath:     "./storage",
				MaxUploadSize: 0,
			},
			wantErr: true,
			errMsg:  "max_upload_size must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSecurityConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SecurityConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with secure token",
			config: SecurityConfig{
				SetupToken: "this-is-a-secure-setup-token-for-testing-purposes",
			},
			wantErr: false,
		},
		{
			name: "empty setup token is valid",
			config: SecurityConfig{
				SetupToken: "",
			},
			wantErr: false,
		},
		{
			name: "insecure default token - changeme",
			config: SecurityConfig{
				SetupToken: "changeme",
			},
			wantErr: true,
			errMsg:  "please set a secure setup token",
		},
		{
			name: "insecure default token - test",
			config: SecurityConfig{
				SetupToken: "test",
			},
			wantErr: true,
			errMsg:  "please set a secure setup token",
		},
		{
			name: "insecure default token - your-secret-setup-token-change-in-production",
			config: SecurityConfig{
				SetupToken: "your-secret-setup-token-change-in-production",
			},
			wantErr: true,
			errMsg:  "please set a secure setup token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEmailConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  EmailConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid smtp config",
			config: EmailConfig{
				Enabled:     true,
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
			},
			wantErr: false,
		},
		{
			name: "unconfigured smtp is valid",
			config: EmailConfig{
				Enabled:  true,
				Provider: "smtp",
			},
			wantErr: false,
		},
		{
			name: "invalid provider",
			config: EmailConfig{
				Enabled:     true,
				Provider:    "invalid",
				FromAddress: "test@example.com",
			},
			wantErr: true,
			errMsg:  "invalid email provider",
		},
		{
			name: "empty provider is valid",
			config: EmailConfig{
				Enabled:     true,
				FromAddress: "test@example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEmailConfig_IsConfigured(t *testing.T) {
	tests := []struct {
		name       string
		config     EmailConfig
		configured bool
	}{
		{
			name: "fully configured smtp",
			config: EmailConfig{
				Enabled:     true,
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
			},
			configured: true,
		},
		{
			name: "smtp missing host",
			config: EmailConfig{
				Enabled:     true,
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPPort:    587,
			},
			configured: false,
		},
		{
			name: "smtp missing port",
			config: EmailConfig{
				Enabled:     true,
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPHost:    "smtp.example.com",
			},
			configured: false,
		},
		{
			name: "email disabled",
			config: EmailConfig{
				Enabled:     false,
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
			},
			configured: false,
		},
		{
			name: "missing from_address",
			config: EmailConfig{
				Enabled:  true,
				Provider: "smtp",
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
			},
			configured: false,
		},
		{
			name: "fully configured sendgrid",
			config: EmailConfig{
				Enabled:        true,
				Provider:       "sendgrid",
				FromAddress:    "test@example.com",
				SendGridAPIKey: "api-key",
			},
			configured: true,
		},
		{
			name: "sendgrid missing api key",
			config: EmailConfig{
				Enabled:     true,
				Provider:    "sendgrid",
				FromAddress: "test@example.com",
			},
			configured: false,
		},
		{
			name: "fully configured mailgun",
			config: EmailConfig{
				Enabled:       true,
				Provider:      "mailgun",
				FromAddress:   "test@example.com",
				MailgunAPIKey: "api-key",
				MailgunDomain: "mg.example.com",
			},
			configured: true,
		},
		{
			name: "mailgun missing domain",
			config: EmailConfig{
				Enabled:       true,
				Provider:      "mailgun",
				FromAddress:   "test@example.com",
				MailgunAPIKey: "api-key",
			},
			configured: false,
		},
		{
			name: "fully configured ses",
			config: EmailConfig{
				Enabled:      true,
				Provider:     "ses",
				FromAddress:  "test@example.com",
				SESAccessKey: "access-key",
				SESSecretKey: "secret-key",
				SESRegion:    "us-east-1",
			},
			configured: true,
		},
		{
			name: "ses missing region",
			config: EmailConfig{
				Enabled:      true,
				Provider:     "ses",
				FromAddress:  "test@example.com",
				SESAccessKey: "access-key",
				SESSecretKey: "secret-key",
			},
			configured: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsConfigured()
			assert.Equal(t, tt.configured, result)
		})
	}
}

func TestAPIConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  APIConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: APIConfig{
				MaxPageSize:     1000,
				MaxTotalResults: 10000,
				DefaultPageSize: 100,
			},
			wantErr: false,
		},
		{
			name: "unlimited values (-1) are valid",
			config: APIConfig{
				MaxPageSize:     -1,
				MaxTotalResults: -1,
				DefaultPageSize: -1,
			},
			wantErr: false,
		},
		{
			name: "zero max page size",
			config: APIConfig{
				MaxPageSize:     0,
				MaxTotalResults: 1000,
				DefaultPageSize: 100,
			},
			wantErr: true,
			errMsg:  "max_page_size must be positive or -1",
		},
		{
			name: "zero max total results",
			config: APIConfig{
				MaxPageSize:     1000,
				MaxTotalResults: 0,
				DefaultPageSize: 100,
			},
			wantErr: true,
			errMsg:  "max_total_results must be positive or -1",
		},
		{
			name: "zero default page size",
			config: APIConfig{
				MaxPageSize:     1000,
				MaxTotalResults: 10000,
				DefaultPageSize: 0,
			},
			wantErr: true,
			errMsg:  "default_page_size must be positive or -1",
		},
		{
			name: "default exceeds max",
			config: APIConfig{
				MaxPageSize:     100,
				MaxTotalResults: 10000,
				DefaultPageSize: 200,
			},
			wantErr: true,
			errMsg:  "default_page_size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestScalingConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ScalingConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid local backend",
			config: ScalingConfig{
				Backend: "local",
			},
			wantErr: false,
		},
		{
			name: "valid postgres backend",
			config: ScalingConfig{
				Backend: "postgres",
			},
			wantErr: false,
		},
		{
			name: "valid redis backend",
			config: ScalingConfig{
				Backend:  "redis",
				RedisURL: "redis://localhost:6379",
			},
			wantErr: false,
		},
		{
			name: "invalid backend",
			config: ScalingConfig{
				Backend: "memcached",
			},
			wantErr: true,
			errMsg:  "invalid scaling backend",
		},
		{
			name: "redis without url",
			config: ScalingConfig{
				Backend:  "redis",
				RedisURL: "",
			},
			wantErr: true,
			errMsg:  "redis_url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoggingConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  LoggingConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: LoggingConfig{
				ConsoleLevel:  "info",
				ConsoleFormat: "console",
				Backend:       "postgres",
				BatchSize:     100,
			},
			wantErr: false,
		},
		{
			name: "invalid console level",
			config: LoggingConfig{
				ConsoleLevel: "verbose",
			},
			wantErr: true,
			errMsg:  "invalid console_level",
		},
		{
			name: "invalid console format",
			config: LoggingConfig{
				ConsoleFormat: "xml",
			},
			wantErr: true,
			errMsg:  "invalid console_format",
		},
		{
			name: "invalid backend",
			config: LoggingConfig{
				Backend: "cloudwatch",
			},
			wantErr: true,
			errMsg:  "invalid logging backend",
		},
		{
			name: "s3 without bucket",
			config: LoggingConfig{
				Backend:  "s3",
				S3Bucket: "",
			},
			wantErr: true,
			errMsg:  "s3_bucket is required",
		},
		{
			name: "negative batch size",
			config: LoggingConfig{
				BatchSize: -1,
			},
			wantErr: true,
			errMsg:  "batch_size cannot be negative",
		},
		{
			name: "negative buffer size",
			config: LoggingConfig{
				BufferSize: -1,
			},
			wantErr: true,
			errMsg:  "buffer_size cannot be negative",
		},
		{
			name: "negative retention days",
			config: LoggingConfig{
				SystemRetentionDays: -1,
			},
			wantErr: true,
			errMsg:  "system_retention_days cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTracingConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TracingConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "disabled tracing doesn't validate",
			config: TracingConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "valid enabled config",
			config: TracingConfig{
				Enabled:    true,
				Endpoint:   "localhost:4317",
				SampleRate: 0.5,
			},
			wantErr: false,
		},
		{
			name: "enabled without endpoint",
			config: TracingConfig{
				Enabled:  true,
				Endpoint: "",
			},
			wantErr: true,
			errMsg:  "tracing endpoint is required",
		},
		{
			name: "sample rate too low",
			config: TracingConfig{
				Enabled:    true,
				Endpoint:   "localhost:4317",
				SampleRate: -0.1,
			},
			wantErr: true,
			errMsg:  "sample_rate must be between 0.0 and 1.0",
		},
		{
			name: "sample rate too high",
			config: TracingConfig{
				Enabled:    true,
				Endpoint:   "localhost:4317",
				SampleRate: 1.5,
			},
			wantErr: true,
			errMsg:  "sample_rate must be between 0.0 and 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFunctionsConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  FunctionsConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: FunctionsConfig{
				FunctionsDir:       "./functions",
				DefaultTimeout:     30,
				MaxTimeout:         300,
				DefaultMemoryLimit: 128,
				MaxMemoryLimit:     1024,
			},
			wantErr: false,
		},
		{
			name: "empty functions dir",
			config: FunctionsConfig{
				FunctionsDir:       "",
				DefaultTimeout:     30,
				MaxTimeout:         300,
				DefaultMemoryLimit: 128,
				MaxMemoryLimit:     1024,
			},
			wantErr: true,
			errMsg:  "functions_dir cannot be empty",
		},
		{
			name: "default timeout exceeds max",
			config: FunctionsConfig{
				FunctionsDir:       "./functions",
				DefaultTimeout:     600,
				MaxTimeout:         300,
				DefaultMemoryLimit: 128,
				MaxMemoryLimit:     1024,
			},
			wantErr: true,
			errMsg:  "default_timeout",
		},
		{
			name: "default memory exceeds max",
			config: FunctionsConfig{
				FunctionsDir:       "./functions",
				DefaultTimeout:     30,
				MaxTimeout:         300,
				DefaultMemoryLimit: 2048,
				MaxMemoryLimit:     1024,
			},
			wantErr: true,
			errMsg:  "default_memory_limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestJobsConfig_Validate(t *testing.T) {
	validConfig := func() JobsConfig {
		return JobsConfig{
			JobsDir:                   "./jobs",
			WorkerMode:                "embedded",
			EmbeddedWorkerCount:       4,
			MaxConcurrentPerWorker:    5,
			MaxConcurrentPerNamespace: 20,
			DefaultMaxDuration:        5 * time.Minute,
			MaxMaxDuration:            time.Hour,
			DefaultProgressTimeout:    5 * time.Minute,
			PollInterval:              time.Second,
			WorkerHeartbeatInterval:   10 * time.Second,
			WorkerTimeout:             30 * time.Second,
		}
	}

	tests := []struct {
		name    string
		modify  func(*JobsConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			modify:  func(c *JobsConfig) {},
			wantErr: false,
		},
		{
			name:    "empty jobs dir",
			modify:  func(c *JobsConfig) { c.JobsDir = "" },
			wantErr: true,
			errMsg:  "jobs_dir cannot be empty",
		},
		{
			name:    "invalid worker mode",
			modify:  func(c *JobsConfig) { c.WorkerMode = "distributed" },
			wantErr: true,
			errMsg:  "invalid worker_mode",
		},
		{
			name:    "valid standalone mode",
			modify:  func(c *JobsConfig) { c.WorkerMode = "standalone" },
			wantErr: false,
		},
		{
			name:    "valid disabled mode",
			modify:  func(c *JobsConfig) { c.WorkerMode = "disabled" },
			wantErr: false,
		},
		{
			name:    "negative worker count",
			modify:  func(c *JobsConfig) { c.EmbeddedWorkerCount = -1 },
			wantErr: true,
			errMsg:  "embedded_worker_count cannot be negative",
		},
		{
			name:    "default duration exceeds max",
			modify:  func(c *JobsConfig) { c.DefaultMaxDuration = 2 * time.Hour },
			wantErr: true,
			errMsg:  "default_max_duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.modify(&config)
			err := config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAIConfig_Validate(t *testing.T) {
	validConfig := func() AIConfig {
		return AIConfig{
			ChatbotsDir:          "./chatbots",
			DefaultMaxTokens:     4096,
			QueryTimeout:         30 * time.Second,
			MaxRowsPerQuery:      1000,
			ConversationCacheTTL: 30 * time.Minute,
			MaxConversationTurns: 50,
		}
	}

	tests := []struct {
		name    string
		modify  func(*AIConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			modify:  func(c *AIConfig) {},
			wantErr: false,
		},
		{
			name:    "empty chatbots dir",
			modify:  func(c *AIConfig) { c.ChatbotsDir = "" },
			wantErr: true,
			errMsg:  "chatbots_dir cannot be empty",
		},
		{
			name:    "zero max tokens",
			modify:  func(c *AIConfig) { c.DefaultMaxTokens = 0 },
			wantErr: true,
			errMsg:  "default_max_tokens must be positive",
		},
		{
			name:    "zero query timeout",
			modify:  func(c *AIConfig) { c.QueryTimeout = 0 },
			wantErr: true,
			errMsg:  "query_timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.modify(&config)
			err := config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMCPConfig_Validate(t *testing.T) {
	validConfig := func() MCPConfig {
		return MCPConfig{
			Enabled:         true,
			BasePath:        "/mcp",
			SessionTimeout:  30 * time.Minute,
			MaxMessageSize:  1024 * 1024,
			RateLimitPerMin: 100,
		}
	}

	tests := []struct {
		name    string
		modify  func(*MCPConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			modify:  func(c *MCPConfig) {},
			wantErr: false,
		},
		{
			name:    "disabled skips validation",
			modify:  func(c *MCPConfig) { c.Enabled = false; c.BasePath = "" },
			wantErr: false,
		},
		{
			name:    "empty base path when enabled",
			modify:  func(c *MCPConfig) { c.BasePath = "" },
			wantErr: true,
			errMsg:  "mcp base_path cannot be empty when enabled",
		},
		{
			name:    "negative session timeout",
			modify:  func(c *MCPConfig) { c.SessionTimeout = -1 * time.Minute },
			wantErr: true,
			errMsg:  "mcp session_timeout cannot be negative",
		},
		{
			name:    "negative max message size",
			modify:  func(c *MCPConfig) { c.MaxMessageSize = -1 },
			wantErr: true,
			errMsg:  "mcp max_message_size cannot be negative",
		},
		{
			name:    "negative rate limit",
			modify:  func(c *MCPConfig) { c.RateLimitPerMin = -1 },
			wantErr: true,
			errMsg:  "mcp rate_limit_per_min cannot be negative",
		},
		{
			name:    "zero session timeout is valid",
			modify:  func(c *MCPConfig) { c.SessionTimeout = 0 },
			wantErr: false,
		},
		{
			name:    "zero max message size is valid",
			modify:  func(c *MCPConfig) { c.MaxMessageSize = 0 },
			wantErr: false,
		},
		{
			name:    "zero rate limit is valid (unlimited)",
			modify:  func(c *MCPConfig) { c.RateLimitPerMin = 0 },
			wantErr: false,
		},
		{
			name:    "with allowed tools",
			modify:  func(c *MCPConfig) { c.AllowedTools = []string{"query", "storage"} },
			wantErr: false,
		},
		{
			name:    "with allowed resources",
			modify:  func(c *MCPConfig) { c.AllowedResources = []string{"schema://", "storage://"} },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.modify(&config)
			err := config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBranchingConfig_Validate(t *testing.T) {
	validConfig := func() BranchingConfig {
		return BranchingConfig{
			Enabled:              true,
			MaxTotalBranches:     50,
			DefaultDataCloneMode: DataCloneModeSchemaOnly,
			AutoDeleteAfter:      24 * time.Hour,
			DatabasePrefix:       "branch_",
			SeedsPath:            "./seeds",
		}
	}

	tests := []struct {
		name    string
		modify  func(*BranchingConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			modify:  func(c *BranchingConfig) {},
			wantErr: false,
		},
		{
			name:    "disabled skips validation",
			modify:  func(c *BranchingConfig) { c.Enabled = false; c.DatabasePrefix = "" },
			wantErr: false,
		},
		{
			name:    "negative max total branches",
			modify:  func(c *BranchingConfig) { c.MaxTotalBranches = -1 },
			wantErr: true,
			errMsg:  "branching max_total_branches cannot be negative",
		},
		{
			name:    "invalid data clone mode",
			modify:  func(c *BranchingConfig) { c.DefaultDataCloneMode = "invalid_mode" },
			wantErr: true,
			errMsg:  "branching default_data_clone_mode must be one of",
		},
		{
			name:    "valid full_clone mode",
			modify:  func(c *BranchingConfig) { c.DefaultDataCloneMode = DataCloneModeFullClone },
			wantErr: false,
		},
		{
			name:    "valid seed_data mode",
			modify:  func(c *BranchingConfig) { c.DefaultDataCloneMode = DataCloneModeSeedData },
			wantErr: false,
		},
		{
			name:    "empty data clone mode defaults to schema_only",
			modify:  func(c *BranchingConfig) { c.DefaultDataCloneMode = "" },
			wantErr: false,
		},
		{
			name:    "negative auto delete after",
			modify:  func(c *BranchingConfig) { c.AutoDeleteAfter = -1 * time.Hour },
			wantErr: true,
			errMsg:  "branching auto_delete_after cannot be negative",
		},
		{
			name:    "zero auto delete after is valid (never)",
			modify:  func(c *BranchingConfig) { c.AutoDeleteAfter = 0 },
			wantErr: false,
		},
		{
			name:    "empty database prefix when enabled",
			modify:  func(c *BranchingConfig) { c.DatabasePrefix = "" },
			wantErr: true,
			errMsg:  "branching database_prefix cannot be empty when enabled",
		},
		{
			name:    "empty seeds path gets default",
			modify:  func(c *BranchingConfig) { c.SeedsPath = "" },
			wantErr: false,
		},
		{
			name:    "zero max total branches is valid",
			modify:  func(c *BranchingConfig) { c.MaxTotalBranches = 0 },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.modify(&config)
			err := config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBranchingConfig_SeedsPathDefault(t *testing.T) {
	t.Run("sets default seeds path when empty", func(t *testing.T) {
		config := BranchingConfig{
			Enabled:        true,
			DatabasePrefix: "branch_",
			SeedsPath:      "",
		}

		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, "./seeds", config.SeedsPath)
	})

	t.Run("preserves custom seeds path", func(t *testing.T) {
		config := BranchingConfig{
			Enabled:        true,
			DatabasePrefix: "branch_",
			SeedsPath:      "/custom/seeds",
		}

		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, "/custom/seeds", config.SeedsPath)
	})
}

func TestDataCloneModeConstants(t *testing.T) {
	t.Run("constants have expected values", func(t *testing.T) {
		assert.Equal(t, "schema_only", DataCloneModeSchemaOnly)
		assert.Equal(t, "full_clone", DataCloneModeFullClone)
		assert.Equal(t, "seed_data", DataCloneModeSeedData)
	})
}

func TestMCPConfig_SetOAuthDefaults(t *testing.T) {
	t.Run("sets default token expiry when zero", func(t *testing.T) {
		config := MCPConfig{}
		config.SetOAuthDefaults()
		assert.Equal(t, time.Hour, config.OAuth.TokenExpiry)
	})

	t.Run("sets default refresh token expiry when zero", func(t *testing.T) {
		config := MCPConfig{}
		config.SetOAuthDefaults()
		assert.Equal(t, 168*time.Hour, config.OAuth.RefreshTokenExpiry)
	})

	t.Run("sets default redirect URIs when empty", func(t *testing.T) {
		config := MCPConfig{}
		config.SetOAuthDefaults()
		assert.NotEmpty(t, config.OAuth.AllowedRedirectURIs)
		assert.Equal(t, DefaultMCPOAuthRedirectURIs(), config.OAuth.AllowedRedirectURIs)
	})

	t.Run("preserves custom token expiry", func(t *testing.T) {
		config := MCPConfig{
			OAuth: MCPOAuthConfig{
				TokenExpiry: 2 * time.Hour,
			},
		}
		config.SetOAuthDefaults()
		assert.Equal(t, 2*time.Hour, config.OAuth.TokenExpiry)
	})

	t.Run("preserves custom refresh token expiry", func(t *testing.T) {
		config := MCPConfig{
			OAuth: MCPOAuthConfig{
				RefreshTokenExpiry: 24 * time.Hour,
			},
		}
		config.SetOAuthDefaults()
		assert.Equal(t, 24*time.Hour, config.OAuth.RefreshTokenExpiry)
	})

	t.Run("preserves custom redirect URIs", func(t *testing.T) {
		customURIs := []string{"http://custom.example.com/callback"}
		config := MCPConfig{
			OAuth: MCPOAuthConfig{
				AllowedRedirectURIs: customURIs,
			},
		}
		config.SetOAuthDefaults()
		assert.Equal(t, customURIs, config.OAuth.AllowedRedirectURIs)
	})
}

func TestDefaultMCPOAuthRedirectURIs(t *testing.T) {
	uris := DefaultMCPOAuthRedirectURIs()

	t.Run("returns non-empty list", func(t *testing.T) {
		assert.NotEmpty(t, uris)
	})

	t.Run("includes Claude Desktop URIs", func(t *testing.T) {
		assert.Contains(t, uris, "https://claude.ai/api/mcp/auth_callback")
		assert.Contains(t, uris, "https://claude.com/api/mcp/auth_callback")
	})

	t.Run("includes Cursor URIs", func(t *testing.T) {
		found := false
		for _, uri := range uris {
			if uri == "cursor://" || uri == "cursor://anysphere.cursor-mcp/oauth/*/callback" {
				found = true
				break
			}
		}
		assert.True(t, found, "should include Cursor redirect URIs")
	})

	t.Run("includes VS Code URIs", func(t *testing.T) {
		assert.Contains(t, uris, "vscode://")
	})

	t.Run("includes localhost wildcards for development", func(t *testing.T) {
		assert.Contains(t, uris, "http://localhost:*")
		assert.Contains(t, uris, "http://127.0.0.1:*")
	})

	t.Run("includes MCP Inspector for development", func(t *testing.T) {
		assert.Contains(t, uris, "http://localhost:6274/oauth/callback")
	})

	t.Run("includes ChatGPT", func(t *testing.T) {
		assert.Contains(t, uris, "https://chatgpt.com/connector_platform_oauth_redirect")
	})
}

func TestMCPConfig_ValidateOAuth(t *testing.T) {
	validConfig := func() MCPConfig {
		return MCPConfig{
			Enabled:  true,
			BasePath: "/mcp",
			OAuth: MCPOAuthConfig{
				Enabled:            true,
				TokenExpiry:        time.Hour,
				RefreshTokenExpiry: 168 * time.Hour,
			},
		}
	}

	tests := []struct {
		name    string
		modify  func(*MCPConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid OAuth config",
			modify:  func(c *MCPConfig) {},
			wantErr: false,
		},
		{
			name:    "negative token expiry",
			modify:  func(c *MCPConfig) { c.OAuth.TokenExpiry = -1 * time.Hour },
			wantErr: true,
			errMsg:  "mcp oauth token_expiry cannot be negative",
		},
		{
			name:    "negative refresh token expiry",
			modify:  func(c *MCPConfig) { c.OAuth.RefreshTokenExpiry = -1 * time.Hour },
			wantErr: true,
			errMsg:  "mcp oauth refresh_token_expiry cannot be negative",
		},
		{
			name:    "zero token expiry is valid",
			modify:  func(c *MCPConfig) { c.OAuth.TokenExpiry = 0 },
			wantErr: false,
		},
		{
			name:    "zero refresh token expiry is valid",
			modify:  func(c *MCPConfig) { c.OAuth.RefreshTokenExpiry = 0 },
			wantErr: false,
		},
		{
			name:    "OAuth disabled skips OAuth validation",
			modify:  func(c *MCPConfig) { c.OAuth.Enabled = false; c.OAuth.TokenExpiry = -1 },
			wantErr: false,
		},
		{
			name:    "DCR enabled is valid",
			modify:  func(c *MCPConfig) { c.OAuth.DCREnabled = true },
			wantErr: false,
		},
		{
			name:    "custom redirect URIs",
			modify:  func(c *MCPConfig) { c.OAuth.AllowedRedirectURIs = []string{"http://example.com"} },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.modify(&config)
			err := config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOAuthProviderConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  OAuthProviderConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid Google provider",
			config: OAuthProviderConfig{
				Name:     "Google",
				ClientID: "client-id-123",
			},
			wantErr: false,
		},
		{
			name: "valid Apple provider",
			config: OAuthProviderConfig{
				Name:     "APPLE",
				ClientID: "com.example.app",
			},
			wantErr: false,
		},
		{
			name: "valid Microsoft provider",
			config: OAuthProviderConfig{
				Name:     "Microsoft",
				ClientID: "client-id-456",
			},
			wantErr: false,
		},
		{
			name: "valid custom provider with issuer URL",
			config: OAuthProviderConfig{
				Name:      "MyIDP",
				ClientID:  "custom-client-id",
				IssuerURL: "https://idp.example.com",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: OAuthProviderConfig{
				Name:     "",
				ClientID: "client-id",
			},
			wantErr: true,
			errMsg:  "oauth provider name is required",
		},
		{
			name: "empty client ID",
			config: OAuthProviderConfig{
				Name:     "google",
				ClientID: "",
			},
			wantErr: true,
			errMsg:  "client_id is required",
		},
		{
			name: "custom provider without issuer URL",
			config: OAuthProviderConfig{
				Name:     "custom-idp",
				ClientID: "client-id",
				// Missing IssuerURL for custom provider
			},
			wantErr: true,
			errMsg:  "issuer_url is required for custom providers",
		},
		{
			name: "name gets normalized to lowercase",
			config: OAuthProviderConfig{
				Name:     "GOOGLE",
				ClientID: "client-id",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOAuthProviderConfig_NameNormalization(t *testing.T) {
	t.Run("normalizes name to lowercase", func(t *testing.T) {
		config := OAuthProviderConfig{
			Name:     "GOOGLE",
			ClientID: "client-id",
		}

		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, "google", config.Name)
	})

	t.Run("normalizes mixed case to lowercase", func(t *testing.T) {
		config := OAuthProviderConfig{
			Name:     "Microsoft",
			ClientID: "client-id",
		}

		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, "microsoft", config.Name)
	})
}

func TestMetricsConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  MetricsConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: MetricsConfig{
				Enabled: true,
				Port:    9090,
				Path:    "/metrics",
			},
			wantErr: false,
		},
		{
			name: "disabled metrics skips validation",
			config: MetricsConfig{
				Enabled: false,
				Port:    0,
				Path:    "",
			},
			wantErr: false,
		},
		{
			name: "port too low",
			config: MetricsConfig{
				Enabled: true,
				Port:    0,
				Path:    "/metrics",
			},
			wantErr: true,
			errMsg:  "metrics port must be between 1 and 65535",
		},
		{
			name: "port too high",
			config: MetricsConfig{
				Enabled: true,
				Port:    70000,
				Path:    "/metrics",
			},
			wantErr: true,
			errMsg:  "metrics port must be between 1 and 65535",
		},
		{
			name: "negative port",
			config: MetricsConfig{
				Enabled: true,
				Port:    -1,
				Path:    "/metrics",
			},
			wantErr: true,
			errMsg:  "metrics port must be between 1 and 65535",
		},
		{
			name: "empty path",
			config: MetricsConfig{
				Enabled: true,
				Port:    9090,
				Path:    "",
			},
			wantErr: true,
			errMsg:  "metrics path cannot be empty",
		},
		{
			name: "path without leading slash",
			config: MetricsConfig{
				Enabled: true,
				Port:    9090,
				Path:    "metrics",
			},
			wantErr: true,
			errMsg:  "metrics path must start with '/'",
		},
		{
			name: "valid min port",
			config: MetricsConfig{
				Enabled: true,
				Port:    1,
				Path:    "/metrics",
			},
			wantErr: false,
		},
		{
			name: "valid max port",
			config: MetricsConfig{
				Enabled: true,
				Port:    65535,
				Path:    "/metrics",
			},
			wantErr: false,
		},
		{
			name: "custom path with leading slash",
			config: MetricsConfig{
				Enabled: true,
				Port:    9090,
				Path:    "/custom/metrics/path",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBranchingConfig_MaxBranchesPerUser(t *testing.T) {
	validConfig := func() BranchingConfig {
		return BranchingConfig{
			Enabled:            true,
			MaxTotalBranches:   50,
			MaxBranchesPerUser: 5,
			DatabasePrefix:     "branch_",
		}
	}

	tests := []struct {
		name    string
		modify  func(*BranchingConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid max branches per user",
			modify:  func(c *BranchingConfig) {},
			wantErr: false,
		},
		{
			name:    "negative max branches per user",
			modify:  func(c *BranchingConfig) { c.MaxBranchesPerUser = -1 },
			wantErr: true,
			errMsg:  "branching max_branches_per_user cannot be negative",
		},
		{
			name:    "zero max branches per user is valid (unlimited)",
			modify:  func(c *BranchingConfig) { c.MaxBranchesPerUser = 0 },
			wantErr: false,
		},
		{
			name:    "high max branches per user is valid",
			modify:  func(c *BranchingConfig) { c.MaxBranchesPerUser = 100 },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.modify(&config)
			err := config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	// Helper to create a valid Config for testing
	validConfig := func() Config {
		return Config{
			EncryptionKey: "12345678901234567890123456789012", // 32 bytes for AES-256
			BaseURL:       "https://example.com",
			Server: ServerConfig{
				Address:      ":8080",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				BodyLimit:    1024 * 1024,
			},
			Database: DatabaseConfig{
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				Database:        "fluxbase",
				SSLMode:         "disable",
				MaxConnections:  50,
				MinConnections:  10,
				MaxConnLifetime: time.Hour,
				MaxConnIdleTime: 30 * time.Minute,
				HealthCheck:     time.Minute,
			},
			Auth: AuthConfig{
				JWTSecret:           "this-is-a-very-secure-secret-key-for-testing-purposes",
				JWTExpiry:           15 * time.Minute,
				RefreshExpiry:       7 * 24 * time.Hour,
				MagicLinkExpiry:     15 * time.Minute,
				PasswordResetExpiry: time.Hour,
				PasswordMinLen:      8,
				BcryptCost:          10,
			},
			Storage: StorageConfig{
				Provider:      "local",
				LocalPath:     "./storage",
				MaxUploadSize: 1024 * 1024,
			},
			Security: SecurityConfig{
				SetupToken: "secure-setup-token-for-testing",
			},
			API: APIConfig{
				MaxPageSize:     1000,
				MaxTotalResults: 10000,
				DefaultPageSize: 100,
			},
			Scaling: ScalingConfig{
				Backend: "local",
			},
			Logging: LoggingConfig{
				ConsoleLevel:  "info",
				ConsoleFormat: "console",
				Backend:       "postgres",
				BatchSize:     100,
			},
		}
	}

	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name:    "missing encryption key",
			modify:  func(c *Config) { c.EncryptionKey = "" },
			wantErr: true,
			errMsg:  "encryption_key is required",
		},
		{
			name:    "encryption key too short",
			modify:  func(c *Config) { c.EncryptionKey = "tooshort" },
			wantErr: true,
			errMsg:  "encryption_key must be exactly 32 bytes",
		},
		{
			name:    "encryption key too long",
			modify:  func(c *Config) { c.EncryptionKey = "123456789012345678901234567890123456" },
			wantErr: true,
			errMsg:  "encryption_key must be exactly 32 bytes",
		},
		{
			name:    "invalid base URL scheme",
			modify:  func(c *Config) { c.BaseURL = "ftp://example.com" },
			wantErr: true,
			errMsg:  "base_url must use http or https scheme",
		},
		{
			name:    "invalid base URL",
			modify:  func(c *Config) { c.BaseURL = "://invalid" },
			wantErr: true,
			errMsg:  "invalid base_url",
		},
		{
			name:    "valid http base URL",
			modify:  func(c *Config) { c.BaseURL = "http://localhost:8080" },
			wantErr: false,
		},
		{
			name:    "empty base URL is valid",
			modify:  func(c *Config) { c.BaseURL = "" },
			wantErr: false,
		},
		{
			name:    "invalid public base URL scheme",
			modify:  func(c *Config) { c.PublicBaseURL = "ftp://example.com" },
			wantErr: true,
			errMsg:  "public_base_url must use http or https scheme",
		},
		{
			name:    "invalid public base URL",
			modify:  func(c *Config) { c.PublicBaseURL = "://invalid" },
			wantErr: true,
			errMsg:  "invalid public_base_url",
		},
		{
			name:    "valid public base URL",
			modify:  func(c *Config) { c.PublicBaseURL = "https://public.example.com" },
			wantErr: false,
		},
		{
			name: "server config error propagates",
			modify: func(c *Config) {
				c.Server.Address = ""
			},
			wantErr: true,
			errMsg:  "server configuration error",
		},
		{
			name: "database config error propagates",
			modify: func(c *Config) {
				c.Database.Host = ""
			},
			wantErr: true,
			errMsg:  "database configuration error",
		},
		{
			name: "auth config error propagates",
			modify: func(c *Config) {
				c.Auth.JWTSecret = ""
			},
			wantErr: true,
			errMsg:  "auth configuration error",
		},
		{
			name: "storage config error propagates",
			modify: func(c *Config) {
				c.Storage.Provider = "invalid"
			},
			wantErr: true,
			errMsg:  "storage configuration error",
		},
		{
			name: "email config error propagates when enabled",
			modify: func(c *Config) {
				c.Email.Enabled = true
				c.Email.Provider = "invalid"
			},
			wantErr: true,
			errMsg:  "email configuration error",
		},
		{
			name: "email config not validated when disabled",
			modify: func(c *Config) {
				c.Email.Enabled = false
				c.Email.Provider = "invalid"
			},
			wantErr: false,
		},
		{
			name: "functions config error propagates when enabled",
			modify: func(c *Config) {
				c.Functions.Enabled = true
				c.Functions.FunctionsDir = ""
			},
			wantErr: true,
			errMsg:  "functions configuration error",
		},
		{
			name: "functions config not validated when disabled",
			modify: func(c *Config) {
				c.Functions.Enabled = false
				c.Functions.FunctionsDir = ""
			},
			wantErr: false,
		},
		{
			name: "jobs config error propagates when enabled",
			modify: func(c *Config) {
				c.Jobs.Enabled = true
				c.Jobs.JobsDir = ""
			},
			wantErr: true,
			errMsg:  "jobs configuration error",
		},
		{
			name: "jobs config not validated when disabled",
			modify: func(c *Config) {
				c.Jobs.Enabled = false
				c.Jobs.JobsDir = ""
			},
			wantErr: false,
		},
		{
			name: "tracing config error propagates when enabled",
			modify: func(c *Config) {
				c.Tracing.Enabled = true
				c.Tracing.Endpoint = ""
			},
			wantErr: true,
			errMsg:  "tracing configuration error",
		},
		{
			name: "tracing config not validated when disabled",
			modify: func(c *Config) {
				c.Tracing.Enabled = false
				c.Tracing.Endpoint = ""
			},
			wantErr: false,
		},
		{
			name: "metrics config error propagates when enabled",
			modify: func(c *Config) {
				c.Metrics.Enabled = true
				c.Metrics.Port = 0
			},
			wantErr: true,
			errMsg:  "metrics configuration error",
		},
		{
			name: "metrics config not validated when disabled",
			modify: func(c *Config) {
				c.Metrics.Enabled = false
				c.Metrics.Port = 0
			},
			wantErr: false,
		},
		{
			name: "ai config error propagates when enabled",
			modify: func(c *Config) {
				c.AI.Enabled = true
				c.AI.ChatbotsDir = ""
			},
			wantErr: true,
			errMsg:  "ai configuration error",
		},
		{
			name: "ai config not validated when disabled",
			modify: func(c *Config) {
				c.AI.Enabled = false
				c.AI.ChatbotsDir = ""
			},
			wantErr: false,
		},
		{
			name: "graphql config error propagates when enabled",
			modify: func(c *Config) {
				c.GraphQL.Enabled = true
				c.GraphQL.MaxDepth = -1
			},
			wantErr: true,
			errMsg:  "graphql configuration error",
		},
		{
			name: "graphql config not validated when disabled",
			modify: func(c *Config) {
				c.GraphQL.Enabled = false
				c.GraphQL.MaxDepth = -1
			},
			wantErr: false,
		},
		{
			name: "mcp config error propagates when enabled",
			modify: func(c *Config) {
				c.MCP.Enabled = true
				c.MCP.BasePath = ""
			},
			wantErr: true,
			errMsg:  "mcp configuration error",
		},
		{
			name: "mcp config not validated when disabled",
			modify: func(c *Config) {
				c.MCP.Enabled = false
				c.MCP.BasePath = ""
			},
			wantErr: false,
		},
		{
			name: "branching config error propagates when enabled",
			modify: func(c *Config) {
				c.Branching.Enabled = true
				c.Branching.DatabasePrefix = ""
			},
			wantErr: true,
			errMsg:  "branching configuration error",
		},
		{
			name: "branching config not validated when disabled",
			modify: func(c *Config) {
				c.Branching.Enabled = false
				c.Branching.DatabasePrefix = ""
			},
			wantErr: false,
		},
		{
			name: "scaling config error propagates",
			modify: func(c *Config) {
				c.Scaling.Backend = "invalid"
			},
			wantErr: true,
			errMsg:  "scaling configuration error",
		},
		{
			name: "logging config error propagates",
			modify: func(c *Config) {
				c.Logging.ConsoleLevel = "invalid"
			},
			wantErr: true,
			errMsg:  "logging configuration error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.modify(&config)
			err := config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_GetPublicBaseURL(t *testing.T) {
	tests := []struct {
		name          string
		baseURL       string
		publicBaseURL string
		expected      string
	}{
		{
			name:          "returns public base URL when set",
			baseURL:       "https://internal.example.com",
			publicBaseURL: "https://public.example.com",
			expected:      "https://public.example.com",
		},
		{
			name:          "falls back to base URL when public not set",
			baseURL:       "https://example.com",
			publicBaseURL: "",
			expected:      "https://example.com",
		},
		{
			name:          "returns empty when both are empty",
			baseURL:       "",
			publicBaseURL: "",
			expected:      "",
		},
		{
			name:          "public URL takes precedence even if empty base URL",
			baseURL:       "",
			publicBaseURL: "https://public.example.com",
			expected:      "https://public.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				BaseURL:       tt.baseURL,
				PublicBaseURL: tt.publicBaseURL,
			}
			result := config.GetPublicBaseURL()
			assert.Equal(t, tt.expected, result)
		})
	}
}
