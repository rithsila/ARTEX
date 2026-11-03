"use client";

import type { ReactNode } from "react";

import Link from "next/link";

import {
  AlertTriangle,
  ArrowUpRight,
  CheckCircle2,
  type LucideIcon,
  PlayCircle,
  Settings,
  ShieldCheck,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

import { DocCallout } from "./doc-callout";
import { DocsNavCards } from "./docs-nav-cards";
import { DocsToc, type TocItem } from "./docs-toc";

export interface UsageStep {
  step: number;
  title: string;
  detail: string;
  tip?: string;
}

export interface ConfigField {
  field: string;
  type?: string;
  description: string;
  defaultVal?: string;
}

export interface GotchaItem {
  title: string;
  description: string;
}

export interface DocModuleTemplateProps {
  title: string;
  description: string;
  groupLabel: "OPERATIONS" | "SYSTEM" | "GETTING STARTED";
  icon: LucideIcon;
  consoleRoute?: string;
  badge?: string;
  purpose: string | ReactNode;
  prerequisites: string | string[];
  usageSteps: UsageStep[];
  configFields: ConfigField[];
  gotchas: GotchaItem[];
}

const MODULE_TOC: TocItem[] = [
  { id: "purpose", title: "Purpose" },
  { id: "prerequisites", title: "Prerequisites" },
  { id: "usage", title: "Step-by-Step Usage" },
  { id: "configuration", title: "Configuration Fields" },
  { id: "gotchas", title: "Common Gotchas" },
];

export function DocModuleTemplate({
  title,
  description,
  groupLabel,
  icon: Icon,
  consoleRoute,
  badge,
  purpose,
  prerequisites,
  usageSteps,
  configFields,
  gotchas,
}: DocModuleTemplateProps) {
  const prereqList = Array.isArray(prerequisites) ? prerequisites : [prerequisites];

  return (
    <div className="flex items-start justify-center gap-8 px-4 py-8 sm:px-6 lg:px-8 xl:gap-12">
      {/* Center Reading Column */}
      <div className="w-full min-w-0 max-w-3xl space-y-10">
        {/* Header Section */}
        <div className="space-y-3.5 border-border/60 border-b pb-6">
          <div className="flex items-center gap-2 font-mono text-primary text-xs uppercase tracking-wider">
            <Icon className="h-4 w-4" />
            <span>{groupLabel}</span>
          </div>

          <div className="flex flex-wrap items-center justify-between gap-3">
            <h1 className="font-bold text-2xl text-foreground tracking-tight sm:text-3xl">{title}</h1>
            {consoleRoute && (
              <Link
                href={consoleRoute}
                className="inline-flex items-center gap-1.5 rounded-md border border-primary/30 bg-primary/10 px-2.5 py-1 font-mono text-primary text-xs transition-colors hover:bg-primary/20"
              >
                <span>Console: {consoleRoute}</span>
                <ArrowUpRight className="h-3.5 w-3.5" />
              </Link>
            )}
          </div>

          <p className="text-muted-foreground text-sm leading-relaxed sm:text-base">{description}</p>

          <div className="flex flex-wrap items-center gap-2 pt-1">
            {badge && (
              <Badge variant="outline" className="border-primary/40 font-mono text-[11px] text-primary">
                {badge}
              </Badge>
            )}
            <Badge variant="outline" className="border-border/60 font-mono text-[11px] text-muted-foreground">
              Module Guide
            </Badge>
          </div>
        </div>

        {/* 1. Purpose */}
        <section id="purpose" className="scroll-mt-20 space-y-3">
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-primary" />
            <h2 className="font-semibold text-foreground text-lg tracking-tight">1. Purpose</h2>
          </div>
          <div className="text-muted-foreground text-xs leading-relaxed sm:text-sm">
            {typeof purpose === "string" ? <p>{purpose}</p> : purpose}
          </div>
        </section>

        {/* 2. Prerequisites */}
        <section id="prerequisites" className="scroll-mt-20 space-y-3">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="h-4 w-4 text-primary" />
            <h2 className="font-semibold text-foreground text-lg tracking-tight">2. Prerequisites</h2>
          </div>
          <div className="space-y-2">
            {prereqList.map((prereq) => (
              <div
                key={prereq}
                className="flex items-start gap-2.5 rounded-lg border border-border/60 bg-card/40 p-3 text-foreground/90 text-xs leading-relaxed"
              >
                <div className="mt-0.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" />
                <span>{prereq}</span>
              </div>
            ))}
          </div>
        </section>

        {/* 3. Step-by-Step Usage */}
        <section id="usage" className="scroll-mt-20 space-y-4">
          <div className="flex items-center gap-2">
            <PlayCircle className="h-4 w-4 text-primary" />
            <h2 className="font-semibold text-foreground text-lg tracking-tight">3. Step-by-Step Usage</h2>
          </div>
          <div className="space-y-3">
            {usageSteps.map((step) => (
              <div key={step.step} className="flex gap-3 rounded-lg border border-border/60 bg-card/30 p-3.5">
                <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/15 font-mono font-semibold text-primary text-xs">
                  {step.step}
                </div>
                <div className="space-y-1 text-xs">
                  <div className="font-medium text-foreground">{step.title}</div>
                  <div className="text-muted-foreground leading-relaxed">{step.detail}</div>
                  {step.tip && <div className="pt-1 font-mono text-[11px] text-primary/90">Tip: {step.tip}</div>}
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* 4. Configuration Fields */}
        <section id="configuration" className="scroll-mt-20 space-y-3">
          <div className="flex items-center gap-2">
            <Settings className="h-4 w-4 text-primary" />
            <h2 className="font-semibold text-foreground text-lg tracking-tight">4. Configuration Fields</h2>
          </div>
          <div className="grid gap-2.5 sm:grid-cols-2">
            {configFields.map((field) => (
              <Card key={field.field} className="border-border/60 bg-card/40">
                <CardContent className="space-y-1.5 p-3.5">
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-mono font-semibold text-foreground text-xs">{field.field}</span>
                    {field.type && (
                      <span className="rounded bg-muted/60 px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
                        {field.type}
                      </span>
                    )}
                  </div>
                  <p className="text-muted-foreground text-xs leading-relaxed">{field.description}</p>
                  {field.defaultVal && (
                    <div className="font-mono text-[10px] text-primary/80">Default: {field.defaultVal}</div>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        {/* 5. Common Gotchas */}
        <section id="gotchas" className="scroll-mt-20 space-y-3">
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-4 w-4 text-amber-500" />
            <h2 className="font-semibold text-foreground text-lg tracking-tight">5. Common Gotchas</h2>
          </div>
          <div className="space-y-2.5">
            {gotchas.map((gotcha) => (
              <DocCallout key={gotcha.title} type="warning" title={gotcha.title} className="text-xs">
                {gotcha.description}
              </DocCallout>
            ))}
          </div>
        </section>

        {/* Bottom Prev/Next Navigation */}
        <DocsNavCards />
      </div>

      {/* Right Column Table of Contents (3-5 items with Back to top) */}
      <DocsToc items={MODULE_TOC} />
    </div>
  );
}
