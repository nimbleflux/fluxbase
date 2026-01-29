---
title: "Configuration Reference"
description: Complete reference for configuring Fluxbase via YAML config file or environment variables including database, authentication, storage, and realtime settings.
---

Complete reference for configuring Fluxbase via configuration file or environment variables.

## Configuration File

Create `fluxbase.yaml` in your working directory:

```yaml
# Server Configuration
server:
  port: 8080
  host: 0.0.0.0
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576 # 1MB

# Database Configuration
database:
  url: postgres://user:password@localhost:5432/fluxbase?sslmode=disable
  max_connections: 100
  idle_connections: 10
  connection_lifetime: 60m
  connection_timeout: 10s

# JWT Authentication
jwt:
  secret: your-secret-key-change-this-in-production
  access_token_expiry: 15m
  refresh_token_expiry: 7d
  issuer: fluxbase
  audience: authenticated

# Storage Configuration
storage:
  provider: local # "local" or "s3"
  local_path: ./storage
  max_upload_size: 10485760 # 10MB in bytes

  # S3 Configuration (when provider: s3)
  s3_endpoint: s3.amazonaws.com
  s3_access_key: ""
  s3_secret_key: ""
  s3_region: us-east-1
  s3_bucket: fluxbase
  s3_use_ssl: true

# Realtime Configuration
realtime:
  enabled: true
  heartbeat_interval: 30s
  max_connections: 1000
  read_buffer_size: 1024
  write_buffer_size: 1024

# Admin UI
admin:
  enabled: true
  path: /admin

# Logging
logging:
  level: info # debug, info, warn, error
  format: json # json or text
  output: stdout # stdout, stderr, or file path

# CORS Configuration
cors:
  enabled: true
  allowed_origins:
    - http://localhost:4000
    - http://localhost:5173
  allowed_methods:
    - GET
    - POST
    - PUT
    - PATCH
    - DELETE
    - OPTIONS
  allowed_headers:
    - Authorization
    - Content-Type
    - Accept
  exposed_headers:
    - Content-Range
    - X-Content-Range
  allow_credentials: true
  max_age: 86400 # 24 hours

# Rate Limiting
rate_limit:
  enabled: false
  requests_per_minute: 100
  burst: 200

# TLS/HTTPS (Upcoming)
tls:
  enabled: false
  cert_file: /path/to/cert.pem
  key_file: /path/to/key.pem
  auto_cert: false # Let's Encrypt
  auto_cert_domain: example.com
```

## Environment Variables

Environment variables take precedence over configuration file values.

### Server

| Variable               | Description      | Default   | Example   |
| ---------------------- | ---------------- | --------- | --------- |
| `PORT`                 | HTTP server port | `8080`    | `8080`    |
| `HOST`                 | HTTP server host | `0.0.0.0` | `0.0.0.0` |
| `SERVER_READ_TIMEOUT`  | Read timeout     | `30s`     | `30s`     |
| `SERVER_WRITE_TIMEOUT` | Write timeout    | `30s`     | `30s`     |
| `SERVER_IDLE_TIMEOUT`  | Idle timeout     | `120s`    | `120s`    |

