# Fluxbase

[![CI](https://github.com/fluxbase-eu/fluxbase/actions/workflows/ci.yml/badge.svg)](https://github.com/fluxbase-eu/fluxbase/actions/workflows/ci.yml)

> **Beta Software**: Fluxbase is currently in beta. While we're working hard to stabilize the API and features, you may encounter breaking changes between versions. We welcome feedback and contributions!

Run `make test-full` to validate all critical flows.

A lightweight, single-binary Backend-as-a-Service (BaaS) alternative to Supabase. Fluxbase provides essential backend services including auto-generated REST APIs, authentication, realtime subscriptions, file storage, and edge functions - all in a single Go binary with PostgreSQL as the only dependency.

## Features

### Core Services

- **PostgREST-compatible REST API**: Auto-generates CRUD endpoints from your PostgreSQL schema
- **GraphQL API**: Full GraphQL support with configurable depth/complexity limits
- **Authentication**: Email/password, magic links, OAuth2 (Google, GitHub, Microsoft, etc.), OIDC, SAML SSO, MFA/TOTP
- **Realtime Subscriptions**: WebSocket-based live data updates using PostgreSQL LISTEN/NOTIFY
- **Storage**: File upload/download with access policies (local filesystem or S3), image transformations
- **Edge Functions**: JavaScript/TypeScript function execution with Deno runtime
- **Background Jobs**: Long-running tasks with progress tracking, retry logic, cron scheduling
- **RPC/Procedures**: SQL-based serverless procedures with scheduling and RBAC
- **Webhooks**: Event-driven webhook delivery for database changes with retries and HMAC signing
- **Vector Search**: pgvector-powered semantic search with automatic embeddings
- **MCP Server**: Model Context Protocol for AI assistant integration

### Key Highlights

- Single binary or container deployment
- PostgreSQL as the only external dependency
- Automatic REST endpoint generation
- Row Level Security (RLS) support
- TypeScript SDK
- Database branching for dev/test environments
- Built-in observability (Prometheus metrics, OpenTelemetry tracing)
- Horizontal scaling with leader election

## Quick Start

For more information about Fluxbase, look into [the docs](https://fluxbase.eu/getting-started/quick-start/).

## Support

For issues, questions, and discussions:

- GitHub Issues: [github.com/fluxbase-eu/fluxbase/issues](https://github.com/fluxbase-eu/fluxbase/issues)
- Documentation: [fluxbase.eu](https://fluxbase.eu)
- Discord: [discord.gg/BXPRHkQzkA](https://discord.gg/BXPRHkQzkA)
