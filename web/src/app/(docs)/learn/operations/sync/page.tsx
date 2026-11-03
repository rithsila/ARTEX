"use client";

import { FolderSync } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function SyncDocPage() {
  return (
    <DocModuleTemplate
      title="Sync"
      description="Bidirectional sync with external tools, wordlist repositories, scanner output importers, and asset exports."
      groupLabel="OPERATIONS"
      icon={FolderSync}
      consoleRoute="/function/sync"
      badge="Data Sync"
      purpose="Enables automated and manual data synchronization between ARTEX and external cybersecurity data sources. Import Nmap XML, Burp Suite state files, and wordlists, or synchronize findings into ticketing systems like Jira and GitHub Issues."
      prerequisites={[
        "Valid export file from external scanner (Nmap XML, Masscan JSON, Burp XML) or configured remote webhook.",
        "PostgreSQL connection active to ingest parsed entities.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Open Sync View",
          detail: "Click 'Sync' under the OPERATIONS menu in the sidebar or browse to /function/sync.",
        },
        {
          step: 2,
          title: "Select Sync Profile",
          detail:
            "Choose between File Import (drag-and-drop scanner outputs) or Remote Provider Sync (scheduled API pulling).",
          tip: "Use the Nmap XML parser to bulk-populate hundreds of target hosts in seconds.",
        },
        {
          step: 3,
          title: "Upload & Map Target Scope",
          detail: "Upload your scanner output file and associate the discovered nodes with an authorized target scope.",
        },
        {
          step: 4,
          title: "Review Sync Report",
          detail: "Verify the summary of newly created hosts, updated ports, and merged vulnerability records.",
        },
      ]}
      configFields={[
        {
          field: "Supported Import Formats",
          type: "string[]",
          description: "File formats accepted for direct ingestion (Nmap XML, Masscan JSON, Burp XML, Sublist3r TXT).",
          defaultVal: "xml, json, txt",
        },
        {
          field: "Auto Conflict Resolution",
          type: "enum",
          description: "Strategy when imported data differs from current state (Overwrite, Merge, Keep Existing).",
          defaultVal: "Merge",
        },
        {
          field: "Remote Sync Schedule",
          type: "cron",
          description: "Cron schedule for pulling external asset catalogs automatically.",
          defaultVal: "0 0 * * *",
        },
        {
          field: "Batch Ingest Chunk Size",
          type: "integer",
          description: "Number of asset records written per SQL transaction during large imports.",
          defaultVal: "500",
        },
      ]}
      gotchas={[
        {
          title: "Large File Upload Timeouts",
          description:
            "Files over 100MB should be split or uploaded directly to data/uploads/ on the host to avoid HTTP proxy timeout limits.",
        },
        {
          title: "Scope Ingestion Mismatch",
          description:
            "Assets imported from third-party scan files that fall outside active Scope rules will be marked as 'Excluded' automatically.",
        },
      ]}
    />
  );
}
