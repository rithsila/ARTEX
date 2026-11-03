"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { ArrowUpRight, MessageSquare, ShieldCheck, Terminal } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

import { DOC_GROUPS } from "./docs-config";
import { DocsSearchTrigger } from "./docs-search-dialog";

export function DocsSidebar() {
  const rawPath = usePathname();
  const currentPath = (rawPath ? decodeURIComponent(rawPath) : "").replace(/\/+$/, "") || "/learn";

  return (
    <aside className="sticky top-0 hidden h-screen w-72 shrink-0 flex-col border-border/60 border-r bg-card/30 backdrop-blur-md md:flex">
      {/* Sidebar Header */}
      <div className="flex h-14 items-center justify-between border-border/60 border-b px-4">
        <Link href="/learn" className="flex items-center gap-2 transition-opacity hover:opacity-85">
          <div className="flex h-6 w-6 items-center justify-center rounded-md bg-primary/10 text-primary">
            <ShieldCheck className="h-4 w-4" />
          </div>
          <span className="font-semibold text-foreground text-sm tracking-tight">ARTEX Docs</span>
        </Link>
        <Badge variant="outline" className="border-primary/30 font-mono text-[10px] text-primary">
          v1.0
        </Badge>
      </div>

      {/* Quick Search in Sidebar */}
      <div className="p-3.5 pb-2">
        <DocsSearchTrigger className="flex h-8 w-full items-center justify-between rounded-md border border-border/70 bg-background/60 px-2.5 transition-colors hover:border-primary/50" />
      </div>

      {/* Docs Navigation Groups */}
      <div className="flex-1 space-y-5 overflow-y-auto px-3.5 py-2">
        {DOC_GROUPS.map((group) => (
          <div key={group.id} className="space-y-1">
            <div className="px-2 py-1 font-mono font-semibold text-[11px] text-muted-foreground/80 uppercase tracking-wider">
              {group.label}
            </div>
            {group.items.map((section) => {
              const Icon = section.icon;
              const secNorm = section.href.replace(/\/+$/, "") || "/learn";
              const isActive = currentPath === secNorm;

              return (
                <Link
                  key={section.id}
                  href={section.href}
                  className={cn(
                    "group flex items-center justify-between gap-3 rounded-lg px-2.5 py-1.5 font-medium text-xs transition-colors",
                    isActive
                      ? "bg-primary/15 font-semibold text-primary shadow-xs ring-1 ring-primary/25"
                      : "text-muted-foreground hover:bg-muted/60 hover:text-foreground",
                  )}
                >
                  <div className="flex min-w-0 items-center gap-2.5">
                    <Icon
                      className={cn(
                        "h-4 w-4 shrink-0 transition-colors",
                        isActive ? "text-primary" : "text-muted-foreground group-hover:text-foreground",
                      )}
                    />
                    <span className="truncate">{section.shortTitle}</span>
                  </div>
                  {section.badge && (
                    <span
                      className={cn(
                        "shrink-0 rounded border px-1.5 py-0.5 font-mono font-semibold text-[10px] leading-none transition-colors",
                        isActive
                          ? "border-primary/40 bg-primary/20 text-primary"
                          : "border-border/60 bg-muted/40 text-muted-foreground group-hover:border-border group-hover:text-foreground",
                      )}
                    >
                      {section.badge}
                    </span>
                  )}
                </Link>
              );
            })}
          </div>
        ))}
      </div>

      {/* Sidebar Footer */}
      <div className="space-y-2 border-border/60 border-t p-3.5 pb-3.5 text-muted-foreground text-xs">
        <div className="flex items-center justify-between text-[11px]">
          <span className="font-mono text-muted-foreground/70 uppercase tracking-wider">Console Links</span>
        </div>
        <div className="grid grid-cols-2 gap-2 font-medium text-xs">
          <Link
            href="/chat"
            className="flex items-center justify-between rounded-md border border-border/50 bg-background/50 px-2.5 py-1.5 text-muted-foreground transition-colors hover:border-primary/40 hover:bg-primary/5 hover:text-foreground"
          >
            <span className="flex items-center gap-1.5 truncate">
              <MessageSquare className="h-3.5 w-3.5 shrink-0 text-primary" />
              <span className="truncate">Chat</span>
            </span>
            <ArrowUpRight className="h-3 w-3 shrink-0 opacity-60" />
          </Link>
          <Link
            href="/function/tasks"
            className="flex items-center justify-between rounded-md border border-border/50 bg-background/50 px-2.5 py-1.5 text-muted-foreground transition-colors hover:border-primary/40 hover:bg-primary/5 hover:text-foreground"
          >
            <span className="flex items-center gap-1.5 truncate">
              <Terminal className="h-3.5 w-3.5 shrink-0 text-primary" />
              <span className="truncate">Tasks</span>
            </span>
            <ArrowUpRight className="h-3 w-3 shrink-0 opacity-60" />
          </Link>
        </div>
      </div>
    </aside>
  );
}
