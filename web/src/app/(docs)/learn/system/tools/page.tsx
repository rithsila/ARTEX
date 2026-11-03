"use client";

import { Wrench } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function ToolsDocPage() {
  return (
    <DocModuleTemplate
      title="Tools"
      description="Inspect atomic command-line tool definitions, host binary installation status, and parameter JSON schemas."
      groupLabel="SYSTEM"
      icon={Wrench}
      consoleRoute="/system/tools"
      badge="Binary Registry"
      purpose="Maintains the low-level registry of all executable tools registered in ARTEX. It verifies whether CLI binaries (like nmap, curl, ffuf, nuclei) exist on the host PATH and displays the strict JSON schema used by AI models to invoke them."
      prerequisites={[
        "Host environment with security binaries installed (or running in Docker with tool packages).",
        "Valid tool registry records parsed from active skills and built-ins.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Navigate to Tools",
          detail: "Select 'Tools' from the SYSTEM menu in the left sidebar or visit /system/tools.",
        },
        {
          step: 2,
          title: "Check Binary Health",
          detail:
            "Review the tool table. Check the green 'Installed' or red 'Missing' status badge for each binary on the server.",
          tip: "Use the filter bar to search for tools by keyword like 'port', 'web', or 'fuzz'.",
        },
        {
          step: 3,
          title: "Inspect Parameter Schemas",
          detail:
            "Click on any tool row to open the schema viewer showing required parameters, types, and validation regexes.",
        },
        {
          step: 4,
          title: "Test Single Tool Run",
          detail:
            "Run a test invocation with sample arguments to verify command execution and output formatting before agent use.",
        },
      ]}
      configFields={[
        {
          field: "Binary Path",
          type: "string",
          description: "Absolute filesystem location of the tool binary on the backend host.",
          defaultVal: "/usr/bin/<tool>",
        },
        {
          field: "Execution Mode",
          type: "enum",
          description: "How commands run (Direct Host Process, Docker Container, or Remote Worker).",
          defaultVal: "Host Process",
        },
        {
          field: "Parameter Strictness",
          type: "enum",
          description: "Rejection policy for unmodeled parameters (Strict JSON Schema, Coerced).",
          defaultVal: "Strict",
        },
        {
          field: "Default Timeout",
          type: "duration",
          description: "Maximum execution limit before tool process receives SIGTERM.",
          defaultVal: "10m",
        },
      ]}
      gotchas={[
        {
          title: "Red 'Missing' Badge",
          description:
            "If a required tool shows Missing, install it on the backend host system (e.g. `apt-get install nmap` or `brew install ffuf`) and refresh the page.",
        },
        {
          title: "Schema Parameter Mismatch",
          description:
            "If an agent repeatedly fails invoking a tool, check that parameter names in the tool schema match the exact flags expected by the CLI binary.",
        },
      ]}
    />
  );
}
