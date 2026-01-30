---
title: "Configuration Reference"
description: Complete reference for configuring Fluxbase via YAML config file or environment variables including database, authentication, storage, and realtime settings.
---

Complete reference for configuring Fluxbase via configuration file or environment variables.

## Configuration File

Create `fluxbase.yaml` in your working directory:

```yaml
# General Configuration
base_url: http://localhost:8080 # Internal base URL (server-to-server)
public_base_url: https://api.example.com # Public base URL (user-facing links, OAuth callbacks)
debug: false

# Server Configuration
server:
  address: ":8080" # Listen address (host:port)
  read_timeout: 300s # 5 min for large file streaming
  write_timeout: 300s # 5 min for large file streaming
  idle_timeout: 120s # 2 min idle timeout
  body_limit: 2147483648 # 2GB global body limit
  allowed_ip_ranges: [] # Global IP allowlist (empty = allow all)

  # Per-endpoint body limits (granular control)
  body_limits:
    enabled: true
    default_limit: 1048576 # 1MB default
    rest_limit: 1048576 # 1MB for REST CRUD
    auth_limit: 65536 # 64KB for auth endpoints
    storage_limit: 524288000 # 500MB for file uploads
    bulk_limit: 10485760 # 10MB for bulk/RPC operations
    admin_limit: 5242880 # 5MB for admin endpoints
    max_json_depth: 64 # Max JSON nesting depth

# Database Configuration
database:
  host: localhost
  port: 5432
  user: postgres # Runtime database user
  admin_user: "" # Admin user for migrations (defaults to user)
  password: postgres
  admin_password: "" # Admin user password (defaults to password)
  database: fluxbase
  ssl_mode: disable # disable, allow, prefer, require, verify-ca, verify-full
  max_connections: 50 # Connection pool max size
  min_connections: 10 # Connection pool min size
  max_conn_lifetime: 1h
  max_conn_idle_time: 30m
  health_check_period: 1m
  user_migrations_path: /migrations/user # Path to user-provided migrations

# Authentication Configuration
auth:
  jwt_secret: your-secret-key-change-in-production
  jwt_expiry: 15m
  refresh_expiry: 168h # 7 days
  service_role_ttl: 24h # Service role token TTL
  anon_ttl: 24h # Anonymous token TTL
  magic_link_expiry: 15m
  password_reset_expiry: 1h
  password_min_length: 12
  bcrypt_cost: 10
  signup_enabled: true
  magic_link_enabled: true
  totp_issuer: Fluxbase # 2FA issuer name shown in authenticator apps
  allow_user_client_keys: true # Allow users to create their own API client keys

  # OAuth/OIDC Providers
  oauth_providers:
    - name: google
      enabled: true
      client_id: ${GOOGLE_CLIENT_ID}
      client_secret: ${GOOGLE_CLIENT_SECRET}
      allow_dashboard_login: false
      allow_app_login: true
    - name: custom-oidc
      enabled: true
      client_id: ${OIDC_CLIENT_ID}
      issuer_url: https://auth.example.com # Auto-discovers .well-known/openid-configuration
      scopes: ["openid", "email", "profile"]
      required_claims:
        roles: ["admin"] # Require specific claim values

  # SAML SSO Providers (Enterprise)
  saml_providers:
    - name: okta
      enabled: true
      idp_metadata_url: https://your-org.okta.com/app/.../sso/saml/metadata
      entity_id: urn:fluxbase:sp
      acs_url: https://api.example.com/auth/saml/okta/callback
      auto_create_users: true
      default_role: authenticated
      allow_idp_initiated: false # Security: disable IdP-initiated SSO
      group_attribute: groups
      required_groups: ["fluxbase-users"]

# Storage Configuration
storage:
  enabled: true
  provider: local # local or s3
  local_path: ./storage
  max_upload_size: 2147483648 # 2GB
  s3_endpoint: ""
  s3_access_key: ""
  s3_secret_key: ""
  s3_bucket: ""
  s3_region: ""
  s3_force_path_style: true # Required for MinIO, R2, Spaces, etc.
  default_buckets: ["uploads", "temp-files", "public"]

  # Image Transformation Settings
  transforms:
    enabled: true
    default_quality: 80
    max_width: 4096
    max_height: 4096
    allowed_formats: ["webp", "jpg", "png", "avif"]
    max_total_pixels: 16000000 # 16 megapixels max
    bucket_size: 50 # Round dimensions to 50px (cache efficiency)
    rate_limit: 60 # Transforms per minute per user
    timeout: 30s
    max_concurrent: 4
    cache_enabled: true
    cache_ttl: 24h
    cache_max_size: 1073741824 # 1GB cache

# Realtime Configuration
realtime:
  enabled: true
  max_connections: 1000
  max_connections_per_user: 10
  max_connections_per_ip: 20
  ping_interval: 30s
  pong_timeout: 60s
  read_buffer_size: 1024
  write_buffer_size: 1024
  message_size_limit: 524288 # 512KB
  channel_buffer_size: 100
  rls_cache_size: 100000
  rls_cache_ttl: 30s
  listener_pool_size: 2 # LISTEN connections for redundancy
  notification_workers: 4
  notification_queue_size: 1000
  client_message_queue_size: 256
  slow_client_threshold: 100
  slow_client_timeout: 30s

# Admin UI
admin:
  enabled: false # Enable React admin dashboard

# Logging
logging:
  console_enabled: true
  console_level: info # trace, debug, info, warn, error
  console_format: console # json or console
  backend: postgres # postgres, s3, local
  s3_bucket: ""
  s3_prefix: logs
  local_path: ./logs
  batch_size: 100
  flush_interval: 1s
  buffer_size: 10000
  pubsub_enabled: true # Enable PubSub for realtime log streaming
  system_retention_days: 7
  http_retention_days: 30
  security_retention_days: 90
  execution_retention_days: 30
  ai_retention_days: 30
  retention_enabled: true
  retention_check_interval: 24h
  custom_categories: []
  custom_retention_days: 30

# CORS Configuration
cors:
  allowed_origins: "http://localhost:5173,http://localhost:8080"
  allowed_methods: "GET,POST,PUT,PATCH,DELETE,OPTIONS"
  allowed_headers: "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-CSRF-Token,Prefer,apikey"
  exposed_headers: "Content-Range,Content-Encoding,Content-Length,X-Request-ID"
  allow_credentials: true
  max_age: 300

# Security Configuration
security:
  setup_token: "" # Required for admin dashboard (openssl rand -base64 32)
  enable_global_rate_limit: false
  admin_setup_rate_limit: 5
  admin_setup_rate_window: 15m
  auth_login_rate_limit: 10
  auth_login_rate_window: 1m
  admin_login_rate_limit: 10
  admin_login_rate_window: 1m

  # CAPTCHA Configuration (bot protection)
  captcha:
    enabled: false
    provider: hcaptcha # hcaptcha, recaptcha_v3, turnstile, cap
    site_key: ""
    secret_key: ""
    score_threshold: 0.5 # For reCAPTCHA v3
    endpoints: ["signup", "login", "password_reset", "magic_link"]
    cap_server_url: "" # For self-hosted Cap provider
    cap_api_key: ""

# Encryption (for sensitive data in database)
encryption_key: "" # 32 bytes for AES-256 (openssl rand -base64 32 | head -c 32)

# Migrations API
migrations:
  enabled: true
  allowed_ip_ranges:
    ["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/16", "127.0.0.0/8"]
  require_service_key: true # Always require service key authentication

# MCP (Model Context Protocol)
mcp:
  enabled: true
  base_path: /mcp
  session_timeout: 30m
  max_message_size: 10485760
  rate_limit_per_min: 100
  allowed_tools: [] # Empty = all tools enabled
  allowed_resources: [] # Empty = all resources enabled
  tools_dir: /app/mcp-tools
  auto_load_on_boot: true

  # MCP OAuth Configuration
  oauth:
    enabled: true # Enable OAuth 2.1 for MCP clients
    dcr_enabled: true # Enable Dynamic Client Registration
    token_expiry: 1h
    refresh_token_expiry: 168h # 7 days
    allowed_redirect_uris: [] # Empty = use defaults

# Branching (Database Branching)
branching:
  enabled: false
  max_branches_per_user: 5
  max_total_branches: 50
  default_data_clone_mode: schema_only # schema_only or full_clone
  auto_delete_after: "0" # 0 = never, or duration like "24h"
  database_prefix: branch_
  admin_database_url: "" # Uses main database URL if empty
  seeds_path: ./seeds
```

