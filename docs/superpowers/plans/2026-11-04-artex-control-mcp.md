# ARTEX Control MCP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let OpenCode read and adjust supported ARTEX configuration through a local MCP companion using ARTEX's existing HTTP API.

**Architecture:** A standalone Go stdio MCP executable calls a fixed ARTEX API origin. Typed reads and immutable, expiring previews feed one approval-gated apply tool. The backend stays the source of truth for persistence and runtime effects.

**Tech Stack:** Go 1.26.3, Go standard-library HTTP/JSON, `github.com/modelcontextprotocol/go-sdk` v1.7.0, OpenCode V2.

**Spec:** `docs/superpowers/specs/2026-11-04-artex-control-mcp-design.md`

**Status:** Proposed plan awaiting design approval. No product code, dependency installation, VM changes, live API requests, or commits are authorized by this planning step.

## Global Constraints

- Use Go 1.26.3 and pin `github.com/modelcontextprotocol/go-sdk` v1.7.0. Keep the default legacy MCP handshake for initial compatibility with OpenCode V2.
- Start in read-only mode. Do not make a write during setup or smoke testing without separate approval.
- Do not import `server`, `agent`, or `db` into the companion.
- No direct database or `config.json` editing, arbitrary HTTP tool, shell execution tool, or implicit task launch.
- HTTP timeout is 15 seconds per request with cancellation propagated. All API bodies are bounded at 8 MiB; tool outputs at 256 KiB. No automatic retries.

## Review Focus

1. An expired or rotated JWT must not leak or trigger a replay; cover in Task 1.
2. Null, omission, false, and IDs above JavaScript's safe integer range must retain meaning; cover in Tasks 2–3.
3. Another operator or a double-click can invalidate a preview; detect stale/duplicate local applies without claiming backend CAS; cover in Task 3.
4. A write can save then fail, time out, or lose its audit result; preserve uncertainty and observed state; cover in Task 3.
5. Empty/huge lists, malicious text, and unavailable ARTEX must not corrupt stdio or cause side effects; cover in Tasks 2 and 4.

## Approval and delivery checklist

- [x] Inspect routes, auth, settings, agent/profile/task behavior, and official OpenCode V2 MCP docs.
- [x] Save the proposed design and implementation plan.
- [x] Complete independent document review and address actionable feedback (concurrency semantics, cancellation cleanup, mismatch outcomes, numeric wire IDs, runtime-effect notes).
- [ ] TRISPER approves design, scope, and implementation method.
- [ ] Implement Tasks 1–5 with RED/GREEN evidence; no live mutation until separately approved.

## File map

| Path | Responsibility |
|---|---|
| `cmd/artex-mcp/main.go`, `main_test.go` | Environment validation, stderr logging, signals, stdio entrypoint |
| `internal/artexcontrol/client/{client,token,errors}.go` and matching tests | Fixed-origin HTTP requests, secure token-file loading, bounded decoding, sanitized errors |
| `internal/artexcontrol/control/{types,reads,validation}.go` and matching tests | Projected DTOs, typed reads, supported fields and input validation |
| `internal/artexcontrol/control/{changes,audit}.go` and matching tests | Preview store, dispatch/read-back workflow, audit intent/outcome |
| `internal/artexcontrol/mcpserver/server.go`, `server_test.go` | SDK registration, schemas, annotations, tool result/error translation |

Additional artifacts owned by Task 5: `examples/opencode/artex-control.jsonc`, `docs/artex-control-mcp.md`, `docs/artex-control-mcp-verification.md`. Task 4 updates `go.mod` and `go.sum` for the pinned SDK. Production backend files and schemas should not change.

## Shared interfaces

These are implementation decisions, not existing code. Use the module prefix `github.com/Autumn-27/artex/` for imports. Keep shared types in the packages below; do not create duplicate competing definitions.

**`client` package:**

```go
type Options struct { BaseURL, TokenFile string; Timeout time.Duration }
type Client struct { /* private state */ }
type APIError struct { Code string; Status int; Dispatched bool }
func New(opts Options) (*Client, error)
func (c *Client) Do(ctx context.Context, method, path string, body any, out any) error
```

