"use client";

import * as React from "react";

import { useRouter } from "next/navigation";

import {
  Activity,
  ArrowRight,
  BookOpen,
  Bot,
  Brain,
  Bug,
  FolderOpen,
  FolderSync,
  LayoutDashboard,
  MessageSquare,
  Network,
  Radio,
  Rocket,
  ScrollText,
  Search,
  Settings2,
  ShieldAlert,
  Sparkles,
  Target,
  Terminal,
  Workflow,
  Wrench,
} from "lucide-react";

import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";

interface DocsSearchDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function DocsSearchDialog(props: DocsSearchDialogProps = {}) {
  const [internalOpen, setInternalOpen] = React.useState(false);
  const router = useRouter();

  const isControlled = props.open !== undefined;
  const open = isControlled ? (props.open as boolean) : internalOpen;

  const setOpen = React.useCallback(
    (val: boolean) => {
      if (isControlled && props.onOpenChange) {
        props.onOpenChange(val);
      } else {
        setInternalOpen(val);
      }
    },
    [isControlled, props.onOpenChange],
  );

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen(!open);
      }
    };
    const handleCustomOpen = () => setOpen(true);

    document.addEventListener("keydown", down);
    window.addEventListener("open-docs-search", handleCustomOpen);
    return () => {
      document.removeEventListener("keydown", down);
      window.removeEventListener("open-docs-search", handleCustomOpen);
    };
  }, [open, setOpen]);

  const runSelect = (href: string) => {
    setOpen(false);
    router.push(href);
  };

  return (
    <CommandDialog
      open={open}
      onOpenChange={setOpen}
      title="Search Documentation"
      description="Quickly find guides, operations, or system references"
    >
      <Command>
        <CommandInput placeholder="Type to search documentation..." />
        <CommandList>
          <CommandEmpty>No matching documentation found.</CommandEmpty>

          {/* Guides */}
          <CommandGroup heading="Guides & Architecture">
            <CommandItem onSelect={() => runSelect("/learn")} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <BookOpen className="h-4 w-4 text-primary" />
                <span>Overview & Architecture</span>
              </div>
              <ArrowRight className="h-3 w-3 text-muted-foreground" />
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/quickstart")} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Rocket className="h-4 w-4 text-primary" />
                <span>6-Step First Run Quickstart</span>
              </div>
              <ArrowRight className="h-3 w-3 text-muted-foreground" />
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/workflow")} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Workflow className="h-4 w-4 text-primary" />
                <span>Autonomous Loop & Intercept Deep Dive</span>
              </div>
              <ArrowRight className="h-3 w-3 text-muted-foreground" />
            </CommandItem>
          </CommandGroup>

          {/* Operations */}
          <CommandGroup heading="Operations Reference">
            <CommandItem onSelect={() => runSelect("/learn/operations#dashboard")} className="flex items-center gap-2">
              <LayoutDashboard className="h-4 w-4 text-primary" />
              <span>Dashboard — Overview & Metrics</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/operations#chat")} className="flex items-center gap-2">
              <MessageSquare className="h-4 w-4 text-primary" />
              <span>Chat — Interactive Mission Console</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/operations#tasks")} className="flex items-center gap-2">
              <Target className="h-4 w-4 text-primary" />
              <span>Tasks — Execution Queue & Inspection</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/operations#findings")} className="flex items-center gap-2">
              <Bug className="h-4 w-4 text-primary" />
              <span>Findings — Vulnerability Evidence Store</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/operations#traffic")} className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-primary" />
              <span>Traffic — MITM Recording Proxy (:8788)</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/operations#commands")} className="flex items-center gap-2">
              <Terminal className="h-4 w-4 text-primary" />
              <span>Exec — Direct Shell Execution</span>
            </CommandItem>
            <CommandItem
              onSelect={() => runSelect("/learn/operations#llm-records")}
              className="flex items-center gap-2"
            >
              <Radio className="h-4 w-4 text-primary" />
              <span>LLM Streams — Token & Prompt Audit</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/operations#assets")} className="flex items-center gap-2">
              <Network className="h-4 w-4 text-primary" />
              <span>Assets — Target Host & Port Inventory</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/operations#sync")} className="flex items-center gap-2">
              <FolderSync className="h-4 w-4 text-primary" />
              <span>Sync — External Scanner Importer</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/operations#workspace")} className="flex items-center gap-2">
              <FolderOpen className="h-4 w-4 text-primary" />
              <span>Workspace — Artifacts & Wordlists</span>
            </CommandItem>
          </CommandGroup>

          {/* System */}
          <CommandGroup heading="System Reference">
            <CommandItem onSelect={() => runSelect("/learn/system#llm")} className="flex items-center gap-2">
              <Brain className="h-4 w-4 text-primary" />
              <span>LLM — Provider Keys & Models</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/system#agents")} className="flex items-center gap-2">
              <Bot className="h-4 w-4 text-primary" />
              <span>Agent — Orchestration & Personas</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/system#mcp")} className="flex items-center gap-2">
              <Wrench className="h-4 w-4 text-primary" />
              <span>MCP — Model Context Protocol Bridge</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/system#skills")} className="flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-primary" />
              <span>Skill — Tool Packs & Workflows</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/system#intercept")} className="flex items-center gap-2">
              <ShieldAlert className="h-4 w-4 text-primary" />
              <span>Rules — Tool Intercept Safety Gates</span>
            </CommandItem>
            <CommandItem
              onSelect={() => runSelect("/learn/system#asset-intercept")}
              className="flex items-center gap-2"
            >
              <Network className="h-4 w-4 text-primary" />
              <span>Scope — Target Authorization Boundaries</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/system#approvals")} className="flex items-center gap-2">
              <ShieldAlert className="h-4 w-4 text-primary" />
              <span>Approvals — Human Decision Gate</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/system#logs")} className="flex items-center gap-2">
              <ScrollText className="h-4 w-4 text-primary" />
              <span>Logs — Live Daemon Diagnostics</span>
            </CommandItem>
            <CommandItem onSelect={() => runSelect("/learn/system#settings")} className="flex items-center gap-2">
              <Settings2 className="h-4 w-4 text-primary" />
              <span>Settings — Admin Password & Preferences</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </CommandDialog>
  );
}

export function DocsSearchTrigger({ onClick, className }: { onClick?: () => void; className?: string }) {
  const handleClick = () => {
    if (onClick) {
      onClick();
    } else {
      window.dispatchEvent(new CustomEvent("open-docs-search"));
    }
  };

  return (
    <button type="button" onClick={handleClick} className={className} aria-label="Search documentation">
      <div className="flex items-center gap-2">
        <Search className="h-3.5 w-3.5 text-muted-foreground" />
        <span className="text-muted-foreground text-xs">Search docs...</span>
      </div>
      <kbd className="pointer-events-none hidden h-5 select-none items-center gap-0.5 rounded border border-border/70 bg-muted/60 px-1.5 font-mono text-[10px] text-muted-foreground sm:inline-flex">
        <span className="text-xs">⌘</span>K
      </kbd>
    </button>
  );
}
