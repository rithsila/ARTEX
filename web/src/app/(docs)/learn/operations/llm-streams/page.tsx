"use client";

import { Radio } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function LLMStreamsDocPage() {
  return (
    <DocModuleTemplate
      title="LLM Streams"
      description="Real-time token audit, prompt inspection, completion latency metrics, and raw model exchange debugging."
      groupLabel="OPERATIONS"
      icon={Radio}
      consoleRoute="/function/llm-records"
      badge="Prompt Audit"
      purpose="Captures and displays the raw JSON payloads exchanged between ARTEX agent orchestrators and upstream AI providers. Allows engineers to inspect exact system instructions, tool definitions, reasoning streams, and token usage statistics."
      prerequisites={[
        "Active AI sessions generated via Chat (/chat) or autonomous background tasks.",
        "LLM logging enabled in system configuration.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Access LLM Streams View",
          detail: "Click 'LLM Streams' under OPERATIONS in the navigation sidebar or visit /function/llm-records.",
        },
        {
          step: 2,
          title: "Filter by Provider or Agent",
          detail:
            "Filter recorded streams by AI Provider (OpenAI, Anthropic, Custom), Agent name, or request timestamp.",
          tip: "Search for specific tool call names like 'run_nmap' to find corresponding reasoning logs.",
        },
        {
          step: 3,
          title: "Inspect Full Prompt Exchange",
          detail:
            "Click on any log record to open side-by-side prompt tabs: System Prompt, Chat History, Injected Context, and Model Response.",
        },
        {
          step: 4,
          title: "Analyze Latency & Token Usage",
          detail:
            "Examine Time-To-First-Token (TTFT), completion duration, input tokens, output tokens, and cost estimates.",
        },
      ]}
      configFields={[
        {
          field: "Record History Limit",
          type: "integer",
          description: "Maximum number of historical LLM stream events kept in the database.",
          defaultVal: "1000",
        },
        {
          field: "Sanitize API Keys",
          type: "boolean",
          description: "Masks authorization headers and sensitive credentials from recorded stream logs.",
          defaultVal: "true",
        },
        {
          field: "Log Tool Schemas",
          type: "boolean",
          description: "Includes complete JSON function schemas in the recorded prompt payload inspector.",
          defaultVal: "true",
        },
        {
          field: "Auto Prune Age",
          type: "days",
          description: "Deletes stream audit records older than this duration to conserve storage.",
          defaultVal: "14",
        },
      ]}
      gotchas={[
        {
          title: "High Database Growth",
          description:
            "Large context models generate megabytes of prompt logs per session. Lower the Record History Limit if database storage is constrained.",
        },
        {
          title: "Streaming Truncation on Abort",
          description:
            "If an operator cancels a chat query mid-turn, the stream record will mark 'aborted' with partial completion tokens.",
        },
      ]}
    />
  );
}