### Database

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_DATABASE_HOST` | PostgreSQL host | `localhost` | `localhost` |
| `FLUXBASE_DATABASE_PORT` | PostgreSQL port | `5432` | `5432` |
| `FLUXBASE_DATABASE_USER` | Runtime database user | `postgres` | `fluxbase` |
| `FLUXBASE_DATABASE_PASSWORD` | Runtime user password | `postgres` | `your-password` |
| `FLUXBASE_DATABASE_DATABASE` | Database name | `fluxbase` | `fluxbase` |
| `FLUXBASE_DATABASE_SSL_MODE` | SSL mode | `disable` | `require` |
| `FLUXBASE_DATABASE_MAX_CONNECTIONS` | Max connection pool size | `25` | `100` |
| `FLUXBASE_DATABASE_MIN_CONNECTIONS` | Min connections in pool | `5` | `5` |
| `FLUXBASE_DATABASE_MAX_CONN_LIFETIME` | Connection max lifetime | `1h` | `1h` |
| `FLUXBASE_DATABASE_MAX_CONN_IDLE_TIME` | Connection max idle time | `30m` | `30m` |
| `FLUXBASE_DATABASE_HEALTH_CHECK_PERIOD` | Health check interval | `1m` | `1m` |
| `FLUXBASE_DATABASE_ADMIN_USER` | Admin user for migrations (defaults to USER) | `""` | `postgres` |
| `FLUXBASE_DATABASE_ADMIN_PASSWORD` | Admin user password (defaults to PASSWORD) | `""` | `admin-password` |
| `FLUXBASE_DATABASE_USER_MIGRATIONS_PATH` | Path to user-provided migrations | `/migrations/user` | `/app/migrations` |

**SSL Modes:**

- `disable` - No SSL (development only)
- `allow` - Prefer SSL if available
- `prefer` - Use SSL if available (default for many clients)
- `require` - Require SSL connection
- `verify-ca` - Require SSL and verify CA certificate
- `verify-full` - Require SSL and verify CA + hostname

### Authentication

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_AUTH_JWT_SECRET` | JWT signing key (min 32 chars) | **(required)** | `openssl rand -base64 32` |
| `FLUXBASE_AUTH_JWT_EXPIRY` | Access token expiration | `15m` | `15m`, `1h` |
| `FLUXBASE_AUTH_REFRESH_EXPIRY` | Refresh token expiration | `168h` (7 days) | `168h`, `720h` |
| `FLUXBASE_AUTH_MAGIC_LINK_EXPIRY` | Magic link expiration | `15m` | `15m` |
| `FLUXBASE_AUTH_PASSWORD_RESET_EXPIRY` | Password reset expiration | `1h` | `1h` |
| `FLUXBASE_AUTH_PASSWORD_MIN_LENGTH` | Minimum password length | `8` | `8` |
| `FLUXBASE_AUTH_BCRYPT_COST` | Bcrypt cost factor (4-31) | `10` | `10` |
| `FLUXBASE_AUTH_SIGNUP_ENABLED` | Enable user registration | `true` | `true`, `false` |
| `FLUXBASE_AUTH_MAGIC_LINK_ENABLED` | Enable magic link auth | `true` | `true`, `false` |
| `FLUXBASE_AUTH_TOTP_ISSUER` | 2FA TOTP issuer name | `Fluxbase` | `MyApp` |

**OAuth/OIDC Providers:**

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_AUTH_GOOGLE_CLIENT_ID` | Google OAuth client ID | `""` | Your Google client ID |
| `FLUXBASE_AUTH_APPLE_CLIENT_ID` | Apple Sign In client ID | `""` | Your Apple Services ID |
| `FLUXBASE_AUTH_MICROSOFT_CLIENT_ID` | Microsoft/Azure AD client ID | `""` | Your Microsoft client ID |

**Security Best Practices:**

- Use a strong, random JWT secret (min 32 characters): `openssl rand -base64 32`
- Rotate JWT secrets periodically
- Use short access token expiry (15-30 minutes)
- Use longer refresh token expiry (7-30 days)

### Storage

| Variable                              | Description                  | Default      | Example                                    |
| ------------------------------------- | ---------------------------- | ------------ | ------------------------------------------ |
| `FLUXBASE_STORAGE_ENABLED`            | Enable storage               | `true`       | `true`, `false`                            |
| `FLUXBASE_STORAGE_PROVIDER`           | Storage backend              | `local`      | `local`, `s3`                              |
| `FLUXBASE_STORAGE_LOCAL_PATH`         | Local storage path           | `./storage`  | `/var/lib/fluxbase/storage`                |
| `FLUXBASE_STORAGE_MAX_UPLOAD_SIZE`    | Max upload size (bytes)      | `2147483648` | `2147483648` (2GB)                         |
| `FLUXBASE_STORAGE_S3_ENDPOINT`        | S3 endpoint                  | -            | `s3.amazonaws.com`                         |
| `FLUXBASE_STORAGE_S3_ACCESS_KEY`      | S3 access key                | -            | `AKIAIOSFODNN7EXAMPLE`                     |
| `FLUXBASE_STORAGE_S3_SECRET_KEY`      | S3 secret key                | -            | `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY` |
| `FLUXBASE_STORAGE_S3_REGION`          | S3 region                    | -            | `us-west-2`                                |
| `FLUXBASE_STORAGE_S3_BUCKET`          | S3 bucket name               | -            | `my-bucket`                                |
| `FLUXBASE_STORAGE_S3_FORCE_PATH_STYLE`| Use path-style S3 addressing | `true`       | `true`, `false`                            |
| `FLUXBASE_STORAGE_DEFAULT_BUCKETS`    | Auto-create these buckets    | `["uploads", "temp-files", "public"]` | -               |

**S3-Compatible Services:**

- AWS S3
- MinIO (local development): `http://localhost:9000`
- DigitalOcean Spaces: `https://nyc3.digitaloceanspaces.com`
- Wasabi: `https://s3.wasabisys.com`
- Backblaze B2: `https://s3.us-west-002.backblazeb2.com`

