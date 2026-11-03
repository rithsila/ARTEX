"use client";

import { Bot } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function AgentDocPage() {
  return (
    <DocModuleTemplate
      title="Agent"
      description="Build and customize autonomous AI agents, their system prompts, specialist roles, and authorized tool packs."
      groupLabel="SYSTEM"
      icon={Bot}
      consoleRoute="/system/agents"
      badge="Orchestration"
      purpose="Enables operators to define distinct AI agent personalities with specialized system prompts, tool permissions, and operational constraints. You can craft specialized agents for fast reconnaissance, web app exploitation, or cloud auditing."
      prerequisites={[
        "At least one active AI provider configured in System > LLM.",
        "Installed tool packs and skills in the system catalog.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Access Agent Manager",
          detail: "Click 'Agent' in the SYSTEM section of the sidebar or navigate to /system/agents.",
        },
        {
          step: 2,
          title: "Create New Agent Profile",
          detail: "Click 'Create Agent' to open the profile builder modal.",
          tip: "Clone an existing agent template to quickly adjust prompts without starting from scratch.",
        },
        {
          step: 3,
          title: "Assign Role & System Prompt",
          detail:
            "Define the agent name, short description, and detailed instructions outlining methodology and reporting rules.",
        },
        {
          step: 4,
          title: "Select Allowed Tools & LLM",
          detail: "Select which LLM provider powers this agent, toggle authorized skills, and click Save Agent.",
        },
      ]}
      configFields={[
        {
          field: "Agent Name",
          type: "string",
          description: "Unique label for identifying the persona in the Chat agent selector dropdown.",
          defaultVal: "Planner",
        },
        {
          field: "System Prompt",
          type: "text",
          description: "Core behavioral instructions and ethical testing constraints injected into the model context.",
        },
        {
          field: "Bound LLM Provider",
          type: "select",
          description: "Specific LLM provider instance used whenever this agent runs.",
          defaultVal: "Default Provider",
        },
        {
          field: "Enabled Skills",
          type: "multiselect",
          description: "List of authorized tool packs and skill suites available to this agent.",
          defaultVal: "All Skills",
        },
      ]}
      gotchas={[
        {
          title: "Too Many Registered Tools",
          description:
            "Enabling 50+ tools simultaneously can overwhelm model attention windows. Restrict each agent persona to 5-15 task-relevant tools.",
        },
        {
          title: "Prompt Hallucinations",
          description:
            "If an agent invents fictional command flags, refine the system prompt with strict instructions: 'Never fabricate CLI flags not listed in tool schemas'.",
        },
      ]}
    />
  );
}