`Do` is private to companion implementation boundaries: never expose it as a generic MCP tool. Paths supplied by control code begin `/api/`, are relative, and cannot override the configured origin/prefix. `Dispatched` is conservative: true once a mutation might have reached the server; it is not proof the server committed it. Reject redirects, bypass environment proxies, and redact error bodies rather than echoing them.

**`control` package:**

```go
type API interface {
    Do(ctx context.Context, method, path string, body any, out any) error
}
type Service struct { /* private API reference */ }
func NewService(api API) *Service
func (s *Service) Health(ctx context.Context) (Health, error)
func (s *Service) Settings(ctx context.Context) (Settings, error)
func (s *Service) Profiles(ctx context.Context) ([]Profile, error)
func (s *Service) Agents(ctx context.Context) ([]Agent, error)
func (s *Service) Agent(ctx context.Context, key string) (AgentDetail, error)
func (s *Service) Tasks(ctx context.Context, offset, limit int) (TaskPage, error)
func (s *Service) Task(ctx context.Context, id string) (Task, error)

type ChangeRequest struct {
    Operation string // setting|agent_config|agent_prompt|profile_activate|task_control
    Target string    // agent key, profile ID, or task ID; empty for setting
    Key string       // field name, or pause|resume for task_control
    Value json.RawMessage // scalar or {template,note}; absent for activation/control
}
type ChangePreview struct {
    ChangeID string `json:"change_id"`
    Operation string `json:"operation"`
    Target string `json:"target"`
    Key string `json:"key"`
    Before json.RawMessage `json:"before"`
    After json.RawMessage `json:"after"`
    Effect string `json:"effect"`
    ExpiresAt time.Time `json:"expires_at"`
}
type ApplyResult struct {
    ChangeID string `json:"change_id"`
    Outcome string `json:"outcome"` // verified|partial_or_failed|outcome_unknown
    Observed json.RawMessage `json:"observed,omitempty"`
    Effect string `json:"effect"`
    AuditIncomplete bool `json:"audit_incomplete"`
}
type AuditEvent struct {
    Timestamp time.Time; OperationID, SessionID, Operation, Target, Key string
    BeforeHash, AfterHash, DispatchState, Outcome string; HTTPStatus int
}
type Audit interface { Append(ctx context.Context, event AuditEvent) error }
type ChangeManager struct { /* private immutable previews and mutex */ }
func NewChangeManager(s *Service, audit Audit, now func() time.Time) *ChangeManager
func (m *ChangeManager) Preview(ctx context.Context, req ChangeRequest) (ChangePreview, error)
func (m *ChangeManager) Apply(ctx context.Context, changeID, sessionID string) (ApplyResult, error)
func NewFileAudit(path string) (Audit, error)
```

Define `Health`, `Settings`, `Profile`, `Agent`, `AgentDetail`, `Task`, and `TaskPage` in `types.go` with exactly the projections from the design. Use concrete fields, not a raw backend map. For `Settings`, include the eleven readable settings with their actual JSON names; concurrency fields are read-only, and only the other nine settings are writable. `AgentDetail` contains `agent` plus current `prompt`; `TaskPage` contains `tasks`, `total`, and nullable `next_offset`. Private agent wire DTOs use `*int64` for numeric/null `llm_profile_id`; public DTOs expose it as a nullable decimal string. Do not decode it through float64 or directly into a string.

**`mcpserver` package:**

```go
type Options struct { WriteEnabled bool }
func New(s *control.Service, changes *control.ChangeManager, opts Options) *mcp.Server
```

`changes` may be nil only in read-only mode. `main` rejects write mode without a valid audit path before starting MCP.

---

## Task 1: Fixed-origin API client and secure token loading

**Files:** Create the `client` files/tests in the file map.

**Interfaces:** Consumes `Options`; produces `New`, `Client.Do`, and typed `APIError` for all later tasks. Does not load ARTEX config files or open any database.

