# api-recon — reference manual

Grep recipes, `config.json` template, troubleshooting. All greps run against the `js/` directory. If a bundle is single-line, run `js-beautify` or `sed 's/}/}\n/g'` first; usually raw grep with a context window is enough.

## Script notes

Every file in `scripts/` is a **reference template**. Adjust it for the target site before running. Typical change points:

| Script | Common changes |
|---|---|
| `harvest_static.py` | endpoint regex, webpack/Vite manifest parsing, micro-frontend publicPath, retry/concurrency |
| `runtime_harvest.js` | neutralize field names and success values, stub match rules and body shape, routes source, WS recording, `waitUntil`/`routeTimeout`/`proxy` |
| `preload.js` | `loginPathRe`, L1 stubs, `neutralize.fields`, `apiPattern`, whether L3 is enabled, `recordDetail`, `observe.*`, `neutralizeVueRouter` |
| `spider_mpa.py` | `--exclude` destructive links, cookie, depth/max, same-domain filter |
| `extract_route_map.py` | `routeMap` / `routeLink` regex, KEY naming pattern |
| `build_perm_tree.py` | `userRouteAuth` parsing, `ROOTS`/`PREFIX_PARENT` hierarchy heuristics, stub outer field names |
| `config.json` | single entry point for all site-specific params above |

Put adjusted files in the task working directory (e.g. `recon/`). State the exact changes against the reference scripts in the report.

---

## A. The three reverse-engineering gates

### A1. Render gate — "how do I tell a session is logged in?"

```bash
grep -rhoaE '.{0,40}(isLogin|isAuthenticated|loggedIn|hasLogin|requireAuth)\b.{0,80}' js | head
grep -rhoaE 'function (getUser|getToken|getAuth)[0-9]?\([^)]*\)\{.{0,200}' js | head
grep -rhoaE '(localStorage|sessionStorage)\.getItem\("[^"]+"\)' js | sort -u
grep -rhoaE '(Cookies?|cookie)\.(get|load)\("[^"]+"\)' js | sort -u
grep -rhoaE '\batob\(|JSON\.parse\(|jwt|decode' js | head
```

Find the chain `isLogin = f(getUser())` → `getUser = decode(storage.read(KEY))`. Pin down the **storage key**, the **container** (Cookie vs localStorage) and the **encoding**:

| Encoding | config forge method |
|---|---|
| plain string / `"1"` / token | `"value": "anything-truthy"` |
| `JSON.parse(x)` | `"value": "json:{\"id\":1,\"username\":\"admin\"}"` |
| `JSON.parse(atob(x))` | `"value": "b64json:{\"id\":1,\"username\":\"admin\"}"` |
| JWT | unsigned / `alg:none` JWT, or sign with a key found in the bundle |
| encrypted (SM2/AES/RSA) | find the hardcoded key; forge when the render gate only needs a decodable blob; otherwise static fallback |

→ Write into `cookies` / `localStorage`.

### A2. Interceptor gate — "what triggers the /login redirect?"

```bash
grep -rhoaE '.{0,60}(interceptors\.response|axios|request\.use).{0,120}' js | head
grep -rhoaE '.{0,40}(response_code|errcode|errno|\bcode\b|\bret\b|\bstatus\b)\s*[=!]==?\s*[\-0-9]{1,4}.{0,60}' js | head -20
grep -rhoaE '.{0,40}(not logged in|please log in again|login expired|unauthorized|login invalid|authorization|token.{0,10}invalid|未登录|请重新登录|登录已过期|登录失效).{0,40}' js | head
grep -rhoaE '.{0,30}(location\.href|router\.(push|replace)|navigate)\([^)]*login[^)]*\)' js | head
```

Determine: **field name**, **success value** (usually `0` or `200`), **failure value that triggers the redirect**. Verify with a junk session:

```bash
curl -sk -X POST -H 'Cookie: <fakekey>=junk' https://target/api/<protected> -d '{}' -H 'Content-Type: application/json'
```

→ Write into `neutralize.fields` + `neutralize.success`.

### A3. Content gate — "where do menus/permissions come from?"

```bash
grep -rhoaE '"/api[^"]*(permission|perm|role|menu|acl|resource|nav)[^"]*"' js | sort -u
grep -rhoaE '.{0,30}(menus|permissions|menuList|routeList|authList|role_permissions)\b.{0,120}' js | head
grep -rhoaE 'userRouteAuth|getResultTree|routeMap|routeLink|hasPermission|checkAuth' js | head
grep -rhoaE '([A-Z_][A-Z0-9_]*):\{name:"[^"]*",link:"/[^"]+"\}' js | head
```

