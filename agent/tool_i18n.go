package agent

import "regexp"

// This file holds the English copy for the built-in tool catalog. The Go tool
// definitions keep their own descriptions; these overrides replace them at seed
// time and in the one-time database refresh, so both a fresh install and an
// existing database show English in the Tools page and to the model.

var cjkRe = regexp.MustCompile(`[\p{Han}]`)

// HasCJK reports whether s contains any Han character.
func HasCJK(s string) bool { return cjkRe.MatchString(s) }

// toolDescEN is the English description for each built-in tool key.
var toolDescEN = map[string]string{
	"add_company_scope": "Add domains, IPs, CIDRs, ICP filings, or company keywords to a company's asset scope. Domains, networks, and ICP entries auto-claim matching assets; keywords only hint the agent about scope. Company names are unique: created if missing, reused if present (scope is merged in). Put one scope value per line; the system detects root domain, URL, single IP, CIDR, ICP filing, or company keyword. Always give reason. Guards: bare TLDs and overly wide ranges are rejected (IPv4 prefix /16-/32, IPv6 /32-/128); bad lines are skipped and returned in errors.",

	"add_hint": "Attach a human or main-agent strategic hint to the exploration graph. The planner reads it on the next intent round. Prefer batch: put several hints in the hints array in one call. Returns an ids array, same length and order as hints (failed item id=0, see errors). For one hint, omit hints and pass top-level text.",

	"add_intent": "Create an exploration direction, write it into the frontier, and link it into the exploration chain. An intent is an open direction, not a fixed type. Use one summary sentence to say what to explore, verify, or exploit. Prefer batch: put several new directions in the intents array in one call. Returns ids in the same order (failed id=0, see errors). For one intent, omit intents and pass top-level summary.",

	"add_task_hint": "Inject a strategic hint into a specific task. That task's planner reads it on the next intent round. Prefer batch: put several hints in the hints array in one call (returns ids in the same order, failed id=0). For one hint, omit hints and pass top-level text.",

	"add_task_scope": "Add test scope to this task. The scope is the task authorization boundary and the denominator for asset test coverage. kind is one of: company (all assets of a company) / root_domain (whole root domain, all subdomains) / subdomain (one exact subdomain) / ip / cidr / icp / keyword. Hosts a worker touches are added automatically as exact subdomains; this tool is for widening scope on purpose: the whole root domain, the whole company, or a specific subdomain or IP. value: company name or id (company must exist); domain for root_domain/subdomain; IP or range for ip/cidr; ICP number or company keyword for icp/keyword. Always give reason for audit. Use the entries array for several items.",

	"bind_finding_traffic": "Attach verified real HTTP traffic to an already recorded finding. finding_id is the separate finding-record id, not an exploration node id. All refs in one batch succeed or fail together. Re-binding does not overwrite an existing note. Binding marks an existing report as needing an update. Do not re-probe only to fill a packet, and do not create the finding again.",

	"create_custom_tool": "Important: when you install a tool the platform does not have yet, call this to put it into the platform so agents can use it. Creates a custom tool (shell/command/script/http). shell is only a bash environment declaration: it needs key, description, and agents, and no exec or schema.",

	"create_mcp": "Create an MCP server (stdio/http/sse). After creation its tools must be granted per agent visibility.",

	"create_skill": "Create a new skill (writes SKILL.md, agentskills.io spec). name is lowercase letters, digits, and hyphens.",

	"delete_assets_by_host": "Delete assets by exact host: removes that host's domain/subdomain plus its services and endpoints. The host match is exact (lowercase, trimmed), not fuzzy and not a wildcard. Passing a root domain (example.com) also deletes its subdomains and their services and endpoints; passing a subdomain (a.example.com) or an IP deletes only that host with its services and endpoints. This is a hard delete on the global asset library (shared across tasks) and cannot be undone.",

	"expand_digest": "Expand a cold digest: returns the compact member list it folded (id/summary/state/confidence), same shape as overview recent_facts/recent_done_intents. For full detail or evidence of one member, use node_detail(member_id).",

	"get_finding_retest_context": "Read the finding evidence snapshot, retest status, extra notes, and current task constraints for this retest session. No parameters. Reads only this session.",

	"get_finding_traffic": "Read the real traffic evidence bound to a finding. Does not depend on capture being on. finding_id is the separate finding-record id from the report_finding JSON result (not the exploration node id on the first line). First call without binding_id to get the list and version. An empty list is normal: for non-HTTP findings such as TCP, or when nothing was captured, you can still write the report from text and command evidence. When bindings exist, read the body in chunks with binding_id, side (request/response), and offset. When saving, pass the version you read as evidence_version to update_finding_report; that tool still takes the exploration node id as finding_id.",

	"get_task_graph": "Read the exploration graph overview for a specific task (same data as graph_overview: asset counts, frontier, findings, coverage). Select the task with task_id.",

	"get_task_node_detail": "Read the full content of one exploration node in a specific task (finding/fact/intent/goal: summary plus detail, evidence, and PoC). id is the exploration node id (from report_finding or list_task_findings). Use this before writing a vulnerability report to get that finding's full evidence.",

	"get_task_worker_trace": "Read how one work (intent) ran inside a specific task. get_task_worker_trace(task_id, intent_id) gives the step summary stream; then pass step_ids=[...] to fetch those steps in full (at most 5 per call; extra ids are ignored).",

	"get_worker_output": "Read the final output of one intent (work) in this task or a directly related task. Results from related tasks carry source_task_id/inherited=true and are read-only. A normal finish returns the summary; a stopped or failed work returns its last output before it stopped.",

	"get_worker_trace": "Read the execution process of one intent (work). get_worker_output only gives the final result; this gives the steps. Three modes: (1) intent_id only returns a summary stream for each step (summary up to 100 characters, with step_id; an outline, not full output); (2) intent_id + q returns only step summaries that match the keyword (searches summary and full output, still returns summary only); (3) intent_id + step_ids returns the full detail of those steps, at most 5 per call, with extras listed in notice/omitted_step_ids. Typical flow: use (1) or (2) to find the step_id, then (3) for full output. Thinking steps are excluded. History from directly related tasks is supported and read-only.",

	"goal_met": "End the whole task now. Call it only when you are sure every goal is truly met and the task is finished as a whole. Reaching one goal, one flag, or one finding is not enough; use prove_goal for that. This is not a way to end the current planning round: if you have no new intent to dispatch, or you are waiting on a worker, just end the round without calling this tool. Zero intents is normal. Prefer proving goals one by one with prove_goal; goal_met bypasses that and finishes from the global view.",

	"graph_overview": "Exploration posture digest: asset counts, sites with no endpoints, frontier, findings, and hints (strategic notes from humans or the main agent that must be considered when generating intents). Call it first when planning.",

	"insert_assets": "Register new assets in bulk; one call can mix several types (see the type enum). Required fields per type: root_domain to domain; ip to ip (IPv4/IPv6, not a hostname); subdomain to domain; app to app_name; service (HTTP) to url; service (non-HTTP) to service_name plus port (ip or domain required); endpoint to url plus method. auth/technologies/params are appended and merged, not replaced. Returns {results:[{index,id,type}], errors:[{index,error}]}.",

	"kill_work": "Terminate one running intent (work). Use it to stop exploration that drifted or has no value. A terminated intent is marked stopped and is not reclaimed automatically. Check get_worker_output first to see what it is doing.",

	"list_assets": "Query the asset library with a DSL expression, or fetch directly by id/ids. Supports paging. Returns only assets inside the test scope of this task and directly related tasks. DSL: field=value fuzzy (ILIKE) | field==value exact | field!=value exclude | numeric fields support > >= < <= | a bare word is a full-text fuzzy match; AND/OR combine (AND binds tighter), parentheses group. Asset type goes in the separate type parameter, not in the DSL. When id/ids are absent, dsl must be non-empty (unconditional full dumps are not allowed). Fields: domain (root/sub/service domain), root_domain, ip, url, page_title, icp, service_name, app_name, method (GET/POST), service_type (http|other), record_type (A/CNAME), technology (array; = fuzzy, == exact), port/status_code/company_id (integers). Example: status_code>=400 AND technology=shiro ; (port==80 OR port==443) AND technology=nginx",

	"list_companies": "List companies in the asset library with their asset scope and owned asset counts. Use it to see which companies exist and to get company_id (for linking apps in insert_assets or filtering list_assets by company_id). Optional search filters by company name, case-insensitive; empty returns all.",

	"list_facts": "Page through exploration facts and conclusions for this task and directly related tasks, newest first (compact: id, summary, state; long summaries are truncated, use node_detail(id) for full text). All parameters are optional: limit (default 20, max 100), before (cursor: pass next_before from the previous page for older rows; omit or 0 for the newest page), q (filter by summary keyword). Returns {facts, total, has_more, next_before}: total is the filtered count; when has_more=true use next_before to continue. Rows from related tasks carry source_task_id/inherited=true and are read-only. For findings use list_findings.",

	"list_findings": "List confirmed findings for this task and directly related tasks (compact: id, task_id, intent_id, vulnclass, severity, summary, status). Rows from related tasks carry source_task_id/inherited=true and are read-only. Findings only; use list_facts for normal exploration facts and node_detail(id) for detail.",

	"list_goals": "List the goal nodes of this task and their status (open/met). Use it to judge whether the goal is met.",

	"list_llm_profiles": "List available LLM profiles: id, name, model, format, and whether each is the active profile. Pass an id as llm_profile_id to spawn_task to give a child task its own LLM (a cheap model for recon, a strong model for exploitation). API keys are never included.",

	"list_task_findings": "Read confirmed findings for a specific task (with flag and PoC; each row has id, task_id, intent_id, vulnclass, severity, summary, status). Select the task with task_id.",

	"list_task_worker_traces": "List which works (intents) ran in a specific task and how many steps each took. Use it to decide which work is worth reading, then get_task_worker_trace.",

	"list_tasks": "List all tasks (id, description, goal, status, run duration, parent task, LLM profile). The orchestration agent uses it for the global view: which tasks are stuck too long and which LLM each one uses. Duration: running means created to now; terminal means created to last activity, in seconds. llm_profile is the profile used by the task planner and workers; (active profile) means it follows the global active profile.",

	"list_untested_assets": "List assets in the scope of this task and directly related tasks that no fact anchor has covered yet. Related scope is read-only and is for your own judgment; it does not decide for you. Optional type filter: root_domain/subdomain/service/app/endpoint/ip. Paging: page starts at 1, page_size defaults to 10. Returns {assets:[{id,type,label}], total, page, page_size}. Only available inside a task context.",

	"node_detail": "Fetch the full content of an exploration node in this task or a directly related task by id. Inherited nodes carry source_task_id/inherited=true and are read-only. Only exploration node ids from list_facts, list_findings, or graph_overview are valid here; for assets use list_assets or asset_neighbors.",

	"pause_task": "Pause a specific task (stops its planner and worker loops).",

	"prove_goal": "Call it when a finding or fact proves that a goal is met: links the evidence node to the goal node and marks the goal met.",

	"record_fact": "Write an exploration fact or conclusion into the exploration graph, linked to the intent that produced it (intent_id). Use it for exploration results, both positive (fingerprint, enumeration) and negative (port closed, parameter not injectable, no login entry found). Fold several observations from one exploration run into ONE fact: summary is one sentence, detail holds the specifics. Example: fingerprint intent to one fact {summary: 'Identified the tech stack and response traits of site X', detail: 'nginx 1.25 / Vue3 / 200 / title=.. / body_len=..'} rather than three separate facts for status code, fingerprint, and title. One intent usually produces one fact; over-splitting inflates the graph. The facts array is for several genuinely different conclusions (each may omit intent_id and use the top-level one). Returns ids in the same order. Write only what you actually saw in tool output, never a guess. evidence is one line: the command plus the one or two output lines that prove it; keep detail in detail. confidence is observed (seen directly) or inferred (inferred from signs). For negative conclusions, state the observation and a tentative reading; the planner decides whether to drop the direction. Mark inferred when the means are not exhausted or the evidence is weak.",

	"record_finding_retest_result": "Save the single verdict for this retest session. The original finding evidence and report are unchanged. If the session ends successfully with verdict fixed, the system sets the finding status to Remediated; other verdicts keep the current status. You must provide the evidence you actually checked; when undecided, state what blocked the verdict.",

	"report_finding": "Record a confirmed finding. Use evidence for checkable proof such as command output or logs. In a task context pass the current intent_id. The returned finding_id is the separate finding-record id; finding_node_id is the exploration node id (the first line keeps that node number).",

	"search_all_worker_traces": "Usually not needed, because most information is already provided. Search the execution steps of the other works in this task by keyword (q), to recover something a worker saw but did not write into a fact (a path, token, or error). Your own intent's steps are excluded automatically. Returns matching step summaries (up to 100 characters) with intent_id; then use get_worker_trace(intent_id, step_ids=[...]) for full content.",

	"search_task_worker_traces": "Search the execution steps of every work in a specific task by keyword. Returns matching step summaries plus intent_id.",

	"set_constraints": "Add operating constraints to this task to bound exploration: type=allow (permitted action) or deny (forbidden action). A constraint is a rule about which actions may or may not be taken (for example: test this port only, do not scan other ports; no writes to the production database; passive recon only). It is not a goal and not an attack step. Prefer batch: put several in the constraints array in one call; returns ids in the same order (failed id=0, see errors). For one constraint, omit constraints and pass top-level text/type. Register only constraints written in the task goal or description; do not invent. When the type is unclear use deny.",

	"set_goals": "Add exploration goals to this task. A goal is a final deliverable or checkable result, not an attack step or recon action. Prefer batch: put several goals in the goals array in one call; returns ids in the same order (failed id=0, see errors). For one goal, omit goals and pass top-level text. vulnclass is optional (SQLi/IDOR and similar); leave it empty for business-logic goals. The system marks a goal met; this tool only adds goals.",

	"spawn_task": "Create a child task and start its exploration engine. Returns task_id. Use it to turn one piece of work (a challenge or a target) into its own task. parent_ref is optional: pass the parent task id from the current orchestration context to link parent and child.",

	"steer_work": "Inject a course correction into a running intent (work) without interrupting it or losing progress: the worker receives your message before its next action and adjusts. Use it for in-intent corrections such as stop doing X and focus on Y. If the whole direction is wrong, use kill_work and dispatch a new intent instead. Check get_worker_output first to see what it is doing.",

	"traffic_blob": "Read the original text of an oversized request or response body in chunks. The part shown in traffic_get as '...[truncated] @blob sha256:<hash>' lives here; pass that hash to get the full content. One call returns at most 8KB; continue with offset (the result reports the total length). Good for backups, leaked source, and large JSON exports that exceed the inline limit.",

	"traffic_get": "Fetch the original request and response of one captured traffic record by id (truncated when too large). Use it with traffic_search to avoid repeating curls.",

	"traffic_search": "Query traffic already captured by the recording proxy. host is required: a bare host, host:port, or a full URL, plus optional URL substring or body keyword filters. When a port is given, only that service's traffic is returned, so different ports on one IP do not mix. body_contains does a full-text search over captured request and response headers and bodies, with any substring (at least 3 characters). Returns a light index only (id/method/url/status/resp_len), no body. Once results are non-empty you must verify each request and response with traffic_get, then hand the ids that truly support the finding to bind_finding_traffic. Returns 3 rows by default, at most 10 per page; page through with page.",

	"update_custom_tool": "Edit an existing custom tool by key.",

	"update_finding_report": "Write or replace the detailed report for a recorded finding (full Markdown, replaces the old body). finding_id is the id returned by report_finding (the number in 'finding recorded: <id>'). The report should cover overview, impact, reproduction steps, evidence and PoC, and remediation.",

	"update_mcp": "Edit an existing MCP server by id.",

	"update_skill_file": "Write or overwrite one file inside a skill (SKILL.md by default). Use it to change skill content or add scripts and references.",
}