### Realtime

| Variable                                      | Description                    | Default  | Example         |
| --------------------------------------------- | ------------------------------ | -------- | --------------- |
| `FLUXBASE_REALTIME_ENABLED`                   | Enable realtime                | `true`   | `true`, `false` |
| `FLUXBASE_REALTIME_MAX_CONNECTIONS`           | Max WebSocket connections      | `1000`   | `5000`          |
| `FLUXBASE_REALTIME_MAX_CONNECTIONS_PER_USER`  | Max connections per user       | `10`     | `20`            |
| `FLUXBASE_REALTIME_MAX_CONNECTIONS_PER_IP`    | Max connections per IP         | `20`     | `50`            |
| `FLUXBASE_REALTIME_PING_INTERVAL`             | Ping interval                  | `30s`    | `30s`           |
| `FLUXBASE_REALTIME_PONG_TIMEOUT`              | Pong timeout                   | `60s`    | `60s`           |
| `FLUXBASE_REALTIME_READ_BUFFER_SIZE`          | WebSocket read buffer          | `1024`   | `2048`          |
| `FLUXBASE_REALTIME_WRITE_BUFFER_SIZE`         | WebSocket write buffer         | `1024`   | `2048`          |
| `FLUXBASE_REALTIME_MESSAGE_SIZE_LIMIT`        | Max message size (bytes)       | `524288` | `1048576`       |
| `FLUXBASE_REALTIME_RLS_CACHE_SIZE`            | RLS permission cache entries   | `100000` | `200000`        |
| `FLUXBASE_REALTIME_RLS_CACHE_TTL`             | RLS cache TTL                  | `30s`    | `60s`           |

### Admin UI

| Variable               | Description     | Default  | Example         |
| ---------------------- | --------------- | -------- | --------------- |
| `FLUXBASE_ADMIN_ENABLED` | Enable Admin UI | `false`  | `true`, `false` |

### Logging

| Variable                              | Description                    | Default    | Example                          |
| ------------------------------------- | ------------------------------ | ---------- | -------------------------------- |
| `FLUXBASE_LOGGING_CONSOLE_ENABLED`    | Enable console logging         | `true`     | `true`, `false`                  |
| `FLUXBASE_LOGGING_CONSOLE_LEVEL`      | Console log level              | `info`     | `debug`, `info`, `warn`, `error` |
| `FLUXBASE_LOGGING_CONSOLE_FORMAT`     | Console log format             | `console`  | `console`, `json`                |
| `FLUXBASE_LOGGING_BACKEND`            | Log storage backend            | `postgres` | `postgres`, `s3`, `local`        |
| `FLUXBASE_LOGGING_SYSTEM_RETENTION_DAYS`    | System log retention     | `7`        | `7`                              |
| `FLUXBASE_LOGGING_HTTP_RETENTION_DAYS`      | HTTP log retention       | `30`       | `30`                             |
| `FLUXBASE_LOGGING_SECURITY_RETENTION_DAYS`  | Security log retention   | `90`       | `90`                             |
| `FLUXBASE_LOGGING_EXECUTION_RETENTION_DAYS` | Execution log retention  | `30`       | `30`                             |

### CORS

