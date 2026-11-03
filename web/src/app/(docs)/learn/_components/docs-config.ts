import { Activity, BookOpen, type LucideIcon, Rocket, Settings2, Workflow } from "lucide-react";

export interface DocSection {
  id: string;
  title: string;
  shortTitle: string;
  description: string;
  href: string;
  icon: LucideIcon;
  badge?: string;
}

export const DOC_SECTIONS: DocSection[] = [
  {
    id: "overview",
    title: "Overview & Architecture",
    shortTitle: "Overview",
    description: "System design, core components, and how ARTEX works end-to-end.",
    href: "/learn",
    icon: BookOpen,
  },
  {
    id: "quickstart",
    title: "6-Step First Run Guide",
    shortTitle: "Quickstart",
    description: "From fresh installation to your first automated security scan in 5 minutes.",
    href: "/learn/quickstart",
    icon: Rocket,
    badge: "5 min",
  },
  {
    id: "operations",
    title: "Operations Reference",
    shortTitle: "Operations",
    description: "Guide to Dashboard, Chat, Tasks, Findings, Traffic, Exec, LLM Streams, Assets, Sync, and Workspace.",
    href: "/learn/operations",
    icon: Activity,
    badge: "10",
  },
  {
    id: "system",
    title: "System Reference",
    shortTitle: "System",
    description: "Guide to LLM, Agent, MCP, Skill, Tools, Alerts, Rules, Scope, Approvals, Logs, and Settings.",
    href: "/learn/system",
    icon: Settings2,
    badge: "11",
  },
  {
    id: "workflow",
    title: "Autonomous Loop & Intercept Deep Dive",
    shortTitle: "Workflow & Intercept",
    description: "Under the hood: agent reasoning loops, tool execution, proxy capture, and human approval gates.",
    href: "/learn/workflow",
    icon: Workflow,
    badge: "Loop",
  },
];

export function getDocsNavigation(pathname: string): {
  current: DocSection;
  prev: DocSection | null;
  next: DocSection | null;
} {
  const normalized = pathname.replace(/\/+$/, "") || "/learn";
  const currentIndex = DOC_SECTIONS.findIndex((section) => {
    const secNorm = section.href.replace(/\/+$/, "") || "/learn";
    return secNorm === normalized;
  });

  const index = currentIndex >= 0 ? currentIndex : 0;
  return {
    current: DOC_SECTIONS[index],
    prev: index > 0 ? DOC_SECTIONS[index - 1] : null,
    next: index < DOC_SECTIONS.length - 1 ? DOC_SECTIONS[index + 1] : null,
  };
}
