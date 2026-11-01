#!/usr/bin/env bash
# Development mode: run backend (:8787) + traffic proxy (:8788) with Next.js dev (:5173).
# Frontend proxies /api to backend; press Ctrl-C to stop all services.
#
# For single-binary mode with embedded UI, see the README; do not use this script.
set -euo pipefail
cd "$(dirname "$0")"

# Terminate all child processes in this process group on exit.
cleanup() { kill 0 2>/dev/null || true; }
trap cleanup EXIT INT TERM

# Backend (standard go run without embedded UI); worker concurrency configured in System Settings.
go run ./cmd/artex -addr :8790 -proxy 127.0.0.1:8788 &

# Frontend live reload (Next.js dev server, proxies /api to :8790).
( cd web && npm run dev ) &

echo "[dev] Backend :8790 / Proxy :8788 / Frontend http://localhost:5173  (Press Ctrl-C to exit)"
wait