**Two data layers** (common in enterprise admin panels):

| API | typical payload | consumer |
|---|---|---|
| `.../role_permissions` | `{ permissions: string[], role_type }` | route guard, button-level ACL |
| `.../permissions/all` | `tree[{ code, position, children }]` | sidebar menu render |
| `userRouteAuth` in bundle | `{ CODE: { url, name? } }` | code → frontend path |
| `routeMap` in bundle | `{ KEY: { name, link } }` | alias resolution (webpack `o.DASHBOARD`) |

Read the consumer code to confirm: how `getResultTree(tree, permissions)` filters, which field `v-if` / `hasAuth(code)` checks.

**Manual forge** (small sites): build a permissive payload → `stubs`.

**Full permission-tree restore** (large sites, sidebar/submodules still blank): see **section I**.

---

## B. config.json template

```json
{
  "baseUrl": "https://target/",
  "runtimeMode": "both",
  "chromium": "/usr/bin/chromium",

  "cookies": [
    { "name": "auth", "value": "b64json:{\"id\":1,\"username\":\"admin\",\"role\":\"admin\",\"func\":{},\"permissions\":[\"*\"]}" }
  ],
  "localStorage": { "token": "faketoken", "isLogin": "1" },

  "neutralize": {
    "fields": ["response_code", "code", "errno", "ret", "status"],
    "success": 0,
    "flags": { "success": true, "message": "ok" }
  },
  "forward": true,
  "loginUrlPattern": "/login",
  "apiPattern": "/api/|/rest/|/graphql",

  "mockTier": "L1+L2",
  "recordDetail": true,
  "observe": {
    "storageReads": false,
    "cookieReads": false,
    "xhrHeaders": true
  },
  "neutralizeVueRouter": true,
  "stubs": [
    {
      "match": "permissions/all|/menu|role_permissions",
      "body": {
        "response_code": 0, "code": 0,
        "data": {
          "permissions": ["*"],
          "menus": [
            { "name": "dashboard", "path": "/dashboard", "show": true, "children": [] },
            { "name": "alert", "path": "/alert", "show": true, "children": [] }
          ]
        }
      }
    }
  ],

  "explore": {
    "clickTabs": true,
    "clickTables": true,
    "pushStateFallback": true,
    "maxMenuItems": 50
  },

  "routes": ["/dashboard", "/alert", "/asset", "/device", "/report", "/config", "/system"],
  "waitMs": 1500, "perRouteMs": 900, "headless": true,
  "waitUntil": "domcontentloaded",
  "routeTimeout": 12000,
  "proxy": "",

  "captureResponses": true, "recordWs": true, "respMax": 600
}
```

Field notes:
- `runtimeMode`: `depth` (Puppeteer), `coverage` (browser MCP), `both`
- `cookies[].value` prefix: `b64json:` → base64(JSON); `json:` → raw JSON; no prefix → literal
- `forward: true` forwards real requests and rewrites the code fields; `false` fully offline stub
- `mockTier`: preload tier enabled in coverage mode, e.g. `L1+L2`, `L1+L2+L3`
- `routes` comes from `routes.txt`; after forging menus the harness auto-appends `<a href>`
- `captureResponses` / `recordWs` only take effect in depth mode
- `waitUntil`: use `domcontentloaded` for large SPAs, avoid `networkidle2` hangs
- `routeTimeout`: per-route `page.goto` timeout (ms)
- `proxy`: Puppeteer `--proxy-server`; `HTTP_PROXY` / `HTTPS_PROXY` also work

### B1. Dual stub template (role_permissions + permissions/all)

```json
"stubs": [
  {
    "match": "role_permissions",
    "body": {
      "response_code": 0,
      "data": {
        "permissions": ["MONITOR", "MONITOR_ALERT", "THREAT", "ASSETS_RISK"],
        "role_type": "SUPER_ADMIN"
      }
    }
  },
  {
    "match": "permissions/all",
    "body": {
      "response_code": 0,
      "data": [
        {
          "code": "MONITOR",
          "position": 1,
          "children": [
            { "code": "MONITOR_ALERT", "position": 1, "children": [] }
          ]
        }
      ]
    }
  }
]
```

Outer field names (`response_code` / `code` / `data`) must match the A2 interceptor gate; `permissions` must cover every leaf code in the tree.

---

## C. coverage mode: preload config

Edit the `CONFIG` object at the top of `scripts/preload.js`, or replace it before CDP injection:

