# ARTEX security audit

## Verdict: Action Required

Revision inspected: `e150b9f7a14dc65b78e098cd1bee2d27283b9a79`, plus the current working tree. This is a local source audit and targeted remediation, not a malware-free certification or a completed exhaustive dependency/binary review. No confirmed malware was identified. No files or stored evidence were deleted. Existing untracked skills were preserved.

## Threat table

Line references below describe the original audited source unless stated otherwise. Open findings are not labeled Removed/Cleaned.

| Severity / threat | File and line | Mechanism | Status |
| --- | --- | --- | --- |
| High: crawler cookie disclosure | `skills/api-recon/scripts/spider_mpa.py:26-30,101` | Prefix-based origin checks accept lookalike hosts; urllib forwards the supplied Cookie on cross-origin redirects. Reproduced with two local HTTP servers. | Cleaned: parsed origins and same-origin redirects; regression tests added |
| High: MCP credential disclosure | `mcphttp/client.go:182-207,503-510` | Remote SSE endpoint announcement can change destination while retaining configured credential headers. | Cleaned: same-origin endpoint/redirect restrictions; regression tests added |
| High, conditional: policy failure opens execution | `intercept/intercept.go:165-169,196-199,244-247,508-512`; `guard/guard.go:87-89`; `server/assembly.go:340-345` | Policy-load failure can become disabled interception/no match; tool-config read failure returns original tools. | Open: needs fail-closed policy handling and database-backed tests |
| High, deployment-dependent: first-run takeover | `server/auth.go:123-125,163-205` | Public setup lets the first caller initialize an unclaimed instance. | Open: require local setup or bootstrap proof |
| Medium: sessions survive password change | `server/auth.go:22,86-103,249-258` | Password update does not revoke already signed seven-day tokens. | Open: persistent session-generation/revocation design needed |
| High: deployment provenance mismatch | `docker-compose.yml:18`; `.github/workflows/release.yml:123` | Compose pulls `tomicdd/artex`, while reviewed CI publishes `autumn27/artex`. Container receives sensitive configuration and writable mounts. | Open: confirm intended publisher and reviewed digest; no evidence either publisher is malicious |
| Medium: update authenticity gap | `selfupdate/stage.go:65-99`; `selfupdate/bootstrap.go:175-177` | Binary and checksum share release authority; downloaded executable runs in smoke test before normal restart. | Open: signed manifest/pinned provenance needed |
| High for sensitive data: raw capture and retained evidence | `traffic/traffic.go:435-482`; `llmrec/capture.go:57-119`; `evidence/store.go:146-178,347-415`; `db/tasks.go:631-646` | Optional raw headers/bodies and evidence may retain credentials or personal data; task deletion is selective, and usage ledger survives. | Open: approved retention/redaction policy and explicit deletion scope required |
| Medium: intentional outbound data | `agent/provider.go:253-283`; `llmpool/pool.go:73-92`; `notify/webhook.go:37-55,102-140`; `enrich/enrich.go:67-72,117-122` | Configured providers/fallbacks, notifications and DNS receive task-related information. | Retained intentional functionality; review approved destinations |
| Medium: weak transport/log privacy defaults | `config/config.go:151-176`; `agent/provider.go:422-449`; `skills/api-recon/scripts/spider_mpa.py:21`; `skills/api-recon/scripts/runtime_harvest.js:35,97-102` | Assembled database config defaults to no TLS; provider error bodies/URLs can enter logs; recon disables certificate verification/browser sandbox. | Open: explicit secure defaults/opt-in policy needed |
| Medium: mutable installation/build inputs | `install.sh:22`; `Dockerfile:13,20,28-30`; `.github/workflows/release.yml:6-7,14,92,127` | Remote shell installers, latest packages/tagged images/actions, broad workflow write permissions. | Open: pin and verify inputs, narrow permissions |
| Medium: untracked skills packaged locally | `build.sh:183`; `Dockerfile:40`; `docker-compose.yml:39` | Local distributions include working-tree skills beyond reviewed commit. | Open: review/package allowlist; original files preserved |
| Low: executable environment file | `reset-password.sh:196-198` | Shell sources `.env`, which can execute arbitrary local shell commands. | Open: parse allowed keys as data |

