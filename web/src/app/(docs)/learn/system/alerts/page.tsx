"use client";

import { BellRing } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function AlertsDocPage() {
  return (
    <DocModuleTemplate
      title="Alerts"
      description="Set up outbound notification channels (webhooks, email, chat bots) when critical findings or approval requests occur."
      groupLabel="SYSTEM"
      icon={BellRing}
      consoleRoute="/system/notify"
      badge="Notifications"
      purpose="Dispatches real-time outbound alert messages to external communication platforms. Notify security engineers instantly when a Critical or High vulnerability is verified, an autonomous scan completes, or an execution approval is waiting."
      prerequisites={[
        "Target notification endpoint (Slack webhook, Discord bot URL, Microsoft Teams connector, or SMTP credentials).",
        "Outbound network access from ARTEX backend to the notification provider.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Navigate to Alerts",
          detail: "Click 'Alerts' in the SYSTEM section of the sidebar or browse to /system/notify.",
        },
        {
          step: 2,
          title: "Create Notification Channel",
          detail: "Click 'Add Channel'. Select Channel Type (Slack, Discord, Custom Webhook, Email SMTP).",
          tip: "Use Custom Webhook to integrate with internal SIEM or ticketing systems like Splunk or Jira.",
        },
        {
          step: 3,
          title: "Configure Webhook URL & Events",
          detail:
            "Paste destination webhook URL and select trigger events: 'Critical Finding', 'Approval Required', and 'Mission Finished'.",
        },
        {
          step: 4,
          title: "Send Test Notification",
          detail: "Click 'Send Test Alert'. Verify receipt in your messaging channel, then click Save.",
        },
      ]}
      configFields={[
        {
          field: "Channel Type",
          type: "enum",
          description: "Integration protocol format (Webhook, Slack, Discord, Email).",
          defaultVal: "Webhook",
        },
        {
          field: "Webhook URL",
          type: "secret",
          description: "Target HTTP endpoint receiving POST requests with formatted alert payloads.",
        },
        {
          field: "Min Severity Trigger",
          type: "enum",
          description: "Minimum vulnerability severity required to fire finding alerts (Low, Med, High, Critical).",
          defaultVal: "High",
        },
        {
          field: "Max Retry Count",
          type: "integer",
          description: "Exponential backoff retries if destination webhook endpoint returns HTTP 5xx.",
          defaultVal: "3",
        },
      ]}
      gotchas={[
        {
          title: "Rate Limiting on Discord/Slack",
          description:
            "High-volume finding bursts can trigger API rate limits on chat webhooks. ARTEX automatically throttles outbound notifications to 1 message per second.",
        },
        {
          title: "Silent Webhook Failures",
          description:
            "If test alerts do not arrive, check System > Logs for HTTP status code errors or TLS certificate rejection on self-hosted webhook endpoints.",
        },
      ]}
    />
  );
}
