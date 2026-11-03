"use client";

import { Plug } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function MCPDocPage() {
  return (
    <DocModuleTemplate
      title="MCP"
      description="Connect external Model Context Protocol (MCP) servers via stdio or SSE transports to extend capabilities."
      groupLabel="SYSTEM"
      icon={Plug}
      consoleRoute="/system/mcp"
      badge="Protocol Bridge"
      purpose="Provides standardized connectivity to third-party MCP servers. External servers expose custom tool schemas, enterprise database connections, code repository access, or specialized scanners directly to ARTEX AI agents."
      prerequisites={[
        "Local MCP server binary executable on host system, or remote SSE HTTP endpoint reachable by ARTEX.",
        "Valid authentication tokens or environment variables required by the target MCP server.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Navigate to MCP Manager",
          detail: "Select 'MCP' from the SYSTEM navigation menu or visit /system/mcp.",
        },
        {
          step: 2,
          title: "Add MCP Server Connection",
          detail:
            "Click 'Add Server'. Choose Transport Type: Stdio (command-line child process) or SSE (Server-Sent Events URL).",
          tip: "For local Dockerized tools, run stdio with 'docker run -i --rm your-mcp-image'.",
        },
        {
          step: 3,
          title: "Configure Arguments & Environment",
          detail:
            "Specify executable path, command line arguments, and required API keys as environment key-value pairs.",
        },
        {
          step: 4,
          title: "Verify Tool Discovery",
          detail:
            "Click 'Connect'. ARTEX performs protocol negotiation, retrieves exposed tool definitions, and marks the server online.",
        },
      ]}
      configFields={[
        {
          field: "Transport Type",
          type: "enum",
          description: "IPC communication protocol (stdio, sse).",
          defaultVal: "stdio",
        },
        {
          field: "Command / URL",
          type: "string",
          description: "Binary executable path for stdio mode or HTTP endpoint URL for SSE mode.",
        },
        {
          field: "Environment Variables",
          type: "key-value",
          description: "Confidential credentials passed directly to the child process environment.",
        },
        {
          field: "Auto Reconnect",
          type: "boolean",
          description: "Automatically restarts failed stdio processes or reconnects dropped SSE streams.",
          defaultVal: "true",
        },
      ]}
      gotchas={[
        {
          title: "Stdio Process Hangs",
          description:
            "Ensure the external MCP server communicates strictly via standard JSON-RPC over stdout. Any unformatted print statements will corrupt the protocol parser.",
        },
        {
          title: "Network Timeouts with Remote SSE",
          description:
            "Keep-alive intervals must be maintained by the remote SSE server to prevent intermediate proxies from terminating idle connections.",
        },
      ]}
    />
  );
}
