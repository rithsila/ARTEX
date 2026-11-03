"use client";

import { Settings2 } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function SettingsDocPage() {
  return (
    <DocModuleTemplate
      title="Settings"
      description="Manage administrator credentials, system branding identity, database connections, and interface theme defaults."
      groupLabel="SYSTEM"
      icon={Settings2}
      consoleRoute="/system/settings"
      badge="Configuration"
      purpose="Houses core instance configuration settings. Update your administrator credentials, configure the system display name, download the MITM Root CA certificate, and toggle interface themes."
      prerequisites={[
        "Administrator login account credentials.",
        "PostgreSQL connection active to persist key-value configuration rows.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Access Settings",
          detail: "Click 'Settings' at the bottom of the SYSTEM section in the sidebar or browse to /system/settings.",
        },
        {
          step: 2,
          title: "Update Admin Password",
          detail: "Under the Security section, enter current password and provide a new strong passphrase.",
          tip: "Use ./reset-password.sh from the host CLI if the admin password is lost.",
        },
        {
          step: 3,
          title: "Download Root CA Certificate",
          detail: "Click 'Download CA Certificate' to retrieve ca.crt for importing into target browser trust stores.",
        },
        {
          step: 4,
          title: "Save & Apply Changes",
          detail:
            "Click 'Save Settings'. Changes take effect immediately across all services without restarting the application.",
        },
      ]}
      configFields={[
        {
          field: "System Name",
          type: "string",
          description: "Display title shown on the navigation bar, login screen, and report covers.",
          defaultVal: "ARTEX Pentest Console",
        },
        {
          field: "Session Expiry",
          type: "duration",
          description: "Duration of inactivity before web JWT session tokens expire.",
          defaultVal: "24h",
        },
        {
          field: "Theme Mode",
          type: "enum",
          description: "Default visual theme preference (Dark, Light, System Sync).",
          defaultVal: "Dark",
        },
        {
          field: "Max Concurrent Scans",
          type: "integer",
          description: "Global ceiling on simultaneous automated penetration testing missions.",
          defaultVal: "2",
        },
      ]}
      gotchas={[
        {
          title: "Database DSN Overrides",
          description:
            "If ARTEX_PG_DSN environment variable is set on the host, it takes precedence over database connection fields saved in the settings UI.",
        },
        {
          title: "Forgotten Password CLI Recovery",
          description:
            "If locked out of the web UI, run `./reset-password.sh -p 'NewPassword'` directly on the server to update the bcrypt hash in PostgreSQL.",
        },
      ]}
    />
  );
}