| Variable                       | Description                       | Default                                       | Example                                 |
| ------------------------------ | --------------------------------- | --------------------------------------------- | --------------------------------------- |
| `FLUXBASE_CORS_ALLOWED_ORIGINS`  | Allowed origins (comma-separated) | `http://localhost:5173,http://localhost:8080` | `http://localhost:3000,https://app.com` |
| `FLUXBASE_CORS_ALLOWED_METHODS`  | Allowed HTTP methods              | `GET,POST,PUT,PATCH,DELETE,OPTIONS`           | `GET,POST,PUT,DELETE`                   |
| `FLUXBASE_CORS_ALLOWED_HEADERS`  | Allowed headers                   | `Origin,Content-Type,Accept,Authorization,...`| `Authorization,Content-Type`            |
| `FLUXBASE_CORS_ALLOW_CREDENTIALS`| Allow credentials                 | `true`                                        | `true`, `false`                         |
| `FLUXBASE_CORS_MAX_AGE`          | Preflight cache time (seconds)    | `300`                                         | `86400`                                 |

### Security

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_SECURITY_SETUP_TOKEN` | Token for admin dashboard setup (required to enable dashboard) | `""` | `openssl rand -base64 32` |
| `FLUXBASE_ENCRYPTION_KEY` | Key for AES-256-GCM encryption of sensitive data (OAuth tokens, client keys). Must be exactly 32 bytes. | `""` | `openssl rand -base64 32 \| head -c 32` |
| `FLUXBASE_SECURITY_ENABLE_GLOBAL_RATE_LIMIT` | Enable global API rate limiting | `false` | `true`, `false` |
| `FLUXBASE_SECURITY_ADMIN_SETUP_RATE_LIMIT` | Max attempts for admin setup | `5` | `5` |
| `FLUXBASE_SECURITY_ADMIN_SETUP_RATE_WINDOW` | Time window for admin setup rate limit | `15m` | `15m` |
| `FLUXBASE_SECURITY_AUTH_LOGIN_RATE_LIMIT` | Max attempts for auth login | `10` | `10` |
| `FLUXBASE_SECURITY_AUTH_LOGIN_RATE_WINDOW` | Time window for auth login rate limit | `1m` | `1m` |
| `FLUXBASE_SECURITY_ADMIN_LOGIN_RATE_LIMIT` | Max attempts for admin login | `10` | `10` |
| `FLUXBASE_SECURITY_ADMIN_LOGIN_RATE_WINDOW` | Time window for admin login rate limit | `1m` | `1m` |

:::caution[Required for Admin Dashboard]
`FLUXBASE_SECURITY_SETUP_TOKEN` must be set to enable the admin dashboard. Generate a secure token with `openssl rand -base64 32`.
:::

### AI Chatbots

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_AI_ENABLED` | Enable AI chatbot functionality | `true` | `true`, `false` |
| `FLUXBASE_AI_CHATBOTS_DIR` | Directory for chatbot definitions | `./chatbots` | `./chatbots` |
| `FLUXBASE_AI_AUTO_LOAD_ON_BOOT` | Load chatbots from filesystem at boot | `true` | `true`, `false` |
| `FLUXBASE_AI_DEFAULT_MAX_TOKENS` | Default max tokens per request | `4096` | `4096` |
| `FLUXBASE_AI_QUERY_TIMEOUT` | SQL query execution timeout | `30s` | `30s` |
| `FLUXBASE_AI_MAX_ROWS_PER_QUERY` | Max rows returned per query | `1000` | `1000` |
| `FLUXBASE_AI_CONVERSATION_CACHE_TTL` | TTL for conversation cache | `30m` | `1h` |
| `FLUXBASE_AI_MAX_CONVERSATION_TURNS` | Max turns per conversation | `50` | `50` |

**AI Provider Configuration:**

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_AI_PROVIDER_ENABLED` | Enable config-based provider | `false` | `true`, `false` |
| `FLUXBASE_AI_PROVIDER_TYPE` | Provider type | `""` | `openai`, `azure`, `ollama` |
| `FLUXBASE_AI_PROVIDER_NAME` | Display name for provider | `""` | `Default Provider` |
| `FLUXBASE_AI_PROVIDER_MODEL` | Default model | `""` | `gpt-4-turbo` |

**OpenAI Settings:**

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_AI_OPENAI_API_KEY` | OpenAI API key | `""` | `sk-...` |
| `FLUXBASE_AI_OPENAI_ORGANIZATION_ID` | OpenAI organization ID | `""` | `org-...` |
| `FLUXBASE_AI_OPENAI_BASE_URL` | Custom base URL (for compatible APIs) | `""` | `https://api.openai.com/v1` |