## Five-phase coverage

### 1. Mining and resource hijacking

Targeted searches found no named XMRig, Stratum, CoinHive, Claymore, CryptoNight or Monero indicators in searched source. Reviewed startup restart loops and development workers have explicit service roles. No hidden miner was established. This does not rule out unknown implementations or malicious installed dependencies.

### 2. Backdoors and execution

Reviewed authentication, execution/configuration paths, MCP transport, interception and updater trust. Shell/Python custom tools, stdio MCP startup and reverse-shell support are explicit pentest functionality, not sufficient evidence of malware. No production master password or fixed attacker callback was found in reviewed source. Base64 and optional UPX usage were not automatically classified as malicious. Application permission bypass and prompt-based scope instructions do not provide OS/network containment.

### 3. Telemetry and retention

No dedicated Google Analytics/Baidu tracker, browser geolocation collector or college/student schema was identified in inspected source. Fonts and theme bootstrap are local. Generic task content and captured bodies can still contain student records. Traffic and full LLM recording have false defaults in `server/manager.go:412-414`; disabling local recording does not stop sending context to the selected LLM. Provider-side retention, live settings, exported records and backups remain unverified. Build verification explicitly disabled Next.js telemetry with `NEXT_TELEMETRY_DISABLED=1`; this is not a permanent application configuration change.

### 4. Archives and supply chain

Reviewer inventoried 607 tracked paths: no tracked executable, shared library, WASM or archive identified. PNG/ICO files are present; two NUL delimiters in a TSX deduplication key are not an executable payload. No tracked submodules or symlinks were identified. Local ignored `artex`, `dist`, generated bundles and installed dependency contents were not exhaustively reverse engineered.

Frontend lock: 772 non-root entries, npm registry URLs with integrity fields. Recon lock: 85 entries, npmmirror URLs with integrity fields. Go requirements have checksum entries and no replace directive; `norma v0.4.3` is a critical orchestration dependency. No obvious name-based typosquat was identified. This is not registry-owner verification, an advisory-database scan, or proof of uncompromised package contents. No external scanner received source or secrets.

80 pre-existing untracked files were Markdown (about 3.57 MB), reviewed by inventory and targeted searches rather than exhaustive semantic inspection. Offensive examples were treated as data, never followed as instructions.

### 5. Remediation and verification

Crawler regression reproduced Cookie arrival at the second local server before the patch. After remediation, no cross-origin request occurs; same-origin redirects still carry the cookie; lookalike hosts, userinfo URLs, downgraded schemes and different ports are rejected. Three Python tests pass. Initial test invocation timed out under workstation proxy settings; tests now exclude local fixtures from proxies.

Backend build passed: `GOPROXY=off GOSUMDB=off go build -o <approved-temp>/artex-security-audit ./cmd/artex`. This is the non-embedded backend build; the existing local binary was not overwritten.

Frontend static build passed: `NEXT_TELEMETRY_DISABLED=1 npm run build:static`, including TypeScript and 31 generated pages. Next warned that static export disables middleware/API routes. Complete `embedui` binary packaging was not run.

Frontend unit tests: `node --test web/src/lib/*.test.mjs`: 7 passed, with Node module-type warnings.

Full Go suite: `GOPROXY=off GOSUMDB=off go test -timeout 90s ./...`: 591 passed, 22 failed, 236 skipped in the reported summary. Every reported failure below cites missing PostgreSQL configuration. No configured/live database was used. This run occurred while the independent MCP remediation was still in progress; targeted and final results are recorded separately.

