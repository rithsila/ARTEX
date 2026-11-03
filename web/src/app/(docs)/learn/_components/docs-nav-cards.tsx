"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { ArrowLeft, ArrowRight } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";

import { getDocsNavigation } from "./docs-config";

export function DocsNavCards() {
  const rawPath = usePathname();
  const currentPath = (rawPath ? decodeURIComponent(rawPath) : "").replace(/\/+$/, "") || "/learn";
  const { prev, next } = getDocsNavigation(currentPath);

  if (!prev && !next) {
    return null;
  }

  return (
    <div className="grid gap-4 pt-10 sm:grid-cols-2">
      {prev ? (
        <Link href={prev.href} className="group block focus:outline-hidden">
          <Card className="h-full border-border/60 transition-all hover:border-primary/50 hover:bg-muted/30">
            <CardContent className="flex flex-col justify-between p-4 sm:p-5">
              <div className="flex items-center gap-1.5 font-mono text-muted-foreground text-xs uppercase tracking-wider group-hover:text-primary">
                <ArrowLeft className="h-3.5 w-3.5 transition-transform group-hover:-translate-x-1" />
                <span>Previous</span>
              </div>
              <div className="pt-2">
                <div className="font-semibold text-foreground text-sm group-hover:text-primary">{prev.title}</div>
                <p className="line-clamp-2 pt-1 text-muted-foreground text-xs leading-relaxed">{prev.description}</p>
              </div>
            </CardContent>
          </Card>
        </Link>
      ) : (
        <div className="hidden sm:block" />
      )}

      {next && (
        <Link href={next.href} className="group block focus:outline-hidden sm:text-right">
          <Card className="h-full border-border/60 transition-all hover:border-primary/50 hover:bg-muted/30">
            <CardContent className="flex flex-col justify-between p-4 sm:p-5">
              <div className="flex items-center gap-1.5 font-mono text-muted-foreground text-xs uppercase tracking-wider group-hover:text-primary sm:justify-end">
                <span>Next</span>
                <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-1" />
              </div>
              <div className="pt-2">
                <div className="font-semibold text-foreground text-sm group-hover:text-primary">{next.title}</div>
                <p className="line-clamp-2 pt-1 text-muted-foreground text-xs leading-relaxed">{next.description}</p>
              </div>
            </CardContent>
          </Card>
        </Link>
      )}
    </div>
  );
}
