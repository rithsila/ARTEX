"use client";

import * as React from "react";

import { ArrowUp, ListTree } from "lucide-react";

import { cn } from "@/lib/utils";

export interface TocItem {
  id: string;
  title: string;
  level?: 2 | 3;
}

interface DocsTocProps {
  items: TocItem[];
}

export function DocsToc({ items }: DocsTocProps) {
  const [activeId, setActiveId] = React.useState<string>(items[0]?.id || "");

  React.useEffect(() => {
    if (items.length === 0) return;

    const headingElements = items
      .map((item) => document.getElementById(item.id))
      .filter((el): el is HTMLElement => el !== null);

    const onScroll = () => {
      const scrollPos = window.scrollY + 120;
      let currentId = items[0]?.id || "";
      for (const el of headingElements) {
        if (el.offsetTop <= scrollPos) {
          currentId = el.id;
        }
      }
      setActiveId(currentId);
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
    return () => window.removeEventListener("scroll", onScroll);
  }, [items]);

  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  if (items.length === 0) {
    return null;
  }

  return (
    <aside className="sticky top-20 hidden h-[calc(100vh-6rem)] w-60 shrink-0 self-start overflow-y-auto px-4 py-2 xl:block">
      <div className="flex items-center gap-2 pb-3 font-semibold text-foreground text-xs uppercase tracking-wider">
        <ListTree className="h-3.5 w-3.5 text-primary" />
        <span>On this page</span>
      </div>

      <nav className="space-y-1 text-xs">
        {items.map((item) => {
          const isActive = activeId === item.id;
          return (
            <a
              key={item.id}
              href={`#${item.id}`}
              className={cn(
                "block py-1 leading-relaxed transition-colors",
                item.level === 3 ? "pl-4 text-[11px]" : "pl-1 font-medium",
                isActive
                  ? "border-primary border-l-2 pl-2 font-semibold text-primary"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {item.title}
            </a>
          );
        })}
      </nav>

      <div className="mt-8 border-border/60 border-t pt-4">
        <button
          type="button"
          onClick={scrollToTop}
          className="inline-flex items-center gap-1.5 text-muted-foreground text-xs transition-colors hover:text-foreground"
        >
          <ArrowUp className="h-3.5 w-3.5 text-primary" />
          <span>Back to top</span>
        </button>
      </div>
    </aside>
  );
}
