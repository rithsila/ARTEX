---
name: api-recon
description: Invoke this skill when collecting website API endpoints.
---

# API Recon (Frontend Interface Recon)

Under an **authorized** scope, discover as completely as possible: **backend API** (path, method, params, response body), **frontend routes**, **UI trigger points** (Tab, modal, table actions, etc.).

---

## Boundaries and Prohibitions (Agent Must Read · Violation Is Out of Scope)

This skill does **API / param surface recon only**. It is not a vulnerability hunting or exploitation phase.

### Task Boundaries

| Scope | Allowed | Forbidden |
|---|---|---|
| **Target** | Enumerate path, method, params, routes, UI trigger points | SQLi/XSS/IDOR/brute force/fuzz bugs, packet tampering attacks, destructive operations |
| **Auth** | Hook + stub/mock to bypass the **client-side** login gate | Asking the user for or guessing credentials; attempting real login form submission |
| **Runtime** | Hook interfaces without credentials, use mock responses to push the SPA into the post-login shell | Flows that require a real backend session to continue |

### Credential-Free Dynamic Analysis (Phase 3 Default)

1. Use `preload.js` / `runtime_harvest.js` to **intercept and stub** login, permission, menu and other bootstrap interfaces;
2. Return **correctly shaped, business-code-success, data may be empty** mock bodies for business query interfaces;
3. Let the frontend render post-login pages with no backend or under 401, so more XHR/fetch/WebSocket fire;
4. **Empty data, blank tables, placeholder UI are all expected** — do not pivot to real login or vulnerability testing because of them.

**One line**: use mocks to prop open frontend routes and component mounting, **record outbound requests only**; what the backend returns does not matter, what matters is which interfaces the frontend **still sends**.

### Hard Process Prohibitions

| Prohibited | Alternative |
|---|---|
| Before Phase 1 completes, grep/curl/Read the main entry `index-*.js` to extract API path | Run `OUTDIR/harvest_static.py` |
| Hand-write `extract_apis.py` or similar scripts to replace harvest | Edit `OUTDIR/harvest_static.py` and rerun |
| Repeat the same grep/command after it fails ≥2 times | Change strategy: read tool_logs, edit harvest, check reference |
| Skip Gate A/B and run the stock `scripts/` directly | Copy to OUTDIR and modify for the target |
| Real username/password, OTP, OAuth or other auth | stub/mock (see above) |
| Skip stub to "get real data", run IDOR/injection tests | Record outbound only; that is the recon boundary |
| Delete, export sensitive data, bulk write or other irreversible operations | Same applies to coverage clicks |
| Claim all pages and interfaces are captured without runtime + dynamic enumeration | See "Definition of Done" or mark limitations |
| Claim all params are known without the param trigger matrix + diff | Phase 3b matrix + Phase 5 diff |
| Infer required/optional from a single runtime sample | Multi-sample diff or validation rules/error reverse inference |

---

## Two-Layer Model + Run Modes

| Layer | Output | Limit |
|---|---|---|
| **Static** (JS bundle) | Full endpoint paths, route draft, request assembly field candidates | No HTTP method; params need Phase 1b; misses runtime-concatenated URLs |
| **Runtime** (live session) | Method + body + response + dynamic URL + WS/SSE; multi-sample diff fills params | Page must actually render to send requests; single sample cannot decide required/optional |

| Run mode | Engine | Fit |
|---|---|---|
| **depth** | `runtime_harvest.js` (Puppeteer) | API inventory, METHOD/params/response body, WS/SSE, reproducible batch runs |
| **coverage** | browser + `preload.js` | Click Tab/modal/table, deeper feature coverage |
| **both** | depth then coverage | Most complete, longest runtime |

**Param methodology** (no generic script): path via harvest/regex; params via **anchor window expansion + UI binding chain + multi-sample diff + error reverse inference** (grep recipes in [reference.md](reference.md) section J).

---

## Definition of Done

Recon may be claimed complete only when all hold:

