# ARTEX Control MCP Companion — Operator Guide

This guide explains how to install, configure, and operate the standalone `artex-mcp` companion server for OpenCode V2.

---

## 1. Overview & Architecture

The `artex-mcp` binary is a dedicated Model Context Protocol (MCP) companion that runs as a local stdio process. It bridges OpenCode to an ARTEX backend without importing or modifying backend or database code.

```
┌─────────────────┐       stdio (JSON-RPC)       ┌─────────────┐       HTTP (REST)       ┌───────────────┐
│   OpenCode V2   │ ◄──────────────────────────► │  artex-mcp  │ ◄─────────────────────► │ ARTEX Backend │
│  (Client / UI)  │   legacy MCP (SDK v1.7.0)    │ (Companion) │   GET / PUT / POST      │    (:8787)    │
└─────────────────┘                              └─────────────┘                         └───────────────┘
                                                        │
                                                        ▼
                                                 Local File Audit
                                                 (audit.jsonl, 0600)
```

### Safety & Guardrails
- **Read-Only by Default**: On startup with `ARTEX_MCP_WRITE_ENABLED=false` (or unset), only seven read tools are exposed. Preview and apply tools are completely absent from the MCP catalog.
- **Two-Phase Commit**: No mutation occurs directly. An explicit preview (`preview_*`) must be created first, returning an immutable 300-second expiring `change_id`.
- **One-Shot Apply with Read-Back**: Applying a change dispatches exactly one mutation, re-reads state, records audit intent and outcome, and invalidates the preview.
- **Client-Side OpenCode Approval**: `apply_change` is gated with `"effect": "ask"` in OpenCode configuration.
- **Strict Isolation**: No task execution or background attack workflows can be initiated through this companion.

---

## 2. Installation & Build

Compile the companion binary to an absolute path of your choice:

```bash
# Build binary
go build -o /usr/local/bin/artex-mcp ./cmd/artex-mcp

# Verify permissions
chmod 0755 /usr/local/bin/artex-mcp
```

Or for user-local installations:
```bash
mkdir -p "$HOME/.local/bin"
go build -o "$HOME/.local/bin/artex-mcp" ./cmd/artex-mcp
chmod 0755 "$HOME/.local/bin/artex-mcp"
```

---

## 3. Authentication & Token Provisioning

`artex-mcp` authenticates to the ARTEX REST API using an administrative JWT bearer token read from a local file.

### Step 1: Obtain a Token
Obtain an admin JWT via the ARTEX login endpoint:
```bash
curl -s -X POST http://10.10.23.5:8787/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"ARTEX","password":"<your_admin_password>"}' | jq -r .token > ~/.config/artex/token.jwt
```

### Step 2: Secure the Token File
Ensure the token file is readable only by the current user:
```bash
chmod 0600 /path/to/artex.jwt
```
The companion verifies file permissions and rejects world- or group-readable token files.

### Token Renewal
JWT tokens expire according to ARTEX configuration. When expired, repeat Step 1 to refresh the token file; the companion reads the file on every request.

---

## 4. Network & SSH Tunneling

If ARTEX runs on a remote host (e.g., remote server or VPS on `:8787`), route traffic through a secure SSH tunnel:

```bash
# Forward local port 18787 to remote ARTEX :8787
ssh -N -L 18787:127.0.0.1:8787 user@remote-host
```

Then set:
```bash
export ARTEX_BASE_URL="http://127.0.0.1:18787"
```

---

## 5. Audit Logging

When write mode is enabled (`ARTEX_MCP_WRITE_ENABLED=true`), `ARTEX_MCP_AUDIT_FILE` is mandatory.
Audit records contain structured JSON lines capturing:
- Timestamp (RFC3339Nano)
- Operation type (`setting`, `agent_config`, `agent_prompt`, `profile_activate`, `task_control`)
- Target identifier
- Action phase (`intent` before dispatch, `outcome` after dispatch)
- SHA-256 fingerprints of current and proposed states
- OpenCode session ID (`ai.opencode/sessionID` if provided by client)
- Final verified outcome (`verified`, `partial_or_failed`, `outcome_unknown`)