**Azure OpenAI Settings:**

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_AI_AZURE_API_KEY` | Azure OpenAI API key | `""` | Your API key |
| `FLUXBASE_AI_AZURE_ENDPOINT` | Azure OpenAI endpoint | `""` | `https://your-resource.openai.azure.com` |
| `FLUXBASE_AI_AZURE_DEPLOYMENT_NAME` | Azure deployment name | `""` | `gpt-4-deployment` |
| `FLUXBASE_AI_AZURE_API_VERSION` | Azure API version | `""` | `2024-02-15-preview` |

**Ollama Settings:**

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_AI_OLLAMA_ENDPOINT` | Ollama endpoint | `""` | `http://localhost:11434` |
| `FLUXBASE_AI_OLLAMA_MODEL` | Ollama model name | `""` | `llama2`, `mistral` |

### Edge Functions

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_FUNCTIONS_ENABLED` | Enable edge functions | `true` | `true`, `false` |
| `FLUXBASE_FUNCTIONS_FUNCTIONS_DIR` | Directory for function files | `./functions` | `./functions` |
| `FLUXBASE_FUNCTIONS_AUTO_LOAD_ON_BOOT` | Load functions from filesystem at boot | `true` | `true`, `false` |
| `FLUXBASE_FUNCTIONS_DEFAULT_TIMEOUT` | Default function timeout (seconds) | `30` | `30` |
| `FLUXBASE_FUNCTIONS_MAX_TIMEOUT` | Maximum function timeout (seconds) | `300` | `300` |
| `FLUXBASE_FUNCTIONS_DEFAULT_MEMORY_LIMIT` | Default memory limit (MB) | `128` | `128` |
| `FLUXBASE_FUNCTIONS_MAX_MEMORY_LIMIT` | Maximum memory limit (MB) | `1024` | `1024` |
| `FLUXBASE_FUNCTIONS_MAX_OUTPUT_SIZE` | Maximum output size (bytes, 0=unlimited) | `10485760` | `10485760` |

### Deno Runtime

Global settings for the Deno runtime used by edge functions and background jobs.

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_DENO_NPM_REGISTRY` | Custom npm registry URL for `npm:` specifiers | `""` | `https://npm.your-company.com/` |
| `FLUXBASE_DENO_JSR_REGISTRY` | Custom JSR registry URL for `jsr:` specifiers | `""` | `https://jsr.your-company.com/` |

