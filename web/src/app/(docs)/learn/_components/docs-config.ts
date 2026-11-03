import {
  Activity,
  Ban,
  BellRing,
  BookOpen,
  Bot,
  Brain,
  Bug,
  ClipboardList,
  FolderOpen,
  FolderSync,
  LayoutDashboard,
  type LucideIcon,
  MessageSquare,
  Network,
  Plug,
  Radio,
  Rocket,
  ScrollText,
  Settings2,
  ShieldAlert,
  Sparkles,
  Target,
  Terminal,
  Workflow,
  Wrench,
} from "lucide-react";

export interface DocSection {
  id: string;
  title: string;
  shortTitle: string;
  description: string;
  href: string;
  icon: LucideIcon;
  badge?: string;
  consoleRoute?: string;
}

export interface DocGroup {
  id: string;
  label: string;
  items: DocSection[];
}

export const DOC_GROUPS: DocGroup[] = [
  {
    id: "getting-started",
    label: "GETTING STARTED",
    items: [
      {
        id: "overview",
        title: "Overview",
        shortTitle: "Overview",
        description: "System design, core components, and how ARTEX works end-to-end.",
        href: "/learn",
        icon: BookOpen,
      },
      {
        id: "quickstart",
        title: "Quickstart",
        shortTitle: "Quickstart",
        description: "From fresh installation to your first automated security scan in 5 minutes.",
        href: "/learn/quickstart",
        icon: Rocket,
      },
      {
        id: "workflow",
        title: "Workflow",
        shortTitle: "Workflow",
        description: "Agent reasoning loops, tool execution, proxy capture, and approval gates.",
        href: "/learn/workflow",
        icon: Workflow,
      },
    ],
  },
  {
    id: "operations",
    label: "OPERATIONS",
    items: [
      {
        id: "op-dashboard",
        title: "Dashboard",
        shortTitle: "Dashboard",
        description: "Engagement overview, active tasks, finding distributions, and traffic volume.",
        href: "/learn/operations/dashboard",
        icon: LayoutDashboard,
        consoleRoute: "/dashboard",
      },
      {
        id: "op-chat",
        title: "Chat",
        shortTitle: "Chat",
        description: "Interactive AI agent mission console, streaming thoughts, and tool feedback.",
        href: "/learn/operations/chat",
        icon: MessageSquare,
        consoleRoute: "/chat",
      },
      {
        id: "op-tasks",
        title: "Tasks",
        shortTitle: "Tasks",
        description: "Decomposed task execution queue, status inspector, and failure restarts.",
        href: "/learn/operations/tasks",
        icon: Target,
        consoleRoute: "/function/tasks",
      },
      {
        id: "op-findings",
        title: "Findings",
        shortTitle: "Findings",
        description: "Confirmed vulnerabilities, CVSS scoring, PoC reproduction steps, and exports.",
        href: "/learn/operations/findings",
        icon: Bug,
        consoleRoute: "/function/findings",
      },
      {
        id: "op-traffic",
        title: "Traffic",
        shortTitle: "Traffic",
        description: "Built-in MITM recording proxy on port 8788 and SQLite request archive.",
        href: "/learn/operations/traffic",
        icon: Activity,
        consoleRoute: "/function/traffic",
      },
      {
        id: "op-exec",
        title: "Exec",
        shortTitle: "Exec",
        description: "Direct shell execution console for human command verification.",
        href: "/learn/operations/exec",
        icon: Terminal,
        consoleRoute: "/function/commands",
      },
      {
        id: "op-llm-streams",
        title: "LLM Streams",
        shortTitle: "LLM Streams",
        description: "Real-time token audit, prompt inspect, and completion latency inspector.",
        href: "/learn/operations/llm-streams",
        icon: Radio,
        consoleRoute: "/function/llm-records",
      },
      {
        id: "op-assets",
        title: "Assets",
        shortTitle: "Assets",
        description: "Discovered hostnames, open ports, web tech stacks, and alive status tracking.",
        href: "/learn/operations/assets",
        icon: Network,
        consoleRoute: "/function/assets",
      },
      {
        id: "op-sync",
        title: "Sync",
        shortTitle: "Sync",
        description: "Bidirectional sync with external tools, wordlists, and scanner outputs.",
        href: "/learn/operations/sync",
        icon: FolderSync,
        consoleRoute: "/function/sync",
      },
      {
        id: "op-workspace",
        title: "Workspace",
        shortTitle: "Workspace",
        description: "Local file system workspace, tool artifacts, screenshots, and logs.",
        href: "/learn/operations/workspace",
        icon: FolderOpen,
        consoleRoute: "/function/workspace",
      },
    ],
  },
  {
    id: "system",
    label: "SYSTEM",
    items: [
      {
        id: "sys-llm",
        title: "LLM",
        shortTitle: "LLM",
        description: "Configure AI providers, API keys, endpoints, and token generation parameters.",
        href: "/learn/system/llm",
        icon: Brain,
        consoleRoute: "/system/llm",
      },
      {
        id: "sys-agent",
        title: "Agent",
        shortTitle: "Agent",
        description: "Manage autonomous agent personas, system prompts, and authorized skills.",
        href: "/learn/system/agent",
        icon: Bot,
        consoleRoute: "/system/agents",
      },
      {
        id: "sys-mcp",
        title: "MCP",
        shortTitle: "MCP",
        description: "Connect external Model Context Protocol stdio or SSE servers.",
        href: "/learn/system/mcp",
        icon: Plug,
        consoleRoute: "/system/mcp",
      },
      {
        id: "sys-skill",
        title: "Skill",
        shortTitle: "Skill",
        description: "Inspect specialized penetration testing skill packs and SKILL.md specs.",
        href: "/learn/system/skill",
        icon: Sparkles,
        consoleRoute: "/system/skills",
      },
      {
        id: "sys-tools",
        title: "Tools",
        shortTitle: "Tools",
        description: "Atomic command-line tool registry, binary health checks, and JSON schemas.",
        href: "/learn/system/tools",
        icon: Wrench,
        consoleRoute: "/system/tools",
      },
      {
        id: "sys-alerts",
        title: "Alerts",
        shortTitle: "Alerts",
        description: "Outbound notification channels (webhooks, email) for critical events.",
        href: "/learn/system/alerts",
        icon: BellRing,
        consoleRoute: "/system/notify",
      },
      {
        id: "sys-rules",
        title: "Rules",
        shortTitle: "Rules",
        description: "Tool execution intercept rules: allow, block, or require human approval.",
        href: "/learn/system/rules",
        icon: ShieldAlert,
        consoleRoute: "/system/intercept",
      },
      {
        id: "sys-scope",
        title: "Scope",
        shortTitle: "Scope",
        description: "Target boundary enforcement: permitted domains, IPs, and exclusion lists.",
        href: "/learn/system/scope",
        icon: Ban,
        consoleRoute: "/system/intercept/assets",
      },
      {
        id: "sys-approvals",
        title: "Approvals",
        shortTitle: "Approvals",
        description: "Human-in-the-loop decision queue for paused high-risk tool operations.",
        href: "/learn/system/approvals",
        icon: ClipboardList,
        consoleRoute: "/system/intercept/approvals",
      },
      {
        id: "sys-logs",
        title: "Logs",
        shortTitle: "Logs",
        description: "Live daemon diagnostics, database traces, proxy logs, and export bundle.",
        href: "/learn/system/logs",
        icon: ScrollText,
        consoleRoute: "/system/logs",
      },
      {
        id: "sys-settings",
        title: "Settings",
        shortTitle: "Settings",
        description: "Admin credentials, system identity, database options, and UI preferences.",
        href: "/learn/system/settings",
        icon: Settings2,
        consoleRoute: "/system/settings",
      },
    ],
  },
];

