# Fluxbase DevContainer Quick Start

## 🚀 Get Started in 3 Steps

### 1. Open in Container

```
VS Code → Reopen in Container
```

Wait 5-10 minutes for first build (subsequent starts: ~30 seconds)

### 2. Verify Setup

```bash
bash .devcontainer/test-setup.sh
```

### 3. Start Developing

```bash
make dev  # Start with hot-reload
```

## 📝 Common Commands

```bash
# Development
make dev              # Start with hot-reload
make build            # Build binary
make test             # Run all tests
make test-unit        # Unit tests only
make test-integration # Integration tests

# Database
make migrate-up       # Apply migrations
make migrate-down     # Rollback migrations
make db-setup         # Setup with example data

# Documentation
make docs-dev         # Start docs server
make docs-build       # Build static docs

# Code Quality
make fmt              # Format code
make lint             # Run linters
make vet              # Go vet

# Docker
make docker-build     # Build Docker image
make docker-run       # Run container

# All Commands
make help             # Show all commands
```

## 🌐 Service URLs

| Service       | URL                   | Credentials               |
| ------------- | --------------------- | ------------------------- |
| Fluxbase API  | http://localhost:8080 | -                         |
| MailHog       | http://localhost:8025 | -                         |
| Documentation | http://localhost:3000 | -                         |

## 🗄️ Database

```bash
# Quick connect
psql -h postgres -U postgres -d fluxbase_dev

# Or use SQLTools in VS Code sidebar
```

**Databases**:

- `fluxbase_dev` - Development
- `fluxbase_test` - Testing

**Credentials**:

- Host: `postgres`
- User: `postgres`
- Password: `postgres`

## 🛠️ Installed Tools

### Go

- gopls, dlv, golangci-lint, air, migrate, swag, mockery, staticcheck

### Node.js

- typescript, eslint, prettier, tsx, nodemon

### Testing

- gotestsum, ginkgo

### Database

- psql, redis-cli, pgAdmin

### Utilities

- git, gh, docker, make, jq, httpie, tree

## 🎨 VS Code Extensions

### Essential

- **Claude Code** - AI assistant
- **Go** - Full Go support
- **SQLTools** - Database management

### Useful Shortcuts

- `Ctrl+` ` - Toggle terminal
- `F5` - Start debugging
- `Ctrl+Shift+P` - Command palette
- `Ctrl+P` - Quick file open

## 📋 Project Structure

```
fluxbase/
├── cmd/fluxbase/       # Main entry point
├── internal/           # Private app code
│   ├── api/           # REST API
│   ├── auth/          # Authentication (TO BUILD)
│   ├── config/        # Configuration
│   ├── database/      # DB layer
│   ├── realtime/      # WebSocket (TO BUILD)
│   └── storage/       # File storage (TO BUILD)
├── pkg/               # Public libraries
├── test/              # Integration tests
├── docs/              # Documentation
├── .devcontainer/     # This setup
├── TODO.md            # Task list
└── Makefile           # Commands
```

## 🎯 Current Sprint: Authentication

Next tasks from `TODO.md`:

1. Implement JWT token utilities
2. Create user registration endpoint
3. Create login endpoint
4. Add auth middleware
5. Session management

See `TODO.md` and `IMPLEMENTATION_PLAN.md` for details.

## 💡 Pro Tips

1. **Use Claude Code**: AI-powered development - just ask!
2. **SQLTools**: Database icon in sidebar for queries
3. **Thunder Client**: Test APIs right in VS Code
4. **GitLens**: See git blame inline
5. **TODO Tree**: Track tasks from code comments
6. **Hot Reload**: Changes apply automatically with `make dev`

## 🐛 Quick Troubleshooting

### Container Issues

```bash
# Rebuild
F1 → "Dev Containers: Rebuild Container"

# Check logs
docker compose logs -f
```

### Database Issues

```bash
# Test connection
pg_isready -h postgres -U postgres

# View logs
docker logs fluxbase-postgres-dev
```

### Go Issues

```bash
# Reinstall dependencies
go mod download
go mod tidy

# Rebuild
go build cmd/fluxbase/main.go
```

## 📚 Documentation

- **This Guide**: Quick start reference
- **Full Docs**: `.devcontainer/README.md`
- **Changes**: `.devcontainer/CHANGELOG.md`
- **Fix Summary**: `DEVCONTAINER_FIXES.md`
- **Dev Guide**: `.claude/instructions.md`
- **Tasks**: `TODO.md`
- **Plan**: `IMPLEMENTATION_PLAN.md`

## ✅ Health Check

Run this to verify everything:

```bash
bash .devcontainer/test-setup.sh
```

Should show all green checkmarks ✓

## 🎉 You're Ready!

Start building:

```bash
make dev
```

Open http://localhost:8080/health - should return `{"status":"ok"}`

You're all set! Check out the documentation in `docs/` to learn more.

---

**Need Help?** Use Claude Code or check `.devcontainer/README.md`
