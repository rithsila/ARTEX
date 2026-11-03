"use client";

import { ScrollText } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function LogsDocPage() {
  return (
    <DocModuleTemplate
      title="Logs"
      description="Monitor live daemon diagnostics, database query traces, proxy errors, and export debug bundles in real time."
      groupLabel="SYSTEM"
      icon={ScrollText}
      consoleRoute="/system/logs"
      badge="Diagnostics"
      purpose="Streams internal application logs emitted by the Go daemon, background worker pools, HTTP API handlers, and MITM proxy services. Essential for diagnosing system errors, tool invocation issues, and database deadlocks."
      prerequisites={[
        "ARTEX backend service running with logging configured.",
        "Active WebSocket connection to the /ws log broadcast topic.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Open Logs Viewer",
          detail: "Click 'Logs' under SYSTEM in the left sidebar or navigate to /system/logs.",
        },
        {
          step: 2,
          title: "Filter by Severity Level",
          detail: "Toggle filter buttons to show only Error, Warning, or Info messages.",
          tip: "Enable Auto-scroll to keep the terminal anchored to the latest log line as events stream in.",
        },
        {
          step: 3,
          title: "Search Log Text",
          detail:
            "Type error codes, task IDs, or package names (e.g. 'agent/orchestrator' or 'traffic/mitm') in the search bar.",
        },
        {
          step: 4,
          title: "Download Diagnostic Bundle",
          detail:
            "Click 'Export Logs' to download a timestamped plaintext log bundle for troubleshooting and bug reports.",
        },
      ]}
      configFields={[
        {
          field: "Log Level",
          type: "enum",
          description: "Minimum verbosity level logged by the Go backend (Debug, Info, Warn, Error).",
          defaultVal: "Info",
        },
        {
          field: "Buffer Size",
          type: "integer",
          description: "Number of recent log lines retained in the in-memory ring buffer.",
          defaultVal: "2000",
        },
        {
          field: "Console ANSI Color",
          type: "boolean",
          description: "Enables terminal color escapes in the web log viewer display.",
          defaultVal: "true",
        },
        {
          field: "Log File Path",
          type: "string",
          description: "Disk path where daemon rotating log files are saved on the server.",
          defaultVal: "data/logs/artex.log",
        },
      ]}
      gotchas={[
        {
          title: "Debug Level Performance Impact",
          description:
            "Setting Log Level to 'Debug' during high-volume port scanning or fuzzing produces gigabytes of disk I/O. Use Debug only when isolating specific defects.",
        },
        {
          title: "WebSocket Disconnection During Heavy Load",
          description:
            "If the log stream freezes, refresh the page to establish a new WebSocket connection to the log broadcaster.",
        },
      ]}
    />
  );
}
