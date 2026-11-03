"use client";

import Link from "next/link";

import {
  ArrowRight,
  Bot,
  Brain,
  CheckCircle2,
  ClipboardList,
  Lock,
  Network,
  Rocket,
  ShieldCheck,
  Target,
  Terminal,
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
  { id: "step-1-password", title: "Step 1: Set Admin Password" },
  { id: "step-2-llm", title: "Step 2: Add AI Provider Key" },
  { id: "step-3-agent", title: "Step 3: Create an Agent" },
  { id: "step-4-scope", title: "Step 4: Define Target Scope" },
  { id: "step-5-chat", title: "Step 5: Launch in Chat" },
  { id: "step-6-monitor", title: "Step 6: Monitor & Approve" },
  { id: "real-world-example", title: "Real-World Example" },
  { id: "internal-mechanisms", title: "Internal Mechanisms" },
];

export default function QuickstartPage() {
  return (
    <div className="flex items-start justify-center gap-8 px-4 py-8 sm:px-6 lg:px-8 xl:gap-12">
      {/* Center Reading Container */}
      <div className="w-full min-w-0 max-w-3xl space-y-12">
        {/* Title Header */}
        <div className="space-y-4 border-border/60 border-b pb-8">
          <div className="flex items-center gap-2 font-mono text-primary text-xs uppercase tracking-wider">
            <Rocket className="h-4 w-4" />
            <span>Quickstart Guide</span>
          </div>
          <h1 className="font-bold text-3xl text-foreground tracking-tight sm:text-4xl">6-Step First Run Guide</h1>
          <p className="text-base text-muted-foreground leading-relaxed sm:text-lg">
            Get your ARTEX instance up and running in under 5 minutes. Follow these six concrete steps to configure
            credentials, define your target boundary, and run your first automated penetration test.
          </p>
          <div className="flex items-center gap-2 pt-2">
            <Badge variant="secondary" className="font-mono text-xs">
              Time to complete: 5 minutes
            </Badge>
            <Badge variant="outline" className="border-primary/40 font-mono text-xs text-primary">
              Beginner Friendly
            </Badge>
          </div>
        </div>

        {/* Purpose & Prerequisites */}
        <section id="purpose-and-prerequisites" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Purpose & Prerequisites</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            This guide takes you through initial setup from first login to watching an agent discover endpoints and
            request approval.
          </p>

          <DocCallout type="info" title="Before you begin">
            Ensure the ARTEX backend daemon is running (<code className="text-foreground">./dev.sh</code> or{" "}
            <code className="text-foreground">./artex</code>) and connected to your PostgreSQL database.
          </DocCallout>
        </section>

        {/* 6 Steps Timeline */}
        <section className="space-y-6">
          <DocSteps>
            {/* Step 1 */}
            <DocStep
              number={1}
              title="Set Your Admin Password"
              description="Secure the console before performing scans"
            >
              <div id="step-1-password" className="scroll-mt-24 space-y-3">
                <p>
                  In the left sidebar under <strong>SYSTEM</strong>, click{" "}
                  <Link href="/system/settings" className="text-primary underline">
                    Settings (/system/settings)
                  </Link>
                  . Enter your strong password and click <strong>Save Settings</strong>.
                </p>
                <DocCodeBlock
                  language="bash"
                  title="Alternative: Reset admin password via terminal script"
                  code={`./reset-password.sh -p 'YourSecurePassword123!'`}
                />
              </div>
            </DocStep>

            {/* Step 2 */}
            <DocStep
              number={2}
              title="Add Your AI Provider Key"
              description="Connect to OpenAI, Anthropic, or an OpenAI-compatible endpoint"
            >
              <div id="step-2-llm" className="scroll-mt-24 space-y-3">
                <p>
                  Click{" "}
                  <Link href="/system/llm" className="text-primary underline">
                    LLM (/system/llm)
                  </Link>
                  . Click <strong>Add Provider</strong>, paste your API key, set the model name (e.g.{" "}
                  <code className="text-foreground">gpt-4o</code>), and click <strong>Test Connection</strong>.
                </p>
                <DocCallout type="tip" title="Supported Models">
                  ARTEX supports GPT-4o, Claude 3.5 Sonnet, DeepSeek V3, and local models served via Ollama or vLLM.
                </DocCallout>
              </div>
            </DocStep>

            {/* Step 3 */}
            <DocStep
              number={3}
              title="Create Your First Agent"
              description="Define persona, model profile, and authorized tool skills"
            >
              <div id="step-3-agent" className="scroll-mt-24 space-y-3">
                <p>
                  Go to{" "}
                  <Link href="/system/agents" className="text-primary underline">
                    Agent (/system/agents)
                  </Link>{" "}
                  and click <strong>Create Agent</strong>. Select your model profile and ensure skill packs{" "}
                  <em>api-recon</em> and <em>scopesentry</em> are enabled.
                </p>
              </div>
            </DocStep>

            {/* Step 4 */}
            <DocStep number={4} title="Define Target Scope" description="Strictly authorize domain and IP boundaries">
              <div id="step-4-scope" className="scroll-mt-24 space-y-3">
                <p>
                  Click{" "}
                  <Link href="/system/intercept/assets" className="text-primary underline">
                    Scope (/system/intercept/assets)
                  </Link>
                  . Click <strong>Add Scope Rule</strong>, enter your target domain (such as{" "}
                  <code className="text-foreground font-mono">testphp.vulnweb.com</code>), and set Action to{" "}
                  <strong>Allow</strong>.
                </p>
                <DocCallout type="warning" title="Strict Scope Guard">
                  Any tool request targeting unlisted hosts will be immediately dropped by the kernel guard.
                </DocCallout>
              </div>
            </DocStep>

            {/* Step 5 */}
            <DocStep
              number={5}
              title="Launch Scan in Chat"
              description="Instruct your agent using plain English security goals"
            >
              <div id="step-5-chat" className="scroll-mt-24 space-y-3">
                <p>
                  Open{" "}
                  <Link href="/chat" className="text-primary underline">
                    Chat (/chat)
                  </Link>
                  , select your agent from the top dropdown, and submit your prompt:
                </p>
                <DocCodeBlock
                  language="markdown"
                  title="Example Chat Prompt"
                  code={`Enumerate web endpoints on testphp.vulnweb.com and check for open directory listings and forms.`}
                />
              </div>
            </DocStep>

            {/* Step 6 */}
            <DocStep
              number={6}
              title="Monitor Tasks, Approvals & Findings"
              description="Track real-time progress and make human-in-the-loop decisions"
            >
              <div id="step-6-monitor" className="scroll-mt-24 space-y-3">
                <p>
                  Check{" "}
                  <Link href="/function/tasks" className="text-primary underline">
                    Tasks (/function/tasks)
                  </Link>{" "}
                  for live tool execution. When an amber badge appears on{" "}
                  <Link href="/system/intercept/approvals" className="text-primary underline">
                    Approvals
                  </Link>
                  , click <strong>Approve</strong>.
                </p>
                <p>
                  View confirmed vulnerabilities in{" "}
                  <Link href="/function/findings" className="text-primary underline">
                    Findings (/function/findings)
                  </Link>{" "}
                  and raw HTTP streams in{" "}
                  <Link href="/function/traffic" className="text-primary underline">
                    Traffic (/function/traffic)
                  </Link>
                  .
                </p>
              </div>
            </DocStep>
          </DocSteps>
        </section>

        {/* Real-World Example */}
        <section id="real-world-example" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <Target className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">
              Real-World Example: First Scan Flow
            </h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            Execution walkthrough on authorized test target:
          </p>

          <DocCodeBlock
            language="bash"
            title="Scan sequence"
            code={`1. User prompts: "Discover subdomains and check robots.txt on testphp.vulnweb.com"
2. Planner creates Task #1: HTTP GET https://testphp.vulnweb.com/robots.txt
3. Guard checks host testphp.vulnweb.com -> ALLOWED
4. Proxy records 200 OK response into SQLite index
5. Discovered endpoints (/admin/, /login.php) added to Assets inventory`}
          />
        </section>

        {/* Internal Mechanisms */}
        <section id="internal-mechanisms" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-primary" />
            <h2 className="font-semibold text-2xl text-foreground tracking-tight">Internal Mechanisms</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed">
            What happens behind the scenes during the 6 steps:
          </p>

          <div className="grid gap-3 sm:grid-cols-2 text-xs text-muted-foreground leading-relaxed">
            <div className="rounded-xl border border-border/60 bg-card p-4 space-y-1">
              <span className="font-semibold text-foreground">PostgreSQL Settings Table</span>
              <p>Passwords and API keys are stored encrypted. Reset is always available via CLI script.</p>
            </div>
            <div className="rounded-xl border border-border/60 bg-card p-4 space-y-1">
              <span className="font-semibold text-foreground">WebSocket Streaming</span>
              <p>Chat requests spawn a goroutine streaming token deltas in real-time over WebSocket.</p>
            </div>
            <div className="rounded-xl border border-border/60 bg-card p-4 space-y-1">
              <span className="font-semibold text-foreground">Process Sandbox</span>
              <p>Tools run with piped stdout/stderr buffers and strict execution timeouts.</p>
            </div>
            <div className="rounded-xl border border-border/60 bg-card p-4 space-y-1">
              <span className="font-semibold text-foreground">Idempotent Asset Upserts</span>
              <p>Discovered hosts and ports are inserted into PostgreSQL with ON CONFLICT DO UPDATE.</p>
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