## Environment Variables

Environment variables take precedence over configuration file values.

### General

| Variable                   | Description                                              | Default                 | Example                                 |
| -------------------------- | -------------------------------------------------------- | ----------------------- | --------------------------------------- |
| `FLUXBASE_BASE_URL`        | Internal base URL for server-to-server communication     | `http://localhost:8080` | `http://fluxbase:8080`                  |
| `FLUXBASE_PUBLIC_BASE_URL` | Public base URL for user-facing links, OAuth callbacks   | `""` (uses BASE_URL)    | `https://api.example.com`               |
| `FLUXBASE_DEBUG`           | Enable debug mode                                        | `false`                 | `true`, `false`                         |
| `FLUXBASE_ENCRYPTION_KEY`  | AES-256-GCM encryption key for sensitive data (32 bytes) | `""`                    | `openssl rand -base64 32 \| head -c 32` |

### Server

| Variable                            | Description                                 | Default            | Example                     |
| ----------------------------------- | ------------------------------------------- | ------------------ | --------------------------- |
| `FLUXBASE_SERVER_ADDRESS`           | Listen address (host:port)                  | `:8080`            | `:8080`, `0.0.0.0:8080`     |
| `FLUXBASE_SERVER_READ_TIMEOUT`      | Read timeout                                | `300s`             | `30s`                       |
| `FLUXBASE_SERVER_WRITE_TIMEOUT`     | Write timeout                               | `300s`             | `30s`                       |
| `FLUXBASE_SERVER_IDLE_TIMEOUT`      | Idle timeout                                | `120s`             | `120s`                      |
| `FLUXBASE_SERVER_BODY_LIMIT`        | Global body size limit (bytes)              | `2147483648` (2GB) | `1073741824` (1GB)          |
| `FLUXBASE_SERVER_ALLOWED_IP_RANGES` | Global IP allowlist (CIDR, comma-separated) | `""` (allow all)   | `10.0.0.0/8,192.168.0.0/16` |

**Per-Endpoint Body Limits:**

| Variable                                     | Description                       | Default             | Example            |
| -------------------------------------------- | --------------------------------- | ------------------- | ------------------ |
| `FLUXBASE_SERVER_BODY_LIMITS_ENABLED`        | Enable per-endpoint body limits   | `true`              | `true`, `false`    |
| `FLUXBASE_SERVER_BODY_LIMITS_DEFAULT_LIMIT`  | Default body limit                | `1048576` (1MB)     | `2097152` (2MB)    |
| `FLUXBASE_SERVER_BODY_LIMITS_REST_LIMIT`     | Limit for REST CRUD operations    | `1048576` (1MB)     | `2097152` (2MB)    |
| `FLUXBASE_SERVER_BODY_LIMITS_AUTH_LIMIT`     | Limit for auth endpoints          | `65536` (64KB)      | `131072` (128KB)   |
| `FLUXBASE_SERVER_BODY_LIMITS_STORAGE_LIMIT`  | Limit for file uploads            | `524288000` (500MB) | `1073741824` (1GB) |
| `FLUXBASE_SERVER_BODY_LIMITS_BULK_LIMIT`     | Limit for bulk operations and RPC | `10485760` (10MB)   | `20971520` (20MB)  |
| `FLUXBASE_SERVER_BODY_LIMITS_ADMIN_LIMIT`    | Limit for admin endpoints         | `5242880` (5MB)     | `10485760` (10MB)  |
| `FLUXBASE_SERVER_BODY_LIMITS_MAX_JSON_DEPTH` | Maximum JSON nesting depth        | `64`                | `32`               |

### Database

| Variable                                 | Description                                  | Default            | Example           |
| ---------------------------------------- | -------------------------------------------- | ------------------ | ----------------- |
| `FLUXBASE_DATABASE_HOST`                 | PostgreSQL host                              | `localhost`        | `localhost`       |
| `FLUXBASE_DATABASE_PORT`                 | PostgreSQL port                              | `5432`             | `5432`            |
| `FLUXBASE_DATABASE_USER`                 | Runtime database user                        | `postgres`         | `fluxbase`        |
| `FLUXBASE_DATABASE_PASSWORD`             | Runtime user password                        | `postgres`         | `your-password`   |
| `FLUXBASE_DATABASE_DATABASE`             | Database name                                | `fluxbase`         | `fluxbase`        |
| `FLUXBASE_DATABASE_SSL_MODE`             | SSL mode                                     | `disable`          | `require`         |
| `FLUXBASE_DATABASE_MAX_CONNECTIONS`      | Max connection pool size                     | `25`               | `100`             |
| `FLUXBASE_DATABASE_MIN_CONNECTIONS`      | Min connections in pool                      | `5`                | `5`               |
| `FLUXBASE_DATABASE_MAX_CONN_LIFETIME`    | Connection max lifetime                      | `1h`               | `1h`              |
| `FLUXBASE_DATABASE_MAX_CONN_IDLE_TIME`   | Connection max idle time                     | `30m`              | `30m`             |
| `FLUXBASE_DATABASE_HEALTH_CHECK_PERIOD`  | Health check interval                        | `1m`               | `1m`              |
| `FLUXBASE_DATABASE_ADMIN_USER`           | Admin user for migrations (defaults to USER) | `""`               | `postgres`        |
| `FLUXBASE_DATABASE_ADMIN_PASSWORD`       | Admin user password (defaults to PASSWORD)   | `""`               | `admin-password`  |
| `FLUXBASE_DATABASE_USER_MIGRATIONS_PATH` | Path to user-provided migrations             | `/migrations/user` | `/app/migrations` |

**SSL Modes:**

- `disable` - No SSL (development only)
- `allow` - Prefer SSL if available
- `prefer` - Use SSL if available (default for many clients)
- `require` - Require SSL connection
- `verify-ca` - Require SSL and verify CA certificate
- `verify-full` - Require SSL and verify CA + hostname

### Authentication

| Variable                               | Description                           | Default         | Example                   |
| -------------------------------------- | ------------------------------------- | --------------- | ------------------------- |
| `FLUXBASE_AUTH_JWT_SECRET`             | JWT signing key (min 32 chars)        | **(required)**  | `openssl rand -base64 32` |
| `FLUXBASE_AUTH_JWT_EXPIRY`             | Access token expiration               | `15m`           | `15m`, `1h`               |
| `FLUXBASE_AUTH_REFRESH_EXPIRY`         | Refresh token expiration              | `168h` (7 days) | `168h`, `720h`            |
| `FLUXBASE_AUTH_SERVICE_ROLE_TTL`       | Service role token TTL                | `24h`           | `24h`, `48h`              |
| `FLUXBASE_AUTH_ANON_TTL`               | Anonymous token TTL                   | `24h`           | `24h`, `48h`              |
| `FLUXBASE_AUTH_MAGIC_LINK_EXPIRY`      | Magic link expiration                 | `15m`           | `15m`                     |
| `FLUXBASE_AUTH_PASSWORD_RESET_EXPIRY`  | Password reset expiration             | `1h`            | `1h`                      |
| `FLUXBASE_AUTH_PASSWORD_MIN_LENGTH`    | Minimum password length               | `12`            | `8`, `16`                 |
| `FLUXBASE_AUTH_BCRYPT_COST`            | Bcrypt cost factor (4-31)             | `10`            | `10`, `12`                |
| `FLUXBASE_AUTH_SIGNUP_ENABLED`         | Enable user registration              | `true`          | `true`, `false`           |
| `FLUXBASE_AUTH_MAGIC_LINK_ENABLED`     | Enable magic link auth                | `true`          | `true`, `false`           |
| `FLUXBASE_AUTH_TOTP_ISSUER`            | 2FA TOTP issuer name                  | `Fluxbase`      | `MyApp`                   |
| `FLUXBASE_AUTH_ALLOW_USER_CLIENT_KEYS` | Allow users to create API client keys | `true`          | `true`, `false`           |

