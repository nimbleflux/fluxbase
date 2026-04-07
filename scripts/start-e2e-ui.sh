#!/bin/bash
set -e

# Start E2E test environment for Playwright admin UI tests.
#
# Starts two processes:
#   1. Go backend on :8082 against fluxbase_playwright database
#   2. Vite dev server on :5050 proxying API calls to :8082
#
# Usage:
#   ./scripts/start-e2e-ui.sh            # start servers (foreground, Ctrl+C to stop)
#   ./scripts/start-e2e-ui.sh --ensure   # start in background if not already running
#   ./scripts/start-e2e-ui.sh --restart  # kill existing servers, start fresh in background
#   ./scripts/start-e2e-ui.sh --clean    # reset playwright DB + restart servers
#   ./scripts/start-e2e-ui.sh --stop     # stop running servers

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

BACKEND_PORT=8082
VITE_PORT=5050
PIDFILE="/tmp/fluxbase-e2e-ui.pid"
LOGFILE="/tmp/fluxbase-e2e-ui.log"

# --- Database Reset ---

reset_playwright_db() {
    echo -e "${YELLOW}Resetting fluxbase_playwright database...${NC}"
    local ADMIN_USER="${FLUXBASE_DATABASE_ADMIN_USER:-${FLUXBASE_DATABASE_USER:-postgres}}"
    local ADMIN_PASSWORD="${FLUXBASE_DATABASE_ADMIN_PASSWORD:-${FLUXBASE_DATABASE_PASSWORD:-postgres}}"
    local DB_HOST="${FLUXBASE_DATABASE_HOST:-localhost}"
    local DB_PORT="${FLUXBASE_DATABASE_PORT:-5432}"

    export PGPASSWORD="$ADMIN_PASSWORD"

    # Drop all non-system schemas (preserves extensions and roles)
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$ADMIN_USER" -d fluxbase_playwright -c \
        "DO \$\$ DECLARE r RECORD; BEGIN FOR r IN SELECT nspname FROM pg_namespace WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast') LOOP EXECUTE 'DROP SCHEMA IF EXISTS ' || quote_ident(r.nspname) || ' CASCADE'; END LOOP; END \$\$;" || {
        unset PGPASSWORD
        echo -e "${RED}Failed to reset database${NC}"
        exit 1
    }

    # Recreate public schema with basic grants
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$ADMIN_USER" -d fluxbase_playwright -c \
        "CREATE SCHEMA IF NOT EXISTS public; GRANT ALL ON SCHEMA public TO public;"

    # Ensure fluxbase_app has BYPASSRLS (server bootstrap expects this)
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$ADMIN_USER" -d fluxbase_playwright -c \
        "ALTER USER fluxbase_app WITH BYPASSRLS;" 2>/dev/null || true

    unset PGPASSWORD
    echo -e "${GREEN}Database reset complete. Server will re-apply schemas on startup.${NC}"
}

# --- Helpers ---

is_backend_running() {
    curl -sf "http://localhost:${BACKEND_PORT}/health" > /dev/null 2>&1
}

is_vite_running() {
    curl -sf "http://localhost:${VITE_PORT}/admin/" > /dev/null 2>&1
}

is_running() {
    is_backend_running && is_vite_running
}

stop_servers() {
    if [ -f "$PIDFILE" ]; then
        while IFS= read -r pid; do
            kill "$pid" 2>/dev/null || true
        done < "$PIDFILE"
        rm -f "$PIDFILE"
    fi
    # Also kill anything on our ports as a fallback
    lsof -ti:$BACKEND_PORT | xargs -r kill 2>/dev/null || true
    lsof -ti:$VITE_PORT | xargs -r kill 2>/dev/null || true
}

wait_for_backend() {
    for i in $(seq 1 30); do
        if is_backend_running; then
            return 0
        fi
        sleep 2
    done
    echo -e "${RED}Backend failed to start within 60 seconds${NC}"
    return 1
}

wait_for_vite() {
    for i in $(seq 1 20); do
        if is_vite_running; then
            return 0
        fi
        sleep 2
    done
    echo -e "${RED}Vite failed to start within 40 seconds${NC}"
    return 1
}

# --- Start logic ---

start_background() {
    # Ensure we're in the project root
    cd "$(dirname "$0")/.."

    # Add Deno to PATH (required for edge functions)
    export PATH="/home/vscode/.deno/bin:$PATH"

    # Database configuration
    # Force playwright database — do NOT inherit from .env or parent shell.
    # The .env file sets FLUXBASE_DATABASE_DATABASE=fluxbase_dev which would
    # otherwise be picked up by both the shell and godotenv.Load().
    export FLUXBASE_DATABASE_DATABASE="fluxbase_playwright"
    export FLUXBASE_DATABASE_HOST="${FLUXBASE_DATABASE_HOST:-localhost}"
    export FLUXBASE_DATABASE_PORT="${FLUXBASE_DATABASE_PORT:-5432}"
    export FLUXBASE_DATABASE_USER="${FLUXBASE_DATABASE_USER:-fluxbase_app}"
    export FLUXBASE_DATABASE_PASSWORD="${FLUXBASE_DATABASE_PASSWORD:-fluxbase_app_password}"
    export FLUXBASE_SECURITY_SETUP_TOKEN="${FLUXBASE_SECURITY_SETUP_TOKEN:-test-setup-token-for-dev-environment-32chars}"
    export FLUXBASE_SERVER_ADDRESS=":${BACKEND_PORT}"

    # --- Start Go backend ---
    if ! is_backend_running; then
        echo -e "${YELLOW}Starting Go backend on :${BACKEND_PORT}...${NC}"
        echo -e "  Database: ${FLUXBASE_DATABASE_DATABASE}"
        GOGC=50 go run -tags "ocr" cmd/fluxbase/main.go >> "$LOGFILE" 2>&1 &
        echo "$!" >> "$PIDFILE"
        if ! wait_for_backend; then
            echo -e "${RED}Check logs: ${LOGFILE}${NC}"
            exit 1
        fi
        echo -e "${GREEN}Backend ready on :${BACKEND_PORT}${NC}"
    else
        echo -e "${GREEN}Backend already running on :${BACKEND_PORT}${NC}"
    fi

    # --- Start Vite dev server ---
    if ! is_vite_running; then
        echo -e "${YELLOW}Starting Vite dev server on :${VITE_PORT}...${NC}"
        (
            cd admin
            unset NODE_OPTIONS
            export VITE_PROXY_TARGET="http://localhost:${BACKEND_PORT}"
            bun run dev --host 0.0.0.0 --port ${VITE_PORT}
        ) >> "$LOGFILE" 2>&1 &
        echo "$!" >> "$PIDFILE"
        if ! wait_for_vite; then
            echo -e "${RED}Check logs: ${LOGFILE}${NC}"
            exit 1
        fi
        echo -e "${GREEN}Vite ready on :${VITE_PORT}${NC}"
    else
        echo -e "${GREEN}Vite already running on :${VITE_PORT}${NC}"
    fi
}

