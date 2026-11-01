"use client";

// LLM 重试配置的共用件：五层重试各自的「次数 + 间隔」。
//
// 五层从内到外：建连(SDK) → 空响应(SDK) → 同 provider 安全窗口 → 轮询熔断 → 意图重跑。
// 前三层跟着端点走，所以每个模型配置都能覆盖全局默认；后两层是进程级的，只有全局一份。
//
// 所有输入都遵循同一套「留空 = 不配置」语义，与后端 db.RetryRule 一致：
//   次数   空/0 = 用内置默认 | -1 = 关闭这层重试 | >0 = 用这个次数
//   间隔   空/0 = 用这层原本的指数退避 | >0 = 改用这个固定毫秒间隔

import * as React from "react";

import { Loader2Icon, SaveIcon } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api";
import type { LLMRetryOverride, LLMRetryPolicy, LLMRetryRule } from "@/lib/types";

export const ZERO_RULE: LLMRetryRule = { attempts: 0, interval_ms: 0 };
export const ZERO_OVERRIDE: LLMRetryOverride = {
  connect: ZERO_RULE,
  empty: ZERO_RULE,
  stream: ZERO_RULE,
};
const ZERO_POLICY: LLMRetryPolicy = {
  ...ZERO_OVERRIDE,
  breaker: ZERO_RULE,
  intent: ZERO_RULE,
};

type LayerMeta = {
  title: string;
  /** 这层重试发生在哪、由谁执行 */
  where: string;
  /** 什么样的错误会走到这层——具体到状态码，别让人猜 */
  trigger: string;
  /** 长得像但【不】走这层的错误，省得填了没反应还以为是 bug */
  skips?: string;
  desc: string;
  attemptsLabel: string;
  /** 次数留空时的默认值，用于占位符 */
  defAttempts: number;
  /** 间隔留空时的默认策略，用于占位符 */
  defInterval: string;
  /** 次数填 -1 的含义 */
  offHint: string;
};

export const RETRY_LAYERS = {
  connect: {
    title: "Connect Retry",
    where: "SDK · Pre-200 Handshake",
    trigger: "Connection resets, read/write timeouts, DNS resolution failures, and HTTP 408, 429, 500, 502, 503, 504.",
    skips: "Deterministic rejections (400, 401, 403, 404, 413, 422) fail immediately and escalate without retrying.",
    desc: "Replays identical request. Once streaming begins (HTTP 200 received), mid-stream dropouts are handled by outer layers.",
    attemptsLabel: "Retry Attempts",
    defAttempts: 3,
    defInterval: "0.5s→1s→2s exponential (cap 8s)",
    offHint: "-1 = disable retry, escalate on first error",
  },
  empty: {
    title: "Empty Response Retry",
    where: "SDK · OpenAI format only",
    trigger:
      "HTTP 200 with normal finish_reason stop, but response content blocks are completely empty (gateway dropped frames, lost reasoning tokens, sampling drops).",
    skips: "Omit token truncation (max_tokens exhaustion must be resolved by increasing max output tokens).",
    desc: "Re-transmits full prompt payload. Consumes tokens on long contexts; keep attempts conservative.",
    attemptsLabel: "Retry Attempts",
    defAttempts: 2,
    defInterval: "0.5s→1s→2s exponential (cap 8s)",
    offHint: "-1 = deliver empty response as-is",
  },
  stream: {
    title: "Stream Safe Window Retry",
    where: "ARTEX Runtime · Prior to token delivery",
    trigger:
      "Stream interrupted post-connect (after 200 OK): mid-stream drops, provider overloaded, or stream error events (429/5xx) before any tokens were delivered to caller.",
    skips:
      "Quota depletion (402, handled by failover pool), context overflow (413, handled by compaction), and 400/401/403/404/422 rejections are not retried.",
    desc: "Replays request on same profile. Safe to replay because no output was committed or acted upon.",
    attemptsLabel: "Retry Attempts",
    defAttempts: 2,
    defInterval: "0.5s→1s exponential (cap 4s)",
    offHint: "-1 = stream drop immediately escalates to intent rerun",
  },
  breaker: {
    title: "Circuit Breaker",
    where: "ARTEX Runtime · Process-level global",
    trigger:
      "Consecutive transient errors (429, 5xx, socket drops) reach threshold. Deterministic errors (402 quota, 401/403 bad key, 404 missing model) trip breaker on first occurrence.",
    skips: "A single successful request clears the error counter.",
    desc: "Tripped profiles enter cooldown and are bypassed by the failover pool. State persists in DB across restarts.",
    attemptsLabel: "Consecutive Failures to Trip",
    defAttempts: 3,
    defInterval: "1min→5min→30min step backoff",
    offHint: "-1 = never trip on transient errors (deterministic errors still trip)",
  },
  intent: {
    title: "Intent Rerun",
    where: "ARTEX Runtime · Process-level global",
    trigger:
      "Worker finishes with model_error after all inner retry layers are exhausted or after stream failed mid-generation (unsafe to replay silently).",
    skips:
      "Quota depletion handled by profile failover. Tasks being paused, stopped, or finalized immediately bypass backoff.",
    desc: "Re-executes current agent intent from start. Inner layer attempts multiply across reruns.",
    attemptsLabel: "Rerun Attempts",
    defAttempts: 2,
    defInterval: "Fixed 3s",
    offHint: "-1 = no rerun, mark intent blocked on error",
  },
} satisfies Record<string, LayerMeta>;

