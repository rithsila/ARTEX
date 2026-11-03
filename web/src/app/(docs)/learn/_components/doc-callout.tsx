import type { ReactNode } from "react";

import { AlertCircle, AlertTriangle, Info, Lightbulb } from "lucide-react";

import { cn } from "@/lib/utils";

export type CalloutType = "info" | "tip" | "warning" | "danger";

interface DocCalloutProps {
  type?: CalloutType;
  title?: string;
  children: ReactNode;
  className?: string;
}

const CALLOUT_STYLES: Record<
  CalloutType,
  {
    container: string;
    border: string;
    icon: typeof Info;
    iconColor: string;
    titleColor: string;
  }
> = {
  info: {
    container: "bg-blue-500/5 text-blue-900 dark:text-blue-200 border-blue-500/20",
    border: "border-l-blue-500",
    icon: Info,
    iconColor: "text-blue-500",
    titleColor: "text-blue-600 dark:text-blue-400",
  },
  tip: {
    container: "bg-emerald-500/5 text-emerald-900 dark:text-emerald-200 border-emerald-500/20",
    border: "border-l-emerald-500",
    icon: Lightbulb,
    iconColor: "text-emerald-500",
    titleColor: "text-emerald-600 dark:text-emerald-400",
  },
  warning: {
    container: "bg-amber-500/5 text-amber-900 dark:text-amber-200 border-amber-500/20",
    border: "border-l-amber-500",
    icon: AlertTriangle,
    iconColor: "text-amber-500",
    titleColor: "text-amber-600 dark:text-amber-400",
  },
  danger: {
    container: "bg-red-500/5 text-red-900 dark:text-red-200 border-red-500/20",
    border: "border-l-red-500",
    icon: AlertCircle,
    iconColor: "text-red-500",
    titleColor: "text-red-600 dark:text-red-400",
  },
};

export function DocCallout({ type = "info", title, children, className }: DocCalloutProps) {
  const style = CALLOUT_STYLES[type];
  const Icon = style.icon;

  return (
    <div
      className={cn(
        "my-4 flex items-start gap-3.5 rounded-lg border border-l-4 p-4 text-xs leading-relaxed transition-colors",
        style.container,
        style.border,
        className,
      )}
    >
      <Icon className={cn("mt-0.5 h-4 w-4 shrink-0", style.iconColor)} />
      <div className="flex-1 space-y-1">
        {title && <div className={cn("font-semibold text-xs", style.titleColor)}>{title}</div>}
        <div className="text-foreground/90">{children}</div>
      </div>
    </div>
  );
}
