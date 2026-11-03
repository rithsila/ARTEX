# ScopeSentry in ARTEX

Use one enabled HTTP MCP connection named **`ScopeSentry`**. It powers direct tools
and one-way asset import. Terminal access is optional and uses a separate CLI binary
and separate credential configuration. Chat is a tool interface, not another server.
Nothing below has been deployed or run against a live scanner by this change.

## 1. Connect and grant access (operator steps)

1. In ARTEX Asset Sync, configure the `ScopeSentry` data source with its full MCP URL,
   for example `http://scopesentry:8082/mcp`, and an API key in the `X-API-Key` header.
   Use an operator-provided key; do not print it in chat, Git, transcripts, or logs.
   Check connectivity and refresh tool discovery. Current source has **19** tools;
   older servers may have 17. Discover names and schemas; do not require an exact count.
2. In agent settings, make the `scopesentry` skill visible to **Auto** for standalone
   Chat. Grant it separately to **Main Agent** (`mainagent`) for the task console.
   A Worker needs its own grant only if it should operate ScopeSentry. Planner and
   custom agent grants are also independent. No defaults or existing visibility rows
   are widened by this change.
3. Start a new chat turn/run and ask it to load `Skill(scopesentry)`. The skill's
   `mcps: ScopeSentry` links the exact existing MCP name. When visible, it makes these
   tools skill-gated: schemas are deferred and calls stay locked until skill load,
   even if direct MCP visibility is also enabled. After load, use `SearchExtraTools`
   to discover schemas and `ExecuteExtraTool` to invoke the revealed tools.
4. For direct MCP access without this skill, explicitly grant the MCP to the chosen
   agent. It becomes globally deferred/discoverable from run start if no visible
   skill gates it. Enabling a server alone does not grant agent visibility.
5. For chat asset import, bind the separate host tool `sync_scopesentry_assets` in the
   Tools catalog to each intended agent. It seeds with **no bindings**, including
   no Auto binding. The skill/MCP grant does not grant this host tool. It imports
   assets into ARTEX but never creates a scan. Existing tool guards still apply.

Read-only first prompt:

> Load the scopesentry skill. Discover the live tools. List projects, online nodes,
> and scan templates. Show IDs and names. Do not create or change anything or scan.

If no tools appear, check exact server name, enabled state, URL/header, per-agent
visibility, skill loading, deferred discovery, and a new turn after settings changes.
Do not fix a missing tool by granting every tool to every agent.

## 2. Optional terminal CLI

Run these build-only commands from the ARTEX checkout. They use the local sibling
checkout (override `--source /path/to/ScopeSentry`) without downloading code/modules,
reading config, or copying credentials. A local Go compiler and cached dependencies
are required. `GOTOOLCHAIN=local`, `GOPROXY=off`, `GOSUMDB=off`, `-mod=readonly`, and
`-trimpath` prevent automatic fetches/module edits and remove local build paths.

```bash
# Native: tools/bin/scopesentry-<host-os>-<host-arch>
bash scripts/build-scopesentry-cli.sh --os native --arch native

# Docker: choose the actual deployment architecture, not the laptop architecture
bash scripts/build-scopesentry-cli.sh --os linux --arch amd64
bash scripts/build-scopesentry-cli.sh --os linux --arch arm64
```

Only Linux/macOS and amd64/arm64 are accepted. `tools/bin/` ignores generated files;
`.gitignore` keeps the directory present in clean Docker build contexts. An optional
`--output-dir PATH` is useful for temporary verification; use the default directory
for Docker. Install the native artifact in an operator-selected PATH directory as
`scopesentry` (do not execute a Linux artifact on macOS).

Docker's existing ARTEX precompiled `dist/${TARGETARCH}/artex` remains required.
For a future operator build, add `--build-arg INSTALL_SCOPESENTRY_CLI=true` to the
usual `docker build` command, matching the image platform (`linux/amd64` or
`linux/arm64`). The Dockerfile selects only `scopesentry-linux-${TARGETARCH}`.
The option defaults to false. A missing matching artifact logs a notice and preserves
MCP-only builds; it never falls back to a native/wrong-architecture artifact. Check
the installed CLI afterwards if terminal access is required. No Docker build or
deployment was performed here.

### Credentials and URLs are separate

| Interface | URL | Credentials |
| --- | --- | --- |
| ARTEX direct MCP and sync | Full endpoint ending in `/mcp` | Stored MCP HTTP headers |
| CLI in ARTEX terminal | Service **base URL**, e.g. `http://scopesentry:8082` | `SCOPESENTRY_API_KEY` or a protected CLI config file |

