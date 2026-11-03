"use client";

import Link from "next/link";

import {
  ArrowRight,
  Cpu,
  Database,
  Lock,
  PauseCircle,
  ShieldAlert,
  ShieldCheck,
  Terminal,
  Workflow,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

import { DocCallout } from "../_components/doc-callout";
import { DocCodeBlock } from "../_components/doc-code-block";
import { DocStep, DocSteps } from "../_components/doc-steps";
import { DocsNavCards } from "../_components/docs-nav-cards";
import { DocsToc, type TocItem } from "../_components/docs-toc";

const TOC_ITEMS: TocItem[] = [
  { id: "purpose-and-prerequisites", title: "Purpose & Prerequisites" },
  { id: "autonomous-loop-phases", title: "Autonomous Loop Phases" },
  { id: "intercept-engine", title: "Intercept Engine & Human Gate" },
  { id: "rule-precedence", title: "Rule Evaluation Precedence" },
  { id: "step-by-step-rules", title: "Step-by-Step: Adding Approval Rules" },
  { id: "real-world-example", title: "Real-World Walkthrough" },
  { id: "backend-mechanisms", title: "Go Backend Implementation Details" },
];

export default function WorkflowPage() {
  return (
    <div className="flex items-start justify-center gap-8 px-4 py-8 sm:px-6 lg:px-8 xl:gap-12">
      {/* Center Reading Container */}
      <div className="w-full min-w-0 max-w-3xl space-y-12">
        {/* Title Header */}
        <div className="space-y-4 border-border/60 border-b pb-8">
          <div className="flex items-center gap-2 font-mono text-primary text-xs uppercase tracking-wider">
            <Workflow className="h-4 w-4" />
            <span>Under the Hood</span>
          </div>
          <h1 className="font-bold text-3xl text-foreground tracking-tight sm:text-4xl">
            Autonomous Loop & Intercept Deep Dive
          </h1>
          <p className="text-base text-muted-foreground leading-relaxed sm:text-lg">
            Understand how ARTEX plans missions, executes command tools, intercepts network traffic, and suspends
            execution for human approval.
          </p>
          <div className="flex flex-wrap gap-2 pt-2">
            <Badge variant="outline" className="border-primary/40 font-mono text-primary text-xs">
              Planner/Worker Cycle
            </Badge>
            <Badge variant="outline" className="border-border/60 font-mono text-muted-foreground text-xs">
              Go Channels
            </Badge>
            <Badge variant="outline" className="border-border/60 font-mono text-muted-foreground text-xs">
              Scope Guard
            </Badge>
            <Badge variant="outline" className="border-border/60 font-mono text-muted-foreground text-xs">
              MITM :8788
            </Badge>
          </div>
        </div>

        {/* 1. Purpose & Prerequisites */}
        <section id="purpose-and-prerequisites" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Purpose & Prerequisites</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            Automated security agents can be risky if they run unconstrained. ARTEX pairs an autonomous reasoning loop
            with an ironclad <strong className="text-foreground">Intercept Guard</strong> so you can automate scanning
            without risking out-of-scope traffic or destructive payloads.
          </p>

          <DocCallout type="info" title="Prerequisites">
            <ul className="list-disc pl-4 space-y-1 text-xs">
              <li>Familiarity with basic console operations (Chat, Tasks, Scope, and Approvals).</li>
              <li>Active Go 1.26 backend and PostgreSQL state database.</li>
              <li>Written authorization for any tested target network.</li>
            </ul>
          </DocCallout>
        </section>

        {/* 2. Autonomous Loop Phases */}
        <section id="autonomous-loop-phases" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Cpu className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Autonomous Loop Phases</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            The core engine follows a strict 5-phase loop executed by the Planner and Worker agents:
          </p>

          <DocSteps>
            <DocStep
              number={0}
              title="Phase 0: Scope & Destination Guard"
              description="Pre-Flight destination boundary verification"
            >
              Before any task is queued, target hostnames and IPs are checked against the Asset Scope table. If a target
              is not listed or explicitly banned, the tool request is rejected immediately with an in-scope error.
            </DocStep>

            <DocStep
              number={1}
              title="Phase 1: Planning & Task Decomposition"
              description="Planner agent reads objectives and schedules work"
            >
              The Planner agent reads your prompt, inspects known assets from PostgreSQL, and breaks the goal down into
              independent Task records. Tasks are inserted into the queue with parameters and assigned tool packs.
            </DocStep>

            <DocStep
              number={2}
              title="Phase 2: Tool Execution & Intercept Check"
              description="Worker goroutine evaluates safety rules"
            >
              A Worker goroutine claims the task. It passes the tool name and command parameters through the Intercept
              Rule Evaluator. If action is <em>Allow</em>, the tool starts. If action is <em>Require Approval</em>, the
              goroutine halts.
            </DocStep>

            <DocStep
              number={3}
              title="Phase 3: Traffic Interception (:8788)"
              description="MITM recording proxy captures complete request/response streams"
            >
              Tools run with HTTP_PROXY environment variables pointing to{" "}
              <code className="font-mono text-foreground">127.0.0.1:8788</code>. Every request and response pair is
              streamed to disk and indexed in SQLite for full-text search.
            </DocStep>

            <DocStep
              number={4}
              title="Phase 4: Output Parsing & Finding Deduplication"
              description="Vulnerability extraction and PostgreSQL evidence storage"
            >
              The agent parses tool output, matches vulnerability signatures, and writes a Finding record into
              PostgreSQL. Deduplication rules ensure the same vulnerability on the same URL is updated rather than
              duplicated.
            </DocStep>
          </DocSteps>
        </section>

        {/* 3. Intercept Engine & Human Gate */}
        <section id="intercept-engine" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <PauseCircle className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Intercept Engine & Human Gate</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            The heart of ARTEX safety is the Human-in-the-Loop Intercept Engine. When a sensitive action is requested,
            the system pauses execution cleanly:
          </p>

          <div className="rounded-xl border border-border/80 bg-zinc-950 p-5 font-mono text-xs text-zinc-300 space-y-3 shadow-sm">
            <div className="rounded-lg border border-zinc-800 bg-zinc-900/80 p-3 space-y-1">
              <div className="font-semibold text-primary">[1] Tool Request Triggered</div>
              <p className="text-zinc-400 text-[11px]">
                Worker prepares to execute:{" "}
                <code className="text-zinc-200">sqlmap -u &quot;https://test.com/id=1&quot; --batch</code>
              </p>
            </div>

            <div className="flex justify-center text-zinc-500 text-xs">↓ Intercept Rule Evaluation</div>

            <div className="rounded-lg border border-amber-500/40 bg-amber-500/10 p-3 text-amber-200 space-y-1">
              <div className="font-semibold text-amber-400">
                [2] Rule Hit: &quot;sqlmap requires human approval&quot;
              </div>
              <p className="text-amber-200/90 text-[11px]">
                Goroutine pauses on Go wait channel:{" "}
                <code className="font-mono text-amber-300">{"decision := <-waitChan"}</code>. Status set to
                &quot;pending&quot;.
              </p>
            </div>

            <div className="flex justify-center text-zinc-500 text-xs">↓ WebSocket Notification</div>

            <div className="rounded-lg border border-zinc-800 bg-zinc-900/80 p-3 space-y-1">
              <div className="font-semibold text-zinc-200">[3] Operator Web Interface</div>
              <p className="text-zinc-400 text-[11px]">
                Amber badge lights up on <strong>Approvals</strong> in sidebar. Operator reviews command parameters and
                clicks <strong>Approve</strong>.
              </p>
            </div>

            <div className="flex justify-center text-zinc-500 text-xs">↓ Channel Resume</div>

            <div className="rounded-lg border border-primary/40 bg-primary/10 p-3 space-y-1">
              <div className="font-semibold text-primary">[4] Goroutine Resumes</div>
              <p className="text-zinc-300 text-[11px]">
                API handler sends <code className="font-mono text-primary">{"waitChan <- ActionApprove"}</code>. Tool
                begins execution safely.
              </p>
            </div>
          </div>
        </section>

        {/* 4. Rule Evaluation Precedence */}
        <section id="rule-precedence" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <ShieldAlert className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Rule Evaluation Precedence</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            Every candidate command is evaluated through a strict cascade:
          </p>

          <div className="space-y-3 text-xs leading-relaxed">
            <div className="flex items-start gap-3 rounded-xl border border-destructive/30 bg-destructive/5 p-4">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-destructive/20 font-bold font-mono text-destructive">
                1
              </span>
              <div>
                <strong className="text-foreground font-semibold">Scope Banned / Out-of-Scope: </strong>
                If the destination IP or domain is not explicitly allowed in <em>System &rarr; Scope</em>, the command
                is immediately aborted. No exceptions.
              </div>
            </div>

            <div className="flex items-start gap-3 rounded-xl border border-destructive/20 bg-card/60 p-4">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-destructive/15 font-bold font-mono text-destructive">
                2
              </span>
              <div>
                <strong className="text-foreground font-semibold">Intercept Block Rules: </strong>
                If any active rule matching tool or parameters has action <em>Block</em>, the command is terminated with
                an error.
              </div>
            </div>

            <div className="flex items-start gap-3 rounded-xl border border-amber-500/30 bg-amber-500/5 p-4">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-amber-500/20 font-bold font-mono text-amber-500">
                3
              </span>
              <div>
                <strong className="text-foreground font-semibold">Require Approval Rules: </strong>
                If any matching rule has action <em>Require Approval</em>, execution halts until an admin clicks Approve
                in the web UI.
              </div>
            </div>

            <div className="flex items-start gap-3 rounded-xl border border-primary/20 bg-card/60 p-4">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 font-bold font-mono text-primary">
                4
              </span>
              <div>
                <strong className="text-foreground font-semibold">Allow (Default): </strong>
                If scope is verified and no block or approval rules trigger, the tool process starts immediately.
              </div>
            </div>
          </div>
        </section>

        {/* 5. Step-by-Step Approval Rule Setup */}
        <section id="step-by-step-rules" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Lock className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">
              Step-by-Step: Adding an Approval Rule
            </h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            How to configure a mandatory approval gate for aggressive port scanning tools:
          </p>

          <DocSteps>
            <DocStep number={1} title="Open Rules in System">
              In the left sidebar, click{" "}
              <Link href="/system/intercept" className="text-primary underline font-medium">
                Rules (/system/intercept)
              </Link>{" "}
              under the <strong>SYSTEM</strong> section.
            </DocStep>
            <DocStep number={2} title="Create Intercept Rule">
              Click the <strong>Create Intercept Rule</strong> button to open the configuration dialog.
            </DocStep>
            <DocStep number={3} title="Set Rule Properties">
              Enter Rule Name: <code className="font-mono text-foreground">Port Scan Gate</code> and Tool Name:{" "}
              <code className="font-mono text-foreground">nmap</code>.
            </DocStep>
            <DocStep number={4} title="Select Action & Save">
              In the Action dropdown, choose <strong>Require Approval</strong>, then click <strong>Save Rule</strong>.
            </DocStep>
          </DocSteps>
        </section>

        {/* 6. Real-World Walkthrough */}
        <section id="real-world-example" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Terminal className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">
              Real-World Walkthrough: Gate in Action
            </h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            What happens when an agent encounters an approval gate during an assessment:
          </p>

          <DocCodeBlock
            language="bash"
            title="Live execution trace"
            code={`1. User prompt: "Test https://app.example.com for input validation flaws"
2. Discovery: Agent browses endpoint and discovers SQL syntax error in search parameter
3. Tool Trigger: Agent prepares command: sqlmap -u "https://app.example.com/search?q=test" --batch
4. Gate Halts: Intercept rule matches "sqlmap" -> status set to "Awaiting Approval"
5. Human Review: Operator inspects parameters on /system/intercept/approvals and clicks "Approve"
6. Execution: Tool runs through proxy :8788 and records confirmed flaw in Findings`}
          />
        </section>

        {/* 7. Go Backend Implementation Details */}
        <section id="backend-mechanisms" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Database className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Go Backend Implementation Details</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            Key files in the Go codebase that power the autonomous loop and safety gates:
          </p>

          <div className="grid gap-3 text-muted-foreground text-xs leading-relaxed sm:grid-cols-2">
            <div className="space-y-1.5 rounded-xl border border-border/60 bg-card/40 p-4">
              <span className="font-mono font-semibold text-primary text-sm">agent/orchestrator.go</span>
              <p>
                Manages the planner and worker loops. Breaks goals into tasks and streams model thought deltas to
                WebSocket clients.
              </p>
            </div>
            <div className="space-y-1.5 rounded-xl border border-border/60 bg-card/40 p-4">
              <span className="font-mono font-semibold text-primary text-sm">guard/intercept.go</span>
              <p>
                Evaluates candidate tool calls against scope assets and intercept rules. Manages the Go wait channel map
                for pending approvals.
              </p>
            </div>
            <div className="space-y-1.5 rounded-xl border border-border/60 bg-card/40 p-4">
              <span className="font-mono font-semibold text-primary text-sm">traffic/proxy.go</span>
              <p>
                HTTP/HTTPS MITM proxy listening on :8788. Dumps raw request/response streams and indexes metadata into
                SQLite.
              </p>
            </div>
            <div className="space-y-1.5 rounded-xl border border-border/60 bg-card/40 p-4">
              <span className="font-mono font-semibold text-primary text-sm">db/schema.sql</span>
              <p>
                Defines PostgreSQL tables for tasks, findings, assets, intercept rules, and settings with idempotent
                migrations.
              </p>
            </div>
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
