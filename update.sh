#!/usr/bin/env bash
# ARTEX update script: 1) Docker update (pull new image)  2) Local source build
# Corresponding to install.sh: install handles first-time setup, update upgrades to new versions.
# Database migrations are run automatically on startup via schema.sql.
set -euo pipefail
cd "$(cd "$(dirname "$0")" && pwd)"

info(){ printf '\033[36m[*]\033[0m %s\n' "$*"; }
ok(){   printf '\033[32m[+]\033[0m %s\n' "$*"; }
warn(){ printf '\033[33m[!]\033[0m %s\n' "$*"; }
die(){  printf '\033[31m[x]\033[0m %s\n' "$*" >&2; exit 1; }
ask(){  local p="$1" d="${2:-}" a; read -rp "$p${d:+ [$d]}: " a; echo "${a:-$d}"; }

# ── Optional: Sync repo to latest code ───────────────────────────────────────
sync_repo(){
  [ -d .git ] && command -v git >/dev/null 2>&1 || { warn "Not a git working tree, skipping git pull"; return; }
  [ "$(ask 'Pull latest git commits (git pull --ff-only)? (y/n)' y)" = y ] || return
  if ! git pull --ff-only; then
    warn "git pull could not fast-forward (local modifications or diverted branch); keeping current code"
  fi
}

# ── 1) Docker Update ────────────────────────────────────────────────────────
update_docker(){
  command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1 \
    || die "docker / docker compose not found; run ./install.sh first"
  [ -f .env ] || die ".env not found; run ./install.sh for initial setup"

  local tag; tag="$(ask 'Target image tag (press Enter for .env / latest)' '')"
  if [ -n "$tag" ]; then
    if grep -q '^ARTEX_TAG=' .env; then
      sed -i.bak "s|^ARTEX_TAG=.*|ARTEX_TAG=${tag}|" .env && rm -f .env.bak
    else
      printf '\nARTEX_TAG=%s\n' "$tag" >> .env
    fi
    ok "Set ARTEX_TAG to ${tag}"
  fi

  info "Pulling latest image (artex service only)..."
  docker compose pull artex
  info "Recreating container (schema migrations apply on startup)..."
  docker compose up -d artex
  ok "Update complete → http://localhost:8787"
  info "View logs: docker compose logs -f artex"
  info "Prune old images (optional): docker image prune -f"
}

# ── 2) Local Source Update ──────────────────────────────────────────────────
update_local(){
  command -v go >/dev/null 2>&1 || die "Go not found (>=1.26 required): https://go.dev/dl/"
  [ -f config.json ] || warn "config.json not found; run ./install.sh for initial setup"
  ok "Go: $(go version)"

  if command -v npm >/dev/null 2>&1; then
    info "Rebuilding frontend static assets..."
    ( cd web && npm ci && npm run build:static )
    rm -rf server/webui/dist && cp -r web/out server/webui/dist
    info "Rebuilding single binary with embedded UI..."
    CGO_ENABLED=0 go build -tags embedui -trimpath -o artex ./cmd/artex
  else
    warn "npm not found: building backend binary WITHOUT embedded UI (run npm run dev separately)"
    CGO_ENABLED=0 go build -o artex ./cmd/artex
  fi
  ok "Build complete → ./artex"
  warn "Restart running artex process to apply changes (schema migrations apply on startup)"
}

echo "=============================="
echo "  ARTEX Update"
echo "  1) Docker update (pull new image and recreate)"
echo "  2) Local update (recompile with go)"
echo "=============================="
case "$(ask 'Select' 1)" in
  1) sync_repo; update_docker ;;
  2) sync_repo; update_local ;;
  *) die "Invalid choice" ;;
esac
