"use client";

import { Target } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function TasksDocPage() {
  return (
    <DocModuleTemplate
      title="Tasks"
      description="Decomposed task execution queue, status inspector, terminal logs, and sub-task failure restarts."
      groupLabel="OPERATIONS"
      icon={Target}
      consoleRoute="/function/tasks"
      badge="Execution Queue"
      purpose="Maintains an inspectable record of every unit of work planned or executed by ARTEX. Operators can view running jobs, inspect stdout/stderr outputs, retry failed tasks, or cancel runaway jobs."
      prerequisites={[
        "Tasks are created automatically by the Planner Agent or manually via tool dispatch.",
        "PostgreSQL backend active to persist task logs and status checkpoints.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Access Tasks View",
          detail: "Click 'Tasks' under the OPERATIONS menu in the sidebar or navigate to /function/tasks.",
        },
        {
          step: 2,
          title: "Filter by Status",
          detail: "Filter tasks by status tabs: All, Pending, Running, Finished, or Failed.",
          tip: "Click on a task row to open the full drawer containing real-time command output.",
        },
        {
          step: 3,
          title: "Inspect Arguments & Output",
          detail: "Review exact CLI parameters, environment variables, exit codes, and execution duration.",
        },
        {
          step: 4,
          title: "Control Execution",
          detail:
            "Click 'Cancel' to kill an active process, or click 'Rerun' to re-dispatch a failed task with identical inputs.",
        },
      ]}
      configFields={[
        {
          field: "Task Timeout",
          type: "duration",
          description: "Maximum execution duration allowed for a single command before being terminated automatically.",
          defaultVal: "10m",
        },
        {
          field: "Concurrency Limit",
          type: "integer",
          description: "Maximum number of parallel tool execution sub-processes running on the host system.",
          defaultVal: "4",
        },
        {
          field: "Retry Max Attempts",
          type: "integer",
          description: "Automatic retry count for network-timeout failures before marking a task as failed.",
          defaultVal: "2",
        },
        {
          field: "Log Retention",
          type: "days",
          description: "Number of days raw command output logs are retained in the database.",
          defaultVal: "30",
        },
      ]}
      gotchas={[
        {
          title: "Tasks Paused in Pending State",
          description:
            "A task remaining in 'Pending' typically indicates that an intercept rule requires operator approval. Check System > Approvals.",
        },
        {
          title: "Exit Code 127 (Command Not Found)",
          description:
            "Indicates that the underlying CLI binary (e.g. nmap, nuclei) is missing on the server host. Install it in the system PATH.",
        },
      ]}
    />
  );
}