| Failing test group | Names |
| --- | --- |
| Finding traffic | TestFindingTrafficToolUpgradePreservesCustomization; TestFindingTrafficAPIAndExport; TestFindingTrafficArchiveV3RoundTripAndRetry; TestFindingTrafficFailedReportDoesNotTrigger; TestFindingTrafficUTF8SegmentsAndInheritedWrites |
| Finding workflow | TestFindingWorkflowAutoHintToPlannerAndSetting; TestFindingWorkflowMigrationPreservesUserConfiguration; TestFindingWorkflowReporterBindsBeforeWritingReport |
| Notifications A | TestNotifyEndToEndRealtimeDelivery; TestNotifyChannelAPIMasksSecretsAndPreservesOnUpdate; TestNotifyChannelAPICreateValidation; TestNotifyFilterBlocksBelowThreshold; TestNotifyDigestBatchesMultipleFindingsIntoOneMessage |
| Notifications B | TestNotifyDisabledChannelDoesNotSend; TestNotifyStatusChangeDelivery; TestNotifyStatusChangeSuppressedByDefault; TestNotifyTestMessageEndpoint; TestNotifyDeliveriesHistoryAndRetry |
| Notifications C | TestNotifyMetaAndSettingsRoundTrip; TestNotifyDeepLinkUsesPublicBaseURL; TestNotifyNoDeepLinkWithoutBaseURL; TestNotifyDigestSegmentsAndDefersRemainder |

Frontend Biome check failed on untouched frontend files: 84 errors, 138 warnings, 734 informational diagnostics across 141 files. No formatting auto-fix was applied. A clean build does not imply a clean lint check or a passing integration suite.

## Remaining decisions

The open findings above remain actionable vulnerabilities or trust gaps, not confirmed malicious code. Silently selecting a Docker publisher, deleting stored evidence, removing external LLM support or changing first-run authentication would require decisions not established by this audit. Use an isolated disposable PostgreSQL database for remaining tests. Do not point destructive integration tests at a production database.

## Final targeted verification and patch

Final full-suite rerun after both fixes: `GOPROXY=off GOSUMDB=off go test -count=1 -timeout 90s ./...` reported **612 passed, 22 failed, 236 skipped**. The same 22 tests listed above failed for absent PostgreSQL configuration. The backend was rebuilt successfully after the final MCP changes; `git diff --check` also passed.

MCP remediation was independently diff-reviewed and rerun with `GOPROXY=off GOSUMDB=off go test -count=1 -timeout 30s ./mcphttp/...`: 22 reported passing tests/subtests. Four cross-origin local-server cases failed before the fix. Coverage includes SSE-announced endpoints, initial SSE/HTTP redirects, message redirects, URL constraints, same-origin credential preservation, and effective-response relative endpoint resolution. Cross-origin configurations previously accepted are deliberately rejected now.

The exact code/test patch is saved in `docs/security-fixes.patch`, including new untracked test files. Audit documentation is separate. No code was removed on a malware classification, because no such classification was confirmed. Remaining vulnerabilities mean the requested complete cleanup and zero-breakage verification are not established.

## Delivery self-check

| Axis | Score | Evidence / improvement |
| --- | --- | --- |
| Accuracy | 4/5 | Two local-server regression families verified; other source findings remain unverified against live deployment. |
| Completeness | 2/5 | All five audit areas addressed, but dependency internals, ignored binaries, deployed images and complete remediation remain incomplete. |
| Clarity | 4/5 | Open and cleaned issues separated; original line references require consulting the patch after edits. |
| Actionability | 4/5 | Exact patch and commands supplied; PostgreSQL and publisher/retention decisions still needed. |
| Conciseness | 4/5 | Detailed evidence kept in this report; table is necessarily longer than the final summary. |

Overall: 3.6/5. Highest-impact follow-up is fail-closed policy enforcement with disposable-database tests, followed by trusted image provenance and retention controls. A truly exhaustive audit requires package/binary review and runtime verification. The partial-completion assessment reflects the requested scope rather than claiming a complete cleanup.
