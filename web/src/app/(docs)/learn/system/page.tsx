"use client";

import * as React from "react";

import Link from "next/link";

import {
  ArrowRight,
  Ban,
  BellRing,
  Bot,
  Brain,
  ClipboardList,
  HelpCircle,
  type LucideIcon,
  Plug,
  ScrollText,
  Search,
  Settings2,
  ShieldAlert,
  Sparkles,
  Wrench,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

import { DocCallout } from "../_components/doc-callout";
import { DocsNavCards } from "../_components/docs-nav-cards";
import { DocsToc, type TocItem } from "../_components/docs-toc";

interface SysModuleDoc {
  id: string;
  name: string;
  route: string;
  icon: LucideIcon;
  badge: string;
  purpose: string;
  prerequisites: string;
  uiInstructions: string[];
  realWorldExample: string;
  internalMechanism: string;
}

const SYSTEM_MODULES: SysModuleDoc[] = [
  {
    id: "llm",
    name: "LLM",
    route: "/system/llm",
    icon: Brain,
    badge: "AI Providers",
    purpose: "Configure AI model providers, API credentials, base URLs, and token generation parameters.",
    prerequisites: "An API key from OpenAI, Anthropic, or an OpenAI-compatible endpoint (like Ollama or vLLM).",
    uiInstructions: [
      "Click LLM (/system/llm) in the SYSTEM sidebar section.",
      "Click Add Provider to create a new provider configuration.",
      "Select your provider type (OpenAI, Anthropic, or Custom).",
      "Paste your API key and set the Model Name (e.g., gpt-4o).",
      "Click Test Connection to verify authentication, then click Save.",
    ],
    realWorldExample:
      "Connecting an OpenAI profile with key sk-proj-... and setting model gpt-4o to serve as the default planner model.",
    internalMechanism:
      "Stores provider configurations in the llm_providers PostgreSQL table with AES-encrypted API keys. Evaluated by llmpool/ at runtime.",
  },
  {
    id: "agents",
    name: "Agent",
    route: "/system/agents",
    icon: Bot,
    badge: "Agent Orchestration",
    purpose:
      "Build and customize autonomous AI agents, their system prompts, personality roles, and authorized skills.",
    prerequisites: "At least one active LLM profile configured in System > LLM.",
    uiInstructions: [
      "Click Agent (/system/agents) in the sidebar.",
      "Click Create Agent to open the agent builder modal.",
      "Enter a Name and Role description (e.g., 'Web Exploitation Specialist').",
      "Select which LLM provider and model this agent uses.",
      "Toggle authorized skills and tool packs, then click Save Agent.",
    ],
    realWorldExample:
      "Creating a 'Fast Recon' agent with a concise system prompt and access only to subdomain and port scanning tools.",
    internalMechanism:
      "Agent definitions are saved to the agents table. When a chat session starts, agent/orchestrator.go binds the agent prompt and tool list into the LLM context.",
  },
  {
    id: "mcp",
    name: "MCP",
    route: "/system/mcp",
    icon: Plug,
    badge: "Protocol Bridge",
    purpose:
      "Connect external Model Context Protocol (MCP) servers to extend ARTEX with third-party tools and data sources.",
    prerequisites: "A local command (stdio) or remote URL (SSE) implementing the Model Context Protocol.",
    uiInstructions: [
      "Click MCP (/system/mcp) in the SYSTEM menu.",
      "Click Add MCP Server.",
      "Enter Server Name and choose Transport Type (Stdio or SSE).",
      "For stdio, specify the command and arguments; for SSE, specify the HTTP endpoint URL.",
      "Click Connect and verify the discovered tools in the tools list.",
    ],
    realWorldExample:
      "Connecting a GitHub MCP server to allow the agent to check open issues or read repositories during white-box assessments.",
    internalMechanism:
      "Launches child processes over standard I/O or connects via Server-Sent Events, registering remote tool schemas into the internal tool catalog.",
  },
  {
    id: "skills",
    name: "Skill",
    route: "/system/skills",
    icon: Sparkles,
    badge: "Tool Packs",
    purpose:
      "Manage specialized penetration testing skill packs that bundle tool definitions, instruction guidelines, and workflows.",
    prerequisites: "Skill directories located in the skills/ repository folder with valid SKILL.md manifests.",
    uiInstructions: [
      "Click Skill (/system/skills) in the left sidebar.",
      "Browse available skill packs such as api-recon, scopesentry, or playwright-cli.",
      "Click on any skill to read its SKILL.md instructions and exported tools.",
      "Toggle the Enable switch to make the skill available to agents.",
    ],
    realWorldExample:
      "Enabling the 'api-recon' skill pack to give your agent automated OpenAPI/Swagger spec parsing capabilities.",
    internalMechanism:
      "The skills/ directory is scanned at startup. Markdown frontmatter and tool specifications are parsed and registered into the agent execution registry.",
  },
  {
    id: "tools",
    name: "Tools",
    route: "/system/tools",
    icon: Wrench,
    badge: "Binary Registry",
    purpose:
      "Inspect all atomic command-line tools and utilities registered in the system, checking binary health and JSON schemas.",
    prerequisites: "Binaries (like nmap, curl, ffuf) installed on the host system PATH.",
    uiInstructions: [
      "Click Tools (/system/tools) in the sidebar.",
      "Review the registered tools table with names, descriptions, and binary status indicators.",
      "Click any tool to view its JSON schema parameters.",
      "Check the Installed status to verify if the required CLI binary exists on the server.",
    ],
    realWorldExample:
      "Verifying that 'nuclei' is installed and healthy, checking that parameter '-target' is marked as required in the schema.",
    internalMechanism:
      "Validates binary existence using exec.LookPath in Go. Tool parameters are validated against JSON schema definitions before process invocation.",
  },
  {
    id: "notify",
    name: "Alerts",
    route: "/system/notify",
    icon: BellRing,
    badge: "Event Notifications",
    purpose:
      "Set up outbound notification channels (webhooks, email, chat bots) when critical findings or approval requests occur.",
    prerequisites: "A webhook endpoint (Slack, Discord, Teams) or SMTP email server credentials.",
    uiInstructions: [
      "Click Alerts (/system/notify) in the sidebar.",
      "Click Add Notification Channel.",
      "Select Channel Type (Webhook, Email, Slack, etc.).",
      "Paste your Webhook URL and select Trigger Events (Critical Finding, Approval Pending, Scan Completed).",
      "Click Send Test Notification, then click Save.",
    ],
    realWorldExample:
      "Configuring a Slack incoming webhook that sends a message into #security-alerts whenever a High or Critical finding is confirmed.",
    internalMechanism:
      "The notify/ package subscribes to internal event channels. Matching events are formatted and dispatched asynchronously with retry backoff.",
  },
  {
    id: "intercept",
    name: "Rules",
    route: "/system/intercept",
    icon: ShieldAlert,
    badge: "Safety Gates",
    purpose:
      "Define granular intercept rules that automatically allow, block, or require human approval on tool calls.",
    prerequisites: "None. Rules take effect immediately upon creation.",
    uiInstructions: [
      "Click Rules (/system/intercept) in the left menu.",
      "Click Create Intercept Rule.",
      "Specify rule criteria: Tool Name (e.g., 'sqlmap'), Command regex, or argument pattern.",
      "Choose Action: Allow, Block, or Require Approval.",
      "Set Priority and toggle the rule Active, then click Save.",
    ],
    realWorldExample:
      "Creating a rule: Tool Name = 'nmap' and arguments match '-p-' (all ports) -> Action: Require Approval, preventing unintentional denial of service.",
    internalMechanism:
      "Evaluated in intercept/eval.go before any tool executes. If an action is 'Require Approval', a pending approval record is created in PostgreSQL.",
  },
  {
    id: "asset-intercept",
    name: "Scope",
    route: "/system/intercept/assets",
    icon: Ban,
    badge: "Target Boundaries",
    purpose:
      "Define the strict target boundary for assessments by specifying permitted hostnames, IP addresses, and CIDR blocks.",
    prerequisites: "Written authorization for all domains or networks added to the scope.",
    uiInstructions: [
      "Click Scope (/system/intercept/assets) in the sidebar.",
      "Click Add Asset Scope.",
      "Enter the target hostname, domain pattern (e.g., '*.example.com'), or CIDR block (192.168.1.0/24).",
      "Select Action: Allow or Exclude (Banned).",
      "Click Save. Any outbound request to an unlisted or banned host is blocked.",
    ],
    realWorldExample:
      "Adding 'example.com' as Allow, and adding 'payments.example.com' as Exclude to keep the payment gateway out of testing scope.",
    internalMechanism:
      "The guard/ package checks destination hostnames and IP addresses against the asset scope table before executing network calls or proxy requests.",
  },
  {
    id: "approvals",
    name: "Approvals",
    route: "/system/intercept/approvals",
    icon: ClipboardList,
    badge: "Human Gate",
    purpose: "Review and decide on paused tool execution requests waiting for human permission.",
    prerequisites: "Requires a running task that triggered an intercept rule with action 'Require Approval'.",
    uiInstructions: [
      "Click Approvals (/system/intercept/approvals) in the sidebar (or click the amber badge).",
      "Inspect the pending request card showing Tool Name, Target, Arguments, and Triggered Rule.",
      "Review the command to ensure parameters are safe.",
      "Click Approve to permit execution immediately, or click Reject to cancel the tool call.",
    ],
    realWorldExample:
      "The agent attempts to execute 'ffuf -w wordlist.txt -u https://example.com/FUZZ'. The approval card shows the request; you review and click Approve.",
    internalMechanism:
      "The executing worker goroutine is paused waiting on a Go channel in guard/wait.go. Submitting Approve sends a resume signal, unblocking the worker.",
  },
  {
    id: "logs",
    name: "Logs",
    route: "/system/logs",
    icon: ScrollText,
    badge: "System Diagnostics",
    purpose: "Monitor live daemon log entries, diagnostic traces, database queries, and proxy errors in real time.",
    prerequisites: "Backend server running with logging enabled.",
    uiInstructions: [
      "Click Logs (/system/logs) in the left menu.",
      "Toggle Auto-scroll to follow new log entries as they arrive.",
      "Filter logs by Level: Info, Warning, or Error.",
      "Use the search box to filter for specific task IDs or error codes.",
      "Click Clear or Download Logs to export the diagnostic log bundle.",
    ],
    realWorldExample:
      "Troubleshooting a failed tool execution: searching the logs for 'exit code 127' reveals that a required CLI dependency is missing on the host.",
    internalMechanism:
      "The Go logger outputs to an in-memory ring buffer while broadcasting structured log events to web clients over the WebSocket /ws endpoint.",
  },
  {
    id: "settings",
    name: "Settings",
    route: "/system/settings",
    icon: Settings2,
    badge: "Core Configuration",
    purpose: "Manage system identity, admin password, database connection info, and interface theme defaults.",
    prerequisites: "Admin access to the ARTEX instance.",
    uiInstructions: [
      "Click Settings (/system/settings) in the sidebar.",
      "Update your Admin Password under the Security section.",
      "Configure System Display Name and default interface preferences.",
      "Click Save Settings to apply the configuration changes.",
    ],
    realWorldExample:
      "Updating the system display name to 'Acme Red Team Pentest' and resetting the admin password before handing off the console.",
    internalMechanism:
      "Stores key-value pairs in the settings PostgreSQL table. Settings changes are applied immediately to running services without requiring a restart.",
  },
];

const TOC_ITEMS: TocItem[] = SYSTEM_MODULES.map((m) => ({
  id: m.id,
  title: m.name,
}));

export default function SystemPage() {
  const [filter, setFilter] = React.useState("");

  const filteredModules = React.useMemo(() => {
    const q = filter.trim().toLowerCase();
    if (!q) return SYSTEM_MODULES;
    return SYSTEM_MODULES.filter(
      (m) =>
        m.name.toLowerCase().includes(q) ||
        m.purpose.toLowerCase().includes(q) ||
        m.realWorldExample.toLowerCase().includes(q),
    );
  }, [filter]);

  return (
    <div className="flex items-start justify-center gap-8 px-4 py-8 sm:px-6 lg:px-8 xl:gap-12">
      {/* Center Reading Container */}
      <div className="w-full min-w-0 max-w-3xl space-y-12">
        {/* Title Header */}
        <div className="space-y-4 border-border/60 border-b pb-8">
          <div className="flex items-center gap-2 font-mono text-primary text-xs uppercase tracking-wider">
            <Settings2 className="h-4 w-4" />
            <span>System Guide</span>
          </div>
          <h1 className="font-bold text-3xl text-foreground tracking-tight sm:text-4xl">System Reference</h1>
          <p className="text-base text-muted-foreground leading-relaxed sm:text-lg">
            Complete guide to all 11 system configuration modules in ARTEX. Covers AI models, agent definitions, MCP
            protocols, skill tool packs, safety intercept gates, scope boundaries, and settings.
          </p>

          {/* Quick search filter */}
          <div className="relative pt-2 sm:w-80">
            <Search className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search system modules..."
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="pl-9 text-xs"
            />
          </div>
        </div>

        {/* Modules List */}
        <div className="space-y-12">
          {filteredModules.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-lg border border-dashed p-10 text-center">
              <HelpCircle className="mb-2 h-8 w-8 text-muted-foreground/60" />
              <p className="font-medium text-sm">{`No system modules match "${filter}"`}</p>
              <Button variant="ghost" size="sm" onClick={() => setFilter("")} className="mt-2 text-xs">
                Clear filter
              </Button>
            </div>
          ) : (
            filteredModules.map((module) => {
              const Icon = module.icon;
              return (
                <section
                  key={module.id}
                  id={module.id}
                  className="scroll-mt-20 space-y-4 rounded-xl border border-border/60 bg-card/40 p-5 transition-colors hover:border-border"
                >
                  {/* Module Title & Link */}
                  <div className="flex items-center justify-between border-border/60 border-b pb-3">
                    <div className="flex items-center gap-3">
                      <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                        <Icon className="h-5 w-5" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <h2 className="font-semibold text-foreground text-xl tracking-tight">{module.name}</h2>
                          <Badge variant="outline" className="font-mono text-[10px]">
                            {module.badge}
                          </Badge>
                        </div>
                        <code className="font-mono text-muted-foreground text-xs">{module.route}</code>
                      </div>
                    </div>
                    <Button variant="outline" size="sm" asChild className="h-8 gap-1.5 text-xs">
                      <Link href={module.route}>
                        <span>Open {module.name}</span>
                        <ArrowRight className="h-3 w-3" />
                      </Link>
                    </Button>
                  </div>

                  {/* 1. Purpose & Prerequisites */}
                  <div className="space-y-2">
                    <p className="text-foreground/90 text-sm leading-relaxed">{module.purpose}</p>
                    <div className="inline-flex items-center gap-1.5 rounded-md border border-border/60 bg-muted/30 px-2.5 py-1 text-xs text-muted-foreground">
                      <span className="font-semibold text-foreground">Prerequisites:</span>
                      <span>{module.prerequisites}</span>
                    </div>
                  </div>

                  {/* 2. Step-by-Step Instructions */}
                  <div className="space-y-2 pt-1">
                    <div className="font-mono font-semibold text-[11px] text-muted-foreground uppercase tracking-wider">
                      How to configure in UI
                    </div>
                    <div className="space-y-1.5">
                      {module.uiInstructions.map((step, idx) => (
                        <div key={step} className="flex items-start gap-2.5 text-xs text-foreground/90">
                          <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-primary/10 font-mono text-[10px] font-bold text-primary">
                            {idx + 1}
                          </span>
                          <span className="leading-relaxed">{step}</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* 3. Real-World Example */}
                  <DocCallout type="tip" title="Real-World Example" className="my-2">
                    {module.realWorldExample}
                  </DocCallout>

                  {/* 4. Internal Mechanism */}
                  <div className="rounded-lg border border-border/50 bg-muted/20 p-3 text-xs text-muted-foreground leading-relaxed">
                    <strong className="font-mono text-[11px] text-foreground">Under the hood: </strong>
                    {module.internalMechanism}
                  </div>
                </section>
              );
            })
          )}
        </div>

        {/* Previous / Next Article Navigation Cards */}
        <DocsNavCards />
      </div>

      {/* Right Desktop Table of Contents */}
      <DocsToc items={TOC_ITEMS} />
    </div>
  );
}