```javascript
const CONFIG = {
  loginPathRe: /\/(login|signin)(\/|$|\?)/i,
  mockTier: 'L1+L2',
  forward: true,
  recordDetail: true,
  extractUrlsFromResponse: true,
  neutralizeVueRouter: true,
  observe: { storageReads: false, cookieReads: false, xhrHeaders: true },
  neutralize: { fields: ['response_code', 'code'], success: 0 },
  stubs: [ /* same as config.json stubs */ ],
  apiPattern: /\/(api|apis|v\d+|dev|internal|graphql)\//i,
};
```

Verify: `window.__API_RECON_PRELOAD__ === true` and the pathname is stable.

Export the recording:

```javascript
JSON.stringify({
  apis: [...window.__API_RECON_LOG__],
  detail: window.__API_RECON_DETAIL__,
  routes: [...(window.__API_RECON_ROUTES__ || [])],
  observe: window.__API_RECON_OBSERVE__,
}, null, 2)
```

---

## D. preload / runtime Hook capabilities

Browser Hook capabilities built into preload (coverage) and runtime_harvest (depth), and what they cover:

| Hook capability | Value for API discovery | Coverage |
|---|---|---|
| Hook fetch / XHR.open | record request URL/method | ✅ `recordDetail` + `__API_RECON_LOG__` |
| Hook XHR.setRequestHeader | discover Authorization and other headers | ✅ `observe.xhrHeaders` |
| Hook localStorage/cookie reads | confirm session key names | ⚠️ optional `observe.storageReads/cookieReads` |
| Vue route access | complete frontendRoutes | ✅ `__API_RECON_ROUTES__` (loaded routes) |
| Neutralize Vue route guard / block login redirect | open up modules to trigger APIs | ✅ `neutralizeVueRouter` + native redirect neutralization |
| React route access | complete routes | ⚠️ static + clicks; no dedicated Hook |
| Page redirect block (login path) | stay on page for analysis | ⚠️ blocks login path only, avoids blocking business navigation |
| Hook crypto libs (CryptoJS/SM etc.) | encrypted params → plaintext API body | ❌ must manually Hook crypto function inputs; write findings to config |
| Anti-debug bypass | otherwise runtime records no APIs | ❌ manual handling; static still works |

---

## E. Endpoint extraction regex (when static yield is too low)

Loosen `extract_endpoints` in `harvest_static.py`, or do it manually:

```bash
grep -rhoaE '"/[a-z][A-Za-z0-9_/\-]{3,}"' js | sort -u
grep -rhoaE '/api/[a-zA-Z0-9_./-]+' js | sort -u
```

---

## F. Troubleshooting

| Symptom | Cause → fix |
|---|---|
| Few static APIs | endpoint dialect mismatch → loosen the regex (section D) |
| chunk count ≪ manifest | CSS-only or undeployed chunk; 404s already retried |
| runtime still shows the login page | render gate wrong → recheck A1: key name, container, encoding, domain |
| shell loads but modules blank | content gate → forge menus (A3); `routes` paths may be wrong |
| only bootstrap/locale per route | permission codes incomplete → section I permission-tree restore; check the `role_permissions` + `permissions/all` dual stub |
| sidebar items present but subpages blank | tree missing intermediate nodes, or code mismatched with `userRouteAuth` |
| every API redirects to login | interceptor gate → verify `neutralize`; nested fields need extended walk logic |
| 0 WS frames | subscribe only after user interaction; raise `perRouteMs` |
| empty response body | real responses only exist with `forward: true` |
| Chromium missing | install chromium or set `config.chromium` / `CHROMIUM` |
| heavy mocking but still back to login | Hook too late or missing `location.href` setter → document-start + preload |
| all lists empty | empty arrays at L3 are normal; keep clicking Tabs/settings/details |
| Redux action mistaken for a route | filter internal paths containing get/set/change/clear/toggle/upload |
| Vue still redirects to login | preload not at document-start → change injection timing; or clear guards manually when `neutralizeVueRouter: false` |
| URLs in responses missing from log | enable `extractUrlsFromResponse`; or extract manually from `__API_RECON_DETAIL__` |
| Authorization header name unknown | enable `observe.xhrHeaders` or check request headers in DevTools |
| runtime very slow / timing out | set `waitUntil: domcontentloaded`; lower `routeTimeout`; do not use `networkidle2` |
| proxy connection failed | check `proxy` / env vars; keep Puppeteer and curl on the same proxy port |

---

