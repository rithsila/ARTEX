---
name: scopesentry
description: Use when managing or querying the headless ScopeSentry security scanning platform via MCP or the standalone scopesentry CLI (projects, scan tasks, templates, asset queries, direct tool execution).
mcps: ScopeSentry
---

# ScopeSentry MCP & CLI Guide

ScopeSentry is a headless, distributed asset mapping and vulnerability scanning platform. All Web UI and browser components have been removed in favor of a native Model Context Protocol (MCP) server, an administrative CLI (`ScopeSentry`), and a lightweight standalone Go CLI client (`scopesentry`).

In ARTEX, grant this skill separately to Auto (standalone Chat), Main Agent (task
console), and any Worker that needs it. Loading `Skill(scopesentry)` unlocks the
enabled `ScopeSentry` MCP's deferred tools; discover schemas via `SearchExtraTools`
and invoke via `ExecuteExtraTool`. Visibility alone is not loading. If the skill is
visible, its MCP tools stay gated until load, even with direct MCP visibility.
No permission grants or visibility migrations happen automatically.
See `SCOPESENTRY_GUIDE.md` for setup and the separate asset import tool.

## Safety: adding targets to an existing project

`create_project` replaces scope for an existing name. Before an “add target” request:

1. Select the exact project ID (`list_projects_data`); read `get_project` for that ID.
2. Read the **COMPLETE current `project.target`**, not a list summary or truncated
   response. If the read fails, is missing, truncated, or incomplete, **STOP**. Do not
   call `create_project`; ask for a complete read instead.
3. Merge all existing targets with new targets, preserving the old scope. Pass the
   existing project name and full merged target string to `create_project`. Never
   pass only the new target for an existing name. Use actual newline characters.
4. Read `get_project` again to verify. Avoid concurrent scope edits: this API has no
   compare-and-swap protection.

Example: old `example.com` + `192.0.2.0/24`, adding `third.example`, must send
`"target":"example.com\n192.0.2.0/24\nthird.example"` in JSON. JSON `\n` decodes to real
newlines; shell double quotes do not. Checked transcript:
`integration/scopesentry/testdata/workflow.json` (NOT an LLM obedience test).

---

## 1. Preparation & Credentials

### 1.1 Health & Service Endpoints

- **Liveness check**: `GET /live` returns HTTP 200 JSON (`{"status":"ok"}`).
- **Readiness check**: `GET /ready` returns HTTP 200 JSON on successful MongoDB and Redis ping (`{"status":"ready"}`).
- **MCP endpoint**: `POST /mcp` (default port: `8082`, e.g. `http://<host>:8082/mcp`).

### 1.2 Headless Admin Commands

All administration is performed via the headless server binary `./ScopeSentry`:

- **Create API Key**:
  ```bash
  ./ScopeSentry apikey create --name <key-name>
  ```
  Generates and returns an API key (e.g. `ssk_...`). Save this key (**shown only once**).
- **List API Keys**:
  ```bash
  ./ScopeSentry apikey list
  ```
- **Delete API Key**:
  ```bash
  ./ScopeSentry apikey delete --id <key-id>
  ```
- **Get Plugin Key**:
  ```bash
  ./ScopeSentry pluginkey
  ```

### 1.3 Configure MCP Clients (Cursor, Claude Desktop, etc.)

