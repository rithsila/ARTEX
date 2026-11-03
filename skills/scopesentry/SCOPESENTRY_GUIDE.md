# ScopeSentry Headless MCP Complete Guide

## 1. Quick Overview

ScopeSentry is a headless, distributed Attack Surface Management (ASM) and vulnerability scanning platform. All Web UI and browser components have been removed in favor of a native Model Context Protocol (MCP) server, an administrative CLI (`ScopeSentry`), and a standalone Go CLI client (`scopesentry`).

### Connection & Service Endpoints
- **MCP Endpoint**: `POST /mcp` (e.g. `http://<host>:8082/mcp`)
- **Authentication**: `X-API-Key: ssk_...` or `Authorization: Bearer ssk_...`
- **Liveness Endpoint**: `GET /live` returns HTTP 200 JSON (`{"status":"ok"}`)
- **Readiness Endpoint**: `GET /ready` returns HTTP 200 JSON on MongoDB and Redis connectivity ping (`{"status":"ready"}`)
- **Output Format**: JSON-RPC 2.0 text blocks (`{ "type": "text", "text": "..." }`)

### Headless Admin Credentials
No browser or Web UI is required to generate or manage credentials:
- **Create API Key**: `./ScopeSentry apikey create --name <key-name>`
- **List API Keys**: `./ScopeSentry apikey list`
- **Delete API Key**: `./ScopeSentry apikey delete --id <key-id>`
- **Get Plugin Key**: `./ScopeSentry pluginkey`

---

## 2. Architecture & Communication Flow

```
+------------------------------------+        HTTP POST /mcp        +----------------------+
| AI Agent / Client                  | ---------------------------> | ScopeSentry Backend  |
| - Standalone CLI (`scopesentry`)   |    Headers: X-API-Key: ssk_  | - Native MCP Server  |
| - Cursor / Claude Desktop / ARTEX  | <--------------------------- | - MongoDB & Redis    |
+------------------------------------+      JSON-RPC Result Text     | - Scanning Nodes     |
                                                                    +----------------------+
```

1. **Client Dispatches Tool Call**:
   - Sends JSON-RPC `tools/call` with tool name and argument JSON object.
2. **Authentication Verification**:
   - Backend verifies the `X-API-Key` or `Bearer` token against the database.
3. **Execution & Data Retrieval**:
   - Queries MongoDB collections (assets, subdomains, tasks, projects, vulnerabilities) or dispatches tasks to scanner nodes.
4. **Structured JSON Response**:
   - Returns structured stringified JSON payload inside MCP text content blocks.

---

## 3. Client Setup & Usage

### Option A: Standalone Go CLI Client (`scopesentry`)
The standalone Go client communicates exclusively over `/mcp` with zero local database dependencies.

#### CLI Output Conventions
- **stdout**: Pure, machine-readable JSON (safe for `jq` and downstream pipes).
- **stderr**: Diagnostic messages, progress indicators, and errors.
- **Exit codes**:
  - `0`: Success (`ExitSuccess`)
  - `1`: Execution error (`ExitExecutionError`)
  - `2`: Usage / flag parsing error (`ExitUsageError`)
  - `3`: Polling timeout (`ExitTimeout`)

#### Core Commands
- **Health Check & Diagnostics**:
  ```bash
  scopesentry doctor
  ```
- **Project Management**:
  ```bash
  scopesentry project list
  scopesentry project create --name "Internal-Net" --target "10.0.0.0/8\ncorp.local"
  scopesentry project get <project-id>
  ```
- **Task Management**:
  ```bash
  scopesentry task list
  scopesentry task create --name "vuln-scan" --node "node-1" --template "<template-id>" --target-source general --target "example.com"
  scopesentry task wait <task-id> --timeout 10m --interval 5s
  ```
- **Asset Queries**:
  ```bash
  scopesentry asset list --type subdomain --search 'domain=^api. && port==443'
  scopesentry asset count --type vulnerability --search 'level==high'
  scopesentry asset get --type asset --id <asset-id>
  ```
- **Direct Tool Calling**:
  ```bash
  scopesentry tools list
  scopesentry tools describe <tool-name>
  scopesentry tools call <tool-name> --args '<json>'
  ```

### Option B: Using in Cursor or Claude Desktop
Add to your `cursor-settings` or `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "scopesentry": {
      "url": "http://<YOUR_SCOPESENTRY_IP>:8082/mcp",
      "headers": {
        "X-API-Key": "ssk_your_api_key_here"
      }
    }
  }
}
```

### Option C: Integration in ARTEX
1. Configure Data Source with URL `http://<host>:8082/mcp`.
2. Supply the API key generated headlessly via `./ScopeSentry apikey create --name artex`.
3. ARTEX performs MCP handshake over `/mcp` and binds the 17 security tools.

---

## 4. MCP Tools Reference

ScopeSentry exposes 17 core scanning and asset management tools. All tool schemas and parameter descriptions are native English.

