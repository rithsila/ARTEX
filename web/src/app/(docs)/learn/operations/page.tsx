"use client";

import Link from "next/link";

import {
  Activity,
  ArrowRight,
  Bug,
  FolderOpen,
  FolderSync,
  LayoutDashboard,
  MessageSquare,
  Network,
  Radio,
  Target,
  Terminal,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

import { DocCallout } from "../_components/doc-callout";
import { DocsNavCards } from "../_components/docs-nav-cards";
import { DocsToc, type TocItem } from "../_components/docs-toc";

interface OpModuleSummary {
  name: string;
  route: string;
  consoleUrl: string;
  icon: typeof LayoutDashboard;
  badge: string;
  description: string;
}

const OPERATIONS_MODULES: OpModuleSummary[] = [
  {
    name: "Dashboard",
    route: "/learn/operations/dashboard",
    consoleUrl: "/dashboard",
    icon: LayoutDashboard,
    badge: "Overview",
    description: "Engagement overview, active tasks, finding distributions, and traffic volume.",
  },
  {
    name: "Chat",
    route: "/learn/operations/chat",
    consoleUrl: "/chat",
    icon: MessageSquare,
    badge: "Control Center",
    description: "Interactive AI agent mission console, streaming thoughts, and tool feedback.",
  },
  {
    name: "Tasks",
    route: "/learn/operations/tasks",
    consoleUrl: "/function/tasks",
    icon: Target,
    badge: "Execution Queue",
    description: "Decomposed task execution queue, status inspector, and failure restarts.",
  },
  {
    name: "Findings",
    route: "/learn/operations/findings",
    consoleUrl: "/function/findings",
    icon: Bug,
    badge: "Evidence Store",
    description: "Confirmed vulnerabilities, CVSS scoring, PoC reproduction steps, and exports.",
  },
  {
    name: "Traffic",
    route: "/learn/operations/traffic",
    consoleUrl: "/function/traffic",
    icon: Activity,
    badge: "MITM :8788",
    description: "Built-in MITM recording proxy on port 8788 and SQLite request archive.",
  },
  {
    name: "Exec",
    route: "/learn/operations/exec",
    consoleUrl: "/function/commands",
    icon: Terminal,
    badge: "Shell Execution",
    description: "Direct shell execution console for human command verification.",
  },
  {
    name: "LLM Streams",
    route: "/learn/operations/llm-streams",
    consoleUrl: "/function/llm-records",
    icon: Radio,
    badge: "Prompt Audit",
    description: "Real-time token audit, prompt inspect, and completion latency inspector.",
  },
  {
    name: "Assets",
    route: "/learn/operations/assets",
    consoleUrl: "/function/assets",
    icon: Network,
    badge: "Attack Surface",
    description: "Discovered hostnames, open ports, web tech stacks, and alive status tracking.",
  },
  {
    name: "Sync",
    route: "/learn/operations/sync",
    consoleUrl: "/function/sync",
    icon: FolderSync,
    badge: "Data Sync",
    description: "Bidirectional sync with external tools, wordlists, and scanner outputs.",
  },
  {
    name: "Workspace",
    route: "/learn/operations/workspace",
    consoleUrl: "/function/workspace",
    icon: FolderOpen,
    badge: "Artifact Storage",
    description: "Local file system workspace, tool artifacts, screenshots, and logs.",
  },
];

const TOC_ITEMS: TocItem[] = [
  { id: "overview", title: "Operations Hub" },
  { id: "modules-directory", title: "Module Directory" },
  { id: "architecture-role", title: "Architecture Role" },
];

export default function OperationsHubPage() {
  return (
    <div className="flex items-start justify-center gap-8 px-4 py-8 sm:px-6 lg:px-8 xl:gap-12">
      <div className="w-full min-w-0 max-w-3xl space-y-10">
        {/* Header */}
        <div id="overview" className="scroll-mt-20 space-y-3.5 border-border/60 border-b pb-6">
          <div className="flex items-center gap-2 font-mono text-primary text-xs uppercase tracking-wider">
            <Activity className="h-4 w-4" />
            <span>Console Mirror</span>
          </div>
          <h1 className="font-bold text-2xl text-foreground tracking-tight sm:text-3xl">Operations Reference</h1>
          <p className="text-muted-foreground text-sm leading-relaxed sm:text-base">
            Detailed guides for all 10 operational modules mirroring the ARTEX console navigation. Explore live
            telemetry, task queues, finding repositories, and MITM proxy logs.
          </p>
          <div className="flex flex-wrap gap-2 pt-1">
            <Badge variant="outline" className="border-primary/40 font-mono text-primary text-xs">
              10 Dedicated Modules
            </Badge>
            <Badge variant="outline" className="border-border/60 font-mono text-muted-foreground text-xs">
              Real-Time Telemetry
            </Badge>
          </div>
        </div>

        {/* Directory Grid */}
        <section id="modules-directory" className="scroll-mt-20 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold text-foreground text-lg tracking-tight">Operations Modules</h2>
            <span className="font-mono text-muted-foreground text-xs">10 Guides</span>
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            {OPERATIONS_MODULES.map((mod) => {
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

        {/* Architecture Role */}
        <section id="architecture-role" className="scroll-mt-20 space-y-3">
          <h2 className="font-semibold text-foreground text-lg tracking-tight">How Operations Work Together</h2>
          <DocCallout type="info" title="Unified Event Pipeline">
            When you type a goal in <strong>Chat</strong>, the Planner Agent breaks it into <strong>Tasks</strong>.
            Tasks invoke tools whose HTTP requests route through the <strong>Traffic</strong> MITM proxy (:8788).
            Discovered nodes are indexed into <strong>Assets</strong>, proven vulnerabilities are saved into{" "}
            <strong>Findings</strong>, and all telemetry reflects instantly on the <strong>Dashboard</strong>.
          </DocCallout>
        </section>

        <DocsNavCards />
      </div>

      <DocsToc items={TOC_ITEMS} />
    </div>
  );
}