**OAuth/OIDC Providers:**

OAuth providers are configured via YAML config file. For simple setups, use environment variables for well-known providers:

| Variable                             | Description                  | Default | Example                   |
| ------------------------------------ | ---------------------------- | ------- | ------------------------- |
| `FLUXBASE_AUTH_GOOGLE_CLIENT_ID`     | Google OAuth client ID       | `""`    | Your Google client ID     |
| `FLUXBASE_AUTH_GOOGLE_CLIENT_SECRET` | Google OAuth client secret   | `""`    | Your Google client secret |
| `FLUXBASE_AUTH_APPLE_CLIENT_ID`      | Apple Sign In client ID      | `""`    | Your Apple Services ID    |
| `FLUXBASE_AUTH_MICROSOFT_CLIENT_ID`  | Microsoft/Azure AD client ID | `""`    | Your Microsoft client ID  |

For custom OIDC providers, use the YAML config file with `oauth_providers` array (see YAML example above).

**SAML SSO Providers (Enterprise):**

SAML providers are configured via YAML config file with `saml_providers` array:

```yaml
auth:
  saml_providers:
    - name: okta # Provider identifier
      enabled: true
      idp_metadata_url: https://... # IdP metadata URL (recommended)
      # OR idp_metadata_xml: "<EntityDescriptor>..."  # Inline metadata XML
      entity_id: urn:fluxbase:sp # Your SP entity ID
      acs_url: https://api.example.com/auth/saml/okta/callback
      auto_create_users: true # Create user if not exists
      default_role: authenticated # Role for new users
      allow_idp_initiated: false # Disable for security
      allow_dashboard_login: false # Allow admin SSO
      allow_app_login: true # Allow app user SSO
      group_attribute: groups # SAML attribute for groups
      required_groups: ["fluxbase-users"] # User must be in one of these
      required_groups_all: [] # User must be in ALL of these
      denied_groups: [] # Deny if in any of these
      attribute_mapping: # Map SAML attributes to user fields
        email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
```

**Security Best Practices:**

- Use a strong, random JWT secret (min 32 characters): `openssl rand -base64 32`
- Rotate JWT secrets periodically
- Use short access token expiry (15-30 minutes)
- Use longer refresh token expiry (7-30 days)
- Disable `allow_idp_initiated` for SAML providers to prevent replay attacks

### Storage

| Variable                               | Description                  | Default                               | Example                                    |
| -------------------------------------- | ---------------------------- | ------------------------------------- | ------------------------------------------ |
| `FLUXBASE_STORAGE_ENABLED`             | Enable storage               | `true`                                | `true`, `false`                            |
| `FLUXBASE_STORAGE_PROVIDER`            | Storage backend              | `local`                               | `local`, `s3`                              |
| `FLUXBASE_STORAGE_LOCAL_PATH`          | Local storage path           | `./storage`                           | `/var/lib/fluxbase/storage`                |
| `FLUXBASE_STORAGE_MAX_UPLOAD_SIZE`     | Max upload size (bytes)      | `2147483648`                          | `2147483648` (2GB)                         |
| `FLUXBASE_STORAGE_S3_ENDPOINT`         | S3 endpoint                  | -                                     | `s3.amazonaws.com`                         |
| `FLUXBASE_STORAGE_S3_ACCESS_KEY`       | S3 access key                | -                                     | `AKIAIOSFODNN7EXAMPLE`                     |
| `FLUXBASE_STORAGE_S3_SECRET_KEY`       | S3 secret key                | -                                     | `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY` |
| `FLUXBASE_STORAGE_S3_REGION`           | S3 region                    | -                                     | `us-west-2`                                |
| `FLUXBASE_STORAGE_S3_BUCKET`           | S3 bucket name               | -                                     | `my-bucket`                                |
| `FLUXBASE_STORAGE_S3_FORCE_PATH_STYLE` | Use path-style S3 addressing | `true`                                | `true`, `false`                            |
| `FLUXBASE_STORAGE_DEFAULT_BUCKETS`     | Auto-create these buckets    | `["uploads", "temp-files", "public"]` | -                                          |

**S3-Compatible Services:**

- AWS S3
- MinIO (local development): `http://localhost:9000`
- DigitalOcean Spaces: `https://nyc3.digitaloceanspaces.com`
- Wasabi: `https://s3.wasabisys.com`
- Backblaze B2: `https://s3.us-west-002.backblazeb2.com`

**Image Transformations:**

| Variable                                       | Description                                 | Default                          | Example            |
| ---------------------------------------------- | ------------------------------------------- | -------------------------------- | ------------------ |
| `FLUXBASE_STORAGE_TRANSFORMS_ENABLED`          | Enable on-the-fly image transformations     | `true`                           | `true`, `false`    |
| `FLUXBASE_STORAGE_TRANSFORMS_DEFAULT_QUALITY`  | Default output quality (1-100)              | `80`                             | `85`               |
| `FLUXBASE_STORAGE_TRANSFORMS_MAX_WIDTH`        | Maximum output width (pixels)               | `4096`                           | `8192`             |
| `FLUXBASE_STORAGE_TRANSFORMS_MAX_HEIGHT`       | Maximum output height (pixels)              | `4096`                           | `8192`             |
| `FLUXBASE_STORAGE_TRANSFORMS_ALLOWED_FORMATS`  | Allowed output formats                      | `["webp", "jpg", "png", "avif"]` | -                  |
| `FLUXBASE_STORAGE_TRANSFORMS_MAX_TOTAL_PIXELS` | Max total pixels (width × height)           | `16000000` (16MP)                | `25000000`         |
| `FLUXBASE_STORAGE_TRANSFORMS_BUCKET_SIZE`      | Dimension bucketing size (cache efficiency) | `50`                             | `100`              |
| `FLUXBASE_STORAGE_TRANSFORMS_RATE_LIMIT`       | Transforms per minute per user              | `60`                             | `120`              |
| `FLUXBASE_STORAGE_TRANSFORMS_TIMEOUT`          | Max transform duration                      | `30s`                            | `60s`              |
| `FLUXBASE_STORAGE_TRANSFORMS_MAX_CONCURRENT`   | Max concurrent transforms                   | `4`                              | `8`                |
| `FLUXBASE_STORAGE_TRANSFORMS_CACHE_ENABLED`    | Enable transform caching                    | `true`                           | `true`, `false`    |
| `FLUXBASE_STORAGE_TRANSFORMS_CACHE_TTL`        | Cache TTL                                   | `24h`                            | `48h`              |
| `FLUXBASE_STORAGE_TRANSFORMS_CACHE_MAX_SIZE`   | Max cache size (bytes)                      | `1073741824` (1GB)               | `2147483648` (2GB) |

### Realtime

