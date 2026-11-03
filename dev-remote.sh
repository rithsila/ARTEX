#!/usr/bin/env bash
# Frontend dev mode against remote backend
set -euo pipefail
cd "$(dirname "$0")/web"

echo "[dev] Using remote backend at 10.10.23.5:8787"
npm run dev
