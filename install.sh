#!/usr/bin/env bash
# ARTEX installation script: 1) Full Docker  2) Local compilation and run
set -euo pipefail
cd "$(cd "$(dirname "$0")" && pwd)"

info(){ printf '\033[36m[*]\033[0m %s\n' "$*"; }
ok(){   printf '\033[32m[+]\033[0m %s\n' "$*"; }
warn(){ printf '\033[33m[!]\033[0m %s\n' "$*"; }
die(){  printf '\033[31m[x]\033[0m %s\n' "$*" >&2; exit 1; }
ask(){  local p="$1" d="${2:-}" a; read -rp "$p${d:+ [$d]}: " a; echo "${a:-$d}"; }
rand(){ head -c 18 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 24; }

# ── Docker environment check / auto-installation ─────────────────────────────
ensure_docker(){
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    ok "Found docker and docker compose"; return
  fi
  warn "docker or docker compose not found"
  case "$(uname -s)" in
    Linux)
      if [ "$(ask 'Install Docker automatically? (y/n)' y)" = y ]; then
        curl -fsSL https://get.docker.com | sh
        sudo usermod -aG docker "$USER" || true
        ok "Docker installation completed (log out and back in to use without sudo)"
      else
        die "Please install docker and retry"
      fi ;;
    Darwin) die "On macOS please install Docker Desktop: https://www.docker.com/products/docker-desktop/" ;;
    *)      die "Please install docker and retry" ;;
  esac
}

# ── 1) Full Docker Installation ─────────────────────────────────────────────
install_docker(){
  ensure_docker
  if [ ! -f .env ]; then
    cp .env.example .env 2>/dev/null || true
    local pw key
    pw="$(ask 'Postgres password (press Enter for random generation)' "$(rand)")"
    key="$(ask 'ANTHROPIC_API_KEY (optional, can configure in UI)' '')"
    sed -i.bak "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=${pw}|" .env
    sed -i.bak "s|^ANTHROPIC_API_KEY=.*|ANTHROPIC_API_KEY=${key}|" .env
    rm -f .env.bak
    ok "Generated .env (POSTGRES_PASSWORD configured)"
  else
    info "Using existing .env"
  fi
  info "Pulling images and starting services..."
  docker compose pull || true
  docker compose up -d
  ok "Services started → http://localhost:8787"
  info "View logs: docker compose logs -f artex"
}

# ── 2) Local Build and Run ──────────────────────────────────────────────────
install_local(){
  echo "Database configuration:"
  echo "  1) Connect to existing PostgreSQL"
  echo "  2) Run PostgreSQL in Docker (requires docker)"
  case "$(ask 'Select' 1)" in
    2)
      ensure_docker
      local pw; pw="$(ask 'Postgres password (press Enter for random)' "$(rand)")"
      docker run -d --name artex-pg -p 5432:5432 \
        -e POSTGRES_USER=artex -e POSTGRES_PASSWORD="$pw" -e POSTGRES_DB=artex \
        -v artex-pg:/var/lib/postgresql/data postgres:16-alpine
      DB_HOST=127.0.0.1 DB_PORT=5432 DB_USER=artex DB_PASS="$pw" DB_NAME=artex DB_SSL=disable ;;
    *)
      DB_HOST="$(ask 'Database host' 127.0.0.1)"
      DB_PORT="$(ask 'Port' 5432)"
      DB_USER="$(ask 'Username' artex)"
      DB_PASS="$(ask 'Password' '')"
      DB_NAME="$(ask 'Database name' artex)"
      DB_SSL="$(ask 'sslmode (disable/require)' disable)" ;;
  esac

  # Generate config.json
  cat > config.json <<JSON
{
  "database": {
    "host": "${DB_HOST}",
    "port": ${DB_PORT},
    "user": "${DB_USER}",
    "password": "${DB_PASS}",
    "dbname": "${DB_NAME}",
    "sslmode": "${DB_SSL}"
  }
}
JSON
  ok "Generated config.json"

  # Go environment check
  command -v go >/dev/null 2>&1 || die "Go not found; please install Go (>=1.26): https://go.dev/dl/"
  ok "Go: $(go version)"

  # Embedded frontend requires Node.js static export
  if command -v npm >/dev/null 2>&1; then
    info "Building frontend static assets..."
    ( cd web && npm ci && npm run build:static )
    rm -rf server/webui/dist && cp -r web/out server/webui/dist
    info "Compiling single binary with embedded UI..."
    CGO_ENABLED=0 go build -tags embedui -trimpath -o artex ./cmd/artex
  else
    warn "npm not found: compiling backend WITHOUT embedded UI (run npm run dev separately)"
    CGO_ENABLED=0 go build -o artex ./cmd/artex
  fi
  ok "Build complete → ./artex"

  info "Starting artex... (Press Ctrl-C to exit)"
  ./artex
}

echo "=============================="
echo "  ARTEX Installation"
echo "  1) Full Docker installation"
echo "  2) Local run (compile with go)"
echo "=============================="
case "$(ask 'Select' 1)" in
  1) install_docker ;;
  2) install_local ;;
  *) die "Invalid choice" ;;
esac