| Variable                                     | Description                  | Default          | Example         |
| -------------------------------------------- | ---------------------------- | ---------------- | --------------- |
| `FLUXBASE_REALTIME_ENABLED`                  | Enable realtime              | `true`           | `true`, `false` |
| `FLUXBASE_REALTIME_MAX_CONNECTIONS`          | Max WebSocket connections    | `1000`           | `5000`          |
| `FLUXBASE_REALTIME_MAX_CONNECTIONS_PER_USER` | Max connections per user     | `10`             | `20`            |
| `FLUXBASE_REALTIME_MAX_CONNECTIONS_PER_IP`   | Max connections per IP       | `20`             | `50`            |
| `FLUXBASE_REALTIME_PING_INTERVAL`            | Ping interval                | `30s`            | `30s`           |
| `FLUXBASE_REALTIME_PONG_TIMEOUT`             | Pong timeout                 | `60s`            | `60s`           |
| `FLUXBASE_REALTIME_READ_BUFFER_SIZE`         | WebSocket read buffer        | `1024`           | `2048`          |
| `FLUXBASE_REALTIME_WRITE_BUFFER_SIZE`        | WebSocket write buffer       | `1024`           | `2048`          |
| `FLUXBASE_REALTIME_MESSAGE_SIZE_LIMIT`       | Max message size (bytes)     | `524288` (512KB) | `1048576`       |
| `FLUXBASE_REALTIME_CHANNEL_BUFFER_SIZE`      | Channel buffer size          | `100`            | `200`           |
| `FLUXBASE_REALTIME_RLS_CACHE_SIZE`           | RLS permission cache entries | `100000`         | `200000`        |
| `FLUXBASE_REALTIME_RLS_CACHE_TTL`            | RLS cache TTL                | `30s`            | `60s`           |

**Advanced Realtime Settings:**

| Variable                                      | Description                                      | Default | Example |
| --------------------------------------------- | ------------------------------------------------ | ------- | ------- |
| `FLUXBASE_REALTIME_LISTENER_POOL_SIZE`        | LISTEN connections for redundancy                | `2`     | `4`     |
| `FLUXBASE_REALTIME_NOTIFICATION_WORKERS`      | Workers for parallel notification processing     | `4`     | `8`     |
| `FLUXBASE_REALTIME_NOTIFICATION_QUEUE_SIZE`   | Queue size per notification worker               | `1000`  | `2000`  |
| `FLUXBASE_REALTIME_CLIENT_MESSAGE_QUEUE_SIZE` | Per-client message queue for async sending       | `256`   | `512`   |
| `FLUXBASE_REALTIME_SLOW_CLIENT_THRESHOLD`     | Queue length threshold for slow client detection | `100`   | `200`   |
| `FLUXBASE_REALTIME_SLOW_CLIENT_TIMEOUT`       | Duration before disconnecting slow clients       | `30s`   | `60s`   |

### Migrations API

| Variable                                  | Description                                     | Default                                                            | Example                    |
| ----------------------------------------- | ----------------------------------------------- | ------------------------------------------------------------------ | -------------------------- |
| `FLUXBASE_MIGRATIONS_ENABLED`             | Enable migrations API                           | `true`                                                             | `true`, `false`            |
| `FLUXBASE_MIGRATIONS_ALLOWED_IP_RANGES`   | IP CIDR ranges allowed to access migrations API | `["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/16", "127.0.0.0/8"]` | -                          |
| `FLUXBASE_MIGRATIONS_REQUIRE_SERVICE_KEY` | Require service key authentication              | `true`                                                             | Always `true` for security |

:::note[Migrations Security]
The migrations API requires both IP allowlist and service key authentication. This ensures only trusted CI/CD pipelines can run migrations.
:::

### Admin UI

| Variable                 | Description     | Default | Example         |
| ------------------------ | --------------- | ------- | --------------- |
| `FLUXBASE_ADMIN_ENABLED` | Enable Admin UI | `false` | `true`, `false` |

### Logging

| Variable                                    | Description             | Default    | Example                          |
| ------------------------------------------- | ----------------------- | ---------- | -------------------------------- |
| `FLUXBASE_LOGGING_CONSOLE_ENABLED`          | Enable console logging  | `true`     | `true`, `false`                  |
| `FLUXBASE_LOGGING_CONSOLE_LEVEL`            | Console log level       | `info`     | `debug`, `info`, `warn`, `error` |
| `FLUXBASE_LOGGING_CONSOLE_FORMAT`           | Console log format      | `console`  | `console`, `json`                |
| `FLUXBASE_LOGGING_BACKEND`                  | Log storage backend     | `postgres` | `postgres`, `s3`, `local`        |
| `FLUXBASE_LOGGING_SYSTEM_RETENTION_DAYS`    | System log retention    | `7`        | `7`                              |
| `FLUXBASE_LOGGING_HTTP_RETENTION_DAYS`      | HTTP log retention      | `30`       | `30`                             |
| `FLUXBASE_LOGGING_SECURITY_RETENTION_DAYS`  | Security log retention  | `90`       | `90`                             |
| `FLUXBASE_LOGGING_EXECUTION_RETENTION_DAYS` | Execution log retention | `30`       | `30`                             |

### CORS

| Variable                          | Description                       | Default                                        | Example                                 |
| --------------------------------- | --------------------------------- | ---------------------------------------------- | --------------------------------------- |
| `FLUXBASE_CORS_ALLOWED_ORIGINS`   | Allowed origins (comma-separated) | `http://localhost:5173,http://localhost:8080`  | `http://localhost:3000,https://app.com` |
| `FLUXBASE_CORS_ALLOWED_METHODS`   | Allowed HTTP methods              | `GET,POST,PUT,PATCH,DELETE,OPTIONS`            | `GET,POST,PUT,DELETE`                   |
| `FLUXBASE_CORS_ALLOWED_HEADERS`   | Allowed headers                   | `Origin,Content-Type,Accept,Authorization,...` | `Authorization,Content-Type`            |
| `FLUXBASE_CORS_ALLOW_CREDENTIALS` | Allow credentials                 | `true`                                         | `true`, `false`                         |
| `FLUXBASE_CORS_MAX_AGE`           | Preflight cache time (seconds)    | `300`                                          | `86400`                                 |

### Security

| Variable                                     | Description                                                    | Default | Example                   |
| -------------------------------------------- | -------------------------------------------------------------- | ------- | ------------------------- |
| `FLUXBASE_SECURITY_SETUP_TOKEN`              | Token for admin dashboard setup (required to enable dashboard) | `""`    | `openssl rand -base64 32` |
| `FLUXBASE_SECURITY_ENABLE_GLOBAL_RATE_LIMIT` | Enable global API rate limiting                                | `false` | `true`, `false`           |
| `FLUXBASE_SECURITY_ADMIN_SETUP_RATE_LIMIT`   | Max attempts for admin setup                                   | `5`     | `5`                       |
| `FLUXBASE_SECURITY_ADMIN_SETUP_RATE_WINDOW`  | Time window for admin setup rate limit                         | `15m`   | `15m`                     |
| `FLUXBASE_SECURITY_AUTH_LOGIN_RATE_LIMIT`    | Max attempts for auth login                                    | `10`    | `10`                      |
| `FLUXBASE_SECURITY_AUTH_LOGIN_RATE_WINDOW`   | Time window for auth login rate limit                          | `1m`    | `1m`                      |
| `FLUXBASE_SECURITY_ADMIN_LOGIN_RATE_LIMIT`   | Max attempts for admin login                                   | `10`    | `10`                      |
| `FLUXBASE_SECURITY_ADMIN_LOGIN_RATE_WINDOW`  | Time window for admin login rate limit                         | `1m`    | `1m`                      |

:::caution[Required for Admin Dashboard]
`FLUXBASE_SECURITY_SETUP_TOKEN` must be set to enable the admin dashboard. Generate a secure token with `openssl rand -base64 32`.
:::

**CAPTCHA Configuration (Bot Protection):**

