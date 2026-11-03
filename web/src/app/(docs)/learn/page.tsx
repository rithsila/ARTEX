"use client";

import Link from "next/link";

import {
  Activity,
  ArrowRight,
  BookOpen,
  CheckCircle2,
  Cpu,
  Database,
  Layers,
  Network,
  Rocket,
  Shield,
  ShieldAlert,
  Terminal,
  Workflow,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

import { DocCallout } from "./_components/doc-callout";
import { DocCodeBlock } from "./_components/doc-code-block";
import { DocStep, DocSteps } from "./_components/doc-steps";
import { DocsNavCards } from "./_components/docs-nav-cards";
import { DocsToc, type TocItem } from "./_components/docs-toc";

const TOC_ITEMS: TocItem[] = [
  { id: "purpose-and-prerequisites", title: "Purpose & Prerequisites" },
  { id: "system-architecture", title: "System Architecture" },
  { id: "core-components", title: "Core Components" },
  { id: "step-by-step-lifecycle", title: "Step-by-Step Lifecycle" },
  { id: "real-world-example", title: "Real-World Example" },
  { id: "internal-mechanisms", title: "Internal Mechanisms" },
  { id: "next-steps", title: "Next Steps" },
];

export default function OverviewPage() {
  return (
    <div className="flex items-start justify-center gap-8 px-4 py-8 sm:px-6 lg:px-8 xl:gap-12">
      {/* Center Reading Container */}
      <div className="w-full min-w-0 max-w-3xl space-y-12">
        {/* Title Header */}
        <div className="space-y-4 border-border/60 border-b pb-8">
          <div className="flex items-center gap-2 font-mono text-primary text-xs uppercase tracking-wider">
            <BookOpen className="h-4 w-4" />
            <span>Architecture & Overview</span>
          </div>
          <h1 className="font-bold text-3xl text-foreground tracking-tight sm:text-4xl">Learn ARTEX</h1>
          <p className="text-base text-muted-foreground leading-relaxed sm:text-lg">
            ARTEX is an autonomous AI penetration testing platform. It plans security assessments, runs security tools,
            inspects web traffic, and asks for your approval before touching dangerous steps.
          </p>
          <div className="flex flex-wrap gap-2 pt-2">
            <Badge variant="outline" className="border-primary/40 font-mono text-xs text-primary">
              Autonomous Planner
            </Badge>
            <Badge variant="outline" className="border-border/60 font-mono text-xs text-muted-foreground">
              MITM Proxy :8788
            </Badge>
            <Badge variant="outline" className="border-border/60 font-mono text-xs text-muted-foreground">
              Human-in-the-Loop Intercept
            </Badge>
            <Badge variant="outline" className="border-border/60 font-mono text-xs text-muted-foreground">
              PostgreSQL 16+
            </Badge>
          </div>
        </div>

        {/* 1. Purpose & Prerequisites */}
        <section id="purpose-and-prerequisites" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Shield className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Purpose & Prerequisites</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            The purpose of ARTEX is to safely automate repetitive vulnerability scanning, reconnaissance, and
            proof-of-concept verification while keeping humans in total control of risky actions.
          </p>

          <div className="grid gap-4 sm:grid-cols-2 pt-2">
            <div className="rounded-xl border border-border/60 bg-card/60 p-5 space-y-2.5">
              <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                <CheckCircle2 className="h-4 w-4 text-primary" />
                What ARTEX Does
              </div>
              <ul className="space-y-1.5 text-xs text-muted-foreground leading-relaxed">
                <li>• Breaks complex penetration testing goals into bite-sized tasks.</li>
                <li>• Runs tools (nmap, ffuf, nuclei, sqlmap, custom scripts) automatically.</li>
                <li>• Records every outbound HTTP request through a dedicated MITM proxy.</li>
                <li>• Deduplicates confirmed security flaws into clear Finding records.</li>
              </ul>
            </div>

            <div className="rounded-xl border border-border/60 bg-card/60 p-5 space-y-2.5">
              <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                <CheckCircle2 className="h-4 w-4 text-primary" />
                System Prerequisites
              </div>
              <ul className="space-y-1.5 text-xs text-muted-foreground leading-relaxed">
                <li>
                  • <strong className="text-foreground">Go backend:</strong> Go 1.26+ listening on{" "}
                  <code className="text-foreground font-mono">:8787</code>.
                </li>
                <li>
                  • <strong className="text-foreground">Database:</strong> PostgreSQL 16+ via{" "}
                  <code className="text-foreground font-mono">ARTEX_PG_DSN</code>.
                </li>
                <li>
                  • <strong className="text-foreground">AI Model:</strong> API key from OpenAI, Anthropic, or custom
                  endpoint.
                </li>
                <li>
                  • <strong className="text-foreground">Target Scope:</strong> Written authorization for all scanned
                  domains.
                </li>
              </ul>
            </div>
          </div>

          <DocCallout type="tip" title="Quick Tip">
            Before launching scans on any network, always add the domain or CIDR block to{" "}
            <strong className="text-foreground">System &rarr; Scope</strong>. ARTEX will strictly block any request to
            unlisted hosts.
          </DocCallout>
        </section>

        {/* 2. System Architecture */}
        <section id="system-architecture" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Layers className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">System Architecture</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            ARTEX is built as a single unified system combining a high-performance Go daemon, an in-process MITM proxy,
            persistent SQL storage, and a modern Next.js frontend.
          </p>

          <div className="rounded-xl border border-border/80 bg-zinc-950 p-5 font-mono text-xs text-zinc-300 space-y-4 shadow-sm">
            <div className="rounded-lg border border-zinc-800 bg-zinc-900/80 p-3.5 space-y-1">
              <div className="flex items-center justify-between text-primary font-semibold">
                <span>[FRONTEND] Next.js 16 + React 19</span>
                <span className="text-zinc-500 font-normal">Browser / Port :5173</span>
              </div>
              <p className="text-zinc-400 text-[11px]">
                Sidebar navigation, real-time WebSocket logs, interactive approvals, traffic viewer, and findings
                manager.
              </p>
            </div>

            <div className="flex justify-center text-zinc-500 text-xs">↓ JSON REST API & WebSocket (/api, /ws)</div>

            <div className="rounded-lg border border-primary/30 bg-primary/5 p-3.5 space-y-2">
              <div className="flex items-center justify-between text-primary font-semibold">
                <span>[BACKEND DAEMON] Go 1.26 Engine</span>
                <span className="text-primary font-normal">HTTP :8787</span>
              </div>
              <div className="grid gap-2 sm:grid-cols-3 pt-1 text-[11px]">
                <div className="rounded border border-zinc-800/80 bg-zinc-900/60 p-2">
                  <div className="font-semibold text-primary">agent/ Orchestrator</div>
                  <p className="text-zinc-400 text-[10px]">Planner loop & worker pools.</p>
                </div>
                <div className="rounded border border-zinc-800/80 bg-zinc-900/60 p-2">
                  <div className="font-semibold text-primary">guard/ & intercept/</div>
                  <p className="text-zinc-400 text-[10px]">Scope checks & approval channels.</p>
                </div>
                <div className="rounded border border-zinc-800/80 bg-zinc-900/60 p-2">
                  <div className="font-semibold text-primary">skills/ & tools/</div>
                  <p className="text-zinc-400 text-[10px]">Tool execution & MCP bridge.</p>
                </div>
              </div>
            </div>

            <div className="flex justify-center text-zinc-500 text-xs">↓ Proxy Capture & State Storage</div>

            <div className="grid gap-3 sm:grid-cols-2">
              <div className="rounded-lg border border-zinc-800 bg-zinc-900/80 p-3 space-y-1">
                <div className="flex items-center justify-between font-semibold text-zinc-200">
                  <span>MITM Proxy :8788</span>
                  <span className="text-zinc-500 text-[10px]">SQLite Index</span>
                </div>
                <p className="text-zinc-400 text-[11px]">
                  Streams requests and indexes headers into{" "}
                  <code className="text-zinc-200">data/traffic/_index/index.sqlite</code>.
                </p>
              </div>

              <div className="rounded-lg border border-zinc-800 bg-zinc-900/80 p-3 space-y-1">
                <div className="flex items-center justify-between font-semibold text-zinc-200">
                  <span>PostgreSQL 16+</span>
                  <span className="text-zinc-500 text-[10px]">Main DB</span>
                </div>
                <p className="text-zinc-400 text-[11px]">
                  Stores assets, tasks, findings, intercept rules, and admin settings.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* 3. Core Components */}
        <section id="core-components" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Cpu className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Core Components</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            Every penetration test in ARTEX involves four core systems working together:
          </p>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-xl border border-border/60 bg-card/40 p-4 space-y-2">
              <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                <Terminal className="h-4 w-4 text-primary" />
                1. Planner & Worker Agents
              </div>
              <p className="text-xs text-muted-foreground leading-relaxed">
                The Planner reads your goal in Chat, queries discovered assets, and breaks the mission into independent
                Tasks. Worker agents pick tasks from the queue and run security tools.
              </p>
            </div>

            <div className="rounded-xl border border-border/60 bg-card/40 p-4 space-y-2">
              <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                <ShieldAlert className="h-4 w-4 text-primary" />
                2. Guard & Intercept Engine
              </div>
              <p className="text-xs text-muted-foreground leading-relaxed">
                Before any tool runs, Guard checks target addresses against allowed Scope. Intercept rules check tool
                names and parameters. If a rule specifies &quot;Require Approval&quot;, execution freezes until you
                approve.
              </p>
            </div>

            <div className="rounded-xl border border-border/60 bg-card/40 p-4 space-y-2">
              <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                <Activity className="h-4 w-4 text-primary" />
                3. MITM Traffic Proxy (:8788)
              </div>
              <p className="text-xs text-muted-foreground leading-relaxed">
                All inspection tools send outbound traffic through HTTP proxy{" "}
                <code className="text-foreground font-mono">127.0.0.1:8788</code>. ARTEX records exact request and
                response headers and bodies, making every finding reproducible.
              </p>
            </div>

            <div className="rounded-xl border border-border/60 bg-card/40 p-4 space-y-2">
              <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                <Database className="h-4 w-4 text-primary" />
                4. Findings & Evidence Store
              </div>
              <p className="text-xs text-muted-foreground leading-relaxed">
                When a tool validates an issue (such as SQL injection or an open admin panel), ARTEX creates a Finding
                with severity, endpoint URL, reproduction steps, and raw request/response proof traces.
              </p>
            </div>
          </div>
        </section>

        {/* 4. Step-by-Step Lifecycle */}
        <section id="step-by-step-lifecycle" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Workflow className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Step-by-Step Lifecycle</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            The exact progression of an assessment inside ARTEX:
          </p>

          <DocSteps>
            <DocStep number={1} title="Define Scope" description="Authorizing target domains and IP boundaries">
              You enter allowed domains or CIDRs in{" "}
              <code className="font-mono text-primary">Scope (/system/intercept/assets)</code>. Unlisted hosts are
              blocked automatically.
            </DocStep>

            <DocStep number={2} title="Instruct in Chat" description="Submitting objectives to the autonomous planner">
              You open <code className="font-mono text-primary">Chat (/chat)</code> and give a testing goal (e.g.,{" "}
              <code className="text-foreground">&quot;Map endpoints and test for authentication flaws&quot;</code>).
            </DocStep>

            <DocStep number={3} title="Planning & Tasks" description="Decomposing the mission into discrete actions">
              The planner creates discrete task records visible on{" "}
              <code className="font-mono text-primary">Tasks (/function/tasks)</code>.
            </DocStep>

            <DocStep
              number={4}
              title="Approval Gate"
              description="Reviewing and granting permission for sensitive commands"
            >
              If a command matches an intercept rule, an amber badge lights up on{" "}
              <code className="font-mono text-primary">Approvals (/system/intercept/approvals)</code>. You inspect
              parameters and approve or deny.
            </DocStep>

            <DocStep
              number={5}
              title="Evidence & Findings"
              description="Deduplicated vulnerability records with raw proxy traces"
            >
              Verified security flaws appear in{" "}
              <code className="font-mono text-primary">Findings (/function/findings)</code> with full HTTP traffic
              captured in <code className="font-mono text-primary">Traffic (/function/traffic)</code>.
            </DocStep>
          </DocSteps>
        </section>

        {/* 5. Real-World Example */}
        <section id="real-world-example" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Network className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">
              Real-World Example: Scanning with Approval Gate
            </h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            A practical walkthrough assessing an authorized staging server with a safety guard:
          </p>

          <DocCodeBlock
            language="bash"
            title="Sample chat instruction"
            code={`# Chat Input:
Scan https://staging.example.com for input vulnerabilities and enumerate endpoints.`}
          />

          <DocCallout type="warning" title="Safety Intercept Triggered">
            When the agent detects database syntax errors on <code className="text-foreground">/api/search?q=</code>, it
            attempts to launch <code className="text-foreground">sqlmap</code>. The intercept rule matches and
            immediately halts execution. The amber badge on <strong className="text-foreground">Approvals</strong>{" "}
            illuminates.
          </DocCallout>

          <p className="text-xs text-muted-foreground leading-relaxed">
            After clicking <strong className="text-foreground">Approve</strong> in the web UI, the worker resumes,
            verifies the SQL injection, and logs a High Severity Finding with the exact HTTP request/response evidence.
          </p>
        </section>

        {/* 6. Internal Mechanisms */}
        <section id="internal-mechanisms" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Database className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Internal Mechanisms</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            How ARTEX executes commands, stores records, and prevents race conditions:
          </p>

          <div className="space-y-3 text-xs text-muted-foreground leading-relaxed">
            <div className="rounded-xl border border-border/60 bg-card/40 p-4 space-y-1.5">
              <div className="font-semibold text-foreground text-sm flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-primary" />
                Go Advisory Lock (7337741001)
              </div>
              <p>
                On startup, the Go backend calls <code className="text-foreground font-mono">db.Open()</code> which
                acquires PostgreSQL advisory lock <code className="text-foreground font-mono">7337741001</code>. This
                synchronizes embedded SQL schema migrations across instances and prevents migration conflicts.
              </p>
            </div>

            <div className="rounded-xl border border-border/60 bg-card/40 p-4 space-y-1.5">
              <div className="font-semibold text-foreground text-sm flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-primary" />
                MITM Proxy &amp; SQLite Fast Indexing
              </div>
              <p>
                The recording proxy in <code className="text-foreground font-mono">traffic/</code> listens on port{" "}
                <code className="text-foreground font-mono">:8788</code>. As requests stream through, body payloads are
                stored to disk while metadata is indexed into a pure-Go SQLite database (
                <code className="text-foreground font-mono">data/traffic/_index/index.sqlite</code>).
              </p>
            </div>

            <div className="rounded-xl border border-border/60 bg-card/40 p-4 space-y-1.5">
              <div className="font-semibold text-foreground text-sm flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-primary" />
                Wait Channels for Human Approval
              </div>
              <p>
                When a tool call triggers an approval gate in{" "}
                <code className="text-foreground font-mono">guard/intercept.go</code>, the executing goroutine pauses on
                an internal Go response channel. Once the user approves via{" "}
                <code className="text-foreground font-mono">/api/intercept/action</code>, the channel unblocks and the
                tool starts.
              </p>
            </div>
          </div>
        </section>

        {/* 7. Next Steps */}
        <section id="next-steps" className="scroll-mt-20 space-y-4 border-border/60 border-t pt-8">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-semibold text-lg text-foreground">Ready to start?</h3>
              <p className="text-muted-foreground text-xs">
                Follow our 6-step walkthrough to get ARTEX running and perform your first scan.
              </p>
            </div>
            <Button asChild className="gap-2">
              <Link href="/learn/quickstart">
                <span>Start Quickstart</span>
                <Rocket className="h-4 w-4" />
              </Link>
            </Button>
          </div>
        </section>

        {/* Previous / Next Article Navigation Cards */}
        <DocsNavCards />
      </div>

      {/* Right Desktop Table of Contents */}
      <DocsToc items={TOC_ITEMS} />
    </div>
  );
}
