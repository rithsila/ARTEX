# ARTEX Control MCP — Proposed Design

**Status:** Draft for approval. No implementation or deployment is approved by this document.
**Date:** 2026-11-04
**Reader:** TRISPER and the engineer implementing the companion plan.

## Intent

Control ARTEX from OpenCode in plain language, without opening an SSH shell to edit settings or asking ARTEX's chat agent to change its own configuration. OpenCode calls named MCP tools; those tools use ARTEX's existing authenticated HTTP API.

TRISPER selected a custom MCP server. The language, transport, first-release scope, and controls below are proposed engineering choices, not yet approved requirements.

## Approaches

| Approach | Benefit | Cost | Decision |
|---|---|---|---|
| Local Go MCP companion → ARTEX HTTP API | Matches this Go repo; no new VM listener; uses existing runtime update behavior | Needs network access and an ARTEX login token | Recommended first release |
| MCP endpoint embedded inside ARTEX | One server deployment; can later support narrow service credentials | Changes the live backend and needs transport/auth design | Later if needed |
| Python/Node MCP wrapper → ARTEX HTTP API | Fast prototype; also avoids backend changes | Adds another runtime and package lifecycle | Viable, but not the proposed default |

The earlier REST-only and direct-file options were discussed; TRISPER chose custom MCP instead. This design recommends a local companion within that choice. It does not claim approval of the companion topology.

## Current implementation evidence

| Concern | Source | Consequence |
|---|---|---|
| API routes and frontend mounting | `server/server.go:655–942` | Existing settings, profile, agent, and task endpoints can be reused. `/api/mcp` manages MCP servers consumed by ARTEX; it is not an MCP control endpoint. |
| Authentication | `server/auth.go:19–23, 86–137, 261–298` | Login returns a seven-day JWT; send it through the Bearer header. No scoped service token or refresh endpoint was found in this auth implementation. |
| Settings | `server/server.go:3285–3321, 3361–3555` | API supports partial updates and runtime effects, but writes are sequential, not one atomic transaction. Even an error can follow a saved change. |
| Worker/concurrency behavior | `server/manager.go:290–336` | `workers` affects tasks started afterwards. Changing concurrency can immediately admit queued tasks; related persistence uses two writes. |
| Agent config | `server/server_mgmt.go:349–466` | Optional fields preserve omitted values; `llm_profile_id` distinguishes absent, null, and numeric ID. Agent caches are invalidated. |
| Profiles | `server/server_mgmt.go:1784–1842, 1927–1945`; `server/dto.go:588–644` | Listing masks keys; activation hot-applies. Saving a profile uses a broader object, not a guaranteed patch; leave profile editing out of v1. |
| Prompts | `server/server_mgmt.go:469–532` | Reading an agent includes its current prompt; saving validates and versions it. The save response does not prove an already-running turn adopted the prompt. |
| Tasks | `server/server.go:1036–1137, 1728–1737`; `server/task_control.go:89–149` | Listing/detail reads and explicit pause/resume exist. `/api/active` can start an engine; it must not be presented as a read-only selection tool. |

Shared project memory also warns that backend integration test runners may clean persistent data. Use fake HTTP servers for the companion, and an explicitly isolated database for any optional backend integration run.

## Goals and boundaries

### First release

Read health, supported settings, existing profiles, agent config/current prompts, and task status. Preview and apply one supported setting or agent field at a time, replace a prompt with a versioned save, activate an existing model profile, and pause/resume an existing task.

All writes require a preview, explicit OpenCode approval, and a read-back check. Start in read-only mode. Do not make a write during setup or smoke testing without separate approval.

### Not in the first release

No direct database or `config.json` editing, arbitrary HTTP tool, shell execution tool, service restart/update, API-key editing, profile create/edit/delete, task creation/deletion, asset/scope editing, intercept approvals, or ARTEX chat delegation. Skill/tool/MCP binding management is a separate extension after this slice works.

This is a focused first release, not a claim of complete ARTEX administration. If full administration is needed immediately, expand and approve the scope before implementation.

## Architecture and deployment

```text
OpenCode on the operator's machine
  → MCP stdio → local artex-mcp binary
  → authenticated ARTEX HTTP API on the VM
  → existing ARTEX persistence and runtime update paths
```

Use Go 1.26.3 and pin `github.com/modelcontextprotocol/go-sdk` v1.7.0. Keep the default legacy MCP handshake for initial compatibility with OpenCode V2. Add `cmd/artex-mcp` and focused packages under `internal/artexcontrol`; do not import `server`, `agent`, or `db` into the companion.

Use an existing HTTPS endpoint with normal certificate verification, or an operator-managed SSH tunnel terminating at the VM's loopback ARTEX listener. Example tunnel: `ssh -N -L 18787:127.0.0.1:8787 <ssh-host>`. This is an optional setup command, not a connection already verified. The local forwarded port is 18787 to avoid the other known local service using 8787. No firewall or VM change is part of this plan.

The actual deployment URL, SSH alias if needed, and installed OpenCode version remain setup inputs. Do not infer the live ARTEX endpoint from historical VM memory.