| Variable                                      | Description                                      | Default                                               | Example                                        |
| --------------------------------------------- | ------------------------------------------------ | ----------------------------------------------------- | ---------------------------------------------- |
| `FLUXBASE_SECURITY_CAPTCHA_ENABLED`           | Enable CAPTCHA verification                      | `false`                                               | `true`, `false`                                |
| `FLUXBASE_SECURITY_CAPTCHA_PROVIDER`          | CAPTCHA provider                                 | `hcaptcha`                                            | `hcaptcha`, `recaptcha_v3`, `turnstile`, `cap` |
| `FLUXBASE_SECURITY_CAPTCHA_SITE_KEY`          | Public site key (for frontend)                   | `""`                                                  | Your site key                                  |
| `FLUXBASE_SECURITY_CAPTCHA_SECRET_KEY`        | Secret key (for server verification)             | `""`                                                  | Your secret key                                |
| `FLUXBASE_SECURITY_CAPTCHA_SCORE_THRESHOLD`   | Min score for reCAPTCHA v3 (0.0-1.0)             | `0.5`                                                 | `0.7`                                          |
| `FLUXBASE_SECURITY_CAPTCHA_ENDPOINTS`         | Endpoints requiring CAPTCHA                      | `["signup", "login", "password_reset", "magic_link"]` | -                                              |
| `FLUXBASE_SECURITY_CAPTCHA_CAP_SERVER_URL`    | URL for self-hosted Cap server                   | `""`                                                  | `http://cap:3000`                              |
| `FLUXBASE_SECURITY_CAPTCHA_CAP_API_KEY`       | API key for Cap server                           | `""`                                                  | Your Cap API key                               |
| `FLUXBASE_SECURITY_CAPTCHA_TEST_BYPASS_TOKEN` | Test token that bypasses verification (dev only) | `""`                                                  | Leave empty in production                      |

**Supported CAPTCHA Providers:**

- **hCaptcha** - Privacy-focused CAPTCHA (recommended)
- **reCAPTCHA v3** - Google's invisible CAPTCHA with risk scoring
- **Turnstile** - Cloudflare's privacy-preserving alternative
- **Cap** - Self-hosted proof-of-work CAPTCHA

### AI Chatbots

| Variable                             | Description                           | Default      | Example         |
| ------------------------------------ | ------------------------------------- | ------------ | --------------- |
| `FLUXBASE_AI_ENABLED`                | Enable AI chatbot functionality       | `true`       | `true`, `false` |
| `FLUXBASE_AI_CHATBOTS_DIR`           | Directory for chatbot definitions     | `./chatbots` | `./chatbots`    |
| `FLUXBASE_AI_AUTO_LOAD_ON_BOOT`      | Load chatbots from filesystem at boot | `true`       | `true`, `false` |
| `FLUXBASE_AI_DEFAULT_MAX_TOKENS`     | Default max tokens per request        | `4096`       | `4096`          |
| `FLUXBASE_AI_QUERY_TIMEOUT`          | SQL query execution timeout           | `30s`        | `30s`           |
| `FLUXBASE_AI_MAX_ROWS_PER_QUERY`     | Max rows returned per query           | `1000`       | `1000`          |
| `FLUXBASE_AI_CONVERSATION_CACHE_TTL` | TTL for conversation cache            | `30m`        | `1h`            |
| `FLUXBASE_AI_MAX_CONVERSATION_TURNS` | Max turns per conversation            | `50`         | `50`            |

**AI Provider Configuration:**

| Variable                       | Description                  | Default | Example                     |
| ------------------------------ | ---------------------------- | ------- | --------------------------- |
| `FLUXBASE_AI_PROVIDER_ENABLED` | Enable config-based provider | `false` | `true`, `false`             |
| `FLUXBASE_AI_PROVIDER_TYPE`    | Provider type                | `""`    | `openai`, `azure`, `ollama` |
| `FLUXBASE_AI_PROVIDER_NAME`    | Display name for provider    | `""`    | `Default Provider`          |
| `FLUXBASE_AI_PROVIDER_MODEL`   | Default model                | `""`    | `gpt-4-turbo`               |

**OpenAI Settings:**

| Variable                             | Description                           | Default | Example                     |
| ------------------------------------ | ------------------------------------- | ------- | --------------------------- |
| `FLUXBASE_AI_OPENAI_API_KEY`         | OpenAI API key                        | `""`    | `sk-...`                    |
| `FLUXBASE_AI_OPENAI_ORGANIZATION_ID` | OpenAI organization ID                | `""`    | `org-...`                   |
| `FLUXBASE_AI_OPENAI_BASE_URL`        | Custom base URL (for compatible APIs) | `""`    | `https://api.openai.com/v1` |

**Azure OpenAI Settings:**

| Variable                            | Description           | Default | Example                                  |
| ----------------------------------- | --------------------- | ------- | ---------------------------------------- |
| `FLUXBASE_AI_AZURE_API_KEY`         | Azure OpenAI API key  | `""`    | Your API key                             |
| `FLUXBASE_AI_AZURE_ENDPOINT`        | Azure OpenAI endpoint | `""`    | `https://your-resource.openai.azure.com` |
| `FLUXBASE_AI_AZURE_DEPLOYMENT_NAME` | Azure deployment name | `""`    | `gpt-4-deployment`                       |
| `FLUXBASE_AI_AZURE_API_VERSION`     | Azure API version     | `""`    | `2024-02-15-preview`                     |

**Ollama Settings:**

| Variable                      | Description       | Default | Example                  |
| ----------------------------- | ----------------- | ------- | ------------------------ |
| `FLUXBASE_AI_OLLAMA_ENDPOINT` | Ollama endpoint   | `""`    | `http://localhost:11434` |
| `FLUXBASE_AI_OLLAMA_MODEL`    | Ollama model name | `""`    | `llama2`, `mistral`      |

**Embedding Configuration (Vector Search):**

| Variable                                      | Description                                   | Default                 | Example                     |
| --------------------------------------------- | --------------------------------------------- | ----------------------- | --------------------------- |
| `FLUXBASE_AI_EMBEDDING_ENABLED`               | Enable embedding generation for vector search | `false`                 | `true`, `false`             |
| `FLUXBASE_AI_EMBEDDING_PROVIDER`              | Embedding provider                            | `""` (uses AI provider) | `openai`, `azure`, `ollama` |
| `FLUXBASE_AI_EMBEDDING_MODEL`                 | Embedding model                               | `""` (provider default) | `text-embedding-3-small`    |
| `FLUXBASE_AI_AZURE_EMBEDDING_DEPLOYMENT_NAME` | Separate Azure deployment for embeddings      | `""`                    | `text-embedding-ada-002`    |

**OCR Configuration (Knowledge Base PDF Extraction):**

| Variable                    | Description                     | Default     | Example                 |
| --------------------------- | ------------------------------- | ----------- | ----------------------- |
| `FLUXBASE_AI_OCR_ENABLED`   | Enable OCR for image-based PDFs | `true`      | `true`, `false`         |
| `FLUXBASE_AI_OCR_PROVIDER`  | OCR provider                    | `tesseract` | `tesseract`             |
| `FLUXBASE_AI_OCR_LANGUAGES` | Default OCR languages           | `["eng"]`   | `["eng", "deu", "fra"]` |

**Sync Security:**

| Variable                             | Description                             | Default                                                            | Example |
| ------------------------------------ | --------------------------------------- | ------------------------------------------------------------------ | ------- |
| `FLUXBASE_AI_SYNC_ALLOWED_IP_RANGES` | IP CIDR ranges allowed to sync chatbots | `["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/16", "127.0.0.0/8"]` | -       |

### Edge Functions

| Variable                                    | Description                              | Default                                                            | Example         |
| ------------------------------------------- | ---------------------------------------- | ------------------------------------------------------------------ | --------------- |
| `FLUXBASE_FUNCTIONS_ENABLED`                | Enable edge functions                    | `true`                                                             | `true`, `false` |
| `FLUXBASE_FUNCTIONS_FUNCTIONS_DIR`          | Directory for function files             | `./functions`                                                      | `./functions`   |
| `FLUXBASE_FUNCTIONS_AUTO_LOAD_ON_BOOT`      | Load functions from filesystem at boot   | `true`                                                             | `true`, `false` |
| `FLUXBASE_FUNCTIONS_DEFAULT_TIMEOUT`        | Default function timeout (seconds)       | `30`                                                               | `30`            |
| `FLUXBASE_FUNCTIONS_MAX_TIMEOUT`            | Maximum function timeout (seconds)       | `300`                                                              | `300`           |
| `FLUXBASE_FUNCTIONS_DEFAULT_MEMORY_LIMIT`   | Default memory limit (MB)                | `128`                                                              | `256`           |
| `FLUXBASE_FUNCTIONS_MAX_MEMORY_LIMIT`       | Maximum memory limit (MB)                | `1024`                                                             | `2048`          |
| `FLUXBASE_FUNCTIONS_MAX_OUTPUT_SIZE`        | Maximum output size (bytes, 0=unlimited) | `10485760` (10MB)                                                  | `20971520`      |
| `FLUXBASE_FUNCTIONS_SYNC_ALLOWED_IP_RANGES` | IP CIDR ranges allowed to sync functions | `["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/16", "127.0.0.0/8"]` | -               |

