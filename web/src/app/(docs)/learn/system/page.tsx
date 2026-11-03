"use client";

import Link from "next/link";

import {
  ArrowRight,
  Ban,
  BellRing,
  Bot,
  Brain,
  ClipboardList,
  Plug,
  ScrollText,
  Settings2,
  ShieldAlert,
  Sparkles,
  Wrench,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

import { DocCallout } from "../_components/doc-callout";
import { DocsNavCards } from "../_components/docs-nav-cards";
import { DocsToc, type TocItem } from "../_components/docs-toc";

interface SysModuleSummary {
  name: string;
  route: string;
  consoleUrl: string;
  icon: typeof Settings2;
  badge: string;
  description: string;
}

const SYSTEM_MODULES: SysModuleSummary[] = [
  {
    name: "LLM",
    route: "/learn/system/llm",
    consoleUrl: "/system/llm",
    icon: Brain,
    badge: "AI Providers",
    description: "Configure AI providers, API keys, endpoints, and token generation parameters.",
  },
  {
    name: "Agent",
    route: "/learn/system/agent",
    consoleUrl: "/system/agents",
    icon: Bot,
    badge: "Orchestration",
    description: "Manage autonomous agent personas, system prompts, and authorized skills.",
  },
  {
    name: "MCP",
    route: "/learn/system/mcp",
    consoleUrl: "/system/mcp",
    icon: Plug,
    badge: "Protocol Bridge",
    description: "Connect external Model Context Protocol stdio or SSE servers.",
  },
  {
    name: "Skill",
    route: "/learn/system/skill",
    consoleUrl: "/system/skills",
    icon: Sparkles,
    badge: "Tool Packs",
    description: "Inspect specialized penetration testing skill packs and SKILL.md specs.",
  },
  {
    name: "Tools",
    route: "/learn/system/tools",
    consoleUrl: "/system/tools",
    icon: Wrench,
    badge: "Binary Registry",
    description: "Atomic command-line tool registry, binary health checks, and JSON schemas.",
  },
  {
    name: "Alerts",
    route: "/learn/system/alerts",
    consoleUrl: "/system/notify",
    icon: BellRing,
    badge: "Notifications",
    description: "Outbound notification channels (webhooks, email) for critical events.",
  },
  {
    name: "Rules",
    route: "/learn/system/rules",
    consoleUrl: "/system/intercept",
    icon: ShieldAlert,
    badge: "Safety Gates",
    description: "Tool execution intercept rules: allow, block, or require human approval.",
  },
  {
    name: "Scope",
    route: "/learn/system/scope",
    consoleUrl: "/system/intercept/assets",
    icon: Ban,
    badge: "Boundaries",
    description: "Target boundary enforcement: permitted domains, IPs, and exclusion lists.",
  },
  {
    name: "Approvals",
    route: "/learn/system/approvals",
    consoleUrl: "/system/intercept/approvals",
    icon: ClipboardList,
    badge: "Human Gate",
    description: "Human-in-the-loop decision queue for paused high-risk tool operations.",
  },
  {
    name: "Logs",
    route: "/learn/system/logs",
    consoleUrl: "/system/logs",
    icon: ScrollText,
    badge: "Diagnostics",
    description: "Live daemon diagnostics, database traces, proxy logs, and export bundle.",
  },
  {
    name: "Settings",
    route: "/learn/system/settings",
    consoleUrl: "/system/settings",
    icon: Settings2,
    badge: "Configuration",
    description: "Admin credentials, system identity, database options, and UI preferences.",
  },
];

const TOC_ITEMS: TocItem[] = [
  { id: "overview", title: "System Hub" },
  { id: "modules-directory", title: "Module Directory" },
  { id: "safety-architecture", title: "Safety Architecture" },
];

export default function SystemHubPage() {
  return (
    <div className="flex items-start justify-center gap-8 px-4 py-8 sm:px-6 lg:px-8 xl:gap-12">
      <div className="w-full min-w-0 max-w-3xl space-y-10">
        {/* Header */}
        <div id="overview" className="scroll-mt-20 space-y-3.5 border-border/60 border-b pb-6">
          <div className="flex items-center gap-2 font-mono text-primary text-xs uppercase tracking-wider">
            <Settings2 className="h-4 w-4" />
            <span>Console Mirror</span>
          </div>
          <h1 className="font-bold text-2xl text-foreground tracking-tight sm:text-3xl">System Reference</h1>
          <p className="text-muted-foreground text-sm leading-relaxed sm:text-base">
            Complete documentation for all 11 system governance and configuration modules in the ARTEX console.
            Configure models, manage skill packs, set up safety gates, and enforce legal testing scope.
          </p>
          <div className="flex flex-wrap gap-2 pt-1">
            <Badge variant="outline" className="border-primary/40 font-mono text-primary text-xs">
              11 Dedicated Modules
            </Badge>
            <Badge variant="outline" className="border-border/60 font-mono text-muted-foreground text-xs">
              Safety & Governance
            </Badge>
          </div>
        </div>

        {/* Directory Grid */}
        <section id="modules-directory" className="scroll-mt-20 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold text-foreground text-lg tracking-tight">System Modules</h2>
            <span className="font-mono text-muted-foreground text-xs">11 Guides</span>
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            {SYSTEM_MODULES.map((mod) => {
              const Icon = mod.icon;
              return (
                <Link key={mod.route} href={mod.route} className="group block focus:outline-hidden">
                  <Card className="h-full border-border/60 transition-all hover:border-primary/50 hover:bg-muted/30">
                    <CardContent className="space-y-2 p-4">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 font-semibold text-foreground text-sm transition-colors group-hover:text-primary">
                          <Icon className="h-4 w-4 text-primary" />
                          <span>{mod.name}</span>
                        </div>
                        <Badge variant="outline" className="font-mono text-[10px] text-muted-foreground">
                          {mod.badge}
                        </Badge>
                      </div>
                      <p className="line-clamp-2 text-muted-foreground text-xs leading-relaxed">{mod.description}</p>
                      <div className="flex items-center gap-1 pt-1 font-mono text-[11px] text-primary">
                        <span>Read guide</span>
                        <ArrowRight className="h-3 w-3 transition-transform group-hover:translate-x-1" />
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              );
            })}
          </div>
        </section>

        {/* Safety Architecture */}
        <section id="safety-architecture" className="scroll-mt-20 space-y-3">
          <h2 className="font-semibold text-foreground text-lg tracking-tight">Governance & Safety Pipeline</h2>
          <DocCallout type="warning" title="Defense in Depth">
            Before an agent tool call executes on the server, it passes two layers of safety gates: First,{" "}
            <strong>Scope</strong> verifies the destination host is permitted. Second, <strong>Rules</strong> evaluate
            arguments against intercept policies. If a rule specifies 'Require Approval', execution pauses and creates
            an entry in <strong>Approvals</strong>.
          </DocCallout>
        </section>

        <DocsNavCards />
      </div>

      <DocsToc items={TOC_ITEMS} />
    </div>
  );
}