- [ ] **Static**: Phase 1 harvest produces `api_static.txt`, `routes.txt`, `js/`
- [ ] **Runtime**: at least one of depth or coverage; coverage/both require **Hook active + dynamic enumeration loop**
- [ ] **Inside shell**: business path does not redirect to `/login` (watch hash routes)
- [ ] **Params**: coverage/both complete the param trigger matrix + `param_samples.json`; Phase 5 merges `params_merged.json`
- [ ] **Depth** (if module pages are blank): Phase 4 permission tree reconstruction, rerun until **module-level API** appears (not just locale/bootstrap)
- [ ] **Delivery**: Phase 5 outputs complete (see Phase 5 output table); `insert_assets` writes service and endpoint assets

---

## Scripts and Gates

`scripts/` holds reference templates only. **Never** run stock copies and treat the result as final.

**Rule**: read → modify for the target → write into `OUTDIR` (e.g. `recon/`) → record in `CHANGES.md`; if it does not match, rewrite per methodology, borrow structure only.

| Gate | When | Reference script → OUTDIR copy | Common required edits |
|---|---|---|---|
| **A (static)** | After Phase 0, before the **first** harvest/spider run | `harvest_static.py` / `spider_mpa.py` | **Most sites run on the default regex as-is**; edit endpoint regex, webpack/Vite `publicPath`, MPA exclude/cookie only when manifest/dialect mismatches |
| **B (runtime)** | After Phase 2, before depth/coverage runs | `runtime_harvest.js` / `preload.js` + `config.json` | Cookie/localStorage keys, neutralize success value, stubs, login regex, api prefix, hash/history |

**SPA mandatory order** (not swappable; Phase numbers take priority over "explore first, script later"):

| Step | Required | Prohibited |
|---|---|---|
| After Phase 0 completes | Next Bash = `python3 OUTDIR/harvest_static.py <URL> OUTDIR` | curl/grep/Read the main entry `index-*.js` (usually >500KB) |
| Gate A | Copy script → small edits as needed → **run immediately** | Manually extract API first, then decide whether to harvest |
| Before Phase 1 completes | `wc -l` to verify output; on 404 edit harvest and retry | Hand-write extract scripts; grep repeatedly against a URL that was never downloaded |
| From Phase 1b on | grep `OUTDIR/js/*.js` only | Substitute the main bundle for harvest |

- ✅ Copy `harvest_static.py` → (optional) edit regex → **run immediately**
- ❌ curl main bundle → grep many times → write temp extract → harvest last
- **MPA**: next Bash after Phase 0 = `python3 OUTDIR/spider_mpa.py ...`

---

## Tool and Output Constraints

| Constraint | Description |
|---|---|
| Large files | `index-*.js` over 100KB **must not** be Read/grep'd into context; batch process with OUTDIR scripts |
| grep output | Must use `\| head -20` or `-m 5`; keep only path summaries in conversation, do not paste bundle fragments |
| Verification | Use `wc -l`, `ls \| wc -l`; do not Read whole directories |
| regex probe | Optional, ≤1 time, only on small chunks ≤50KB or HTML; official static pass is harvest |
| reference | Recipes/templates/troubleshooting in [reference.md](reference.md); do not inline the full text again |

---

## Execution Roadmap

```
Phase 0 classification + OUTDIR
  → Gate A → Phase 1 harvest (★ run immediately ★)
  → Phase 1b param reverse engineering
  → Phase 2 three auth gates → config.json
  → Gate B → Phase 3 runtime + param matrix
  → Phase 4 permission tree (if needed) → rerun Phase 3
  → Phase 5 merge report + insert_assets bulk insert of all discovered service and endpoint api assets; never omit a discovered asset on insert
```

Check in order; **do not enter the next Phase until the previous item is complete**.

