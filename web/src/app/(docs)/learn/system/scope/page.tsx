"use client";

import { Ban } from "lucide-react";

import { DocModuleTemplate } from "../../_components/doc-module-template";

export default function ScopeDocPage() {
  return (
    <DocModuleTemplate
      title="Scope"
      description="Define strict target boundaries: permitted hostnames, CIDR blocks, wildcard domains, and forbidden exclusions."
      groupLabel="SYSTEM"
      icon={Ban}
      consoleRoute="/system/intercept/assets"
      badge="Boundaries"
      purpose="Enforces the legal and contractual testing boundary for all automated agent actions. Outbound requests or tool arguments pointing to out-of-scope or explicitly excluded IP addresses and hostnames are intercepted and dropped before transmission."
      prerequisites={[
        "Explicit written authorization for every domain, hostname, and IP range added to the scope.",
        "System administrator or lead tester role to modify scope boundaries.",
      ]}
      usageSteps={[
        {
          step: 1,
          title: "Access Scope Manager",
          detail: "Click 'Scope' under SYSTEM in the navigation sidebar or browse to /system/intercept/assets.",
        },
        {
          step: 2,
          title: "Add Target Scope Entry",
          detail:
            "Click 'Add Scope'. Enter target hostname, domain pattern (e.g. '*.staging.example.com'), or CIDR block (e.g. 10.0.0.0/24).",
          tip: "Add wildcard subdomains to automatically cover dynamically discovered subdomains.",
        },
        {
          step: 3,
          title: "Configure Exclusions (Banned Targets)",
          detail:
            "Explicitly exclude sensitive third-party services (e.g. payment gateways, production login portals, cloud SSO).",
        },
        {
          step: 4,
          title: "Verify Active Enforcement",
          detail:
            "Review the Scope table. Ensure target domains are marked 'Allow' and critical exclusions are marked 'Exclude'.",
        },
      ]}
      configFields={[
        {
          field: "Pattern",
          type: "string",
          description: "Domain, exact hostname, IP address, or CIDR notation defining the target.",
        },
        {
          field: "Scope Type",
          type: "enum",
          description: "Policy classification: 'Allow' (in-scope target) or 'Exclude' (strictly forbidden).",
          defaultVal: "Allow",
        },
        {
          field: "Strict DNS Check",
          type: "boolean",
          description: "Resolves hostnames to IPs before tool invocation to ensure target IP is within allowed CIDRs.",
          defaultVal: "true",
        },
        {
          field: "Block Out-of-Scope Requests",
          type: "boolean",
          description: "Drops all network connections through the MITM proxy targeting unlisted hosts.",
          defaultVal: "true",
        },
      ]}
      gotchas={[
        {
          title: "Cloud Provider & CDN IP Ranges",
          description:
            "If your target is hosted on Cloudflare or AWS CloudFront, scoping by IP CIDR can unintentionally include other customers. Scope by hostname pattern instead.",
        },
        {
          title: "Third-Party OAuth Redirects",
          description:
            "Testing login flows often redirects to accounts.google.com or github.com. If strict blocking is enabled, login flows may fail unless OAuth domains are handled.",
        },
      ]}
    />
  );
}
