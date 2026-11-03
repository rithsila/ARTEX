"use client";

import { FolderOpen } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function WorkspaceDocPage() {
  return (
    <DocModuleTemplate
      title="Workspace"
      description="Local filesystem workspace explorer, raw tool artifact inspector, screenshots, evidence files, and report generators."
      groupLabel="OPERATIONS"
      icon={FolderOpen}
      consoleRoute="/function/workspace"
      badge="Artifact Storage"
      purpose="Provides a secure file manager to browse, preview, and download artifacts produced by agents during scans. View screenshots captured by Playwright, download raw fuzzer outputs, inspect temporary wordlists, and access final report files."
      prerequisites={[
        "Active filesystem storage path configured in system settings (default: data/workspace).",
        "At least one task executed that wrote output artifacts to disk.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Access Workspace",
          detail: "Select 'Workspace' from the OPERATIONS sidebar or browse directly to /function/workspace.",
        },
        {
          step: 2,
          title: "Browse Task Directories",
          detail:
            "Navigate directory trees organized by Task ID, target domain, or file category (artifacts, screenshots, logs).",
          tip: "Click on any image file (.png, .jpg) to open an inline lightbox preview.",
        },
        {
          step: 3,
          title: "Preview File Contents",
          detail:
            "View plain text files, markdown summaries, JSON schemas, and command outputs directly inside the browser editor.",
        },
        {
          step: 4,
          title: "Download or Delete Artifacts",
          detail:
            "Select individual files or whole directories to download as ZIP archives, or prune old files to save disk space.",
        },
      ]}
      configFields={[
        {
          field: "Workspace Root Path",
          type: "string",
          description: "Filesystem path on the host where task outputs and generated artifacts are stored.",
          defaultVal: "data/workspace",
        },
        {
          field: "Max File Preview Size",
          type: "bytes",
          description: "Maximum file size loaded into the browser text editor preview.",
          defaultVal: "2MB",
        },
        {
          field: "Allowed File Extensions",
          type: "string[]",
          description: "Whitelist of file extensions permitted for browser download and preview.",
          defaultVal: "txt, log, json, md, png, jpg, html",
        },
        {
          field: "Auto Purge Age",
          type: "days",
          description: "Number of days before temporary fuzzer artifacts are automatically purged.",
          defaultVal: "30",
        },
      ]}
      gotchas={[
        {
          title: "Permission Denied Errors",
          description:
            "Ensure the ARTEX system user has read and write permissions to the data/workspace directory on the host filesystem.",
        },
        {
          title: "Symlink Security Protection",
          description:
            "The workspace explorer blocks path traversal and symlinks pointing outside the configured workspace directory root.",
        },
      ]}
    />
  );
}