## Authentication and process settings

| Variable | Rule |
|---|---|
| `ARTEX_BASE_URL` | Required absolute HTTP(S) URL ending at the host or deployment path prefix, not `/api`; reject query, fragment, or embedded credentials. Plain HTTP permitted only for loopback/tunnel use. |
| `ARTEX_TOKEN_FILE` | Required absolute path to an operator-managed file holding only the JWT, with trailing whitespace allowed; regular file owned by the current account, mode 0600 on Unix; reject symlinks. Re-read before each API request to support token rotation without restart. |
| `ARTEX_MCP_WRITE_ENABLED` | Defaults to `false`; only exact `true` exposes write previews and `apply_change`. |
| `ARTEX_MCP_AUDIT_FILE` | Required when write mode is enabled; absolute path outside the repo, append-only JSONL from this process, mode 0600 on Unix. Failure to record an intent blocks the write. |

Token provisioning and renewal remain a human action through the existing ARTEX login flow. Do not ask for tokens/passwords in chat or place them in tool arguments, logs, or source files. On HTTP 401, report `auth_expired_or_invalid` and the token-file renewal action. Do not automate login or store an admin password in v1.

The JWT still has ARTEX's existing broad admin authority. The companion's allowlist limits its own behavior; it does not create backend-enforced roles or revoke a stolen JWT. A narrow, revocable service identity requires later backend work.

## MCP tool contract

The server name in OpenCode is `artex`. Names below are MCP-local names; OpenCode permission names have the `artex_` prefix.

| Group | Tools | API mapping |
|---|---|---|
| Reads | `health`, `get_settings` | `GET /api/health`, `GET /api/settings` |
| Reads | `list_profiles`, `list_agents`, `get_agent` | `GET /api/llm/profiles`, `GET /api/agents`, `GET /api/agents/{key}` |
| Reads | `list_tasks`, `get_task` | `GET /api/tasks`, `GET /api/tasks/{id}` |
| Previews | `preview_setting_change`, `preview_agent_change`, `preview_prompt_change` | Read current state, validate locally, return an immutable preview; no mutation request. |
| Previews/apply | `preview_profile_activation`, `preview_task_control`, `apply_change` | Preview by reading; apply maps to `POST /api/llm/profiles/active`, `POST /api/tasks/{id}/control`, or the matching config/prompt write. |

Five preview operations: `setting`, `agent_config`, `agent_prompt`, `profile_activate`, `task_control`. The preview returns an opaque random `change_id`, expires after 300 seconds, and binds the operation, target, before-state fingerprint, and exact new value. `apply_change` accepts only `{change_id}`. Pending changes are process-local, capped at 100, and disappear on restart. Apply never accepts replacement payloads.

Expose read-only/idempotent MCP annotations accurately, but never rely on annotations as approval enforcement. `apply_change` is non-idempotent and must prompt through OpenCode permissions.

### Supported settings

For settings writes, allow only `workers`, `traffic_capture`, `llm_record`, `web_search_enabled`, `llm_pool_enabled`, `llm_pool_bind_fallback`, `noa_compaction`, `notify_enabled`, and `notify_digest_interval_min`. Read results also include `task_concurrency_enabled` and `task_concurrency_limit`, but both concurrency fields are read-only in v1. When disabled, the backend reports effective limit 5 rather than a hidden stored limit, and enabling can overwrite that stored limit. Do not offer a misleading verified edit for these fields without first addressing that backend contract in a separately approved change.

Booleans must be JSON booleans. Integer settings must be positive signed 32-bit integers; `notify_digest_interval_min` must be 1–1440. Read results project these eleven fields only. Label concurrency limit as the API-reported effective value, not necessarily the stored value. Keep proxy URLs, interpreter paths, search keys, callback URLs, and other settings out of v1. Do not turn on a feature as part of setup.

`preview_setting_change` accepts `{key, value}` with strict validation and one key only; apply sends `PUT /api/settings` containing that field only. The backend remains the final authority, and one-field requests still do not imply atomic runtime/persistence changes.

### Supported agent edits

Allow exactly one of `max_turns`, `run_seconds`, `web_search`, and `llm_profile_id` per preview. Budgets are nonnegative signed 32-bit integers. `llm_profile_id` accepts null (unbind) or a positive decimal ID string (bind), converted losslessly to an integer for the backend JSON body. Omission is never converted to null. Validate agent keys against `^[a-z][a-z0-9_]*$` and confirm the agent/profile exists during preview.

Prompt replacement is separate: `{agent_key, template, note}`; require a nonempty template at most 65536 UTF-8 bytes and a note at most 256 UTF-8 bytes. Save through `PUT /api/agents/{key}/prompt` with `template` and `note`; the server performs final template validation/versioning. Prompt runtime effect is `saved; running-turn adoption not verified`, not `effective immediately`.

### IDs, output, and bounded reads

