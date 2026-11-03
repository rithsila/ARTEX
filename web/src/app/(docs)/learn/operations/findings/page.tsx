"use client";

import { Bug } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function FindingsDocPage() {
  return (
    <DocModuleTemplate
      title="Findings"
      description="Confirmed vulnerabilities, CVSS 3.1 scoring, HTTP request/response evidence, PoC reproduction steps, and report exports."
      groupLabel="OPERATIONS"
      icon={Bug}
      consoleRoute="/function/findings"
      badge="Evidence Store"
      purpose="Stores verified vulnerability evidence discovered by ARTEX tools and agent reasoning. Each finding records target URL, affected parameters, raw HTTP request/response proof, CVSS metrics, and actionable remediation instructions."
      prerequisites={[
        "Completed task runs or agent workflows that identified potential vulnerabilities.",
        "PostgreSQL findings table populated with structured vulnerability records.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Open Findings Tab",
          detail: "Select 'Findings' from the OPERATIONS sidebar section or navigate directly to /function/findings.",
        },
        {
          step: 2,
          title: "Search & Filter Results",
          detail:
            "Filter findings by severity (Critical, High, Medium, Low), target domain, or vulnerability category (SQLi, XSS, SSRF).",
          tip: "Use the search bar to locate specific CVE numbers or endpoint paths.",
        },
        {
          step: 3,
          title: "Inspect Evidence Drawer",
          detail:
            "Click on any finding card to open the detail panel displaying reproduction steps, raw HTTP payloads, and matched regex patterns.",
        },
        {
          step: 4,
          title: "Export Audit Report",
          detail:
            "Click 'Export' in the top action bar to generate Markdown, JSON, or CSV audit deliverables for stakeholders.",
        },
      ]}
      configFields={[
        {
          field: "Verification Required",
          type: "boolean",
          description: "Requires agent to produce reproducible PoC evidence before registering a finding as Confirmed.",
          defaultVal: "true",
        },
        {
          field: "Deduplication Window",
          type: "hours",
          description: "Consolidates identical findings on the same endpoint within this time window into one record.",
          defaultVal: "24h",
        },
        {
          field: "Default CVSS Version",
          type: "string",
          description: "Base calculation formula used for automated risk scoring (CVSS 3.1 / CVSS 4.0).",
          defaultVal: "CVSS 3.1",
        },
        {
          field: "Auto Notify Critical",
          type: "boolean",
          description: "Triggers immediate alerts via configured webhook channels when a Critical finding is saved.",
          defaultVal: "true",
        },
      ]}
      gotchas={[
        {
          title: "False Positive Management",
          description:
            "If an automated tool flags a non-exploitable condition, click 'Mark as False Positive' to exclude it from reports and train future agent runs.",
        },
        {
          title: "Missing Raw Traffic Evidence",
          description:
            "Raw HTTP request/response tabs require tools to route traffic through the internal MITM proxy on port 8788.",
        },
      ]}
    />
  );
}
