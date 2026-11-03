import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

interface DocStepsProps {
  children: ReactNode;
  className?: string;
}

interface DocStepProps {
  number: number | string;
  title: string;
  description?: string;
  children: ReactNode;
  className?: string;
}

export function DocSteps({ children, className }: DocStepsProps) {
  return <div className={cn("my-6 space-y-6", className)}>{children}</div>;
}

export function DocStep({ number, title, description, children, className }: DocStepProps) {
  return (
    <div className={cn("relative pl-10 md:pl-12", className)}>
      {/* Vertical connection line */}
      <div className="absolute top-8 bottom-0 left-[15px] w-px bg-border/60 md:left-[19px]" />

      {/* Number Badge */}
      <div className="absolute top-0 left-0 flex h-8 w-8 items-center justify-center rounded-full border border-primary/40 bg-card font-mono text-xs font-semibold text-primary shadow-xs ring-4 ring-background md:h-10 md:w-10 md:text-sm">
        {number}
      </div>

      {/* Step Header */}
      <div className="space-y-1 pt-0.5">
        <h3 className="font-semibold text-base text-foreground tracking-tight sm:text-lg">{title}</h3>
        {description && <p className="text-muted-foreground text-xs leading-relaxed">{description}</p>}
      </div>

      {/* Step Content */}
      <div className="pt-3 text-xs leading-relaxed text-foreground/90">{children}</div>
    </div>
  );
}