### Deno Runtime

Global settings for the Deno runtime used by edge functions and background jobs.

| Variable                     | Description                                   | Default | Example                         |
| ---------------------------- | --------------------------------------------- | ------- | ------------------------------- |
| `FLUXBASE_DENO_NPM_REGISTRY` | Custom npm registry URL for `npm:` specifiers | `""`    | `https://npm.your-company.com/` |
| `FLUXBASE_DENO_JSR_REGISTRY` | Custom JSR registry URL for `jsr:` specifiers | `""`    | `https://jsr.your-company.com/` |

**Air-Gapped Environments:** Set these to your private registry URLs for environments without internet access. See [Edge Functions Air-Gapped Guide](/guides/edge-functions/#air-gapped--private-registry-environments).

### RPC (Remote Procedures)

| Variable                                  | Description                               | Default                                                            | Example         |
| ----------------------------------------- | ----------------------------------------- | ------------------------------------------------------------------ | --------------- |
| `FLUXBASE_RPC_ENABLED`                    | Enable RPC functionality                  | `true`                                                             | `true`, `false` |
| `FLUXBASE_RPC_PROCEDURES_DIR`             | Directory for RPC procedure definitions   | `./rpc`                                                            | `./rpc`         |
| `FLUXBASE_RPC_AUTO_LOAD_ON_BOOT`          | Load procedures from filesystem at boot   | `true`                                                             | `true`, `false` |
| `FLUXBASE_RPC_DEFAULT_MAX_EXECUTION_TIME` | Default max execution time                | `30s`                                                              | `60s`           |
| `FLUXBASE_RPC_MAX_MAX_EXECUTION_TIME`     | Maximum allowed execution time            | `5m`                                                               | `10m`           |
| `FLUXBASE_RPC_DEFAULT_MAX_ROWS`           | Default max rows returned                 | `1000`                                                             | `5000`          |
| `FLUXBASE_RPC_SYNC_ALLOWED_IP_RANGES`     | IP CIDR ranges allowed to sync procedures | `["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/16", "127.0.0.0/8"]` | -               |

### Background Jobs

| Variable                                     | Description                                   | Default    | Example                              |
| -------------------------------------------- | --------------------------------------------- | ---------- | ------------------------------------ |
| `FLUXBASE_JOBS_ENABLED`                      | Enable background jobs                        | `true`     | `true`, `false`                      |
| `FLUXBASE_JOBS_JOBS_DIR`                     | Directory for job definitions                 | `./jobs`   | `./jobs`                             |
| `FLUXBASE_JOBS_AUTO_LOAD_ON_BOOT`            | Load jobs from filesystem at boot             | `true`     | `true`, `false`                      |
| `FLUXBASE_JOBS_WORKER_MODE`                  | Worker mode                                   | `embedded` | `embedded`, `standalone`, `disabled` |
| `FLUXBASE_JOBS_EMBEDDED_WORKER_COUNT`        | Number of embedded workers                    | `4`        | `8`                                  |
| `FLUXBASE_JOBS_MAX_CONCURRENT_PER_WORKER`    | Max concurrent jobs per worker                | `5`        | `10`                                 |
| `FLUXBASE_JOBS_MAX_CONCURRENT_PER_NAMESPACE` | Max concurrent jobs per namespace             | `20`       | `50`                                 |
| `FLUXBASE_JOBS_DEFAULT_MAX_DURATION`         | Default job timeout                           | `5m`       | `10m`                                |
| `FLUXBASE_JOBS_MAX_MAX_DURATION`             | Maximum allowed job timeout                   | `1h`       | `2h`                                 |
| `FLUXBASE_JOBS_DEFAULT_PROGRESS_TIMEOUT`     | Progress reporting timeout                    | `5m`       | `10m`                                |
| `FLUXBASE_JOBS_POLL_INTERVAL`                | Worker poll interval                          | `1s`       | `500ms`                              |
| `FLUXBASE_JOBS_WORKER_HEARTBEAT_INTERVAL`    | Worker heartbeat interval                     | `10s`      | `15s`                                |
| `FLUXBASE_JOBS_WORKER_TIMEOUT`               | Worker considered dead after                  | `30s`      | `60s`                                |
| `FLUXBASE_JOBS_GRACEFUL_SHUTDOWN_TIMEOUT`    | Time to wait for running jobs during shutdown | `5m`       | `10m`                                |

**Execution Log Retention:**

| Variable                                      | Description                                  | Default | Example |
| --------------------------------------------- | -------------------------------------------- | ------- | ------- |
| `FLUXBASE_JOBS_FUNCTIONS_LOGS_RETENTION_DAYS` | Retention for function execution logs (days) | `30`    | `60`    |
| `FLUXBASE_JOBS_RPC_LOGS_RETENTION_DAYS`       | Retention for RPC execution logs (days)      | `30`    | `60`    |
| `FLUXBASE_JOBS_JOBS_LOGS_RETENTION_DAYS`      | Retention for job execution logs (days)      | `30`    | `60`    |

**Sync Security:**

| Variable                               | Description                         | Default                                                            | Example |
| -------------------------------------- | ----------------------------------- | ------------------------------------------------------------------ | ------- |
| `FLUXBASE_JOBS_SYNC_ALLOWED_IP_RANGES` | IP CIDR ranges allowed to sync jobs | `["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/16", "127.0.0.0/8"]` | -       |

### Horizontal Scaling

| Variable                                            | Description                                     | Default | Example                      |
| --------------------------------------------------- | ----------------------------------------------- | ------- | ---------------------------- |
| `FLUXBASE_SCALING_WORKER_ONLY`                      | Disable API server, only run job workers        | `false` | `true`, `false`              |
| `FLUXBASE_SCALING_DISABLE_SCHEDULER`                | Disable cron job scheduler on this instance     | `false` | `true`, `false`              |
| `FLUXBASE_SCALING_DISABLE_REALTIME`                 | Disable realtime/WebSocket listener             | `false` | `true`, `false`              |
| `FLUXBASE_SCALING_ENABLE_SCHEDULER_LEADER_ELECTION` | Enable PostgreSQL advisory lock leader election | `false` | `true`, `false`              |
| `FLUXBASE_SCALING_BACKEND`                          | Distributed state backend                       | `local` | `local`, `postgres`, `redis` |
| `FLUXBASE_SCALING_REDIS_URL`                        | Redis/Dragonfly connection URL                  | `""`    | `redis://dragonfly:6379`     |

**Backend Options:**

- `local` - In-memory storage (single instance only, default)
- `postgres` - Uses PostgreSQL for distributed state (no extra dependencies)
- `redis` - Uses Redis-compatible backend (Dragonfly recommended for performance)

**What's Distributed:**

| Feature                | Description                                   |
| ---------------------- | --------------------------------------------- |
| Rate limiting          | Shared counters across all instances          |
| Realtime broadcasts    | Cross-instance pub/sub for application events |
| Scheduler coordination | Leader election prevents duplicate cron jobs  |
| Nonce validation       | PostgreSQL-backed for stateless auth flows    |

**CLI Flags:**

| Flag                       | Description                                         |
| -------------------------- | --------------------------------------------------- |
| `--worker-only`            | Disable API server, only run background job workers |
| `--disable-scheduler`      | Disable cron job scheduler on this instance         |
| `--disable-realtime`       | Disable realtime/WebSocket listener                 |
| `--enable-leader-election` | Enable PostgreSQL advisory lock leader election     |

**Example Production Configuration:**

```bash
# Multi-instance with PostgreSQL backend
FLUXBASE_SCALING_BACKEND=postgres
FLUXBASE_SCALING_ENABLE_SCHEDULER_LEADER_ELECTION=true

# Or with Redis/Dragonfly for high-scale (1000+ req/s)
FLUXBASE_SCALING_BACKEND=redis
FLUXBASE_SCALING_REDIS_URL=redis://:password@dragonfly:6379
FLUXBASE_SCALING_ENABLE_SCHEDULER_LEADER_ELECTION=true
```

### OpenTelemetry Tracing

| Variable                        | Description                  | Default          | Example         |
| ------------------------------- | ---------------------------- | ---------------- | --------------- |
| `FLUXBASE_TRACING_ENABLED`      | Enable OpenTelemetry tracing | `false`          | `true`, `false` |
| `FLUXBASE_TRACING_ENDPOINT`     | OTLP gRPC endpoint           | `localhost:4317` | `jaeger:4317`   |
| `FLUXBASE_TRACING_SERVICE_NAME` | Service name for traces      | `fluxbase`       | `fluxbase`      |
| `FLUXBASE_TRACING_ENVIRONMENT`  | Environment name             | `development`    | `production`    |
| `FLUXBASE_TRACING_SAMPLE_RATE`  | Sample rate (0.0-1.0)        | `1.0`            | `0.1` (10%)     |
| `FLUXBASE_TRACING_INSECURE`     | Use insecure connection      | `true`           | `false`         |

### API Pagination

| Variable                         | Description                                             | Default | Example |
| -------------------------------- | ------------------------------------------------------- | ------- | ------- |
| `FLUXBASE_API_MAX_PAGE_SIZE`     | Max rows per request (-1 = unlimited)                   | `1000`  | `1000`  |
| `FLUXBASE_API_MAX_TOTAL_RESULTS` | Max total retrievable rows (-1 = unlimited)             | `10000` | `10000` |
| `FLUXBASE_API_DEFAULT_PAGE_SIZE` | Auto-applied limit when not specified (-1 = no default) | `1000`  | `100`   |

### GraphQL

| Variable                          | Description                    | Default | Example         |
| --------------------------------- | ------------------------------ | ------- | --------------- |
| `FLUXBASE_GRAPHQL_ENABLED`        | Enable GraphQL API endpoint    | `true`  | `true`, `false` |
| `FLUXBASE_GRAPHQL_MAX_DEPTH`      | Maximum query depth            | `10`    | `15`            |
| `FLUXBASE_GRAPHQL_MAX_COMPLEXITY` | Maximum query complexity score | `1000`  | `2000`          |
| `FLUXBASE_GRAPHQL_INTROSPECTION`  | Enable GraphQL introspection   | `true`  | `false`         |

### Prometheus Metrics

| Variable                   | Description                        | Default    | Example         |
| -------------------------- | ---------------------------------- | ---------- | --------------- |
| `FLUXBASE_METRICS_ENABLED` | Enable Prometheus metrics endpoint | `true`     | `true`, `false` |
| `FLUXBASE_METRICS_PORT`    | Port for metrics server            | `9090`     | `9090`          |
| `FLUXBASE_METRICS_PATH`    | Path for metrics endpoint          | `/metrics` | `/metrics`      |

### MCP (Model Context Protocol)

| Variable                          | Description                       | Default           | Example                   |
| --------------------------------- | --------------------------------- | ----------------- | ------------------------- |
| `FLUXBASE_MCP_ENABLED`            | Enable MCP server                 | `true`            | `true`, `false`           |
| `FLUXBASE_MCP_BASE_PATH`          | Base path for MCP endpoints       | `/mcp`            | `/mcp`                    |
| `FLUXBASE_MCP_SESSION_TIMEOUT`    | Session timeout                   | `30m`             | `1h`                      |
| `FLUXBASE_MCP_MAX_MESSAGE_SIZE`   | Max message size (bytes)          | `10485760` (10MB) | `20971520`                |
| `FLUXBASE_MCP_RATE_LIMIT_PER_MIN` | Rate limit per minute per client  | `100`             | `200`                     |
| `FLUXBASE_MCP_ALLOWED_TOOLS`      | Allowed tools (empty = all)       | `[]`              | `["query", "storage"]`    |
| `FLUXBASE_MCP_ALLOWED_RESOURCES`  | Allowed resources (empty = all)   | `[]`              | `["schema", "functions"]` |
| `FLUXBASE_MCP_TOOLS_DIR`          | Directory for custom MCP tools    | `/app/mcp-tools`  | `./mcp-tools`             |
| `FLUXBASE_MCP_AUTO_LOAD_ON_BOOT`  | Auto-load custom tools on startup | `true`            | `true`, `false`           |

**MCP OAuth Configuration (OAuth 2.1 for MCP Clients):**

| Variable                                   | Description                                  | Default         | Example                  |
| ------------------------------------------ | -------------------------------------------- | --------------- | ------------------------ |
| `FLUXBASE_MCP_OAUTH_ENABLED`               | Enable OAuth 2.1 for MCP clients             | `true`          | `true`, `false`          |
| `FLUXBASE_MCP_OAUTH_DCR_ENABLED`           | Enable Dynamic Client Registration           | `true`          | `true`, `false`          |
| `FLUXBASE_MCP_OAUTH_TOKEN_EXPIRY`          | Access token lifetime                        | `1h`            | `2h`                     |
| `FLUXBASE_MCP_OAUTH_REFRESH_TOKEN_EXPIRY`  | Refresh token lifetime                       | `168h` (7 days) | `720h` (30 days)         |
| `FLUXBASE_MCP_OAUTH_ALLOWED_REDIRECT_URIS` | Allowed redirect URIs (empty = use defaults) | `[]`            | `["http://localhost:*"]` |

:::note[Zero-Config MCP Clients]
MCP OAuth is enabled by default with Dynamic Client Registration (DCR). This allows MCP clients like Claude Desktop to authenticate automatically without pre-registering client credentials.
:::

### Database Branching

| Variable                                     | Description                      | Default       | Example                     |
| -------------------------------------------- | -------------------------------- | ------------- | --------------------------- |
| `FLUXBASE_BRANCHING_ENABLED`                 | Enable database branching        | `false`       | `true`, `false`             |
| `FLUXBASE_BRANCHING_MAX_BRANCHES_PER_USER`   | Max branches per user            | `5`           | `10`                        |
| `FLUXBASE_BRANCHING_MAX_TOTAL_BRANCHES`      | Max total branches               | `50`          | `100`                       |
| `FLUXBASE_BRANCHING_DEFAULT_DATA_CLONE_MODE` | Default data clone mode          | `schema_only` | `schema_only`, `full_clone` |
| `FLUXBASE_BRANCHING_AUTO_DELETE_AFTER`       | Auto-delete branches after       | `0` (never)   | `24h`, `168h`               |
| `FLUXBASE_BRANCHING_DATABASE_PREFIX`         | Prefix for branch database names | `branch_`     | `branch_`                   |
| `FLUXBASE_BRANCHING_SEEDS_PATH`              | Path to seed data files          | `./seeds`     | `./seeds`                   |

### TLS/HTTPS (Upcoming)

| Variable               | Description          | Default | Example              |
| ---------------------- | -------------------- | ------- | -------------------- |
| `TLS_ENABLED`          | Enable TLS           | `false` | `true`, `false`      |
| `TLS_CERT_FILE`        | Path to certificate  | -       | `/etc/certs/tls.crt` |
| `TLS_KEY_FILE`         | Path to private key  | -       | `/etc/certs/tls.key` |
| `TLS_AUTO_CERT`        | Enable Let's Encrypt | `false` | `true`, `false`      |
| `TLS_AUTO_CERT_DOMAIN` | Domain for auto cert | -       | `example.com`        |

## Production Configuration

### Recommended Production Settings

```yaml
# General
base_url: http://fluxbase:8080
public_base_url: https://api.example.com
encryption_key: ${ENCRYPTION_KEY} # 32 bytes for AES-256

server:
  address: ":8080"
  read_timeout: 300s
  write_timeout: 300s

database:
  host: postgres
  port: 5432
  user: fluxbase
  password: ${DB_PASSWORD}
  database: fluxbase
  ssl_mode: require
  max_connections: 100
  min_connections: 20
  max_conn_lifetime: 30m

auth:
  jwt_secret: ${JWT_SECRET}
  jwt_expiry: 15m
  refresh_expiry: 168h # 7 days
  password_min_length: 12

storage:
  provider: s3
  max_upload_size: 2147483648 # 2GB
  s3_endpoint: s3.amazonaws.com
  s3_access_key: ${S3_ACCESS_KEY}
  s3_secret_key: ${S3_SECRET_KEY}
  s3_region: us-east-1
  s3_bucket: my-production-bucket
  s3_force_path_style: false # Use virtual-hosted style for AWS S3

realtime:
  enabled: true
  max_connections: 5000
  max_connections_per_user: 20

admin:
  enabled: false # Disable in production or protect behind VPN

security:
  setup_token: ${SETUP_TOKEN}
  enable_global_rate_limit: true

logging:
  console_level: info
  console_format: json

cors:
  allowed_origins: "https://app.example.com,https://www.example.com"
  allow_credentials: true

scaling:
  backend: postgres # or redis for high-scale
  enable_scheduler_leader_election: true
```

### Environment Variables (Production)

```bash
# .env.production
FLUXBASE_DATABASE_HOST=postgres
FLUXBASE_DATABASE_PORT=5432
FLUXBASE_DATABASE_USER=fluxbase
FLUXBASE_DATABASE_PASSWORD=${DB_PASSWORD}
FLUXBASE_DATABASE_DATABASE=fluxbase
FLUXBASE_DATABASE_SSL_MODE=require

FLUXBASE_AUTH_JWT_SECRET=${JWT_SECRET}

FLUXBASE_STORAGE_PROVIDER=s3
FLUXBASE_STORAGE_S3_ACCESS_KEY=${S3_ACCESS_KEY}
FLUXBASE_STORAGE_S3_SECRET_KEY=${S3_SECRET_KEY}
FLUXBASE_STORAGE_S3_BUCKET=production-bucket

FLUXBASE_LOGGING_CONSOLE_LEVEL=info
FLUXBASE_LOGGING_CONSOLE_FORMAT=json

FLUXBASE_CORS_ALLOWED_ORIGINS=https://app.example.com,https://www.example.com

FLUXBASE_SECURITY_ENABLE_GLOBAL_RATE_LIMIT=true
```

## Development Configuration

### Recommended Development Settings

```yaml
# General
base_url: http://localhost:8080
debug: true

server:
  address: ":8080"

database:
  host: localhost
  port: 5432
  user: fluxbase
  password: fluxbase
  database: fluxbase
  ssl_mode: disable
  max_connections: 20
  min_connections: 5

auth:
  jwt_secret: dev-secret-change-in-production
  jwt_expiry: 24h # Longer for development
  refresh_expiry: 720h # 30 days

storage:
  provider: local
  local_path: ./storage

realtime:
  enabled: true
  max_connections: 100

admin:
  enabled: true

security:
  setup_token: dev-setup-token-change-in-production
  enable_global_rate_limit: false

logging:
  console_level: debug
  console_format: console

cors:
  allowed_origins: "http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000"
  allow_credentials: true
```

## Docker Configuration

### Docker Compose Example

```yaml
version: "3.8"

services:
  fluxbase:
    image: ghcr.io/fluxbase-eu/fluxbase:latest
    environment:
      # Database
      FLUXBASE_DATABASE_HOST: postgres
      FLUXBASE_DATABASE_PORT: 5432
      FLUXBASE_DATABASE_USER: fluxbase
      FLUXBASE_DATABASE_PASSWORD: password
      FLUXBASE_DATABASE_DATABASE: fluxbase
      FLUXBASE_DATABASE_SSL_MODE: disable

      # Authentication
      FLUXBASE_AUTH_JWT_SECRET: ${JWT_SECRET}
      FLUXBASE_AUTH_JWT_EXPIRY: 15m
      FLUXBASE_AUTH_REFRESH_EXPIRY: 168h

      # Storage (MinIO)
      FLUXBASE_STORAGE_PROVIDER: s3
      FLUXBASE_STORAGE_S3_ENDPOINT: http://minio:9000
      FLUXBASE_STORAGE_S3_ACCESS_KEY: minioadmin
      FLUXBASE_STORAGE_S3_SECRET_KEY: minioadmin
      FLUXBASE_STORAGE_S3_BUCKET: fluxbase
      FLUXBASE_STORAGE_S3_FORCE_PATH_STYLE: true
      FLUXBASE_STORAGE_S3_REGION: us-east-1

      # Realtime
      FLUXBASE_REALTIME_ENABLED: true
      FLUXBASE_REALTIME_MAX_CONNECTIONS: 1000

      # Logging
      FLUXBASE_LOGGING_CONSOLE_LEVEL: info
      FLUXBASE_LOGGING_CONSOLE_FORMAT: json

      # CORS
      FLUXBASE_CORS_ALLOWED_ORIGINS: http://localhost:3000

    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./fluxbase.yaml:/app/fluxbase.yaml

  postgres:
    image: ghcr.io/fluxbase-eu/fluxbase-postgres:18
    environment:
      POSTGRES_DB: fluxbase
      POSTGRES_USER: fluxbase
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U fluxbase"]
      interval: 5s
      timeout: 5s
      retries: 5

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data

volumes:
  postgres_data:
  minio_data:
```

## Kubernetes Configuration

Helm chart configuration will be available in a future release.

## Configuration Priority

Configuration is loaded in the following order (later sources override earlier ones):

1. Default values (built-in)
2. Configuration file (`fluxbase.yaml`)
3. Environment variables
4. Command-line flags (if applicable)

## Validation

Fluxbase validates configuration on startup and will fail fast if:

- Required values are missing (e.g., `DATABASE_URL`, `JWT_SECRET`)
- Values are invalid (e.g., negative numbers, invalid formats)
- Database connection fails
- Storage backend is unreachable

## Hot Reload

Currently, Fluxbase does not support hot reloading of configuration. Restart the server after making configuration changes.

## Security Considerations

### Secrets Management

**Never commit secrets to version control!**

Use environment variables or secret management tools:

```bash
# Good: Load from environment
export JWT_SECRET=$(openssl rand -hex 32)
export DATABASE_URL="postgres://user:$(cat /run/secrets/db_password)@localhost/fluxbase"

# Bad: Hardcode in config file
jwt:
  secret: my-secret-key  # ❌ Don't do this!
```

### Production Secrets

Use a secrets management solution:

- **Kubernetes**: Use Secrets and ConfigMaps
- **Docker Swarm**: Use Docker Secrets
- **AWS**: Use AWS Secrets Manager or Parameter Store
- **HashiCorp Vault**: Enterprise secret management
- **Environment**: Use `.env` files (not in git) with proper permissions

## Troubleshooting

### Configuration Not Loading

```bash
# Check if file exists
ls -la fluxbase.yaml

# Validate YAML syntax
yamllint fluxbase.yaml

# Check environment variables
env | grep FLUXBASE
```

### Database Connection Issues

```bash
# Test connection
psql "$DATABASE_URL"

# Check connection string format
echo $DATABASE_URL
```

### CORS Issues

If you see CORS errors in the browser:

1. Check `CORS_ALLOWED_ORIGINS` includes your frontend URL
2. Ensure `CORS_ALLOW_CREDENTIALS` is `true` if sending cookies
3. Check browser console for specific CORS error

## Next Steps

- [Quick Start](/getting-started/quick-start/) - Get Fluxbase running in 5 minutes
- [Authentication](/guides/authentication/) - Set up JWT authentication