- [ ] **Step 1 — Write failing tests:** `TestNewValidatesEndpoint` rejects non-loopback plain HTTP, credentials/query/fragment, and missing host. `TestDoUsesFixedOriginAndPrefix` asserts a base `/artex` yields `/artex/api/health` and caller-supplied absolute paths cannot override it. `TestTokenRotation` writes two 0600 token files in `t.TempDir`, changes the same file's contents between calls, and asserts the fake server sees the new Bearer value without restarting.
- [ ] **Step 2 — Prove RED:** Run `go test ./internal/artexcontrol/client -count=1`. Expect missing implementation failures, not a database or live-network dependency.
- [ ] **Step 3 — Implement:** Use `net/http`, a direct transport, `CheckRedirect` rejection, 15-second default timeout, context propagation, 8 MiB bounded decoder, and token-file regular-file/ownership/mode checks. Read from the opened file descriptor and validate it, rather than relying on a path-only precheck; reject symlinks and empty tokens. Never include tokens, file contents, upstream bodies, or authorization headers in errors.
- [ ] **Step 4 — Extend and prove GREEN:** Add `TestTokenFileRejectsUnsafeInputs` (symlink, directory, permissive mode, empty file), `TestRedirectNeverForwardsToken`, `TestEnvironmentProxyIgnored`, `Test401Sanitized` (code `auth_expired_or_invalid`), `TestCancellation`, `TestOversizeBody`, `TestNonJSONBody`, and `TestWriteNotRetried`. Assert exactly one mutation reaches a deliberately failing fake handler. Run `go test -race ./internal/artexcontrol/client -count=1`; expect PASS.
- [ ] **Step 5 — Checkpoint:** Review the staged diff, then commit only Task 1 files if implementation commits were approved: `feat(mcp): add ARTEX control API client`.

## Task 2: Typed reads and strict domain validation

**Files:** Create `control/types.go`, `reads.go`, `validation.go`, and their tests.

**Interfaces:** Consumes `client.Client` through `API`; produces `Service` and all seven read methods plus projected DTOs for Tasks 3–4. Internal helper `validateChange(req ChangeRequest) error` validates the design's allowlists and exact scalar types; `ChangeRequest` is declared now in `types.go` for Task 3.

- [ ] **Step 1 — Write failing tests:** Table-test `TestReadRouteMapping` for each method against its exact design route. `TestReadProjectionDropsSecrets` seeds fake profile/settings JSON with `api_key`, `api_key_hint`, proxy userinfo, unknown fields, and auth-shaped fields; assert none appears in marshaled tool DTOs. `TestLargeIDsPreserved` uses string `"9007199254740993"` and expects the same ID back. `TestTaskPage` checks numeric ID ordering (`2` before `10`), empty arrays, offsets, total, and limit 1–100.
- [ ] **Step 2 — Prove RED:** Run `go test ./internal/artexcontrol/control -run 'Test(Read|LargeIDs|TaskPage)' -count=1`; expect absent read implementation failures.
- [ ] **Step 3 — Implement:** Decode the real wrappers: `{profiles:[...]}`, `{agents:[...]}`, `{tasks:[...],active:...}`; task detail is unwrapped, and agent detail contains `{agent,prompt,...}`. Read only explicit routes. Validate agent keys and ID strings before constructing paths. Sort/slice task output locally, but never suggest backend pagination. Enforce a 256 KiB marshaled output limit without silently truncating current state.
- [ ] **Step 4 — Extend and prove GREEN:** `TestValidationRejectsUnknownOrWrongTypes` covers secret/proxy keys, multi-field attempts, boolean versus string false, null budgets, negative/overflow/exponent IDs, decimal budgets, unsupported operations, and out-of-range digest minutes. `TestConcurrencyReadOnly` rejects preview writes to either concurrency field without making a mutation, even when the fake server hides a stored limit while disabled. `TestProfileBindingThreeStates` distinguishes no requested edit, explicit null unbind, and large decimal ID string; decode backend numeric `llm_profile_id:9007199254740993` and null into exact public values, and later dispatch the exact numeric ID. `TestOutputLimit` rejects huge prompts/lists. `TestReadsDoNotMutate` asserts every request is GET. Run `go test -race ./internal/artexcontrol/client ./internal/artexcontrol/control -count=1`; expect PASS.
- [ ] **Step 5 — Checkpoint:** Review only the new control read/validation files; approved commit message: `feat(mcp): expose typed ARTEX control reads`.

## Task 3: Immutable previews, one-shot apply, read-back, and audit

**Files:** Create `control/changes.go`, `audit.go`, `changes_test.go`, `audit_test.go`; extend `types.go` only for the shared interfaces already specified.