**Air-Gapped Environments:** Set these to your private registry URLs for environments without internet access. See [Edge Functions Air-Gapped Guide](/guides/edge-functions/#air-gapped--private-registry-environments).

### RPC (Remote Procedures)

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_RPC_ENABLED` | Enable RPC functionality | `true` | `true`, `false` |
| `FLUXBASE_RPC_PROCEDURES_DIR` | Directory for RPC procedure definitions | `./rpc` | `./rpc` |
| `FLUXBASE_RPC_AUTO_LOAD_ON_BOOT` | Load procedures from filesystem at boot | `true` | `true`, `false` |
| `FLUXBASE_RPC_DEFAULT_MAX_EXECUTION_TIME` | Default max execution time | `30s` | `30s` |
| `FLUXBASE_RPC_MAX_MAX_EXECUTION_TIME` | Maximum allowed execution time | `5m` | `5m` |
| `FLUXBASE_RPC_DEFAULT_MAX_ROWS` | Default max rows returned | `1000` | `1000` |

### Background Jobs

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_JOBS_ENABLED` | Enable background jobs | `true` | `true`, `false` |
| `FLUXBASE_JOBS_DIR` | Directory for job definitions | `./jobs` | `./jobs` |
| `FLUXBASE_JOBS_AUTO_LOAD_ON_BOOT` | Load jobs from filesystem at boot | `true` | `true`, `false` |
| `FLUXBASE_JOBS_WORKER_MODE` | Worker mode | `embedded` | `embedded`, `standalone`, `disabled` |
| `FLUXBASE_JOBS_EMBEDDED_WORKER_COUNT` | Number of embedded workers | `4` | `4` |
| `FLUXBASE_JOBS_MAX_CONCURRENT_PER_WORKER` | Max concurrent jobs per worker | `5` | `5` |
| `FLUXBASE_JOBS_MAX_CONCURRENT_PER_NAMESPACE` | Max concurrent jobs per namespace | `20` | `20` |
| `FLUXBASE_JOBS_DEFAULT_MAX_DURATION` | Default job timeout | `5m` | `5m` |
| `FLUXBASE_JOBS_MAX_MAX_DURATION` | Maximum allowed job timeout | `1h` | `1h` |
| `FLUXBASE_JOBS_DEFAULT_PROGRESS_TIMEOUT` | Progress reporting timeout | `5m` | `5m` |
| `FLUXBASE_JOBS_POLL_INTERVAL` | Worker poll interval | `1s` | `1s` |
| `FLUXBASE_JOBS_WORKER_HEARTBEAT_INTERVAL` | Worker heartbeat interval | `10s` | `10s` |
| `FLUXBASE_JOBS_WORKER_TIMEOUT` | Worker considered dead after | `30s` | `30s` |

### Horizontal Scaling

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_SCALING_BACKEND` | Distributed state backend | `local` | `local`, `postgres`, `redis` |
| `FLUXBASE_SCALING_REDIS_URL` | Redis/Dragonfly connection URL | `""` | `redis://dragonfly:6379` |
| `FLUXBASE_SCALING_ENABLE_SCHEDULER_LEADER_ELECTION` | Enable scheduler leader election | `false` | `true`, `false` |

**Backend Options:**

- `local` - In-memory storage (single instance only, default)
- `postgres` - Uses PostgreSQL for distributed state (no extra dependencies)
- `redis` - Uses Redis-compatible backend (Dragonfly recommended for performance)

**What's Distributed:**

| Feature | Description |
|---------|-------------|
| Rate limiting | Shared counters across all instances |
| Realtime broadcasts | Cross-instance pub/sub for application events |
| Scheduler coordination | Leader election prevents duplicate cron jobs |
| Nonce validation | PostgreSQL-backed for stateless auth flows |

**CLI Flags:**

| Flag | Description |
|------|-------------|
| `--worker-only` | Disable API server, only run background job workers |
| `--disable-scheduler` | Disable cron job scheduler on this instance |
| `--disable-realtime` | Disable realtime/WebSocket listener |
| `--enable-leader-election` | Enable PostgreSQL advisory lock leader election |

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

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_TRACING_ENABLED` | Enable OpenTelemetry tracing | `false` | `true`, `false` |
| `FLUXBASE_TRACING_ENDPOINT` | OTLP gRPC endpoint | `localhost:4317` | `jaeger:4317` |
| `FLUXBASE_TRACING_SERVICE_NAME` | Service name for traces | `fluxbase` | `fluxbase` |
| `FLUXBASE_TRACING_ENVIRONMENT` | Environment name | `development` | `production` |
| `FLUXBASE_TRACING_SAMPLE_RATE` | Sample rate (0.0-1.0) | `1.0` | `0.1` (10%) |
| `FLUXBASE_TRACING_INSECURE` | Use insecure connection | `true` | `false` |

### API Pagination

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_API_MAX_PAGE_SIZE` | Max rows per request (-1 = unlimited) | `1000` | `1000` |
| `FLUXBASE_API_MAX_TOTAL_RESULTS` | Max total retrievable rows (-1 = unlimited) | `10000` | `10000` |
| `FLUXBASE_API_DEFAULT_PAGE_SIZE` | Auto-applied limit when not specified (-1 = no default) | `1000` | `100` |

### GraphQL

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_GRAPHQL_ENABLED` | Enable GraphQL API endpoint | `true` | `true`, `false` |
| `FLUXBASE_GRAPHQL_MAX_DEPTH` | Maximum query depth | `10` | `15` |
| `FLUXBASE_GRAPHQL_MAX_COMPLEXITY` | Maximum query complexity score | `1000` | `2000` |
| `FLUXBASE_GRAPHQL_INTROSPECTION` | Enable GraphQL introspection | `true` | `false` |

### Prometheus Metrics

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_METRICS_ENABLED` | Enable Prometheus metrics endpoint | `true` | `true`, `false` |
| `FLUXBASE_METRICS_PORT` | Port for metrics server | `9090` | `9090` |
| `FLUXBASE_METRICS_PATH` | Path for metrics endpoint | `/metrics` | `/metrics` |

### MCP (Model Context Protocol)

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_MCP_ENABLED` | Enable MCP server | `true` | `true`, `false` |
| `FLUXBASE_MCP_BASE_PATH` | Base path for MCP endpoints | `/mcp` | `/mcp` |
| `FLUXBASE_MCP_SESSION_TIMEOUT` | Session timeout | `30m` | `1h` |
| `FLUXBASE_MCP_MAX_MESSAGE_SIZE` | Max message size (bytes) | `10485760` | `10485760` |
| `FLUXBASE_MCP_RATE_LIMIT_PER_MIN` | Rate limit per minute per client | `100` | `200` |
| `FLUXBASE_MCP_ALLOWED_TOOLS` | Allowed tools (empty = all) | `[]` | `["query", "storage"]` |
| `FLUXBASE_MCP_ALLOWED_RESOURCES` | Allowed resources (empty = all) | `[]` | `["schema", "functions"]` |
| `FLUXBASE_MCP_TOOLS_DIR` | Directory for custom MCP tools | `/app/mcp-tools` | `./mcp-tools` |
| `FLUXBASE_MCP_AUTO_LOAD_ON_BOOT` | Auto-load custom tools on startup | `true` | `true`, `false` |

### Database Branching

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FLUXBASE_BRANCHING_ENABLED` | Enable database branching | `false` | `true`, `false` |
| `FLUXBASE_BRANCHING_MAX_BRANCHES_PER_USER` | Max branches per user | `5` | `10` |
| `FLUXBASE_BRANCHING_MAX_TOTAL_BRANCHES` | Max total branches | `50` | `100` |
| `FLUXBASE_BRANCHING_DEFAULT_DATA_CLONE_MODE` | Default data clone mode | `schema_only` | `schema_only`, `full_clone` |
| `FLUXBASE_BRANCHING_AUTO_DELETE_AFTER` | Auto-delete branches after | `0` (never) | `24h`, `168h` |
| `FLUXBASE_BRANCHING_DATABASE_PREFIX` | Prefix for branch database names | `branch_` | `branch_` |
| `FLUXBASE_BRANCHING_SEEDS_PATH` | Path to seed data files | `./seeds` | `./seeds` |

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
server:
  port: 443 # HTTPS
  host: 0.0.0.0
  read_timeout: 30s
  write_timeout: 30s

database:
  url: postgres://fluxbase:password@postgres:5432/fluxbase?sslmode=require
  max_connections: 200
  idle_connections: 20
  connection_lifetime: 30m

jwt:
  secret: ${JWT_SECRET} # From environment
  access_token_expiry: 15m
  refresh_token_expiry: 7d

storage:
  provider: s3
  max_upload_size: 52428800 # 50MB
  s3_endpoint: s3.amazonaws.com
  s3_access_key: ${S3_ACCESS_KEY}
  s3_secret_key: ${S3_SECRET_KEY}
  s3_region: us-east-1
  s3_bucket: my-production-bucket

realtime:
  enabled: true
  heartbeat_interval: 30s
  max_connections: 5000

admin:
  enabled: false # Disable in production or protect behind VPN

logging:
  level: info
  format: json
  output: /var/log/fluxbase/app.log

cors:
  enabled: true
  allowed_origins:
    - https://app.example.com
    - https://www.example.com
  allow_credentials: true

rate_limit:
  enabled: true
  requests_per_minute: 1000
  burst: 2000

tls:
  enabled: true
  auto_cert: true
  auto_cert_domain: api.example.com
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
server:
  port: 8080
  host: 127.0.0.1

database:
  url: postgres://fluxbase:fluxbase@localhost:5432/fluxbase?sslmode=disable
  max_connections: 20
  idle_connections: 5

jwt:
  secret: dev-secret-change-in-production
  access_token_expiry: 24h # Longer for development
  refresh_token_expiry: 30d

storage:
  provider: local
  local_path: ./storage
  max_upload_size: 10485760 # 10MB

realtime:
  enabled: true
  max_connections: 100

admin:
  enabled: true
  path: /admin

logging:
  level: debug
  format: text
  output: stdout

cors:
  enabled: true
  allowed_origins:
    - http://localhost:3000
    - http://localhost:5173
    - http://127.0.0.1:3000
  allow_credentials: true

rate_limit:
  enabled: false # Disable in development

tls:
  enabled: false # Use HTTP in development
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
