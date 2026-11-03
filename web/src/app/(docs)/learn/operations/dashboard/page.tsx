"use client";

import { LayoutDashboard } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function DashboardDocPage() {
  return (
    <DocModuleTemplate
      title="Dashboard"
      description="Engagement overview, real-time metrics, active task counts, finding severity distribution, and traffic volume."
      groupLabel="OPERATIONS"
      icon={LayoutDashboard}
      consoleRoute="/dashboard"
      badge="Overview"
      purpose="Provides a unified high-level mission control dashboard for pen testing teams. It displays active planner tasks, confirmed findings aggregated by CVSS severity, intercepted requests, and system health status."
      prerequisites={[
        "ARTEX daemon running on port 8787 with active PostgreSQL database.",
        "Valid web login session to the ARTEX management console.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Navigate to Dashboard",
          detail:
            "Click 'Dashboard' in the left console sidebar under the OPERATIONS section or visit /dashboard directly.",
        },
        {
          step: 2,
          title: "Review Metric KPI Cards",
          detail: "Inspect the top stat cards: Active Tasks, Total Targets, Confirmed Findings, and Captured Requests.",
          tip: "Clicking any stat card navigates directly to that filtered view.",
        },
        {
          step: 3,
          title: "Analyze Severity Breakdown",
          detail:
            "Check the visual severity distribution chart for Critical, High, Medium, Low, and Informational findings.",
        },
        {
          step: 4,
          title: "Track Recent Activity Feed",
          detail:
            "Monitor real-time security events, completed agent sub-tasks, and urgent approval alerts as they happen.",
        },
      ]}
      configFields={[
        {
          field: "Auto Refresh",
          type: "boolean",
          description: "Controls whether KPI cards and event feeds refresh automatically over WebSocket.",
          defaultVal: "true",
        },
        {
          field: "Date Range Filter",
          type: "string",
          description: "Filters aggregate statistics by time window (Today, Last 7 Days, Full Mission).",
          defaultVal: "Full Mission",
        },
        {
          field: "Severity Threshold",
          type: "enum",
          description: "Sets the minimum vulnerability severity level highlighted in the top alert banner.",
          defaultVal: "High",
        },
        {
          field: "Target Scope Filter",
          type: "string",
          description: "Limits dashboard statistics to a specific asset or wildcard domain group.",
          defaultVal: "All Scopes",
        },
      ]}
      gotchas={[
        {
          title: "Zero Metrics on Fresh Install",
          description:
            "If all KPI numbers show zero after initial deployment, start an automated scan from Chat (/chat) to produce initial tasks and telemetry.",
        },
        {
          title: "WebSocket Connection Lost",
          description:
            "When the status indicator shows disconnected, verify port 8787 is reachable without reverse proxy buffering on /ws.",
        },
      ]}
    />
  );
}
