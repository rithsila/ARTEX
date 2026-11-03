"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { ChevronDown, LayoutDashboard } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";

import { DOC_GROUPS, getDocsNavigation } from "./docs-config";
import { DocsSearchTrigger } from "./docs-search-dialog";

export function DocsHeaderBar() {
  const rawPath = usePathname();
  const currentPath = (rawPath ? decodeURIComponent(rawPath) : "").replace(/\/+$/, "") || "/learn";
  const { current } = getDocsNavigation(currentPath);

  return (
    <div className="sticky top-0 z-20 flex h-14 w-full items-center justify-between border-border/60 border-b bg-card/60 px-4 backdrop-blur-md md:px-6">
      {/* Breadcrumb / Current page selector */}
      <div className="flex items-center gap-2">
        <span className="font-mono text-muted-foreground text-xs uppercase tracking-wider">Docs</span>
        <span className="text-muted-foreground/60 text-xs">/</span>

        {/* Mobile dropdown selector */}
        <div className="md:hidden">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" className="h-7 gap-1 px-2 font-semibold text-xs">
                <span>{current?.shortTitle || "Overview"}</span>
                <ChevronDown className="h-3.5 w-3.5 opacity-60" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="max-h-96 w-64 overflow-y-auto">
              {DOC_GROUPS.map((group, idx) => (
                <DropdownMenuGroup key={group.id}>
                  {idx > 0 && <DropdownMenuSeparator />}
                  <DropdownMenuLabel className="font-mono text-[10px] text-muted-foreground uppercase">
                    {group.label}
                  </DropdownMenuLabel>
                  {group.items.map((section) => {
                    const Icon = section.icon;
                    const secNorm = section.href.replace(/\/+$/, "") || "/learn";
                    const isSelected = currentPath === secNorm;
                    return (
                      <DropdownMenuItem key={section.id} asChild>
                        <Link
                          href={section.href}
                          className={cn(
                            "flex items-center justify-between gap-2 text-xs",
                            isSelected && "font-semibold text-primary",
                          )}
                        >
                          <span className="flex items-center gap-2">
                            <Icon className="h-3.5 w-3.5" />
                            {section.shortTitle}
                          </span>
                        </Link>
                      </DropdownMenuItem>
                    );
                  })}
                </DropdownMenuGroup>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Desktop breadcrumb title */}
        <span className="hidden font-semibold text-foreground text-xs md:inline-block">
          {current?.title || "Documentation"}
        </span>
      </div>

      {/* Right Actions: Quick Search Trigger + Back to Console */}
      <div className="flex items-center gap-2.5">
        <DocsSearchTrigger className="flex h-8 w-36 items-center justify-between rounded-md border border-border/70 bg-background/60 px-2.5 transition-colors hover:border-primary/50 sm:w-52" />

        <Button
          variant="outline"
          size="sm"
          asChild
          className="h-8 gap-1.5 px-3 font-medium text-xs hover:border-primary/50"
        >
          <Link href="/dashboard" title="Return to ARTEX Console">
            <LayoutDashboard className="h-3.5 w-3.5 text-primary" />
            <span>Back to Console</span>
          </Link>
        </Button>
      </div>
    </div>
  );
}