### Project Management
- `list_projects`: Lists projects grouped by tag.
- `list_projects_data`: Paged search for projects by name.
- `get_project`: Fetch single project details (`id` required).
- `create_project`: Create a new project. *Note*: If the project name already exists and the cleaned target contains valid entries, it replaces the existing target scope.

### Asset & Target Queries
- `list_assets`: Query assets with pagination, search DSL, and exact filters.
- `count_assets`: Get exact total count for an asset query directly without fetching data rows.
- `get_asset_detail`: Fetch complete raw details for an asset or vulnerability.
- `add_asset_tag`: Attach custom metadata tags to an asset.

### Scan & Task Automation
- `list_tasks`: List all scan tasks.
- `get_task`: Check task status and progress.
- `create_scan_task`: Launch new scan tasks on target domains/assets.
- `list_nodes`: List online scanner worker nodes.

### Plugins & Templates
- `list_scan_templates`: View existing scan profiles.
- `get_scan_template`: Inspect template module configuration.
- `create_scan_template`: Build customized multi-module scan profiles.
- `list_plugin_modules`: List available engine modules (`SubdomainScan`, `PortScan`, etc.).
- `list_plugins`: List available plugins and their parameters.

---

## 5. Query Syntax & Performance Rules

### Asset Types (`asset_type`)
- `asset`: Web targets and services (domain, IP, port, title, banner).
- `subdomain`: Subdomain records (A, CNAME, etc.).
- `IPAsset`: Host IPs, open ports, and running daemons.
- `vulnerability`: Detected vulnerabilities with request/response evidence.
- `SensitiveResult`: Leaked keys, tokens, and sensitive patterns.
- `UrlScan`: Discovered URLs and paths.
- `SubdomainTakerResult`: Vulnerable DNS records eligible for takeover.
- `RootDomain`, `app`, `mp`, `crawler`, `PageMonitoring`, `DirScanResult`.

### Search DSL Rules
- `==` : Exact match (**uses MongoDB index**, fast). Example: `port==443`
- `^`  : Prefix match (**uses MongoDB index**). Example: `domain=^api.`
- `=`   : Regex fuzzy match (does not use index, slow on large datasets). Example: `title=login`
- `!=`  : Exclude match. Example: `port!="80"`
- `&&` / `||`: Boolean logic. Example: `domain=^example.com && port==80`

### Performance Best Practices
1. Always specify `filter.project: ["<PROJECT_OBJECT_ID>"]` when project context exists.
2. Use `count_assets` first to check result size before paging through `list_assets`.
3. Never put `project` inside the `search` string; use `filter.project`.
4. Prefer `==` and `^` prefix over unindexed `=` fuzzy expressions on large collections.

---

## 6. Two-Phase Scanning Workflow

When scanning root domains, do not run all modules in one heavy scan. Run two phases to optimize worker node distribution:

1. **Phase 1: Subdomain Discovery**
   - Tool: `create_scan_task`
   - Target source: `general`
   - Target: Root domains
   - Template: Only `SubdomainScan` and `SubdomainSecurity`.
   - Poll completion with `scopesentry task wait <task-id>`.
2. **Phase 2: Service & Vulnerability Scan**
   - Tool: `create_scan_task`
   - Target source: `subdomain`
   - Search: `task=="<PHASE_1_TASK_NAME>"`
   - Template: `PortScan`, `AssetMapping`, `VulnerabilityScan`.
   - Subdomains are distributed across worker nodes as independent targets.

---

## 7. Ready-to-Use Analysis Prompts

### Prompt 1: Attack Surface Triage & High-Risk Profiling
```text
You are an offensive security reconnaissance analyst.
Analyze the provided ScopeSentry asset inventory for target: {{TARGET_PROJECT}}.

Inventory Data:
{{SCOPESENTRY_ASSET_JSON}}

Tasks:
1. Identify high-value targets (Admin portals, API gateways, staging/dev environments).
2. Spot anomalies (uncommon ports like 8080, 8443, 8888, 9000, 5000; outdated web servers; internal naming schemes).
3. Check for exposed sensitive findings or takeover indicators.
4. Output a prioritized list of top 5 attack entry points with exact URLs/IPs and recommended testing steps.
```

### Prompt 2: Vulnerability & Sensitive Data Correlation
```text
You are an expert penetration tester.
Correlate the following ScopeSentry findings:
- Vulnerabilities: {{VULN_DATA}}
- Sensitive Results: {{SENSITIVE_DATA}}
- Open Services: {{SERVICE_DATA}}

Output a structured exploitation roadmap:
1. Critical & High findings that allow direct access or authentication bypass.
2. Exposed credentials/tokens and which discovered services they likely unlock.
3. Chained attack vectors (e.g. Info leak -> Service enumeration -> Exploitation).
4. Concrete reproduction steps for the top 3 highest-impact findings.
```
