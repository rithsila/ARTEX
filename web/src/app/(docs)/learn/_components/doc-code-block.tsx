"use client";

import * as React from "react";

import { Check, Copy, Terminal } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface DocCodeBlockProps {
  code: string;
  language?: string;
  title?: string;
  className?: string;
}

export function DocCodeBlock({ code, language = "bash", title, className }: DocCodeBlockProps) {
  const [copied, setCopied] = React.useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // fallback
    }
  };

  return (
    <div
      className={cn(
        "group relative my-4 overflow-hidden rounded-xl border border-border/80 bg-zinc-950 text-zinc-100 shadow-sm",
        className,
      )}
    >
      {/* Header Bar */}
      <div className="flex h-9 items-center justify-between border-zinc-800/80 border-b bg-zinc-900/60 px-3.5 text-xs text-zinc-400">
        <div className="flex items-center gap-2 font-mono text-[11px]">
          <Terminal className="h-3.5 w-3.5 text-primary" />
          <span>{title ?? language}</span>
        </div>
        <div className="flex items-center gap-2">
          {language && <span className="font-mono text-[10px] text-zinc-500 uppercase">{language}</span>}
          <Button
            variant="ghost"
            size="icon"
            onClick={handleCopy}
            className="h-6 w-6 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100"
            title="Copy snippet"
          >
            {copied ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3" />}
            <span className="sr-only">Copy code</span>
          </Button>
        </div>
      </div>

      {/* Code Area */}
      <div className="overflow-x-auto p-4 font-mono text-xs text-zinc-200 leading-relaxed">
        <pre className="selection:bg-primary/30">
          <code>{code}</code>
        </pre>
      </div>
    </div>
  );
}
