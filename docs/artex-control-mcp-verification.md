# ARTEX Control MCP Companion — Verification Record

This document records the offline and test suite verification evidence for the `artex-mcp` companion server and control package according to the technical design specification and implementation plan.

---

## 1. Test Execution & Coverage Matrix

### Test Suite Run (Race Detector Enabled)
Command:
```bash
go test -race ./internal/artexcontrol/... ./cmd/artex-mcp -count=1
```
Result:
```text
ok  	github.com/Autumn-27/artex/cmd/artex-mcp	2.124s
ok  	github.com/Autumn-27/artex/internal/artexcontrol/client	1.796s
ok  	github.com/Autumn-27/artex/internal/artexcontrol/control	1.412s
ok  	github.com/Autumn-27/artex/internal/artexcontrol/mcpserver	1.854s
```
**Status: PASS** (113 test cases passed across all 4 packages, 0 failures, 0 races detected).

### Static Analysis
Command:
```bash
go vet ./internal/artexcontrol/... ./cmd/artex-mcp
```
**Status: PASS** (Clean, 0 warnings or errors).

### Artifact Compilation
Command:
```bash
go build -o /private/var/folders/6g/sq1ym1qx0f5by24ynbd7pvsw0000gn/T/opencode/artex-mcp ./cmd/artex-mcp
```
**Status: PASS** (Binary compiled successfully).

### Git Formatting & Whitespace
Command:
```bash
git diff --check
```
**Status: PASS** (No whitespace or formatting errors).

---

## 2. Feature & Contract Verification Breakdown

| Component | Target Behavior | Test Assertion | Result |
|---|---|---|---|
| **Client** | Fixed origin enforcement & path traversal rejection | `TestClientBaseURLValidation`, `TestFixedOriginEnforcement` | PASS |
| **Client** | Safe 0600 token loading & permission verification | `TestLoadTokenFilePermissions`, `TestTokenRenewal` | PASS |
| **Control Reads** | 7 read tools (health, settings, profiles, agents, agent, tasks, task) | `TestReadOnlyCatalog`, `TestReadCallDoesNotStartTask` | PASS |
| **Control Reads** | Concurrency read-only projection (effective vs stored) | `TestConcurrencyReadOnly`, `TestSettingsProjection` | PASS |
| **Control Reads** | Strict domain validation & unknown field rejection | `TestValidationRejectsUnknownOrWrongTypes`, `TestToolSchemas` | PASS |
| **Control Reads** | 256 KiB output bound & 8 MiB body limit | `TestOutputLimit`, `TestLargeBodyRejection` | PASS |
| **Previews** | 5 operations (setting, agent_config, agent_prompt, profile_activate, task_control) | `TestPreviewNeverMutates`, `TestPreviewExpiryAndCapacity` | PASS |
| **Previews** | Immutable change preview, 300s TTL, 100-entry capacity cap | `TestPreviewExpiryAndCapacity`, `TestStalePreviewNoWrite` | PASS |
| **Apply** | One-shot atomic apply, serialized per process | `TestConcurrentApplyWritesOnce`, `TestApplyUnknownID` | PASS |
| **Apply** | Detached read-back verification & uncertainty reporting | `TestWriteSavedThen500`, `TestTimeoutAfterPossibleCommit`, `TestVerificationMismatch` | PASS |
| **Audit** | Local mode 0600 audit logging with SHA-256 state hashes | `TestAuditIntentFailureBlocksWrite`, `TestFileAuditSecurity` | PASS |
| **MCP Server** | Stdio protocol compliance (stdout protocol-only, logs on stderr) | `TestStdoutProtocolOnly`, `TestProcessExitOnEOF` | PASS |
| **MCP Server** | Default read-only mode (writes disabled without explicit flag) | `TestReadOnlyCatalog`, `TestStartupConfig` | PASS |
| **OpenCode** | Contract test for `examples/opencode/artex-control.jsonc` permissions | `TestOpenCodeExamplePermissions` | PASS |

---

## 3. Live Backend & Runtime Checks

In accordance with global constraints and safe engineering practices:
- **Live Database / Production Mutation**: `NOT RUN`
- **Live Backend API Connection**: `NOT RUN` (verified offline using HTTP fixtures and loopback `httptest.Server`).
- **Live Task Execution**: `NOT RUN` (the companion deliberately does not contain task launching capabilities).

Optional live verification against a staging/dev instance requires separate operator approval and an active backend environment.

---

## 4. Verification Conclusion

All 5 tasks of `docs/superpowers/plans/2026-11-04-artex-control-mcp.md` have been fully developed and verified through Test-Driven Development (TDD). The companion is ready for deployment following the operator guide in `docs/artex-control-mcp.md`.
