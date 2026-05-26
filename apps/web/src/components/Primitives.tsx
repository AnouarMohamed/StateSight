import { AlertCircle } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { ButtonHTMLAttributes, HTMLAttributes, ReactNode } from "react";

export function PageHeader({
  label,
  title,
  actions,
  children
}: {
  label?: string;
  title: string;
  actions?: ReactNode;
  children?: ReactNode;
}) {
  return (
    <header className="flex flex-wrap items-start justify-between gap-4 border-b border-ops-border-muted pb-3">
      <div className="min-w-0">
        {label ? <p className="console-label mb-1 text-ops-dim">{label}</p> : null}
        <h1 className="break-words text-base font-medium text-ops-text">{title}</h1>
        {children ? <div className="console-copy mt-2 flex flex-wrap items-center gap-2 text-ops-muted">{children}</div> : null}
      </div>
      {actions ? <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div> : null}
    </header>
  );
}

export function Panel({ children, className = "", ...props }: HTMLAttributes<HTMLElement>) {
  return (
    <section className={`overflow-hidden border border-ops-border bg-ops-panel ${className}`} {...props}>
      {children}
    </section>
  );
}

const buttonVariants = {
  primary: "border-ops-action-border bg-ops-action text-ops-bg hover:bg-ops-action-hover active:bg-ops-action-hover",
  secondary: "border-ops-border bg-ops-panel text-ops-text hover:bg-ops-shell-hover active:bg-ops-panel-alt",
  quiet: "border-transparent bg-transparent text-ops-accent hover:bg-ops-accent-soft active:bg-ops-accent-soft",
  danger: "border-ops-bad-border bg-ops-bad-soft text-ops-bad hover:bg-ops-bad-border hover:text-ops-text active:bg-ops-bad-border",
  dangerQuiet: "border-transparent bg-transparent text-ops-bad hover:bg-ops-bad-soft active:bg-ops-bad-soft"
} as const;

export function ActionButton({
  icon: Icon,
  variant = "secondary",
  className = "",
  children,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  icon?: LucideIcon;
  variant?: keyof typeof buttonVariants;
}) {
  return (
    <button
      className={`console-action console-transition touch-target inline-flex h-8 items-center justify-center gap-2 border px-3 disabled:cursor-not-allowed disabled:opacity-55 ${buttonVariants[variant]} ${className}`}
      {...props}
    >
      {Icon ? <Icon aria-hidden="true" className="h-4 w-4 shrink-0" /> : null}
      {children}
    </button>
  );
}

export const actionLinkClass =
  "console-action console-transition touch-target inline-flex h-8 items-center justify-center gap-2 border border-ops-border bg-ops-panel px-3 text-ops-text hover:bg-ops-shell-hover active:bg-ops-panel-alt";

export function LoadState({ children = "Loading..." }: { children?: ReactNode }) {
  return (
    <Panel className="p-5" aria-busy="true" role="status">
      <span className="sr-only">{children}</span>
      <div className="space-y-3">
        <div className="h-3 w-40 animate-pulse bg-ops-border-muted" />
        <div className="h-9 animate-pulse bg-ops-panel-alt" />
        <div className="h-9 animate-pulse bg-ops-panel-alt" />
      </div>
    </Panel>
  );
}

export function ErrorState({ children, action }: { children: ReactNode; action?: ReactNode }) {
  return (
    <Panel className="flex flex-wrap items-center justify-between gap-3 border-ops-bad-border bg-ops-bad-soft p-4 text-sm text-ops-bad" role="alert">
      <span className="flex min-w-0 items-center gap-2">
        <AlertCircle aria-hidden="true" className="h-4 w-4 shrink-0" />
        <span className="break-words">{children}</span>
      </span>
      {action}
    </Panel>
  );
}

export function EmptyState({ children }: { children: ReactNode }) {
  return <p className="console-copy px-4 py-8 font-mono text-ops-muted">{children}</p>;
}