CLI precedence: flags > environment > `.scopesentry.yaml` in the working directory >
`~/.config/scopesentry/config.yaml` > defaults. File fields are `server_url`, `api_key`,
and optional duration `timeout`. Use protected files outside Git (mode 0600); do not
put project-local CLI credentials into an unignored file. `--api-key` is supported
but exposes secrets in process arguments/history; prefer the operator-managed secret
environment/file. ARTEX's saved MCP headers are **not** exported to a shell. This
integration never reads a DB key into environment variables. Configure credentials
in the actual terminal execution environment; host env is not automatically container env.

After the operator configures secrets, these commands are read-only:

```bash
scopesentry --url http://scopesentry:8082 doctor
scopesentry tools list
scopesentry tools describe create_scan_task
scopesentry project list
scopesentry tools call list_assets --input '{"asset_type":"subdomain","pageIndex":1,"pageSize":10}'
```

stdout is JSON; stderr is progress/errors. `--input-file ./request.json` is supported
for direct calls. Global URL/key flags must precede the command. The CLI has no DB
dependency and must not be confused with the headless server/admin binary.

## 3. Safe project → scan → results workflow (live steps; NOT run here)

1. Select the exact project ID with `list_projects_data`, then `get_project`. To add
   targets, read the **COMPLETE current `project.target`**. Stop if failed, missing,
   truncated, or incomplete. Merge old targets with additions and send the full scope
   using the existing name to `create_project`; never send only the new target for an
   existing name. Read it back to verify. Preserve existing metadata you intend to
   retain. Avoid concurrent edits: there is no compare-and-swap guarantee.
2. Discover `list_nodes`, `list_scan_templates`, and `get_scan_template`. Select only
   online node names and an existing template ObjectID. Review target scope/template
   with the operator. MCP accepts one-shot scans only; recurring/scheduled requests
   are rejected. `allNode` means online nodes only. REST scheduling is unchanged.
3. Only after explicit scan approval, call `create_scan_task` and keep the returned
   task **ID** plus the chosen unique **name**. Poll `get_task` by ID. A stop request
   `stop_scan_task({"id":"<task ObjectID>"})` or resume request
   `resume_scan_task({"id":"<task ObjectID>"})` accepts exactly one ID. CLI equivalents
   are `scopesentry task stop <task-id>` and `scopesentry task resume <task-id>`.
   Acceptance is not acknowledgement by scanner nodes; check status afterwards.
4. Query results with `list_assets` / `count_assets`: `filter.project` is an array of
   project **ObjectIDs**, `filter.task` is an array of task **names**, not IDs. Other
   exact filters are string arrays supported by that asset collection (such as port,
   service, app, level). Do not put project in the search DSL. Use live schemas;
   vulnerability and other result collections can be queried but are not synced.
5. Sync is a **separate write action**, not automatic after a scan. Use Asset Sync or
   the explicitly bound host tool. `dimension: "task"` takes task names; `"project"`
   takes project ObjectIDs. `asset_types` defaults to subdomain/service/app; service
   maps to ScopeSentry's `asset`. `create_company` applies to project roots only.
   `page_size` accepts 1–500, otherwise 100. Cap: 5,000 per target/type. Check `synced`,
   `companies`, `warnings`, and `errors` for partial failure/truncation. Related IPs
   are imported; vulnerability findings are not. Cancellation stops further work,
   not rows already written or a synchronous store operation already in progress.

Shell multi-target example for a **new** project (Bash/Zsh actual newlines):

```bash
scopesentry project create --name NewExample --target $'example.com\n192.0.2.0/24'
```

For existing projects, read and merge first. In JSON, `"target":"a.example\nb.example"`
decodes correctly; shell `"a.example\nb.example"` passes a literal backslash-n instead.

Separate import prompt (after results are reviewed):

> Import subdomains, services and apps from ScopeSentry task named recon-example
> into ARTEX using sync_scopesentry_assets. Do not scan. Report counts and warnings.

## 4. Local evidence and limits

`integration/scopesentry/testdata/workflow.json` is a checked illustrative transcript:
select project → read full scope → preserve two old targets while adding a third →
discover/validate node and template → create scan → inspect task → filter results by
task name and project ObjectID → separate sync. Failure/incomplete-read branches stop
without a replacement call. Its annotations are not promised MCP response fields.
Tests check these payload contracts; they **do not execute an MCP workflow or test
LLM obedience**. The full read/merge rule is skill guidance, not runtime enforcement.

Safe local tests:

```bash
go test ./integration/scopesentry -count=1
go test server/sync_scopesentry_operation.go server/sync_scopesentry_ingest.go server/sync_scopesentry_operation_test.go -count=1
```

Server/agent/db package tests have `TestMain` database setup/cleanup. Do not run them,
even with `-run`, without an explicit dedicated test DSN. Live UI permissions,
credentials, running-server tool discovery, scanner acknowledgements, DB ingestion,
Docker execution, and LLM behavior require separate operator verification.
