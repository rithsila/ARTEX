# Side Question Long Conversations & Context Budget

Fixed on 2026-09-11. The original implementation treated the character count of request JSON directly as tokens and inherited the main task's 32K output reservation, causing normal side questions to be prematurely rejected when HTML, JS, or tool results were voluminous.

## Open-Source Implementation Comparison

- [Grok CLI Side Question Context](https://github.com/superagent-ai/grok-cli/blob/fb97af83f06dca873281d60168430f06c8de6324/src/agent/agent.ts#L739): Extracts snippets from recent user and assistant text, with a character budget around 2,000 and at most 400 characters per message. Continuous side question history is not maintained on this path.
- [Grok CLI Standalone Request](https://github.com/superagent-ai/grok-cli/blob/fb97af83f06dca873281d60168430f06c8de6324/src/utils/side-question.ts): Independent cancellation signal; output capped at 2,048 tokens when supported by the model; no tools provided.
- [Grok CLI Main Session Compaction](https://github.com/superagent-ai/grok-cli/blob/fb97af83f06dca873281d60168430f06c8de6324/src/agent/compaction.ts): Estimates tokens, retains recent content, folds new content into prior summaries, and handles cross-turn truncation.
- [OpenCode Session Compaction](https://github.com/anomalyco/opencode/blob/b3f1a96c6dd7adeb28b36dd11add1998fc84d67b/packages/core/src/session/compaction.ts): Full request estimation, output/buffer reservation, recent content plus rolling summary, toolless summary requests; defaults to 8,000 recent budget and 4,096 tokens max summary output.
- [OpenCode Overflow Recovery](https://github.com/anomalyco/opencode/blob/b3f1a96c6dd7adeb28b36dd11add1998fc84d67b/packages/core/src/session/runner/llm.ts): Attempts overflow recovery only when assistant output has not yet started; retried calls never re-enter the overflow recovery path.

ARTEX incorporates independent output budgets, recent content with rolling summaries, and bounded recovery. It preserves norma v0.3.6's structured messages and tool pairing without adopting Grok's raw text snippet extraction; it does not record OpenCode-style main session compaction events into the ARTEX main transcript.

## Request Budget & Execution

- Messages follow norma's UTF-8 byte block estimate with a 4/3 safety margin; additional overhead for system prompts, tool schemas, and message envelope is factored in. The estimate is not an exact model token count.
- Side question output defaults to at most 8,192 tokens and never exceeds the configured limit of the main profile. The server environment variable `ARTEX_BTW_MAX_OUTPUT_TOKENS` can set a ceiling between 256 and 32,768; it does not alter the product's default model or main task parameters.
- The input budget is the context window minus the output limit and safety margin; unknown windows fall back to the platform default of 200K. The safety margin is 5% of the window, clamped between 128 and 8,192 tokens.
- Successful exchanges are loaded in ascending ordinal order in batches of up to 20. At most 20 raw exchanges are retained, whose token budget is capped at 1/4 of the input budget and no more than 16K tokens.
- Excess exchanges are folded into the rolling summary. Summaries include historical origin and context timestamp; historical assistant answers are not equivalent to new tool evidence, and the latest main snapshot takes precedence in conflicts.
- When the main context remains oversized, only older messages in the side question replica are summarized, retaining up to 8K tokens of recent content; truncation points never split tool calls and results. An oversized single exchange enters the summary in its entirety.
- Summary input is chunked safely along UTF-8 boundaries according to actual remaining window capacity, with a 2,048-token output ceiling; empty summaries, truncation, tool calls, or budget overruns are never cached. A single side question invocation permits at most 12 summary calls bounded by the shared 120-second timeout; reaching this limit fails cleanly without infinite loops.
- If the model returns context length exceeded on the first attempt before emitting any text or tool calls, it retries at most once after further compaction; if the estimated size does not decrease, recovery stops immediately. Other model errors and partial streaming outputs do not trigger recovery.
- All incurred usage—including summaries, failed attempts, and cancellations—is aggregated into the side question request. If the provider returns no usage, zero is recorded; estimates are never misrepresented as actual usage.

## Persistence & Interface

`side_question_sessions.memory` persists old exchange summaries, the covered ordinal, and cached main context summaries keyed by snapshot identity. `side_question_requests.context_info` records preparation phase, actual replayed exchange count, summary usage status, and budget estimates.

Summaries are saved only when the original request is still active and clear versions match; clearing wipes the cache simultaneously, preventing late writes from reviving cleared data. New snapshots do not reuse old snapshot summaries. Summary fields are preserved across v3 task archives; missing fields in legacy v3 archives are restored with empty objects, and v1/v2 compatibility is maintained.

POST accepts and returns the request upfront; preparation and compaction run asynchronously in the background without holding admission locks or database transactions. SSE/history displays stages: preparing, folding exchanges, compacting snapshot, answering; compaction failures are saved as terminal failure states for the request. The frontend preserves errors and restores the failed question draft without obstructing the input box with toast notifications; history polling never wipes submission errors.

## Validation Notes

- 19/20/21/50 exchange replay, preservation of earlier decisions across 20+ rounds, summary cache reuse across restarts: automated tests pass.
- Long answers and code contexts, chunked request budgeting, tool pairing, snapshot immutability, cache invalidation on new snapshot: automated tests pass.
- Summary failure/cancellation/truncation/oversize/tool call, clear race conditions, call limit, single overflow recovery, and no-retry on partial streams: automated tests pass.
- Dedicated PostgreSQL pagination, restarts, v1/v2/v3 archiving, summary cache and budget metadata archive restoration, legacy v3 missing fields, 20 parent sessions sharing 4 concurrency slots: passed.
- Go candidate service build, frontend TypeScript checks, Biome checks on modified components, and Next.js production build in an isolated directory: passed.
- Embedded browser using isolated UI fixtures verified preparation phases, summary range indicators, post-failure draft recovery, no floating error toasts, no horizontal overflow, and clean console logs at 1280×720 and 390×844. Temporary fixtures removed.
- Read-only replay of existing local Worker snapshots passed new budget checks; for example, Worker #3's 293,085-character snapshot is no longer falsely rejected by local character count. No external model calls made.
- Real conversation test sending private Worker snapshots to Grok was rejected by automated approval review and was not executed; not marked as passed.
- 2026-09-11 00:37 local backend restarted as requested, reusing existing database, data directory, and auth configuration. Running binary matched candidate SHA-256; `/api/health` on backend and frontend proxy returned healthy.

Validation commands (isolated test database only):

```sh
go test -race ./sidequestion ./db ./server -run 'TestSide|TestCheckpoint|TestSnapshot|TestBuildRequest|TestService|TestMainSide|TestTaskArchive' -count=1
go build ./cmd/artex
npx tsc --noEmit
npm run build -- --webpack
```

Frontend production build used an isolated copy to avoid overwriting the current preview `.next`. Candidate binary was at `/private/tmp/artex-btw-budget-candidate`, copied to `/private/tmp/artex-btw-preview/artex` and launched; original binary backed up as `artex.before-context-budget` in the same directory.