**Interfaces:** Consumes all Task 2 reads and validation. Produces `ChangeManager.Preview`, `Apply`, `NewFileAudit`, and the exact types in Shared interfaces. `now` is injectable; audit test doubles can independently fail intent or outcome writes.

- [ ] **Step 1 — Write failing tests:** `TestPreviewNeverMutates` for all five operations; assert no PUT/POST, current/proposed values, 300-second expiry, and a nonempty random ID. `TestApplyDispatchMapping` asserts: one-key setting PUT; one-field agent config PUT; prompt PUT with only `template`/`note`; activation POST with lossless numeric `id`; control POST with `action` equal to pause or resume. Use `json.Decoder.UseNumber` in fake-server body assertions for large IDs.
- [ ] **Step 2 — Prove RED:** Run `go test ./internal/artexcontrol/control -run 'Test(Preview|Apply)' -count=1`; expect missing manager behavior failures.
- [ ] **Step 3 — Implement:** Store copied immutable previews with a 100-entry cap; remove expired entries before checking capacity. Serialize Apply in this process. Re-read fingerprinted state, reject stale/expired/used changes, write audit intent, consume preview, send exactly one mutation, then re-read even after a dispatched failure. After dispatch, use `context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)` for read-back and final audit, while retaining the per-request 15-second bound. Graceful shutdown drains cleanup; forced termination leaves audit intent unresolved. Populate verified/partial_or_failed/outcome_unknown exactly as the design requires. A 2xx followed by definite mismatch is partial_or_failed. Resume verification accepts observed unpaused queued/running/terminal state after a successful control response; pause requires paused=true. Hash canonical typed state; prompt fingerprint is SHA-256 of exact UTF-8 text. Preserve the documented remote-writer race; do not implement fake CAS through a GET followed by PUT.
- [ ] **Step 4 — Extend and prove GREEN:** `TestPreviewExpiryAndCapacity`, `TestStalePreviewNoWrite`, `TestConcurrentApplyWritesOnce`, `TestUnknownIDNoWrite`, `TestAuditIntentFailureBlocksWrite`, `TestOutcomeAuditFailureReported`, `TestWriteSavedThen500`, `TestTimeoutAfterPossibleCommit`, `TestReadBackFailureUnknown`, `TestVerificationMismatch`, `TestQueuedResumeReported`, `TestTerminalOrAlreadyPausedRejected`, and `TestPromptValidationFailure`. Assert a rejected/no-dispatch apply cannot be confused with verified success; a server error after save is partial_or_failed even when observed state matches; timeout is outcome_unknown; no automatic retry or rollback occurs. File-audit tests assert mode 0600, unsafe path rejection, hashes only, and absence of raw prompts/tokens/notes. Run `go test -race ./internal/artexcontrol/... -count=1`; expect PASS.
- [ ] **Step 5 — Checkpoint:** Review error and audit paths separately before an approved commit: `feat(mcp): add previewed and verified ARTEX changes`.

Additional Task 3 assertions: `TestCallerCancellationAfterCommit` verifies one mutation, a consumed preview, detached bounded read-back and outcome audit; `TestGracefulShutdownDrainsCleanup` checks cleanup completion before exit; `TestResumeImmediatelyCompleted` returns verified with observed terminal/unpaused state; `TestVerificationMismatch` expects partial_or_failed on a 2xx plus wrong value. Tests must not wait real expiry/cleanup durations; use injected clocks and cancelable fake endpoints.

Runtime-effect copy is part of implementation: workers → `new tasks`; noa_compaction → `next agent run`; prompt → `saved; running-turn adoption not verified`; agent config/activation → `saved state verified; runtime adoption not verified`. Test these exact messages in `TestRuntimeEffectNotes`. Concurrency is read-only in v1. Verification proves observable saved state, not successful runtime rebuilds or downstream LLM calls.

## Task 4: SDK tools and runnable stdio companion

**Files:** Create `mcpserver/server.go`, `server_test.go`, `cmd/artex-mcp/main.go`, `main_test.go`; modify `go.mod` and `go.sum`.

**Interfaces:** Consumes Tasks 1–3. Produces `mcpserver.New` and `artex-mcp` as a stdio executable. Use SDK `mcp.NewServer`, typed `mcp.AddTool`, and `mcp.StdioTransport`; no network MCP listener.