type LayerKey = keyof typeof RETRY_LAYERS;

/** 毫秒的人话，只用于在输入框旁边回显，免得数零。 */
function humanMs(ms: number) {
  if (!Number.isFinite(ms) || ms <= 0) return "";
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60_000) return `${Number((ms / 1000).toFixed(2))}s`;
  return `${Number((ms / 60_000).toFixed(2))}min`;
}

/** 受控数字输入：空串 ↔ 0，中间态（"-"、"1e"）原样留在本地，不打扰父级。 */
function NumField({
  id,
  value,
  onChange,
  placeholder,
  min,
}: {
  id: string;
  value: number;
  onChange: (n: number) => void;
  placeholder: string;
  min: number;
}) {
  const [text, setText] = React.useState(value === 0 ? "" : String(value));
  // 父级换了一整套值（读取到策略、切换配置）时跟上；自己敲字时不会走到这里，
  // 因为那时 value 已经等于本地文本 parse 后的结果。
  React.useEffect(() => {
    const incoming = value === 0 ? "" : String(value);
    setText((cur) => (Number(cur || 0) === value ? cur : incoming));
  }, [value]);
  return (
    <Input
      id={id}
      type="number"
      min={min}
      className="w-28 shrink-0"
      value={text}
      placeholder={placeholder}
      onChange={(e) => {
        setText(e.target.value);
        const n = Number(e.target.value);
        onChange(e.target.value.trim() === "" || !Number.isFinite(n) ? 0 : Math.trunc(n));
      }}
    />
  );
}

/** 一层重试的两个旋钮。idPrefix 用来在同一页出现多次时保住 label 的 htmlFor。 */
export function RetryRuleFields({
  layer,
  idPrefix,
  value,
  onChange,
  compact,
}: {
  layer: LayerKey;
  idPrefix: string;
  value: LLMRetryRule;
  onChange: (r: LLMRetryRule) => void;
  /** true = 配置抽屉里的紧凑版：省掉展开说明，只留「什么错误会走到这层」这一句 */
  compact?: boolean;
}) {
  const meta = RETRY_LAYERS[layer];
  const human = humanMs(value.interval_ms);
  return (
    <div className={compact ? "grid gap-2" : "grid gap-3 rounded-lg border p-3"}>
      <div className="grid gap-0.5">
        <div className="flex flex-wrap items-baseline gap-2">
          <Label className="text-sm">{meta.title}</Label>
          <span className="text-muted-foreground text-xs">{meta.where}</span>
        </div>
        {/* 哪些错误会走到这层，具体到状态码——填了旋钮却看不到效果，多半是错误压根不落在这层。 */}
        <p className="text-muted-foreground text-xs">
          <span className="font-medium text-foreground">Trigger</span>: {meta.trigger}
        </p>
        {!compact && meta.skips && (
          <p className="text-muted-foreground text-xs">
            <span className="font-medium text-foreground">Bypass</span>: {meta.skips}
          </p>
        )}
        {!compact && <p className="text-muted-foreground text-xs">{meta.desc}</p>}
      </div>
      <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
        <div className="flex items-center gap-2">
          <Label htmlFor={`${idPrefix}-${layer}-n`} className="text-muted-foreground text-xs">
            {meta.attemptsLabel}
          </Label>
          <NumField
            id={`${idPrefix}-${layer}-n`}
            min={-1}
            value={value.attempts}
            placeholder={`Default ${meta.defAttempts}`}
            onChange={(n) => onChange({ ...value, attempts: n })}
          />
        </div>
        <div className="flex items-center gap-2">
          <Label htmlFor={`${idPrefix}-${layer}-ms`} className="text-muted-foreground text-xs">
            Interval ms
          </Label>
          <NumField
            id={`${idPrefix}-${layer}-ms`}
            min={0}
            value={value.interval_ms}
            placeholder="Default backoff"
            onChange={(n) => onChange({ ...value, interval_ms: n })}
          />
          <span className="text-muted-foreground text-xs">{human ? `Fixed ${human}` : meta.defInterval}</span>
        </div>
      </div>
      {!compact && <p className="text-muted-foreground text-xs">Blank = use default; {meta.offHint}.</p>}
    </div>
  );
}