// toolParamDescEN is the English help for each tool parameter, keyed by tool
// then parameter name.
var toolParamDescEN = map[string]map[string]string{
	"add_company_scope": {
		"logo":    "Company logo URL (optional; only used when the company is created)",
		"scope":   "Asset scope, one entry per line: domain / URL / IP / CIDR / ICP filing / company keyword",
		"reason":  "Why this belongs to the company (evidence or source). Required.",
		"company": "Company name (created if missing, reused if present; names are unique)",
	},
	"add_hint": {
		"text":         "[single] Hint text, for example: focus on authenticated APIs",
		"hints":        "[preferred] Hints to add, processed in order. Each element has the same fields as the top level (text/asset_ids/traffic_refs). Returns ids in the same order.",
		"asset_ids":    "Asset ids to anchor (optional, 0/1/many)",
		"traffic_refs": "Optional: verified traffic refs that support a specific finding in this hint, order preserved. A later report_finding can carry them via evidence_hint_id.",
	},
	"add_intent": {
		"intents":    "[preferred] Exploration directions to add, processed in order. Each element has the same fields as the top level (summary/asset_ids/parent_ids/priority). Returns ids in the same order.",
		"summary":    "[single] One sentence describing the direction: what to do and why. It does not need asset ids.",
		"priority":   "Priority 0-10, default 5",
		"asset_ids":  "Target asset ids for this direction (prefer passing them: 0/1/many; these are asset ids from list_assets, not exploration node ids). Pass them whenever the direction is about concrete assets (site, endpoint, parameter, host); they mark what this exploration attacks and drive coverage dedup and the asset link. Leave empty only for global recon with no concrete asset.",
		"parent_ids": "Upstream anchor ids (optional, 0/1/many): which confirmed facts or findings this direction came from. Only existing fact/finding node ids are valid, not intents, goals, or hints. An intent must be anchored in confirmed knowledge. Pass several when several facts combine into one intent. Leave empty for a brand-new top-level recon direction (it attaches to the task origin fact automatically).",
	},
	"add_task_hint": {
		"text":         "[single] Hint text",
		"hints":        "[preferred] Hint array; each element has the same fields as the top level (text/asset_ids/traffic_refs).",
		"task_id":      "Task id",
		"asset_ids":    "Asset ids to anchor (optional, 0/1/many; asset ids inside that task)",
		"traffic_refs": "Optional: verified traffic refs that support a specific finding in this hint, order preserved. A later report_finding can carry them via evidence_hint_id.",
	},
	"add_task_scope": {
		"kind":    "[single] company / root_domain / subdomain / ip / cidr / icp / keyword",
		"value":   "[single] Company name or id / domain / IP / CIDR / ICP / keyword",
		"reason":  "Why it was added (for audit). Required.",
		"entries": "Batch: [{kind, value}]. kind is company/root_domain/subdomain/ip/cidr/icp/keyword.",
	},
	"bind_finding_traffic": {
		"finding_id":   "Separate finding-record id, taken from the finding_id field of list_task_findings or get_task_node_detail",
		"traffic_refs": "Optional: verified traffic refs for this finding, order preserved.",
	},
	"create_custom_tool": {
		"key":         "Tool key (starts with a lowercase letter; letters, digits, underscore)",
		"exec":        "Execution spec (not needed for shell): command to {command}; script to {code}; http to {method,url,headers,body,proxy,use_recording_proxy}",
		"kind":        "shell | command | script (Python only) | http. shell is only a bash environment declaration (tells the model the tool can be called directly in bash; no exec or schema). The other three need exec.",
		"agents":      "Agent keys to bind (optional)",
		"schema":      "Parameter JSON-Schema (may be empty for shell/command/script; required for http, must contain properties)",
		"enabled":     "Enabled (default true)",
		"deferred":    "Deferred (no effect for shell; only for rarely used command/script/http tools)",
		"description": "Description shown to the model",
	},
	"create_mcp": {
		"env":       "Environment variables {KEY:VALUE}",
		"url":       "URL for http/sse",
		"args":      "Command argument array",
		"name":      "MCP server name",
		"command":   "stdio start command (for example npx)",
		"enabled":   "Enabled (default true)",
		"insecure":  "http: skip TLS certificate check (true for self-signed certificates, default false)",
		"transport": "stdio | http / sse",
	},
	"create_skill": {
		"name":         "Skill name (starts with a lowercase letter; letters, digits, hyphen)",
		"description":  "Skill description (required; what it does and when to use it)",
		"instructions": "Markdown body (optional)",
	},
	"delete_assets_by_host": {
		"host": "Host to delete: domain, subdomain, or IP. Exact match, for example example.com, a.example.com, or 1.2.3.4",
	},
	"expand_digest": {
		"id": "Digest node id (from overview cold_digests)",
	},
	"get_finding_traffic": {
		"side":       "request or response, default response",
		"binding_id": "Binding id from the list; omit to return the list",
		"finding_id": "Separate finding-record id",
	},
	"get_task_graph": {
		"task_id": "Task id",
	},
	"get_task_node_detail": {
		"id":      "Exploration node id (not an asset id)",
		"task_id": "Task id",
	},
	"get_task_worker_trace": {
		"task_id":   "Task id",
		"step_ids":  "Optional: step ids to fetch in full (at most 5 per call; extra ids are listed in omitted_step_ids)",
		"intent_id": "Intent id (the work inside that task)",
	},
	"get_worker_output": {
		"intent_id": "Intent id (the work handle)",
	},
	"get_worker_trace": {
		"q":         "Keyword: return only steps whose summary or full output matches (optional; mutually exclusive with step_ids)",
		"limit":     "Return limit for the summary stream or search (optional)",
		"step_ids":  "Step ids to fetch in full (from mode 1 or 2; at most 5 per call, extras are listed in omitted_step_ids)",
		"intent_id": "Intent id (the work handle)",
	},
	"goal_met": {
		"reason": "Why the task is done. Must be evidence that the goals are truly met, not a reason like no new direction this round.",
	},
	"insert_assets": {
		"assets": "Asset array; each element becomes one asset record",
	},
	"kill_work": {
		"intent_id": "Intent id to terminate (the work handle)",
	},
	"list_assets": {
		"id":     "Fetch by a single asset id (optional; exclusive with dsl/type)",
		"dsl":    "DSL query expression (syntax and fields in the tool description). Required when id/ids are absent.",
		"ids":    "Fetch by several asset ids (optional; exclusive with dsl/type)",
		"type":   "Asset type filter: root_domain|ip|subdomain|app|service|endpoint (separate field, stacks with dsl; type alone is not enough, dsl is still required)",
		"limit":  "Return limit, default 10 (optional)",
		"offset": "Page offset, default 0 (optional)",
	},
	"list_companies": {
		"search": "Fuzzy company-name filter (optional, case-insensitive); empty returns all",
	},
	"list_facts": {
		"q":      "Filter by fact summary keyword (case-insensitive); omit to return all",
		"limit":  "Number of rows, default 20, max 100",
		"before": "Page cursor: return facts with id lower than this value; omit or 0 for the newest page",
	},
	"list_task_findings": {
		"task_id": "Task id",
	},
	"list_task_worker_traces": {
		"task_id": "Task id",
	},
	"list_untested_assets": {
		"page":      "Page number, starting at 1 (default 1)",
		"type":      "Asset type filter (optional): root_domain/subdomain/service/app/endpoint/ip",
		"page_size": "Rows per page (default 10)",
	},
	"node_detail": {
		"id": "Exploration node id (not an asset id)",
	},
	"pause_task": {
		"task_id": "Task id to pause",
	},
	"prove_goal": {
		"reason":      "Why this evidence satisfies the goal",
		"goal_id":     "Goal node id",
		"evidence_id": "Finding or fact node id that proves it",
	},
	"record_fact": {
		"facts":      "[use when you have several different conclusions] Fact array; each element has the same fields as the top level (summary/detail/evidence/confidence/intent_id/asset_ids). Omit intent_id to use the top-level one. Returns ids in the same order.",
		"detail":     "Details for this fact: write the several observations from this exploration run here",
		"summary":    "One summary sentence for this exploration conclusion (a summary of detail)",
		"evidence":   "One line of key evidence: the command plus the one or two output lines that prove it. Keep it short; put detail in detail.",
		"asset_ids":  "Related asset ids (optional, 0/1/many): which assets this fact covers",
		"intent_id":  "Intent id that produced this fact (the intent you were given; used as the default for batch items)",
		"confidence": "observed (seen directly in output) | inferred (inferred from signs). Mark negative conclusions honestly.",
	},
	"record_finding_retest_result": {
		"summary":  "Summary of this retest verdict",
		"evidence": "Markdown: the steps you ran, what you observed, controls used, and why the verdict follows; if undecided, list what you checked and what blocked you",
	},
	"report_finding": {
		"name":             "Finding name",
		"summary":          "Finding summary",
		"evidence":         "Evidence or PoC text",
		"severity":         "critical|high|medium|low",
		"asset_ids":        "Affected asset ids",
		"intent_id":        "Intent id in the current task",
		"vulnclass":        "Vulnerability class",
		"traffic_refs":     "Optional; for HTTP/HTTPS findings, search and verify each request and response against the finding first, then fill real ids in reproduction order. For non-HTTP findings such as TCP, or when nothing was captured or no exact record exists, omit or pass []; this does not block the report. Explain why in evidence and provide other checkable evidence. Do not guess ids, infer a link from host or time, or re-probe only to fill a packet. purpose: baseline (normal control) / proof / verification / supporting.",
		"evidence_hint_id": "Optional: the hint node id in this task that corresponds to this finding; its structured traffic_refs are carried automatically. Cannot reference an inherited hint or another finding's hint.",
	},
	"search_all_worker_traces": {
		"q":     "Keyword (searched in the summary and full output of every work step)",
		"limit": "Return limit, default 100 (optional)",
	},
	"search_task_worker_traces": {
		"q":       "Search keyword",
		"task_id": "Task id",
	},
	"set_constraints": {
		"text":        "[single] One operating constraint",
		"type":        "[single] allow or deny; defaults to deny",
		"constraints": "[preferred] Constraints to add, processed in order. Each element: text (required) plus type (allow|deny). Returns ids in the same order.",
	},
	"set_goals": {
		"text":      "[single] One independent, checkable final goal",
		"goals":     "[preferred] Goals to add, processed in order. Each element: text (required, one independent checkable goal) plus optional vulnclass. Returns ids in the same order.",
		"vulnclass": "[single] Vulnerability class when clear, such as SQLi or IDOR; leave empty for business-logic goals",
	},
	"spawn_task": {
		"goal":                   "Task goal (what must be achieved)",
		"parent_ref":             "Optional: parent task id (links parent and child)",
		"description":            "Task description (short title)",
		"llm_profile_id":         "Optional: LLM profile id for this child task's planner and workers (see list_llm_profiles); empty inherits the parent, then the global active profile",
		"source_task_ids":        "Optional: read-only source task ids to inherit from (at most 8). The child can read those tasks' known assets and conclusions as a starting point. Unlike parent_ref, which is only a pointer, this inherits content.",
		"timeout_seconds":        "Optional: task timeout in seconds. On expiry the task wraps up gracefully and ends as timeout; empty or 0 means no limit",
		"seed_first_intent":      "Optional: for simple tasks, dispatch one seed intent at creation (content = description plus goal) so a worker starts testing without waiting for the first planner round. Default false (plan first, then execute).",
		"plan_heartbeat_seconds": "Optional: planner heartbeat interval in seconds. When this much time passes since the last planning round or task start with no trigger, a planning round fires (deadlock fallback and supervision of running workers). Empty or 0 means the default 600 (10 min).",
	},
	"steer_work": {
		"message":   "Correction for the worker: what to stop and what to switch to",
		"intent_id": "Intent id to correct (the work handle)",
	},
	"traffic_blob": {
		"hash":   "The 64-character hex value after @blob sha256: in traffic_get",
		"length": "Bytes to read, default and maximum 8192",
		"offset": "Start byte offset, default 0",
	},
	"traffic_get": {
		"id": "Id returned by traffic_search",
	},
	"traffic_search": {
		"host":          "Filter by host (required; for example 107.172.96.177, 107.172.96.177:8082, or http://107.172.96.177:8082/path)",
		"page":          "Page number, starting at 0, default 0 (paged by ts descending)",
		"limit":         "Rows per page, default 3, max 10",
		"contains":      "URL substring filter (optional, for example api or login)",
		"body_contains": "Full-text body search (optional, at least 3 characters) over request and response headers and bodies, for example password or root:x:0",
	},
	"update_custom_tool": {
		"key":         "Custom tool key to edit",
		"exec":        "Execution spec (not needed for shell): command to {command}; script to {code}; http to {method,url,headers,body,proxy,use_recording_proxy}",
		"kind":        "shell | command | script (Python only) | http. shell is only a bash environment declaration (tells the model the tool can be called directly in bash; no exec or schema). The other three need exec.",
		"agents":      "Agent keys to bind (optional)",
		"schema":      "Parameter JSON-Schema (may be empty for shell/command/script; required for http, must contain properties)",
		"enabled":     "Enabled (default true)",
		"deferred":    "Deferred (no effect for shell; only for rarely used command/script/http tools)",
		"description": "Description shown to the model",
	},
	"update_finding_report": {
		"report":           "Full report text, Markdown",
		"finding_id":       "Target finding id (the id returned by report_finding)",
		"evidence_version": "Evidence version returned by get_finding_traffic; prevents overwriting a newer evidence change",
	},
	"update_mcp": {
		"id":        "MCP server id to edit",
		"env":       "Environment variables {KEY:VALUE}",
		"url":       "URL for http/sse",
		"args":      "Command argument array",
		"name":      "MCP server name",
		"command":   "stdio start command (for example npx)",
		"enabled":   "Enabled (default true)",
		"insecure":  "http: skip TLS certificate check (true for self-signed certificates, default false)",
		"transport": "stdio | http / sse",
	},
	"update_skill_file": {
		"file":    "Relative path (optional, default SKILL.md, for example scripts/run.py)",
		"name":    "Skill name",
		"content": "Full file content",
	},
}

