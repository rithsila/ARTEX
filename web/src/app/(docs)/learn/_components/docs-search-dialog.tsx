"use client";

import * as React from "react";

import { useRouter } from "next/navigation";

import { ArrowRight, Search } from "lucide-react";

import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";

import { DOC_GROUPS } from "./docs-config";

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

          {DOC_GROUPS.map((group) => (
            <CommandGroup key={group.id} heading={group.label}>
              {group.items.map((item) => {
                const Icon = item.icon;
                return (
                  <CommandItem
                    key={item.id}
                    value={`${item.title} ${item.description}`}
                    onSelect={() => runSelect(item.href)}
                    className="flex items-center justify-between py-2"
                  >
                    <div className="flex min-w-0 items-center gap-2.5">
                      <Icon className="h-4 w-4 shrink-0 text-primary" />
                      <div className="flex min-w-0 flex-col">
                        <span className="truncate font-medium text-foreground text-xs">{item.title}</span>
                        <span className="truncate text-[11px] text-muted-foreground">{item.description}</span>
                      </div>
                    </div>
                    <ArrowRight className="h-3 w-3 shrink-0 text-muted-foreground opacity-70" />
                  </CommandItem>
                );
              })}
            </CommandGroup>
          ))}
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