## G. hardened targets

When the server validates the session step by step (unforgeable signed cookie, server-rendered menus that cannot be stubbed), runtime stalls at the shell. Expected behavior:

- **Static is enough for endpoint enumeration** — module paths live in the code
- If authorization allows, run the same harness with a **real session**: `forward: true`, no neutralize, capture real methods/params/responses

---

## H. Single-task checklist

1. Confirm the authorization scope
2. **Read** `scripts/harvest_static.py` → adjust for the target → run → review `api_static.txt`, `routes.txt`
3. **Phase 1b**: path anchor window expansion + binding layer → `param_candidates.json` (section J)
4. Reverse A1/A2/A3 → write the site-specific `config.json`
5. **Read and adjust** `runtime_harvest.js` / `preload.js` before running
6. `runtimeMode=depth`: `npm install` → run the adjusted harvest script
7. `runtimeMode=coverage/both`: inject the adjusted preload at document-start → browser MCP dynamic enumeration + **param trigger matrix**
8. Modules not rendering → **section I permission-tree restore** → patch stubs → rerun
9. Multi-sample param diff + reverse inference from errors → `params_merged.json`
10. Merge → `site_map.json` + `api_merged.txt`; honestly mark coverage, gaps and script change points

---

## I. Permission-tree restore (Phase 4 deep dive)

Use when a simple forge `menus: [{ path, show: true }]` fails and submodules still do not mount.

### I1. Locate the auth module

```bash
grep -l 'userRouteAuth' js/*.js
grep -l 'routeMap\|routeLink' js/*.js
grep -rhoaE 'getResultTree|role_permissions|permissions/all' js | head
```

Record: **permission API path**, **response field names**, **consuming chunk filename**.

### I2. Extract routeMap

```bash
python3 scripts/extract_route_map.py recon/js recon/
# produces recon/route_map.json
```

If `[!] no routeMap pattern found`: loosen the regex in `extract_route_map.py`, or grep manually:

```bash
grep -rhoaE '([A-Z_][A-Z0-9_]*):\{name:"[^"]*",link:"/[^"]+"\}' js | head -20
```

### I3. Build the permission tree + stub

```bash
python3 scripts/build_perm_tree.py recon/js recon/ --config recon/config.json
```

Script logic:
1. Parse `userRouteAuth={MONITOR:{url:...},...}` (including webpack aliases like `He=o.DASHBOARD`)
2. Resolve alias → real path via `route_map.json`
3. Infer parent from the code prefix (`MONITOR_ALERT` → `MONITOR`)
4. Emit `permissions_tree.json`, `permissions_all_stub.json`, `role_permissions_stub.json`
5. With `--config`, write `stubs` into `config.json` and extend `routes` automatically

**Adjust per target** (top of the script):
- `DEFAULT_ROOTS`: top-level module code list
- `DEFAULT_PREFIX_PARENT`: `PREFIX_` → parent mapping
- `DEFAULT_EXTRA_PARENT`: orphan nodes not covered by prefix rules

### I4. Verify stub consistency

```bash
# permissions count should ≈ userRouteAuth entry count
wc -l recon/perm_codes_all.txt
# routes should cover every link in route_map
python3 -c "import json; m=json.load(open('recon/route_map.json')); r=set(json.load(open('recon/config.json'))['routes']); print('missing', [v['link'] for v in m.values() if v['link'] not in r])"
```

### I5. Rerun runtime and compare

```bash
node recon/runtime_harvest.js recon/config.json
# compare runtime_api.json counts before vs after forge; check whether /attack, /asset etc. now expose module APIs
```

| before forge | after forge (success) |
|---|---|
| same 3–5 bootstrap entries per route | different routes trigger different module APIs |
| only `/api/locale/language` | module endpoints such as `/api/web/...` appear |
| `routes.txt` has single-digit routes | `routes` 80–110+ from route_map |

### I6. Still failing

- **coverage mode**: click the sidebar + Tabs; permission gating may only request after interaction
- **stub fields**: compare the nesting of the real API (curl + real session) against the stub
- **extra guards**: grep `hasPermission|checkRole|func.` and other button-level checks, extend `role_permissions.permissions`
- **static fallback**: module API paths are still in `api_static.txt`; runtime only fills in METHOD/body; keep params in `param_candidates.json` + recorded samples

---

## J. Param reverse engineering (Phase 1b / 5b / 5c)

**Methodology, not a generic script.** Find paths with regex; find params with anchor window expansion + UI binding chain + multi-sample diff + reverse inference from errors.

