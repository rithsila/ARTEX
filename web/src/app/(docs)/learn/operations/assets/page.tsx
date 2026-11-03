"use client";

import { Network } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function AssetsDocPage() {
  return (
    <DocModuleTemplate
      title="Assets"
      description="Discovered hostnames, IP addresses, open ports, identified web technologies, and alive status inventory."
      groupLabel="OPERATIONS"
      icon={Network}
      consoleRoute="/function/assets"
      badge="Attack Surface"
      purpose="Serves as the centralized attack surface asset inventory. As reconnaissance agents and tools discover subdomains, open TCP/UDP ports, HTTP headers, and software tech stacks, structured asset nodes are registered and updated here."
      prerequisites={[
        "Completed reconnaissance or port scanning scans against targeted scopes.",
        "PostgreSQL assets and ports tables operational.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Navigate to Assets",
          detail: "Select 'Assets' from the OPERATIONS sidebar or visit /function/assets.",
        },
        {
          step: 2,
          title: "Filter by Asset Type",
          detail:
            "Filter assets by domain name, IP address, operating system, or discovered web server technology (e.g. Nginx, Apache).",
          tip: "Click on any IP address to see all associated virtual hostnames.",
        },
        {
          step: 3,
          title: "Inspect Port & Service Details",
          detail:
            "Expand an asset row to view open ports, listening service banners, TLS certificate expiration, and HTTP title tags.",
        },
        {
          step: 4,
          title: "Tag or Prioritize",
          detail:
            "Apply custom tags (e.g., 'production', 'high-priority', 'out-of-scope') to guide agent targeting decisions.",
        },
      ]}
      configFields={[
        {
          field: "Auto Port Merge",
          type: "boolean",
          description: "Automatically updates existing host records when subsequent scans discover new open ports.",
          defaultVal: "true",
        },
        {
          field: "DNS Re-check Interval",
          type: "hours",
          description: "Periodicity for automated DNS resolution checks on discovered subdomains.",
          defaultVal: "48h",
        },
        {
          field: "Max Assets Per Scope",
          type: "integer",
          description: "Safety ceiling on total discovered asset records per engagement.",
          defaultVal: "10000",
        },
        {
          field: "Technology Fingerprinting",
          type: "boolean",
          description: "Enables automated Wappalyzer-style HTTP header and HTML signature analysis on HTTP services.",
          defaultVal: "true",
        },
      ]}
      gotchas={[
        {
          title: "DNS Wildcard False Positives",
          description:
            "Domains with wildcard DNS can generate thousands of false subdomains. Configure wildcard filtering in the initial recon command.",
        },
        {
          title: "Stale Asset Records",
          description:
            "Temporary cloud instances change IPs frequently. Use the 'Re-verify Alive Status' batch action before deep exploitation phases.",
        },
      ]}
    />
  );
}
