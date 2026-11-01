# AGENTS.md — ARTEX Repository Guide

Guidance for automated agents working on ARTEX (AI autonomous penetration testing system).

---

## 1. Stack & Architecture

- **Backend**: Go 1.26+ (`go.mod`). Entrypoint at `cmd/artex/main.go`.
  - `agent/`: Multi-agent orchestration, planner/worker loops (built on `github.com/Autumn-27/norma`).
  - `server/`: HTTP JSON API and WebSocket server on `:8787`.
  - `traffic/`: MITM recording proxy on `:8788`, metadata indexed into pure-Go SQLite (`data/traffic/_index/index.sqlite`).
  - `db/`: PostgreSQL storage for assets, findings, tasks, intercept rules, and settings.
  - `guard/`, `intercept/`: Tool invocation gating and human-in-the-loop intercept evaluation.
  - `skills/`: Tool packs with `SKILL.md` (e.g. `api-recon`, `playwright-cli`, `scopesentry`).
  - `data/`: Local persistent store (`jwt.key`, sqlite, uploads).
- **Frontend**: Next.js 16 (App Router), React 19, Tailwind CSS, shadcn/ui, Biome in `web/`.

---

## 2. Dev & Build Commands

- **Local Dev (Concurrent Backend + Frontend)**:
  ```bash
  ./dev.sh
  ```
  Runs backend on `:8787`, proxy on `:8788`, and Next.js dev server on `:5173` (proxies `/api` to `:8787`).

- **Frontend Build & Embed Quirk**:
  - `go build ./cmd/artex` does **NOT** embed the UI by default (serves a 404 stub).
  - To build the complete single binary with embedded frontend:
    ```bash
    cd web && npm ci && npm run build:static && cd ..
    cp -r web/out server/webui/dist
    CGO_ENABLED=0 go build -tags embedui -o artex ./cmd/artex
    ```
  - Or run the cross-platform builder:
    ```bash
    ./build.sh                     # build for host platform
    ./build.sh --release           # package all release targets
    ```

---

## 3. Database & Migrations

- **Database Engine**: PostgreSQL 16+ is required for the backend state.
- **Connection Config**: `ARTEX_PG_DSN` environment variable takes precedence over `config.json` (`database` field).
- **No Manual Migrations**:
  - `db/schema.sql` is embedded (`//go:embed schema.sql`) and automatically executed on startup via `db.Open()`.
  - Migrations must be written idempotently (`IF NOT EXISTS`, `OR REPLACE`) directly in `db/schema.sql`.
  - Startup schema migration is synchronized across instances via PG advisory lock `7337741001`.

---

## 4. Testing Rules & Prerequisites

- **Running Unit Tests (No Postgres Required)**:
  Packages that do not touch the database or that mock storage pass without a database:
  ```bash
  go test ./guard/... ./notify/... ./llmpool/... ./config/... ./selfupdate/... ./traffic/...
  ```
- **Integration Tests (Postgres Required)**:
  `server/` and `agent/` tests require a running PostgreSQL instance (`ARTEX_PG_DSN` or root `config.json`). Without it, integration tests fail with `未找到数据库配置`.
- **Test Concurrency Advisory Lock**:
  `db`, `agent`, and `server` package test runners share PG advisory lock `7337741002` in `TestMain` to avoid table cleanup races during `go test ./...`.
- **Run a Single Test**:
  ```bash
  go test -v -run <TestName> ./<package_path>
  ```

---

## 5. Frontend (`web/`) Workflow

- **Prerequisite**: Dependencies must be installed in `web/`:
  ```bash
  cd web && npm ci
  ```
- **Linter & Formatter**: Uses Biome (not ESLint/Prettier):
  ```bash
  cd web && npm run check        # Check lint and formatting
  cd web && npm run check:fix    # Auto-fix lint and formatting
  cd web && npm run format       # Format files
  ```

---

## 6. Operational Gotchas

- **Start Script Exit Code Handover**:
  `start.sh` watches for exit code `75` (`selfupdate.ExitRestart`). When self-update or rollback completes, the binary exits with `75` so `start.sh` immediately relaunches the new binary.
- **Admin Password Reset**:
  Use `./reset-password.sh -p '<new_password>'` (writes bcrypt hash directly to PostgreSQL `settings` table).