// LocalizeToolDesc returns the English description for a built-in tool key, or
// the original description when no translation exists.
func LocalizeToolDesc(key, desc string) string {
	if en, ok := toolDescEN[key]; ok && en != "" {
		return en
	}
	return desc
}

// toolParamDescCommon covers parameters shared by several tools, including the
// nested item fields inside array parameters (assets, hints, traffic_refs).
var toolParamDescCommon = map[string]string{
	"description":      "Description",
	"name":             "Name",
	"type":             "Asset type",
	"domain":           "Root domain or subdomain (required for root_domain/subdomain)",
	"root_domain":      "Root domain",
	"subdomain":        "Subdomain",
	"ip":               "IP address. Must be IPv4 or IPv6, not a hostname (use type=subdomain with domain for hostnames). Required for type ip; optional for service/endpoint to link an IP.",
	"ip_domains":       "Domain names bound to this IP (optional for type ip)",
	"ports":            "Open port list (optional for type ip)",
	"port":             "Port number",
	"url":              "Full URL with scheme and port (required for HTTP services; service_type is set to http automatically)",
	"app_name":         "Application name (required for type app)",
	"app_category":     "Application category (optional)",
	"app_desc":         "Application description (optional)",
	"category":         "Application category (optional)",
	"open_ports":       "Open port list (optional for type ip)",
	"bound_domains":    "Domain names bound to this IP (optional for type ip)",
	"record_value":     "DNS record values (optional for subdomain, for example [\"1.2.3.4\",\"2.3.4.5\"])",
	"content_length":   "HTTP response body length in bytes (optional)",
	"service":          "Service name, such as http/ssh/mysql (optional)",
	"app_icp":          "Application ICP filing (optional)",
	"company_id":       "Owning company id (optional for type app; apps cannot be attributed by scope, so set it explicitly. The id comes from add_company_scope)",
	"icp":              "ICP filing number (optional)",
	"service_name":     "Service name, such as ssh/mysql/redis (required for non-HTTP service)",
	"method":           "HTTP method: GET/POST/PUT/PATCH/DELETE and similar (required for endpoint)",
	"status_code":      "HTTP response status code, such as 200/301/403/404 (optional)",
	"body_len":         "HTTP response body length in bytes (optional)",
	"page_title":       "Page title element content (optional)",
	"favicon_mmh3":     "favicon MMH3 hash (optional)",
	"record_type":      "DNS record type: A/AAAA/CNAME/MX and similar (optional for subdomain)",
	"record_values":    "DNS record value list (optional for subdomain, for example [\"1.2.3.4\",\"2.3.4.5\"])",
	"technologies":     "Fingerprint and tech stack list, for example [\"Nginx\",\"Vue\",\"Bootstrap\"] (optional)",
	"auth":             "Discovered authentication entries; each has type/username/password and similar fields (optional, appended not replaced)",
	"params":           "Request parameter list; each has location (query/body/header/path)/name/value/type (optional, appended not replaced)",
	"bundle_id":        "Bundle ID (optional for type app)",
	"task_id":          "Task id",
	"intent_id":        "Intent id (the work handle)",
	"step_ids":         "Step ids to fetch in full (at most 5 per call; extra ids are listed in omitted_step_ids)",
	"goal_id":          "Goal node id",
	"evidence_id":      "Finding or fact node id that proves it",
	"reason":           "Reason (for audit). Required.",
	"q":                "Search keyword",
	"limit":            "Return limit",
	"offset":           "Start offset",
	"page":             "Page number",
	"hash":             "The 64-character hex value after @blob sha256: in traffic_get",
	"length":           "Bytes to read",
	"binding_id":       "Binding id from the list; omit to return the list",
	"side":             "request or response, default response",
	"finding_id":       "Separate finding-record id",
	"traffic_id":       "Real traffic id returned by traffic_search",
	"note":             "How this traffic supports the finding",
	"purpose":          "What the traffic supports: baseline / proof / verification / supporting",
	"hints":            "[preferred] Hint array; each element has the same fields as the top level (text/asset_ids/traffic_refs).",
	"text":             "Text",
	"message":          "Correction for the worker: what to stop and what to switch to",
	"code":             "Code",
	"content":          "File content",
	"file":             "Relative path (optional, default SKILL.md, for example scripts/run.py)",
	"skill_name":       "Skill name",
	"kind":             "Kind",
	"value":            "Value",
	"enabled":          "Enabled (default true)",
	"deferred":         "Deferred",
	"agents":           "Agent keys to bind (optional)",
	"key":              "Key",
	"args":             "Command argument array",
	"env":              "Environment variables {KEY:VALUE}",
	"command":          "stdio start command (for example npx)",
	"insecure":         "http: skip TLS certificate check (true for self-signed certificates, default false)",
	"transport":        "stdio | http / sse",
	"severity":         "critical|high|medium|low",
	"vulnclass":        "Vulnerability class",
	"summary":          "Summary",
	"evidence":         "Evidence or PoC text",
	"report":           "Full report text, Markdown",
	"evidence_version": "Evidence version returned by get_finding_traffic; prevents overwriting a newer evidence change",
	"assessed":         "Assessment",
	"confidence":       "observed (seen directly in output) | inferred (inferred from signs)",
	"goal":             "Task goal (what must be achieved)",
	"dsl":              "DSL query expression (syntax and fields in the tool description). Required when id/ids are absent.",
	"search":           "Fuzzy name filter (optional, case-insensitive); empty returns all",
}

