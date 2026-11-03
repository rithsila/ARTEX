"use client";

import { Terminal } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function ExecDocPage() {
  return (
    <DocModuleTemplate
      title="Exec"
      description="Direct shell execution console for human command verification, interactive tooling, and manual testing."
      groupLabel="OPERATIONS"
      icon={Terminal}
      consoleRoute="/function/commands"
      badge="Shell Execution"
      purpose="Provides a direct terminal execution bridge for operators to run ad-hoc command-line tools directly on the backend host environment. Output streams live into the console window and commands are audited into system logs."
      prerequisites={[
        "Admin authorization or operator role with command execution privileges.",
        "Target binaries installed and accessible in the system environment PATH.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Open Exec Console",
          detail: "Click 'Exec' in the OPERATIONS sidebar section or navigate to /function/commands.",
        },
        {
          step: 2,
          title: "Enter CLI Command",
          detail:
            "Type your security command (e.g. 'curl -I https://example.com' or 'dig target.domain ANY') into the command terminal prompt.",
          tip: "Environment variables configured in system settings are automatically inherited.",
        },
        {
          step: 3,
          title: "Execute & Monitor Stream",
          detail: "Press Enter or click Run. View colorized ANSI terminal output streamed over WebSocket in real time.",
        },
        {
          step: 4,
          title: "Terminate or Export",
          detail: "Use Ctrl+C or the stop button to interrupt running commands. Export command logs for documentation.",
        },
      ]}
      configFields={[
        {
          field: "Default Shell",
          type: "string",
          description: "Path to system shell binary used to execute command strings (/bin/bash, /bin/zsh, /bin/sh).",
          defaultVal: "/bin/sh",
        },
        {
          field: "Command Timeout",
          type: "duration",
          description: "Safety timeout after which long-running commands are automatically sent SIGKILL.",
          defaultVal: "15m",
        },
        {
          field: "Working Directory",
          type: "string",
          description: "Default filesystem working directory for spawned processes.",
          defaultVal: "data/workspace",
        },
        {
          field: "Audit Logging",
          type: "boolean",
          description: "Records operator username, timestamp, and command string in immutable database audit logs.",
          defaultVal: "true",
        },
      ]}
      gotchas={[
        {
          title: "Non-Interactive Commands Only",
          description:
            "Commands that require interactive keyboard input (like sudo prompts or SSH key passwords) will hang. Pass non-interactive flags (e.g., -y or batch mode).",
        },
        {
          title: "Scope Restrictions Still Apply",
          description:
            "Even manual commands executed through Exec are checked against active Scope rules in System > Scope.",
        },
      ]}
    />
  );
}