/** 模型配置抽屉里的三层覆盖（跟着端点走的那三层）。 */
export function ProfileRetryFields({
  value,
  onChange,
}: {
  value: LLMRetryOverride;
  onChange: (o: LLMRetryOverride) => void;
}) {
  return (
    <div className="grid gap-3 rounded-lg border p-3">
      <div className="grid gap-0.5">
        <Label className="text-sm">Retry Overrides</Label>
        <p className="text-muted-foreground text-xs">
          Profile-specific overrides for global retry defaults. Blank = inherit global; -1 = disable layer; explicit
          interval replaces exponential backoff with fixed duration. Circuit breaker and intent rerun are process-level
          and configured globally.
        </p>
      </div>
      {(["connect", "empty", "stream"] as const).map((k) => (
        <div key={k} className="border-t pt-3 first:border-t-0 first:pt-0">
          <RetryRuleFields
            compact
            layer={k}
            idPrefix="pf"
            value={value[k]}
            onChange={(r) => onChange({ ...value, [k]: r })}
          />
        </div>
      ))}
    </div>
  );
}

/** 「重试与退避」tab：五层的全局默认值。 */
export function RetryPolicyPanel() {
  const [policy, setPolicy] = React.useState<LLMRetryPolicy>(ZERO_POLICY);
  const [loading, setLoading] = React.useState(true);
  const [saving, setSaving] = React.useState(false);

  const load = React.useCallback(async () => {
    setLoading(true);
    try {
      const p = await api.llmRetryPolicy();
      setPolicy({ ...ZERO_POLICY, ...p });
    } catch (e) {
      toast.error(`Failed to load retry policy: ${(e as Error).message}`);
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    void load();
  }, [load]);

  async function save() {
    if (saving) return;
    setSaving(true);
    try {
      // 后端会把越界值夹回区间并回传，直接用回传值刷新，所见即所存。
      const saved = await api.saveLLMRetryPolicy(policy);
      setPolicy({ ...ZERO_POLICY, ...saved });
      toast.success("Saved. Active immediately (running calls retain previous parameters)");
    } catch (e) {
      toast.error(`Failed to save: ${(e as Error).message}`);
    } finally {
      setSaving(false);
    }
  }

  const set = (k: LayerKey) => (r: LLMRetryRule) => setPolicy((p) => ({ ...p, [k]: r }));

  if (loading) {
    return (
      <div className="flex items-center gap-2 rounded-lg border border-dashed p-10 text-muted-foreground text-sm">
        <Loader2Icon className="size-4 animate-spin" /> Loading retry policy...
      </div>
    );
  }

  return (
    <div className="grid gap-4">
      <div className="rounded-lg border bg-muted/30 p-3 text-muted-foreground text-xs leading-relaxed">
        Model invocation failures cascade through five defensive retry layers, from inner to outer:
        <span className="text-foreground">
          {" "}
          Connect Handshake → Empty Response → Stream Safe Window → Circuit Breaker → Intent Rerun
        </span>
        . Inner retries must exhaust before escalating outward; attempt counts are
        <span className="text-foreground"> multiplicative</span>. Blank values use built-in defaults. The first three
        layers can be overridden per profile.
      </div>

      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {(Object.keys(RETRY_LAYERS) as LayerKey[]).map((k) => (
          <RetryRuleFields key={k} layer={k} idPrefix="gl" value={policy[k]} onChange={set(k)} />
        ))}
      </div>

      <div className="flex gap-2">
        <Button onClick={save} disabled={saving}>
          {saving ? <Loader2Icon className="animate-spin" /> : <SaveIcon />}
          Save
        </Button>
        <Button variant="outline" onClick={() => setPolicy(ZERO_POLICY)} disabled={saving}>
          Reset All Defaults
        </Button>
      </div>
    </div>
  );
}