1. [ ] **Phase 0**: probe SPA/MPA; create `OUTDIR` → [Phase 0](#phase-0--classification)
2. [ ] **Gate A + Phase 1**: copy script → harvest **immediately** → `wc -l` verify → [Phase 1](#phase-1--static)
3. [ ] **Phase 1b**: anchor window expansion + binding layer → `param_candidates.json` → [Phase 1b](#phase-1b--param-reverse-engineering)
4. [ ] **Phase 2**: three auth gates → `config.json` → [Phase 2](#phase-2--three-auth-gates)
5. [ ] **Gate B**: adjust runtime script → [Phase 3](#phase-3--runtime)
6. [ ] **Phase 3**: depth / coverage / both; confirm shell entry; param trigger matrix → `param_samples.json`
7. [ ] **Phase 4** (if needed): permission tree → patch stubs → rerun Phase 3 → [Phase 4](#phase-4--permission-tree-reconstruction)
8. [ ] **Phase 5**: merge outputs + report + `insert_assets` → [Phase 5](#phase-5--merge-and-report)

---

## Phase 0 — Classification

Fetch the entry HTML and **create `OUTDIR`** (do not modify `scripts/` inside the skill):

- **SPA**: empty shell + `<div id=app>` + chunk → Phase 1–5
- **MPA**: SSR + `<form>`, no endpoint bundle → after Gate A:

```bash
python3 recon/spider_mpa.py <BASE_URL> <OUTDIR> [--cookie "session=..."] [--max 300] [--depth 5] [--exclude "logout|delete|destroy"]
```

Outputs `forms.txt`, `links.txt`, `api_inline.txt`. If SPA forms ≈ 0 → switch to Phase 1.

---

## Phase 1 — Static

Follow [Scripts and Gates](#scripts-and-gates) · [Tool and Output Constraints](#tool-and-output-constraints).

```bash
python3 recon/harvest_static.py <BASE_URL> <OUTDIR>
```

harvest: parse HTML script → webpack/Vite manifest → download all lazy chunks → produce `js/`, `api_static.txt`, `routes.txt`, `chunkmap.txt`.

```bash
wc -l OUTDIR/api_static.txt OUTDIR/routes.txt
ls OUTDIR/js | wc -l
```

- chunk count vs manifest: on 404 edit harvest and retry, do not curl chunks one by one by hand
- `api_static.txt` too small → loosen the endpoint regex inside OUTDIR and rerun (see reference)

### Phase 1b — Param Reverse Engineering

path comes from Phase 1; param fields need separate recon. grep rules in [Tool and Output Constraints](#tool-and-output-constraints).

**Completion standard**: for important interfaces you can answer — field names, transport location, type inference, required or not, sample values, confidence.

#### 1b.0 — Transport Shape

| Shape | Where params live | Static first look |
|---|---|---|
| REST JSON | body + query | `(params\|data\|body)\s*:\s*\{` next to the path anchor |
| GraphQL | `variables` | gql template, `$page: Int` |
| Legacy form | urlencoded | `<form>`, `FormData` |
| File upload | multipart | `FormData.append` |
| Path params | `/user/:id` | route table + `useParams` / `$route.params` |
| Encrypted/signed | wrapped in `sign`/`data` | Hook the encrypt function input (reference section D) |

Output: annotate each interface `transport: query|json|form|graphql|encrypted`.

#### 1b.1 — Anchor Window Expansion

Use a known path as anchor, expand the window to find the request assembly object:

```bash
grep -n '"/api/user/list"' OUTDIR/js/*.js | head -20
grep -rhoaE '.{0,120}("/api[^"]+").{0,200}' OUTDIR/js/*.js | head -20
grep -rhoaE '(params|data|body|payload)\s*:\s*\{' OUTDIR/js/*.js | head -20
```

| Wrapper layer | Param clue |
|---|---|
| axios instance | `data` / `params` |
| Unified request | interceptor injects global fields |
| OpenAPI client | generated method signature |
| React Query / SWR | hook second argument |
| Vue composable | composable input |

Type residue: `yup`/`zod`/rules, `Form.Item name=`, embedded Swagger.

→ `param_candidates.json`: `{ path, fields[], source: "static-callsite", confidence }`

#### 1b.2 — Binding Layer

```
Form field → onFinish/handleSubmit → transform → API payload
```

| Binding source | Technique |
|---|---|
| Form submit | follow submit → transform → API |
| Table search | `getFieldsValue()` → `params` |
| Route | `:id` / `?tab=` |
| Interceptor | global `tenantId`, pagination, sign |
| Enum select | `options` → API enum values |

DevTools call stack: walk up from `fetch`/`XHR.send` to the assembly function.

#### 1b.3 — Assembly Three Questions (≠ Phase 2 three auth gates)

| Question | What to answer |
|---|---|
| **Assembly** | where payload is built, transform traces |
| **Validation** | required, pattern, enum |
| **Transport** | path / query / body / multipart / headers |

The interceptor gate (Phase 2) also reads globally injected fields (Authorization, `X-Tenant-Id`, sign).

#### 1b.4 — Handoff to Phase 3

Candidate fields come from static/binding layer; **required/optional/conditional dependency** need Phase 3 param matrix + diff + Phase 5 error reverse inference.

---

## Phase 2 — Three Auth Gates

grep inside `OUTDIR/js/` (with `head`), write `config.json` (recipes in reference):

| Gate | Question | Keywords |
|---|---|---|
| **Render gate** | How is logged-in determined? | `isLogin`, `getToken`, Cookie/localStorage |
| **Interceptor gate** | What triggers the jump to `/login`? | `response_code`, `errno`, axios interceptor |
| **Content gate** | Where do menu/permissions come from? | `menu`, `permission`, `role`, `acl`, `routes` |

Never treat a localStorage key name as a credential — confirm from chunk/request chain.

**Exit = Gate B**: conclusions land in `config.json`, and `OUTDIR/runtime_harvest.js` / `preload.js` are edited.

### Phase 2b — API Observation (Optional)

Use `preload.js` inside OUTDIR to confirm session key names, Authorization, nested API URLs:

| Config | Output |
|---|---|
| `recordDetail: true` | `__API_RECON_DETAIL__` |
| `observe.xhrHeaders: true` | header observation |
| `extractUrlsFromResponse: true` | sub-API inside responses |
| `observe.storageReads/cookieReads: true` | backfill config |
| `neutralizeVueRouter: true` | `__API_RECON_ROUTES__` |

Export after each coverage round: `__API_RECON_LOG__`, `__API_RECON_DETAIL__`, `__API_RECON_ROUTES__`, `__API_RECON_OBSERVE__`.

---

## Phase 3 — Runtime

Gate B must already be passed; follow [Boundaries and Prohibitions](#boundaries-and-prohibitions-agent-must-read--violation-is-out-of-scope) · credential-free mock strategy.

In `config.json` set `"runtimeMode": "depth" | "coverage" | "both"` (template in reference).

### Hook and stub (shared by depth + coverage)

| Layer | Scope | Purpose |
|---|---|---|
| L1 precise | auth/permission/bootstrap stub | pass first-screen auth |
| L2 negative correction | all JSON responses | not-logged-in code → success |
| L3 fallback | `/api` and similar not hit by L1 | empty success body, prop open the UI |

- **depth**: fake auth + `forward` to rewrite business code + `stubs`; iterate `routes` (hash/history); produce `runtime_api.json`
- **coverage**: inject `preload.js` at **document-start** (CDP `addScriptToEvaluateOnNewDocument` or Userscript)

Verify: `window.__API_RECON_PRELOAD__` exists; business path does not return to `/login`.

```bash
cd recon && npm install
node runtime_harvest.js config.json
```

### 3b — coverage Dynamic Enumeration (Required)

1. Main nav/sidebar — click each item, wait 1–3s for network
2. Tab — `role=tab`, `.ant-tabs-tab`
3. Table — first row view/edit/detail
4. Toolbar — export, filter, create (**avoid irreversible deletes**)
5. Each module entered — merge API/route
6. SPA — controlled `pushState` for paths in `routes.txt` not yet covered (forbidden on MPA)

**Param trigger matrix** (required): record once per operation type per module, **diff multiple samples**:

| Operation | Params usually added |
|---|---|
| List first screen | pagination + default filters |
| Click search | keyword, filter |
| Advanced filter | more optional |
| Create/edit | full entity |
| Bulk/export/sort | `ids[]`, `exportType`, `sortField` |

**Outbound body/headers stay real under stub** — trust the request. Record → `scan_raw.json`, `param_samples.json`, `api_detail.json`.

- **Vue**: `neutralizeVueRouter: true` + document-start preload
- **React**: `routes.txt` + sidebar clicks + `pushState`
- **both**: 3a depth first, then 3b coverage

---

## Phase 4 — Permission Tree Reconstruction

**Trigger**: module pages blank / each route returns only bootstrap (e.g. locale) → content gate not passed.

| Symptom | Meaning |
|---|---|
| Shell entry succeeds | render gate + interceptor gate passed |
| Missing sidebar items / blank on click | stub shape or permission codes incomplete |
| Same few APIs on every route | `v-if permission` not passing |
| `routes.txt` far smaller than bundle | must be completed from the auth module |

```bash
grep -rhoaE '"/api[^"]*(permission|perm|role|menu|acl)[^"]*"' OUTDIR/js/*.js | sort -u | head -30
grep -rhoaE 'userRouteAuth|getResultTree|routeMap|routeLink|menuList|authList' OUTDIR/js/*.js | head -20
```

Typical chain: `role_permissions` (flat codes) + `permissions/all` (tree) → `getResultTree` → `userRouteAuth[CODE].url`.

```bash
python3 recon/extract_route_map.py recon/js recon/
python3 recon/build_perm_tree.py recon/js recon/ --config recon/config.json
```

Intermediate output: `route_map.json`, `userRouteAuth.json`, `permissions_tree.json`, `*_stub.json`, `perm_codes_all.txt`.

stub checks: outer `response_code` matches the interceptor gate; flat codes align with tree; `routes` covers every link in `route_map`.

After updating `config.json`, **rerun Phase 3**. For large SPAs tune `waitUntil`, `routeTimeout`, `perRouteMs` (see reference sections A3/I).

---

## Phase 5 — Merge and Report

### Output Table

| File | Phase | Content |
|---|---|---|
| `js/`, `api_static.txt`, `routes.txt`, `chunkmap.txt` | 1 | static bundle and path |
| `param_candidates.json` | 1b | static param field candidates |
| `config.json` | 2 | three gates + runtime config |
| `runtime_api.json` | 3a | depth detailed recording (incl. WS/SSE) |
| `param_samples.json`, `scan_raw.json`, `api_detail.json` | 3b | multi-sample, click log, detail |
| `route_map.json` etc. | 4 | permission tree intermediate files (if run) |
| `params_merged.json` | 5 | merged param fields + confidence |
| `api_merged.txt` | 5 | `METHOD /path [params] [static\|runtime\|both]` |
| `site_map.json` | 5 | routes, APIs, params, feature points, limitations |
| **insert_assets** | 5 | write all service and endpoint assets into the asset store |

### 5b — Param Merge

Diff from `param_samples.json`. **No generic merge script**. Confidence rules in reference J7 (high/medium/low/pending trigger).

### 5c — Error Reverse Inference

Within authorized scope, send incomplete requests to read 400s (**param recon, not vulnerability testing**): `field 'x' is required`, enum errors, etc. Watch the `data` wrapper, `variables`, and pre-encryption `bizData`.

The report must state: runtimeMode, static/runtime API counts, param confidence, uncovered modules, `CHANGES.md` summary against the reference scripts.

`site_map.json` suggested structure:

```json
{
  "site": "https://example.com",
  "runtimeMode": "both",
  "appType": "vue-spa",
  "routeGuardStrategy": ["nav-neutralize", "L1-auth", "L2-patch", "forward"],
  "apisFromStatic": [],
  "apisFromRuntime": [],
  "apis": [],
  "params": [{ "method": "POST", "path": "/api/user/list", "transport": "json", "fields": [] }],
  "frontendRoutes": [],
  "routesVerifiedByClick": [],
  "featuresTriggered": [],
  "limitations": ""
}
```

More fields and grep recipes in [reference.md](reference.md).

---

## General Notes

- **Framework agnostic**: webpack/Vite/Angular lazy load work the same way
- **Transport**: REST/JSON, GraphQL, WebSocket, SSE; gRPC-web out of scope
- **SSR**: client-side fetch is recordable; RSC/Server Actions not fully enumerable
- **Blind spots**: JSVMP, WASM, strong HMAC/mTLS checks → static plus marked limitations
- **Param blind spots**: conditional linkage, hidden params, WASM assembly → "pending trigger" / "unreachable"
- **Static is the safety net**: when runtime is blocked, static still enumerates endpoints

---

## Additional Resources

- Grep recipes, `config.json` template, troubleshooting, Hook, param reverse engineering section J, site_map template: **[reference.md](reference.md)**
- Reference script paths in the [Scripts and Gates](#scripts-and-gates) table