### J1. Anchor window expansion — from path to the packaging object

```bash
# anchor on a known Phase 1 path
grep -n '"/api/user/list"' js/*.js
grep -rhoaE '.{0,120}("/api[^"]+").{0,200}' js | head
grep -rhoaE '(params|data|body|payload)\s*:\s*\{' js | head
grep -rhoaE '(get|post|put|delete|patch)\([^,]+,\s*\{' js | head
```

### J2. Wrapper layer and transport shape

```bash
# axios / unified request
grep -rhoaE '(axios|request)\.(get|post|put|delete|patch)\(' js | head
grep -rhoaE 'interceptors\.(request|response)' js | head

# GraphQL
grep -rhoaE '(query|mutation)\s+\w+|gql`|graphql\(' js | head
grep -rhoaE '\$[a-zA-Z_]+\s*:\s*(Int|String|Boolean|\[)' js | head

# FormData / multipart
grep -rhoaE 'FormData|\.append\(' js | head

# path params
grep -rhoaE 'path:\s*"/[^"]*:[^"]+"' js | head
grep -rhoaE 'useParams|route\.params|\$route\.params' js | head
```

### J3. Validation gate — required / format / enum

```bash
grep -rhoaE '(required|message|pattern|enum|validator)\s*:' js | head
grep -rhoaE 'yup\.|zod\.|async-validator|Form\.Item|a-form-item|el-form-item' js | head
grep -rhoaE 'rules\s*:\s*\[|name:\s*["\'][a-zA-Z_]+["\']' js | head
grep -rhoaE 'label.*value|options\s*:\s*\[' js | head
```

### J4. Binding layer — form → API

```bash
grep -rhoaE 'onFinish|handleSubmit|getFieldsValue|validateFields' js | head
grep -rhoaE '(pick|omit|transform|dayjs|moment)\(' js | head
```

runtime fill-in: DevTools → Network → request → **Initiator** (call stack); walk up from `fetch`/`send` to the packaging function.

### J5. Encrypted params

```bash
grep -rhoaE 'encrypt|decrypt|sign|CryptoJS|sm2|sm3|sm4|RSA|AES' js | head
```

**Do not guess fields from ciphertext** — Hook the crypto function **inputs** and record the plaintext payload before encryption; write findings to `config.json` / `param_candidates.json`.

### J6. Param trigger matrix (mandatory in Phase 3)

Record each operation once per module, then diff request body/query:

| Operation | Watch |
|---|---|
| list first screen | pagination defaults |
| search | keyword, filters |
| advanced filter | optional fields |
| create/edit | full entity |
| batch/export | `ids[]`, `exportType` |
| sort/paging | `sortField`, `order` |

Output `param_samples.json`: `[{ "path", "method", "action": "search", "body", "query", "headers" }]`

### J7. Confidence rules

| Confidence | Condition |
|---|---|
| **high** | static callsite + ≥2 consistent runtime samples |
| **medium** | static only, or a single runtime hit |
| **low** | inferred from responses/errors, not re-verified |
| **pending trigger** | field known statically, UI/permissions never reached it |

### J8. Scenario quick config

| Scenario | Order |
|---|---|
| REST list page | J1 packaging object → J6 four diffs → J3 rules |
| create/edit form | J3 Form name → J4 submit chain → runtime submit + deliberate blank to see 400 |
| GraphQL | J2 variables declaration → runtime record variables per operation |
| encrypted body | J5 Hook inputs → pre-encryption fields are the real params |

### J9. Mapping to api-recon phases

| api-recon | param recon |
|---|---|
| Phase 1 static | J1 anchor window expansion |
| Phase 2 A2 interceptor | global injected fields (tenantId, sign) |
| Phase 3 runtime | J6 trigger matrix + `param_samples.json` |
| Phase 4 permission tree | different modules, different forms → full fields only trigger with enough permission |
| Phase 5 merge | `params_merged.json` + confidence; never mark a field required from a single sample |

### J10. Troubleshooting

| Symptom | Fix |
|---|---|
| field name present statically but never at runtime | mark as "pending trigger"; complete the permission tree / click advanced filters / drive linked selects through each option |
| same path, different body shapes | normal — record one entry per `action`, do not force-merge the schema |
| stub response is fake but params are needed | **read the outbound request** body/headers, do not infer from the stub response |
| 400 reports a nested field | watch the outer wrapper `data`/`bizData`/`variables` |
| GraphQL shows only the operation name | expand the `variables` JSON; find `$var: Type` statically |

---