start_foreground() {
    stop_servers

    # Ensure we're in the project root
    cd "$(dirname "$0")/.."

    trap stop_servers SIGINT SIGTERM

    export PATH="/home/vscode/.deno/bin:$PATH"
    # Force playwright database — do NOT inherit from .env or parent shell.
    export FLUXBASE_DATABASE_DATABASE="fluxbase_playwright"
    export FLUXBASE_DATABASE_HOST="${FLUXBASE_DATABASE_HOST:-localhost}"
    export FLUXBASE_DATABASE_PORT="${FLUXBASE_DATABASE_PORT:-5432}"
    export FLUXBASE_DATABASE_USER="${FLUXBASE_DATABASE_USER:-fluxbase_app}"
    export FLUXBASE_DATABASE_PASSWORD="${FLUXBASE_DATABASE_PASSWORD:-fluxbase_app_password}"
    export FLUXBASE_SECURITY_SETUP_TOKEN="${FLUXBASE_SECURITY_SETUP_TOKEN:-test-setup-token-for-dev-environment-32chars}"
    export FLUXBASE_SERVER_ADDRESS=":${BACKEND_PORT}"

    echo -e "${YELLOW}Starting Go backend on :${BACKEND_PORT}...${NC}"
    echo -e "  Database: ${FLUXBASE_DATABASE_DATABASE}"
    GOGC=50 go run -tags "ocr" cmd/fluxbase/main.go &
    GO_PID=$!
    echo "$GO_PID" > "$PIDFILE"

    if ! wait_for_backend; then
        exit 1
    fi
    echo -e "${GREEN}Backend ready on :${BACKEND_PORT}${NC}"

    echo -e "${YELLOW}Starting Vite dev server on :${VITE_PORT}...${NC}"
    cd admin
    unset NODE_OPTIONS
    export VITE_PROXY_TARGET="http://localhost:${BACKEND_PORT}"
    bun run dev --host 0.0.0.0 --port ${VITE_PORT} &
    VITE_PID=$!
    cd ..
    echo "$VITE_PID" >> "$PIDFILE"

    echo ""
    echo -e "${GREEN}E2E test environment ready!${NC}"
    echo -e "  Backend:  http://localhost:${BACKEND_PORT}"
    echo -e "  Frontend: http://localhost:${VITE_PORT}/admin/"
    echo -e "  Login:    http://localhost:${VITE_PORT}/admin/login"
    echo -e "  Logs:     ${LOGFILE}"
    echo ""
    echo -e "${YELLOW}Run tests in another terminal: make test-e2e-ui-run${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
    echo ""

    wait
}

# --- Main ---

case "${1:-}" in
    --ensure)
        if is_running; then
            echo -e "${GREEN}E2E servers already running${NC}"
            exit 0
        fi
        rm -f "$PIDFILE"
        start_background
        echo ""
        echo -e "${GREEN}E2E servers ready. Logs: ${LOGFILE}${NC}"
        echo -e "${GREEN}Frontend: http://localhost:${VITE_PORT}/admin/${NC}"
        ;;
    --restart)
        echo -e "${YELLOW}Restarting E2E servers...${NC}"
        stop_servers
        sleep 2
        rm -f "$PIDFILE" "$LOGFILE"
        start_background
        echo ""
        echo -e "${GREEN}E2E servers restarted. Logs: ${LOGFILE}${NC}"
        ;;
    --clean)
        echo -e "${YELLOW}Clean start: resetting database and restarting servers...${NC}"
        stop_servers
        sleep 2
        rm -f "$PIDFILE" "$LOGFILE"
        reset_playwright_db
        start_background
        echo ""
        echo -e "${GREEN}E2E servers ready with fresh database. Logs: ${LOGFILE}${NC}"
        echo -e "${GREEN}Frontend: http://localhost:${VITE_PORT}/admin/${NC}"
        ;;
    --stop)
        echo -e "${YELLOW}Stopping E2E servers...${NC}"
        stop_servers
        echo -e "${GREEN}Stopped${NC}"
        ;;
    "")
        # No args: foreground mode (blocking)
        start_foreground
        ;;
    *)
        echo "Usage: $0 [--ensure|--restart|--clean|--stop]"
        exit 1
        ;;
esac