// paramDesc resolves the English help for one parameter name, preferring the
// per-tool map and falling back to the shared map.
func paramDesc(toolKey, name string) string {
	if m, ok := toolParamDescEN[toolKey]; ok {
		if en, ok := m[name]; ok {
			return en
		}
	}
	return toolParamDescCommon[name]
}

// localizeSchemaNode walks the schema and rewrites Chinese descriptions in
// place. Structure, defaults, enums, and required lists are never touched.
func localizeSchemaNode(toolKey, name string, node any) {
	switch n := node.(type) {
	case map[string]any:
		if d, ok := n["description"].(string); ok && HasCJK(d) {
			if en := paramDesc(toolKey, name); en != "" {
				n["description"] = en
			}
		}
		for k, v := range n {
			localizeSchemaNode(toolKey, k, v)
		}
	case []any:
		for _, v := range n {
			localizeSchemaNode(toolKey, name, v)
		}
	}
}

// LocalizeToolSchema replaces Chinese parameter descriptions with English ones,
// keeping the structure, defaults, and every other field untouched.
func LocalizeToolSchema(key string, schema map[string]any) map[string]any {
	if len(schema) == 0 {
		return schema
	}
	if _, ok := toolParamDescEN[key]; !ok {
		return schema
	}
	localizeSchemaNode(key, "", schema)
	return schema
}

// LocalizeToolSchemaJSON runs LocalizeToolSchema and reports whether anything changed.
func LocalizeToolSchemaJSON(key string, schema map[string]any) (map[string]any, bool) {
	before := countCJK(schema)
	out := LocalizeToolSchema(key, schema)
	return out, countCJK(out) != before
}

func countCJK(v any) int {
	switch t := v.(type) {
	case string:
		if HasCJK(t) {
			return 1
		}
		return 0
	case map[string]any:
		n := 0
		for _, x := range t {
			n += countCJK(x)
		}
		return n
	case []any:
		n := 0
		for _, x := range t {
			n += countCJK(x)
		}
		return n
	}
	return 0
}

// SchemaHasCJK reports whether any string in the schema still has Chinese text.
func SchemaHasCJK(schema map[string]any) bool { return countCJK(schema) > 0 }
