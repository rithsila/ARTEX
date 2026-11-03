"use client";

import { ClipboardList } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function ApprovalsDocPage() {
  return (
    <DocModuleTemplate
      title="Approvals"
      description="Review and decide on paused tool execution requests waiting for human permission before execution."
      groupLabel="SYSTEM"
      icon={ClipboardList}
      consoleRoute="/system/intercept/approvals"
      badge="Human Gate"
      purpose="Serves as the Human-in-the-Loop (HITL) authorization cockpit. When an agent attempts an action that triggers an approval rule (e.g., destructive fuzzing or exploit payloads), execution pauses until an operator explicitly approves or rejects the request."
      prerequisites={[
        "At least one active intercept rule with action set to 'Require Approval'.",
        "A running agent task that triggered this rule.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Navigate to Approvals",
          detail: "Select 'Approvals' under SYSTEM in the sidebar or browse to /system/intercept/approvals.",
        },
        {
          step: 2,
          title: "Inspect Pending Requests",
          detail:
            "Examine the cards in the Pending queue showing Agent Name, Tool Name, Target Host, and Triggered Rule.",
          tip: "A glowing amber badge in the sidebar indicates active requests awaiting your decision.",
        },
        {
          step: 3,
          title: "Audit Exact CLI Arguments",
          detail:
            "Review the full command string, headers, and parameter payloads to verify safety against the target environment.",
        },
        {
          step: 4,
          title: "Approve or Reject",
          detail:
            "Click 'Approve' to unblock worker execution, or click 'Reject' to cancel the tool call and instruct the agent to pick an alternative path.",
        },
      ]}
      configFields={[
        {
          field: "Auto Reject Timeout",
          type: "duration",
          description: "Duration a paused request waits for human review before being rejected automatically.",
          defaultVal: "30m",
        },
        {
          field: "Notification on Trigger",
          type: "boolean",
          description: "Immediately sends an alert via configured notification channels when approval is needed.",
          defaultVal: "true",
        },
        {
          field: "Require Justification Reason",
          type: "boolean",
          description: "Requires the operator to type an audit justification note when approving risky commands.",
          defaultVal: "false",
        },
        {
          field: "Batch Decision Allowed",
          type: "boolean",
          description:
            "Permits bulk 'Approve All' action on multiple pending requests matching identical rule patterns.",
          defaultVal: "true",
        },
      ]}
      gotchas={[
        {
          title: "Timeout Stalling Agent Tasks",
          description:
            "If an approval request expires before human review, the task fails with 'approval timeout' and the agent will revise its plan.",
        },
        {
          title: "Goroutine Resource Hold",
          description:
            "Paused tasks hold background execution goroutines in memory. Avoid leaving dozens of pending approvals unaddressed.",
        },
      ]}
    />
  );
}