export const DOC_SECTIONS: DocSection[] = DOC_GROUPS.flatMap((g) => g.items);

const OPERATIONS_HUB_SECTION: DocSection = {
  id: "operations-hub",
  title: "Operations Reference",
  shortTitle: "Operations",
  description: "Directory of all 10 operational modules mirroring the console sidebar.",
  href: "/learn/operations",
  icon: Activity,
};

const SYSTEM_HUB_SECTION: DocSection = {
  id: "system-hub",
  title: "System Reference",
  shortTitle: "System",
  description: "Directory of all 11 system governance and configuration modules.",
  href: "/learn/system",
  icon: Settings2,
};

export function getDocsNavigation(pathname: string): {
  current: DocSection;
  prev: DocSection | null;
  next: DocSection | null;
} {
  const normalized = pathname.replace(/\/+$/, "") || "/learn";

  if (normalized === "/learn/operations") {
    return {
      current: OPERATIONS_HUB_SECTION,
      prev: DOC_SECTIONS.find((s) => s.href === "/learn/workflow") || null,
      next: DOC_SECTIONS.find((s) => s.href === "/learn/operations/dashboard") || null,
    };
  }

  if (normalized === "/learn/system") {
    return {
      current: SYSTEM_HUB_SECTION,
      prev: DOC_SECTIONS.find((s) => s.href === "/learn/operations/workspace") || null,
      next: DOC_SECTIONS.find((s) => s.href === "/learn/system/llm") || null,
    };
  }

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
