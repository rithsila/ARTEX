# `/btw` Validation Record

Date: 2026-09-10. Branch: `codex/btw-side-question`. Baseline: `8dae851b9b622f2ff2631f332fde9719d0b16fba`.

Conducted with dedicated PostgreSQL test databases and data directories; real model credentials were injected only into isolated test environments, never written into code or this document, and default product models were unchanged. Go 1.26.3, norma v0.3.6, Next.js 16.2.9.

Actual model conversations, returned response objects, engineering assertions, and raw Qwen review output are stored in [validation-2026-09-10.json](validation-2026-09-10.json), with zero API credentials.

## Engineering Verification Matrix

| Scope | Result | Evidence |
| --- | --- | --- |
| Deep copy of structured messages and tool parameters | Passed | `TestCheckpointDeepCopyAndBoundaries` |
| Compaction/summary requests do not overwrite snapshot; full responses and terminal states published; partial replies excluded | Passed | `TestCheckpointDeepCopyAndBoundaries`, `TestSnapshotExcludesPartialStreamAndSelectsPoolMember` |
| Actual model pool member identity selection | Passed | `TestSnapshotExcludesPartialStreamAndSelectsPoolMember` |
| Tool pairing, 20-exchange replay, budget trimming, and budget overflow errors | Passed | `TestBuildRequestCompactionToolPairingAndBudget` |
| Concurrent main and side sessions, bidirectional cancellation isolation | Passed | Blocking provider, `TestMainSideConcurrencyAndIndependentCancellation` |
| Tool execution denied, streaming/non-streaming, usage recorded on failure | Passed | `TestServiceNoToolsAndUsageOnFailure` |
| Real norma ChatAgent + local Read tool, transcript & activity isolation | Passed | `TestSideActualChatCheckpointToolResultAndTranscriptIsolation`, streaming and non-streaming subcases |
| Persistence, pagination, idempotency, restart recovery with partial answers | Passed | `TestSideHistoryIdempotencyPagingAndRecovery` |
| Clear vs. late writer race conditions, parent resource deletion, version comparisons | Passed | `TestSideClearLateWritersAndDeletedParent` |
| Main Agent / Worker archiving and recovery across v1/v2/v3 | Passed | `TestSideTaskArchiveVersions` |
| Three parent endpoints, authentication, resource ownership, logical worker deletion | Passed | `TestSideHTTPGlobalLimitTaskWorkerAndDeletion`, `TestSideCheckpointPersistsBeforeAdmissionAndRestart` |
| Busy main session allows side questions; independent SSE reconnect/disconnect, cancel, and clear | Passed | `TestSideHTTPBusyIsolationClearAndReconnect` |
| Concurrency limit: 1 per parent session / 4 globally | Passed | Both `TestSideHTTP…` test cases |
| Snapshot persisted before admission, follow-ups after restart, old session cannot spoof snapshot | Passed | `TestSideCheckpointPersistsBeforeAdmissionAndRestart` |
| Rejection when cached profile is deleted or model identity changes | Passed | `TestSideRejectsDeletedOrChangedCachedProfile` |
| Cancellation and drain before archive; terminal answers and usage persisted | Passed | `TestSideTaskDrainPersistsBeforeArchive` |
| Streaming consumer early cancellation records usage exactly once with side attribution | Passed | `TestSideUsageRecordedOnceOnConsumerCancellation` |
| Restored worker runtime after restart / deadline context continues publishing new snapshots | Passed | `TestSideRestoredWorkerRuntimePublishesNewCheckpoint` |
| Race detector across related packages | Passed | Commands below |
| TypeScript typecheck and production build | Passed | `npx tsc --noEmit`, `npm run build` |
| New frontend modules Biome linting | Passed | `biome check`, 3 new modules |

Automated checks can be reproduced by configuring `ARTEX_PG_DSN` pointing to an isolated, disposable test database (never point to production):

```sh
go test -race ./agent ./db ./server ./sidequestion ./llmrec ./llmpool \
  -run 'Test(Side|Checkpoint|Snapshot|BuildRequest|Service|MainSide|CaptureRun|TaskArchive|CompleteForwards|StopIntent|CancelIntent)' -count=1
cd web
npx tsc --noEmit
npx biome check src/lib/side-questions.ts src/hooks/use-side-questions.ts src/components/side-question-workspace.tsx
npm run build
```

Full Go regression is not entirely green: the `server` package has two pre-existing tests that fail during temporary directory cleanup, both reporting `TempDir RemoveAll … directory not empty`:

- `TestInheritedActivityDetailAndRelationDeletion`
- `TestTaskMetadataPatchReturnsRenameAndPin`

