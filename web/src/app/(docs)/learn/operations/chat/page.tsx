"use client";

import { MessageSquare } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function ChatDocPage() {
  return (
    <DocModuleTemplate
      title="Chat"
      description="Interactive AI agent mission console, streaming reasoning thoughts, tool execution triggers, and conversational guidance."
      groupLabel="OPERATIONS"
      icon={MessageSquare}
      consoleRoute="/chat"
      badge="Control Center"
      purpose="Acts as the primary operator interface for commanding autonomous AI security agents. You describe high-level testing goals in natural language, and the agent decomposes goals into actionable sub-tasks, dispatches tools, and displays live feedback."
      prerequisites={[
        "At least one active AI provider configured in System > LLM (/system/llm).",
        "At least one agent persona profile active in System > Agent (/system/agents).",
        "Permission to execute penetration tests against authorized target boundaries.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Select Agent Persona",
          detail:
            "Open the agent dropdown menu in the upper-left of the Chat view to choose your specialist (e.g., Default Planner, Web Exploiter, Fast Recon).",
        },
        {
          step: 2,
          title: "Issue Mission Objective",
          detail:
            "Type your security testing goal into the prompt bar, such as: 'Perform full port scan and directory discovery on target.example.com'.",
          tip: "Include scope constraints like 'exclude /admin' directly in the prompt.",
        },
        {
          step: 3,
          title: "Follow Streaming Thoughts",
          detail:
            "Watch the agent's internal reasoning chain, tool selection decisions, and parameter justifications stream live.",
        },
        {
          step: 4,
          title: "Provide Guidance or Intervene",
          detail:
            "Send follow-up instructions to refine target priority, pause specific tasks, or provide authentication credentials when requested.",
        },
      ]}
      configFields={[
        {
          field: "Agent Profile",
          type: "select",
          description: "Selects the system prompt, tool permissions, and model parameters powering the conversation.",
          defaultVal: "Default Planner",
        },
        {
          field: "Temperature Override",
          type: "float",
          description:
            "Adjusts model creativity (0.0 for strict deterministic scans, 0.7 for creative exploit chaining).",
          defaultVal: "0.2",
        },
        {
          field: "Max Autonomous Steps",
          type: "number",
          description:
            "Limits how many sequential tool invocations the agent may run before pausing for operator input.",
          defaultVal: "25",
        },
        {
          field: "Stream Reasoning",
          type: "boolean",
          description: "Toggles live streaming of internal chain-of-thought tokens in the message bubble.",
          defaultVal: "true",
        },
      ]}
      gotchas={[
        {
          title: "Missing API Key Error",
          description:
            "If messages fail to send with 'No active LLM provider found', navigate to System > LLM and add an API key with tested connectivity.",
        },
        {
          title: "Loop Stalling on Broad Prompts",
          description:
            "Vague goals like 'hack this server' cause excessive scanning. Anchor prompts with specific targets and explicit testing phases.",
        },
      ]}
    />
  );
}