All exposed IDs are positive decimal strings within signed 64-bit range. The backend agent DTO's `llm_profile_id` is a number or null, unlike its string `id`: decode with a private `*int64` wire field and project to a nullable decimal string, never through float64. Reject path separators, negative IDs, exponent notation, and overflow before making a request. `list_tasks` accepts `{offset: 0, limit: 20}`, with limit 1–100, locally sorts by numeric ID, then slices and returns `total` and `next_offset`. The current backend returns the whole list; this is output pagination, not backend pagination.

All API bodies are bounded at 8 MiB; tool outputs at 256 KiB. Project DTOs rather than returning raw JSON. Tasks expose only ID, name, status, paused, queued; agents expose key, ID, name, supported config fields, and the requested current prompt. Profiles expose ID, name, model, format, and `is_default`, not key hints, URLs, or proxies. Oversize results return `response_too_large`, never silently truncate config used for previews.

## Write lifecycle and honest failure handling

1. Preview: read current state; validate the exact edit; return before/after and effect note. ARTEX prompt/task text is untrusted data, not instructions.
2. Approval: OpenCode asks before `artex_apply_change`; the server also requires write mode. A preview ID is not proof of human approval and is not an authorization credential.
3. Apply: serialize this companion's apply calls; re-read relevant state and reject expired/used/changed previews before writing. After those checks, persist an audit intent and consume the preview before dispatch. Never retry a mutation automatically.
4. Verify: read again after any dispatched mutation, including a non-2xx response or lost response. Return `verified` only when the intended observable state matches and the API did not report an error. Report `partial_or_failed` on a non-2xx even if read-back shows the requested value, or on a 2xx with a definite read-back mismatch; report `outcome_unknown` on transport loss/failed read-back. Return observed state when available. Never claim a failed request made no change. After dispatch, verification and final audit use a separate context detached from caller cancellation, with one shared 20-second cleanup deadline; each HTTP request still has its 15-second limit. Graceful shutdown allows this cleanup to finish; forced termination cannot guarantee verification or an outcome record, so an audit intent without outcome means unresolved, not failed/no change.
5. Audit: append the final outcome; never auto-rollback. If a post-dispatch audit append fails, surface `audit_incomplete` alongside the observed operation result. A deliberate revert needs a fresh preview and approval.

Fingerprints cover the edited field for config, the full current prompt for prompt replacement, the active profile ID for activation, and status/paused/queued for task control. Local serialization and pre-write fingerprints detect many stale edits, but another API client can still write between check and apply: the backend has no demonstrated compare-and-swap contract. State this race plainly; do not promise global conflict prevention.

Config/prompt verification compares the exact edited value or prompt. Profile activation verifies the selected profile has `is_default=true`. Task control permits only `pause` and `resume`; terminal tasks are rejected locally and backend conflicts remain visible. Pause verification requires observed `paused=true`. Resume requires a successful control response and observed `paused=false`; a subsequent queued, running, or terminal state is reported as observed rather than required to equal a predicted running state. Reject a pause already paused or a resume not paused before dispatch. Re-check the actual backend state after applying.

Effect notes distinguish saved state from runtime adoption: `workers` affects new tasks, `noa_compaction` affects the next agent run, prompt replacement is saved but running-turn adoption is unverified, and agent config/profile activation report `saved state verified; runtime adoption not verified`. Some backend rebuild paths log or ignore errors, so read-back alone cannot prove adoption.

Audit stores timestamp, random operation ID, optional OpenCode session correlation, operation, target, field name, SHA-256 before/after fingerprints, dispatch state, HTTP status, and outcome. It never stores tokens, raw bodies, prompts, notes, or task content. Session IDs are correlation only, not identity. A local audit file is not a tamper-proof server audit.

HTTP timeout is 15 seconds per request with cancellation propagated. Reject redirects; do not send the JWT to a different host or inherited environment proxy. Set the companion's HTTP proxy behavior explicitly to direct; use the tunnel or HTTPS endpoint. Reads and mutations have no automatic retries in v1.

## OpenCode setup and acceptance

Document a project-local example under `examples/opencode/artex-control.jsonc`; do not alter the active OpenCode config during development. Use V2 `mcp.servers.artex`, a local command containing an absolute installed binary path, `environment`, and legacy protocol. Start with write mode false. Permission rules allow the seven reads and preview tools but ask before `artex_apply_change`; ordered rules must not accidentally override that ask.

The setup guide must explain that environment variables belong to OpenCode's background service/process environment, which may differ from a new terminal. Include `opencode mcp list` and `/mcps` checks and preserve unrelated config. Config reload is automatic; do not tell the operator to restart ARTEX for these API changes.

First acceptance is offline: SDK handshake, schemas, read mapping, preview/no-write behavior, apply approval policy example, expiry/stale detection, redaction, and fake HTTP error cases. Live acceptance is separately approved: health and settings read, then one reversible selected setting edit, verify in ARTEX UI, then a fresh approved preview to restore it. No real task is resumed during smoke testing.

## Approval needed

Approve or revise the proposed local Go companion and v1 tool scope. Before live setup, provide the actual API URL (or SSH alias for a tunnel) and confirm the OpenCode major version. Backend service credentials, remote HTTP MCP, complete administration, and ARTEX chat bridging are future extensions, not hidden tasks.