Raw prompt templates, auth tokens, and sensitive strings are excluded from audit records.

---

## 6. OpenCode V2 Configuration

OpenCode V2 supports declarative MCP server definitions and fine-grained permissions.

### Check OpenCode Version
Ensure OpenCode V2 is installed:
```bash
opencode --version
```
*Note: V2 configuration syntax is not backward-compatible with V1.*

### Configure Server and Permissions
Add the server entry to your project or user OpenCode config (`~/.config/opencode/opencode.json` or `opencode.jsonc`). A complete reference template is available at `examples/opencode/artex-control.jsonc`:

```jsonc
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
    // Catch-all prompt for any unlisted tool
    {"action": "artex_*", "resource": "*", "effect": "ask"},

    // Allow read-only inspections without prompting
    {"action": "artex_health", "resource": "*", "effect": "allow"},
    {"action": "artex_get_settings", "resource": "*", "effect": "allow"},
    {"action": "artex_list_profiles", "resource": "*", "effect": "allow"},
    {"action": "artex_list_agents", "resource": "*", "effect": "allow"},
    {"action": "artex_get_agent", "resource": "*", "effect": "allow"},
    {"action": "artex_list_tasks", "resource": "*", "effect": "allow"},
    {"action": "artex_get_task", "resource": "*", "effect": "allow"},

    // Allow generating mutation previews without prompting
    {"action": "artex_preview_setting_change", "resource": "*", "effect": "allow"},
    {"action": "artex_preview_agent_change", "resource": "*", "effect": "allow"},
    {"action": "artex_preview_prompt_change", "resource": "*", "effect": "allow"},
    {"action": "artex_preview_profile_activation", "resource": "*", "effect": "allow"},
    {"action": "artex_preview_task_control", "resource": "*", "effect": "allow"},

    // Require operator confirmation on every actual mutation
    {"action": "artex_apply_change", "resource": "*", "effect": "ask"}
  ]
}
```

---

## 7. Tool Catalog

### Read Tools (Always Available)
| Tool | Description |
|---|---|
| `health` | Read backend status and version (`GET /api/health`). |
| `get_settings` | Get supported system settings (concurrency fields are read-only). |
| `list_profiles` | List LLM profiles without leaking provider credentials. |
| `list_agents` | List registered agent keys and names. |
| `get_agent` | Read agent configuration and current system prompt. |
| `list_tasks` | List background pentest tasks sorted numerically with local paging. |
| `get_task` | Inspect task execution status and paused/queued indicators. |

### Preview & Mutation Tools (Write Mode Only)
| Tool | Description |
|---|---|
| `preview_setting_change` | Generate preview for updating an allowed setting (`key`, `value`). |
| `preview_agent_change` | Generate preview for updating an agent budget or LLM profile binding. |
| `preview_prompt_change` | Generate preview for updating an agent prompt template and note. |
| `preview_profile_activation`| Generate preview for activating an existing LLM profile. |
| `preview_task_control` | Generate preview for pausing or resuming an existing task. |
| `apply_change` | Apply a previously generated `change_id` after operator approval. |

---

## 8. Known Limitations & Constraints

1. **Admin Credentials**: The companion requires an admin-scoped JWT; ARTEX does not currently feature granular RBAC for the REST API.
2. **Manual Token Renewal**: When the JWT expires, the operator must re-login and update the token file.
3. **No Global CAS**: The backend does not support compare-and-swap headers. The companion verifies fingerprint match before mutation, but concurrent out-of-band writers can race.
4. **Local Audit Log**: Audit logs are written to a local file descriptor (mode 0600) and are not externally notarized.
5. **No Profile Authoring**: LLM profile creation and credential editing are intentionally omitted from v1.
