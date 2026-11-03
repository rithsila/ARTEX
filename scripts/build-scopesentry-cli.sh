#!/usr/bin/env bash
# Build only from an operator-selected local checkout. Never read runtime config.
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source_dir="$root/../ScopeSentry"
target_os=native
target_arch=native
output_dir="$root/tools/bin"
usage() {
  echo 'Usage: bash scripts/build-scopesentry-cli.sh [--source PATH] [--os native|linux|darwin] [--arch native|amd64|arm64] [--output-dir PATH]' >&2
}
while [[ $# -gt 0 ]]; do
  case "$1" in
    --source|--os|--arch|--output-dir)
      [[ $# -ge 2 && -n "$2" && "$2" != --* ]] || { usage; exit 2; }
      case "$1" in
        --source) source_dir="$2" ;;
        --os) target_os="$2" ;;
        --arch) target_arch="$2" ;;
        --output-dir) output_dir="$2" ;;
      esac
      shift 2 ;;
    --help|-h) usage; exit 0 ;;
    *) usage; exit 2 ;;
  esac
done
# Validate before invoking Go or creating output directories.
case "$target_os" in native|linux|darwin) ;; *) echo "Unsupported OS: $target_os" >&2; exit 2 ;; esac
case "$target_arch" in native|amd64|arm64) ;; *) echo "Unsupported architecture: $target_arch" >&2; exit 2 ;; esac
[[ -f "$source_dir/go.mod" && -d "$source_dir/cmd/scopesentry" ]] || { echo 'Source must be a local ScopeSentry checkout with cmd/scopesentry.' >&2; exit 2; }
grep -Eq '^module github.com/Autumn-27/ScopeSentry[[:space:]]*$' "$source_dir/go.mod" || { echo 'Unexpected source module.' >&2; exit 2; }
source_dir="$(cd "$source_dir" && pwd)"
# No network/toolchain downloads. Populate the local Go module cache separately
# if needed; fail rather than fetching missing modules.
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
# Native means the machine running the compiler, not inherited cross-build env.
[[ "$target_os" != native ]] || target_os="$(go env GOHOSTOS)"
[[ "$target_arch" != native ]] || target_arch="$(go env GOHOSTARCH)"
case "$target_os/$target_arch" in linux/amd64|linux/arm64|darwin/amd64|darwin/arm64) ;; *) echo "Unsupported target: $target_os/$target_arch" >&2; exit 2 ;; esac
mkdir -p "$output_dir"
output_dir="$(cd "$output_dir" && pwd)"
artifact="$output_dir/scopesentry-$target_os-$target_arch"
(
  cd "$source_dir"
  CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" go build -mod=readonly -trimpath -o "$artifact" ./cmd/scopesentry
)
chmod 0755 "$artifact"
echo "$artifact"
