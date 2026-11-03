"use client";

import { ShieldAlert } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function RulesDocPage() {
  return (
    <DocModuleTemplate
      title="Rules"
      description="Define granular intercept rules that automatically allow, block, or require human approval on tool calls."
      groupLabel="SYSTEM"
      icon={ShieldAlert}
      consoleRoute="/system/intercept"
      badge="Safety Gates"
      purpose="Provides policy-based execution guardrails over autonomous AI agents. Intercept rules evaluate tool names, command patterns, and arguments before any process spawns, enforcing Allow, Block, or Require Approval actions."
      prerequisites={[
        "Familiarity with the tools and arguments your agents are authorized to run.",
        "Rules take effect immediately upon creation without daemon restart.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Navigate to Rules",
          detail: "Select 'Rules' under SYSTEM in the left sidebar or visit /system/intercept.",
        },
        {
          step: 2,
          title: "Create Intercept Rule",
          detail:
            "Click 'New Rule'. Enter a descriptive name like 'Block Dangerous SQLMap Dumps' or 'Require Approval on Mass Scans'.",
          tip: "Order rules by priority: specific restrictions first, broad fallbacks last.",
        },
        {
          step: 3,
          title: "Define Match Conditions",
          detail:
            "Set match filters: Tool Name (e.g. 'sqlmap'), Command regex pattern, or specific target argument flags.",
        },
        {
          step: 4,
          title: "Select Policy Action",
          detail:
            "Choose policy outcome: 'Allow' (bypass further checks), 'Block' (instant rejection), or 'Require Approval' (pause for human review).",
        },
      ]}
      configFields={[
        {
          field: "Rule Name",
          type: "string",
          description: "Descriptive label indicating the safety purpose of this policy.",
        },
        {
          field: "Action",
          type: "enum",
          description: "Policy verdict when rule conditions match (Allow, Block, Require Approval).",
          defaultVal: "Require Approval",
        },
        {
          field: "Tool Filter",
          type: "string",
          description: "Specific tool name or wildcard pattern ('*' matches all tools).",
          defaultVal: "*",
        },
        {
          field: "Argument Regex",
          type: "regex",
          description: "Regular expression tested against the full command string and argument list.",
        },
      ]}
      gotchas={[
        {
          title: "Overly Aggressive Block Rules",
          description:
            "A regex like `.*` with Block action will halt all autonomous tasks. Always test regexes with specific command samples first.",
        },
        {
          title: "Rule Priority Evaluation Order",
          description:
            "Rules evaluate top-down by Priority number. The first matching rule decides the action and stops subsequent rule evaluation.",
        },
      ]}
    />
  );
}
