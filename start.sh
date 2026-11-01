#!/bin/sh
# ARTEX supervisor launch script (Linux / macOS / Docker ENTRYPOINT)
#
# Usage:
#   ./start.sh                       Run in foreground (press Ctrl-C to stop)
#   nohup ./start.sh >artex.log 2>&1 &   Run in background as daemon
#   ./start.sh -addr :9000           Pass extra arguments directly to artex
#
# Purpose: launches artex and decides whether to restart based on exit code:
#   0      Clean shutdown by user   → Exit loop
#   75     Restart requested        → Restart immediately (after selfupdate or rollback)
#   other  Crash or abnormal exit  → Backoff and restart (1→2→4... up to 60s)
#
# Note: Download, SHA256 validation, and binary replacement are handled inside Go
# (selfupdate package) on startup to ensure reliability across platforms.
set -u

cd "$(dirname "$0")" || exit 1

BIN=./artex
[ -x "$BIN" ] || { echo "[artex] binary not found or not executable: $BIN" >&2; exit 1; }

RESTART_CODE=75
MAX_DELAY=60

child=0
stopping=0

# Forward termination signals to artex.
# Required under Docker: docker stop sends SIGTERM only to PID 1.
# Forwarding allows artex to perform graceful shutdown before SIGKILL.
forward() {
	stopping=1
	if [ "$child" -ne 0 ]; then
		kill -TERM "$child" 2>/dev/null || true
	fi
}
trap forward INT TERM

delay=1
while :; do
	"$BIN" "$@" &
	child=$!

	# A signal interrupts wait and returns >128 while child is shutting down gracefully.
	# Wait once more to obtain the true exit status code.
	wait "$child"
	code=$?
	if [ "$code" -gt 128 ]; then
		wait "$child"
		code=$?
	fi
	child=0

	if [ "$stopping" -eq 1 ]; then
		echo "[artex] stopped"
		exit 0
	fi

	case "$code" in
		0)
			echo "[artex] clean shutdown"
			exit 0
			;;
		"$RESTART_CODE")
			# Update or rollback staged: restarting artex applies the change on boot.
			echo "[artex] restart requested (applying new version)..."
			delay=1
			;;
		*)
			echo "[artex] abnormal exit (code=$code), restarting in ${delay}s" >&2
			sleep "$delay"
			delay=$((delay * 2))
			[ "$delay" -gt "$MAX_DELAY" ] && delay=$MAX_DELAY
			;;
	esac
done