- [ ] **Step 1 — Write failing tests:** `TestReadOnlyCatalog` expects exactly seven reads and no apply/preview tools. `TestWriteCatalog` expects those seven reads, the five named previews, and `apply_change`. `TestToolSchemas` checks each required field, scalar/nullable types, unknown-field rejection, and that `apply_change` accepts only `change_id`. `TestStartupConfig` covers missing endpoint/token file, invalid write toggle, and missing audit file in write mode.
- [ ] **Step 2 — Prove RED and pin dependency:** Add SDK v1.7.0 using `go get github.com/modelcontextprotocol/go-sdk@v1.7.0`, inspect the `go.mod`/`go.sum` delta, then run `go test ./internal/artexcontrol/mcpserver ./cmd/artex-mcp -count=1`; expect missing registration/entrypoint behavior. Do not upgrade unrelated dependencies. The SDK version exists and declares Go 1.25.0, below this repo's 1.26.3 floor.
- [ ] **Step 3 — Implement:** Register exact tool names from the design, with typed inputs and explicit schema validation (SDK inferred schemas alone must be checked for unwanted unknown-field acceptance). Map previews to `ChangeRequest`; register none in read-only mode. Reject write mode when `changes` is nil. Log only to stderr, never stdout. Propagate signals/cancellation. Convert domain/API errors into sanitized MCP errors; encode structured results plus a concise text fallback. Treat returned ARTEX prompt/name text as data; tool descriptions must not invite instruction-following from target content.
- [ ] **Step 4 — Extend and prove GREEN:** Use SDK in-memory transports for a real initialize/list/call cycle. `TestReadCallDoesNotStartTask`, `TestMCPErrorRedaction`, `TestApplyUnknownID`, `TestStdoutProtocolOnly`, and `TestProcessExitOnEOF` cover API outage, hostile prompt text, no task creation/selection, and clean shutdown. Metadata `ai.opencode/sessionID` is optional audit correlation, never authentication. Run `go test -race ./internal/artexcontrol/... ./cmd/artex-mcp -count=1`, `go vet ./internal/artexcontrol/... ./cmd/artex-mcp`, and `go build -o "$ARTEX_MCP_TEST_BIN" ./cmd/artex-mcp` with an executor-chosen output path in the approved temp directory; expect all pass. Tests may execute only this companion against fake loopback HTTP fixtures.
- [ ] **Step 5 — Checkpoint:** Review catalog annotations and default-off write mode; approved commit message: `feat(mcp): add ARTEX stdio control server`.

## Task 5: OpenCode setup, operator guide, and final evidence

**Files:** Create `examples/opencode/artex-control.jsonc`, `docs/artex-control-mcp.md`, `docs/artex-control-mcp-verification.md`; extend `mcpserver/server_test.go` with the configuration-contract test.

**Interfaces:** Consumes the built companion and its exact catalog. Produces an inactive example config, repeatable setup steps, and honest offline/live verification status. Do not edit `~/.config/opencode` or the active project config without a separate request.