Re-running the `server` package in the same isolated environment directly against the unmodified baseline reproduced these exact cleanup failures. The baseline run also exhibited an assertion failure in `TestCoreTaskLifecyclePG` regarding target node counts; the final modified `server` regression did not experience this assertion failure. All other packages passed, and all side question tests and race detection passed. Baseline issues were not marked as passed in this acceptance, nor were existing assertions altered to mask them.

The Next.js build emitted known multiple lockfile / workspace root inference warnings; compilation completed and all pages generated successfully.

## Browser Verification

Manual automation was run using Codex In-app Browser, connected to an isolated local Go service and Next.js development server. Desktop and 390 × 844 mobile viewport checks verified screenshots and browser logs:

- Entering `/btw` while standard chat is active displays both main content and side panel concurrently; desktop sidebar behaves normally.
- Rapid follow-up questions; stopping a side question preserves partially generated content; main flow continues uninterrupted.
- Closing the panel lets requests continue; reopening restores completed answers; refreshing the page with empty `/btw` restores history.
- Narrow viewport Drawer inputs, buttons, history, and close interactions work without horizontal overflow.
- Clearing requires a confirmation dialog; clearing deletes history while preserving main transcript and snapshots.
- Asking questions and switching between Main Agent and two Workers showed proper agent tab separation and no cross-talk in history.
- A blocking local model fixture kept a Worker running; submitting `/btw` from the Worker input box and stopping the side question left the Worker in real-time execution with its own pause button, while saving the partial side answer.
- Browser error and warning logs remained empty.

Controlled fixtures were used to verify concurrent timing without relying on external model generation speeds. During debugging, two Worker runs failed to establish a concurrent window (task finished / answer ended prematurely); after adjusting the fixture, tests were rerun and passed. These initial failed attempts were not counted as valid passes.

## Real Model Conversations

Probed `grok-4.6` first via OpenAI-compatible endpoint `http://127.0.0.1:12580/tingly/openai`. Probe returned HTTP 200, model name `grok-4.6`, and `READY` in 2.82 seconds. Since the primary model was available, fallback chains (Tingly `glm` and Zhipu `glm-5.3`) were not engaged or verified.

| Scenario | Actual Result |
| --- | --- |
| Querying asset, target, and tag during active main session | Returned `redhaze.top`, homepage fetch and summary goals, `BTW-REAL-0910`; side question finished in 16.97s |
| Querying tool evidence after homepage fetch completes in main session | Accurately cited WebFetch 200, curl redirects 301 → 302 → 200, page title; 7.24s |
| Side question asking Bash to create a test file | Denied execution, target file not created; 7.74s |
| Completed side question leaves main context unaltered | Main transcript SHA-256 and main activity logs match exactly; side tool executions count is 0 |
| Follow-up after full stop / restart of Go service | Preserved previous 3 side history items; answered asset, tag, and title directly from persisted snapshot without rerunning Main Agent |
| New session using Grok non-streaming profile | Correctly answered asset and `ATOMIC-0910`; returned and saved usage: input 11,734, output 138, cache_read 11,520 |

In the asset case, the main session used WebFetch and Bash/curl to read a public homepage, landing on `https://id.redhaze.top/home` with title "RedHaze Technology RedHaze Group · Global Comprehensive Group Portal". Bash buffered the response in a local temporary file; no remote write operations were performed. This fact was verified separately from "side question executed no tools".

Main transcript checksum: `e7e61f135a4a120954b539f357e8c4205d7d5cd7460dcaf3dc0fd066463e1d00`.

**Usage limits:** Tingly's Grok streaming response returns no usage frames. Direct verification sending `stream_options.include_usage=true` returned HTTP 200, 12 data frames, and 0 usage frames. Thus, 0 usage in streaming tests indicates lack of provider metrics, not free billing. Non-streaming usage as well as fixture failure/cancellation usage were properly captured.

## Qwen Review

Review model `qwen-flash` evaluated via OpenAI-compatible endpoint `https://dashscope.aliyuncs.com/compatible-mode/v1`, returning HTTP 200. Evaluated the first three real side question exchanges, main session tool evidence, and engineering assertions; returned `verdict: accept`, `concerns: []`, noting answers matched asset, tag, and page read evidence, with tool denials adhering to constraints. Review usage: prompt 6,625, completion 312, total 6,937 tokens.

This Qwen review did not include the subsequent service restart and non-streaming tests. Qwen's summary of "no writes" was slightly broad: the main session's curl did write a temporary response file locally, as documented above. Concurrency, zero tool executions, and transcript isolation were proven via engineering assertions; model review served only as an auxiliary assessment of answer quality.