Add to your MCP configuration (`cursor-settings` or `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "scopesentry": {
      "url": "http://<your-host>:8082/mcp",
      "headers": {
        "X-API-Key": "ssk_your-key"
      }
    }
  }
}
```

Authentication supports either `X-API-Key: ssk_...` or `Authorization: Bearer ssk_...`.

### 1.4 Configure in ARTEX (Data Source & Skill Grants)

1. **Data Source Setup**:
   - In ARTEX Settings / Asset Sync, set **MCP Endpoint**: `http://<host>:8082/mcp`.
   - Set **API Key**: `ssk_...` (in the `X-API-Key` field).
   - Click **Save & Enable**.
   - If previously connected and showing 17 tools, click the **Refresh** button in the top-right corner of the card to re-fetch the live tool catalog (discovers all **19** tools including `stop_scan_task` and `resume_scan_task`).
2. **Grant Skill to Agents**:
   - Under Agent Settings, enable the `scopesentry` skill for **Auto** (Chat) and/or **Main Agent**.
3. **Safe Verification Prompt**:
   - In Chat, run:
     > Load the scopesentry skill. Discover the live tools. List projects, online nodes, and scan templates. Show IDs and names. Do not create or change anything or scan.

---

## 2. Standalone Go CLI Client (`scopesentry`)

`scopesentry` is a standalone Go binary connecting exclusively over `/mcp` using `github.com/modelcontextprotocol/go-sdk`. It has zero direct database dependencies (does not import MongoDB, Redis, or Gin).

### 2.1 CLI Conventions & Exit Codes

- **stdout**: Pure, machine-readable JSON. Safe to pipe directly to `jq` or downstream tools.
- **stderr**: Diagnostic messages, progress logs, and errors.
- **Exit codes**:
  - `0`: Success (`ExitSuccess`)
  - `1`: Execution error (`ExitExecutionError`)
  - `2`: Usage / flag parsing error (`ExitUsageError`)
  - `3`: Polling timeout (`ExitTimeout`)

### 2.2 CLI Configuration

Configure via CLI flags, environment variables, or config files:
- **Flags**: `--url <server_url>` (default: `http://localhost:8082`), `--api-key <api_key>`
- **Environment variables**: `SCOPESENTRY_URL`, `SCOPESENTRY_API_KEY`
- **Config files**: `.scopesentry.yaml` or `~/.config/scopesentry/config.yaml`

### 2.3 CLI Command Reference

#### Health Diagnostics
```bash
scopesentry doctor
```
Checks `/live`, `/ready`, and connects to `/mcp` to report total available tools.

#### Projects
```bash
# List projects (grouped by tag or paged table)
scopesentry project list
scopesentry project list --by-tag
scopesentry project list --search "internal" --page 1 --size 20

# Get project details
scopesentry project get <project-id>

# Create or update project (replaces scope targets if project name already exists)
scopesentry project create --name "Corp-Infra" --target $'example.com\n192.0.2.0/24' --tag "Prod"
```

#### Scan Tasks
```bash
# List tasks
scopesentry task list
scopesentry task list --search "recon"

# Get task details
scopesentry task get <task-id>

# Create scan task
scopesentry task create --name "recon-task" --node "node-1" --template "<template-id>" --target-source general --target "example.com"

# Wait / poll for task completion
scopesentry task wait <task-id> --timeout 10m --interval 5s

# Request acceptance, not scanner acknowledgement; inspect task status afterwards
scopesentry task stop <task-id>
scopesentry task resume <task-id>
```

#### Assets & Vulnerabilities
```bash
# List assets with DSL search
scopesentry asset list --type subdomain --search 'domain=^api. && port==443' --page 1 --size 20

# Count assets directly without fetching rows
scopesentry asset count --type vulnerability --search 'level==high'

# Get asset detail
scopesentry asset get --type asset --id <asset-id>

# Tag an asset
scopesentry asset add-tag --type asset --id <asset-id> --tag "triage"
```

#### Direct Tool Calling
```bash
# Discover live MCP tools (19 after stop/resume; older deployments differ)
scopesentry tools list

# Describe a specific tool schema
scopesentry tools describe create_scan_task

# Execute tool with JSON arguments
scopesentry tools call list_projects_data --input '{"search":"","pageIndex":1,"pageSize":10}'
scopesentry tools call create_scan_task --input-file ./task.json
```

---

## 3. MCP Tool Overview

The current source exposes 19 tools with English schemas. Always discover the live
list instead of requiring an exact count; older deployments may lack lifecycle tools.

| Tool | Purpose |
| --- | --- |
| `list_projects` | Project tree grouped by tag (includes project ID) |
| `list_projects_data` | Paged project list, searchable by name |
| `get_project` | Project detail |
| `create_project` | Create project (replaces target scope if name exists) |
| `list_tasks` | Scan task list |
| `get_task` | Task detail |
| `list_scan_templates` | Scan template list |
| `get_scan_template` | Template detail |
| `list_plugin_modules` | Scan pipeline module names |
| `list_plugins` | Available plugins (with hash and default parameters) |
| `create_scan_template` | Create scan template |
| `create_scan_task` | Create scan task |
| `stop_scan_task` | Request stop for one task: `{"id":"<task ObjectID>"}` |
| `resume_scan_task` | Request resume for one task: `{"id":"<task ObjectID>"}` |
| `list_assets` | Query assets of each type (paged list) |
| `count_assets` | Count assets matching query without fetching rows |
| `get_asset_detail` | Asset or vulnerability detail |
| `add_asset_tag` | Add a tag to an asset |
| `list_nodes` | Scan node list |

*Scope replacement behavior*: `create_project` requires `name` and valid `target`.
For existing names it replaces scope. Follow the complete-read/merge/verify procedure
above; stop on failed or incomplete reads.

---

## 4. Common Workflows

### 4.1 Query Assets by Project

When a project context exists, pass `filter.project` with the project **ObjectID** to narrow scope and maximize performance.

```json
{
  "asset_type": "asset",
  "pageIndex": 1,
  "pageSize": 20,
  "search": "domain=^example.com",
  "filter": {
    "project": ["<project ObjectID>"]
  }
}
```

### 4.2 Create a Scan Task

1. `list_nodes` to get online worker node names.
2. `list_scan_templates` to get the desired template **ObjectID**.
3. `create_scan_task` with `name`, `node`, `template`, and `targetSource`.

Validate the template ObjectID with `get_scan_template` and choose online node names
before creating a task. `allNode` resolves online nodes only. MCP accepts one-shot
scans only: recurring/scheduled requests are rejected, not silently run once. REST
scheduling remains separate. Stop/resume return acceptance, not worker acknowledgement;
use `get_task` for subsequent status. Do not start a scan without approved scope.

**Target sources (`targetSource`):**

| targetSource | Description | Required parameters |
| --- | --- | --- |
| `general` | Direct target entry | `target` (actual newlines for multiple targets) |
| `project` | Read targets from an existing project | `project` (array of project ObjectIDs) |
| `asset` | Search asset inventory | `search`; optional `project`, `filter`, `targetNumber` |
| `RootDomain` | Search root domain inventory | `search`; optional `project`, `filter`, `targetNumber` |
| `subdomain` | Search subdomain inventory | `search`; optional `project`, `filter`, `targetNumber` |
| `UrlScan` | Search URL scan results | `search`; optional `project`, `filter`, `targetNumber` |
| `*Source` (e.g. `subdomainSource`) | Create from previous asset query results | `search` when `targetTp=search`; `targetIds` when `targetTp=select` |

**Example — scan root domains directly:**
```json
{
  "name": "recon-subdomains",
  "node": ["node-1"],
  "template": "<template ObjectID>",
  "targetSource": "general",
  "target": "example.com\nfoo.com",
  "project": ["<project ObjectID>"]
}
```

**Example — downstream scan from subdomain inventory:**
```json
{
  "name": "recon-ports-and-vulns",
  "node": ["node-1"],
  "template": "<downstream template ObjectID>",
  "targetSource": "subdomain",
  "search": "task==\"recon-subdomains\"",
  "project": ["<project ObjectID>"]
}
```

### 4.3 Two-Phase Root Domain Recon

For root domains, avoid running all scan modules in a single pass. Distributed tasks are dispatched per single target; running all modules at once overburdens the single node assigned to the root domain.

**Recommended two-phase pattern:**

1. **Phase 1 — Subdomain Collection Only**:
   - `targetSource`: `general`
   - `target`: Root domains (multiline)
   - Template: Enable only `SubdomainScan` and `SubdomainSecurity`
   - Use `get_task` or `scopesentry task wait <task-id>` to wait for completion.
2. **Phase 2 — Downstream Modules (Ports, Fingerprints, Vulns)**:
   - `targetSource`: `subdomain`
   - `search`: `task=="<Phase 1 task name>"`
   - Template: Enable `PortScan`, `AssetMapping`, `VulnerabilityScan`, etc.
   - Harvested subdomains are distributed across all nodes as independent targets.

---

## 5. Asset Queries (`list_assets` / `count_assets`)

`count_assets` and `list_assets` accept the same `asset_type`, `search`, and `filter`. Use `count_assets` to retrieve `{ "total": N }` before paging `list_assets`.

### 5.1 Asset Types (`asset_type`)

- Core types: `asset`, `subdomain`, `IPAsset`, `vulnerability`, `SensitiveResult`
- Additional types: `RootDomain`, `app`, `mp`, `UrlScan`, `DirScanResult`, `crawler`, `PageMonitoring`, `SubdomainTakerResult`
- Aliases supported: `web` → `asset`, `vuln` → `vulnerability`, `ip` → `IPAsset`, `url` → `UrlScan`

### 5.2 Search DSL

Custom DSL expressions evaluated against MongoDB:

| Operator | Meaning | Index Usage | Example |
| --- | --- | --- | --- |
| `==` | Exact equality match | **Uses index** | `port==443` |
| `^` | Prefix match | **Uses index** | `domain=^example.com` |
| `=` | Fuzzy match (regex) | Does not use index | `domain=example` |
| `!=` | Exclude | — | `port!="80"` |
| `&&` | Logical AND | — | `domain==example.com && port==443` |
| `||` | Logical OR | — | `title=admin || body=login` |

*Fields common to all asset types*: `tag`, `task` (task name), `rootDomain`.
Results use the task **name**, not the ID from `create_scan_task`. Use
`filter.task: ["<task name>"]` and `filter.project: ["<project ObjectID>"]` together
to narrow results. `get_task` and stop/resume use task **IDs**.

*Search fields by asset type*:
- `asset`: domain, ip, port, service, app, title, statuscode, icon, banner, type, body, header
- `subdomain`: domain, ip, type, value
- `IPAsset`: ip, domain, port, service, webServer, app
- `vulnerability`: url, vulname, matched, request, response, level
- `SensitiveResult`: url, sname, body, info, md5
- `RootDomain`: domain, icp, company
- `UrlScan`: url, input, source, resultId, type
- `DirScanResult`: url, statuscode, redirect, length
- `crawler`: url, method, body, resultId
- `PageMonitoring`: url, hash, diff, response
- `SubdomainTakerResult`: domain, value, type, response

*Important*: Do not put `project` in the `search` string. Use `filter.project` instead.

### 5.3 Exact Filter (`filter`)

A JSON map of string arrays: values under one key are OR'd; different keys are AND'd.
Use only fields supported by the selected asset collection; this is not a universal
set of filters. Inspect live schemas and use the DSL for type-specific fields.

| filter key | Meaning | Value format |
| --- | --- | --- |
| `project` | Owning project | MongoDB **ObjectID** (array) |
| `task` | Source task | Task names (array, not IDs) |
| `port` | Port number | e.g. `["443"]` |
| `service` | Service / protocol | e.g. `["https"]` |
| `app` | Fingerprint | e.g. `["Nginx"]` |
| `level` | Vulnerability severity | e.g. `["high"]` |
| `statuscode`| HTTP status code | e.g. `["200"]` |
| `status` | Finding handling status / URL status | array of strings |
| `sname` | Sensitive rule name | array of strings |
| `tags` | Assigned tags | array of strings |

**Filter example:**
```json
{
  "project": ["<project ObjectID>"],
  "port": ["443", "8443"]
}
```

### 5.4 Sorting (`sort`)

Supported only on `UrlScan` and `DirScanResult`:
```json
{
  "length": "ascending"
}
```

---

## 6. Scan Template Module Names

`TargetHandler`, `SubdomainScan`, `SubdomainSecurity`, `PortScanPreparation`, `PortScan`, `PortFingerprint`, `AssetMapping`, `AssetHandle`, `URLScan`, `WebCrawler`, `URLSecurity`, `DirScan`, `VulnerabilityScan`, `PassiveScan`.

---

## 7. Troubleshooting

| Symptom | Cause | Solution |
| --- | --- | --- |
| MCP connection refused | ScopeSentry server not running | Check `./ScopeSentry` process and `/live` / `/ready` endpoints |
| 401 / 403 Unauthorized | Missing or invalid API key | Recreate key via `./ScopeSentry apikey create --name <name>` |
| Assets not found | Project specified by name or in search | Use `filter.project` with project ObjectID |
| Task/template creation failed | Invalid template or offline node | Query `list_scan_templates` and `list_nodes` for valid IDs and names |
| Query slow or timed out | Unindexed `=` fuzzy queries on large datasets | Use `==` or `^` prefix on indexed fields; supply `filter.project` |