- [ ] **Step 1 — Write the failing setup-contract test:** `TestOpenCodeExamplePermissions` loads the example as JSON (JSON is valid JSONC), asserts `mcp.servers.artex.type=local`, command is `"{env:ARTEX_MCP_BIN}"`, protocol `legacy`, environment write-enabled value `"false"`, exactly the design's read/preview allows, and final `artex_apply_change` rule `ask`. Assert there is no literal token and no broad allow after the ask. Also assert configured server name yields the expected normalized permission names.
- [ ] **Step 2 — Prove RED:** Run `go test ./internal/artexcontrol/mcpserver -run TestOpenCodeExamplePermissions -count=1`; expect missing example failure.
- [ ] **Step 3 — Write setup artifacts:** Use the skeleton below, enumerating the seven actual reads and five previews as individual allow rules between the broad ask and final apply ask. Document build/install to an operator-chosen absolute path, token provisioning/renewal through existing login, mode 0600, absolute audit path, process environment inheritance, optional SSH tunnel using local port 18787, and preserving existing OpenCode config. Explain V2-only syntax and verify the installed major version before merging any config; if V1 is installed, revisit setup rather than silently mixing schemas.

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "servers": {
      "artex": {
        "type": "local",
        "command": ["{env:ARTEX_MCP_BIN}"],
        "protocol": "legacy",
        "environment": {
          "ARTEX_BASE_URL": "{env:ARTEX_BASE_URL}",
          "ARTEX_TOKEN_FILE": "{env:ARTEX_TOKEN_FILE}",
          "ARTEX_MCP_WRITE_ENABLED": "false",
          "ARTEX_MCP_AUDIT_FILE": "{env:ARTEX_MCP_AUDIT_FILE}"
        }
      }
    }
  },
  "permissions": [
    {"action": "artex_*", "resource": "*", "effect": "ask"},
    {"action": "artex_apply_change", "resource": "*", "effect": "ask"}
  ]
}
```

This is a skeleton for the implementer, not the final example: the twelve explicit read/preview allows must be present in the created example, tested in Step 1. Default read-only startup hides the preview/apply tools until the operator deliberately enables writes. Explain that OpenCode approval is client-side and the JWT is still an admin credential; a preview ID alone does not establish approval.

- [ ] **Step 4 — Prove GREEN and record evidence:** Run the focused normal/race tests, focused vet, build, and `git diff --check`; record actual commands and outcomes. Do not run `go test ./...`, or server/agent/db test runners with inferred config. Optional backend checks require an explicit disposable test DSN and separate approval. Mark live checks `NOT RUN` unless actually approved and performed. After separate setup approval, verify `opencode mcp list`, health, and supported settings read; a live mutation needs separate approval for the named field and a second approval to restore it. Verify in the ARTEX UI as well as API read-back; never resume a real task as a connectivity test.
- [ ] **Step 5 — Review and handoff:** Request independent whole-change review, fix confirmed issues, and attach redacted evidence. Approved final commit message: `docs(mcp): document ARTEX OpenCode control setup`. No push, deployment, release, or changelog change is implicit.

## Acceptance examples

| Request in OpenCode | Expected behavior |
|---|---|
| “Show ARTEX settings and model profiles.” | GET-only calls; supported fields; no secrets or changes. |
| “Set ARTEX workers to 4.” | In read-only mode, explain writes are disabled. In approved write mode, preview before/after, ask before apply, verify saved value, explain only new tasks use it. |
| “Use profile 12 for the worker agent.” | Read the existing agent/profile, preserve all omitted fields, preview only `llm_profile_id`, ask, apply, re-read. |
| “Pause task 15.” | Read/preview/approval; preserve backend lifecycle conflicts and report actual paused/queued state. |
| “Update the planner prompt.” | Read current prompt, preview exact full replacement, approval, validated versioned save, re-read; no claim an in-flight run changed. |

## Self-review record

Spec coverage maps client/auth to Task 1; reads/validation to Task 2; edits, runtime-effect notes, audit and uncertainty to Task 3; catalog/protocol/default-off behavior to Task 4; permissions/operator setup/acceptance to Task 5. Interfaces are defined once above. No new backend route or schema is necessary for this slice. Known limitations remain explicit: admin-scoped JWT, manual renewal, no global CAS, local non-tamper-proof audit, no profile editing, and no chat bridge.

Before execution: obtain design approval and select native or subagent-driven implementation. Before live setup: confirm actual endpoint/tunnel and OpenCode version. Before any live mutation: approve that exact operation.

## Planning self-evaluation

Summary: a repo-backed, reviewed draft; implementation and live acceptance remain untested.

| Axis | Score | Evidence and remaining improvement |
|---|---|---|
| Accuracy | 4/5 | Routes and DTOs were checked in source; review corrected hidden concurrency state and numeric bindings. Verify SDK schemas during implementation. |
| Completeness | 4/5 | Five tasks cover the scoped release and failure cases; complete administration is deliberately deferred and the scope needs approval. |
| Clarity | 4/5 | Separate design, file map, and task steps; the document is technical, so the handoff must give a short plain-English summary. |
| Actionability | 4/5 | Exact file paths, interfaces, commands and expected results; actual endpoint and OpenCode version are still setup inputs. |
| Conciseness | 4/5 | Shared interfaces are centralized; security/uncertainty rules appear in both the design and their owning test steps for executor context. |

Overall: **4.0/5**. Highest-impact next improvements: approve the companion/scope; validate the real SDK schemas with offline contract tests; confirm endpoint/version only before setup. Self-check: TRISPER can approve the short summary without reading every implementation detail; the full documents support a fresh engineer.
