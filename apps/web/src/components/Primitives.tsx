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
    <header className="flex flex-wrap items-start justify-between gap-4">
      <div className="min-w-0">
        {label ? <p className="mb-1.5 text-xs font-medium text-ops-muted">{label}</p> : null}
        <h1 className="break-words text-xl font-semibold text-ops-text sm:text-2xl">{title}</h1>
        {children ? <div className="mt-2 flex flex-wrap items-center gap-2 text-sm text-ops-muted">{children}</div> : null}
      </div>
      {actions ? <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div> : null}
    </header>
  );
}

export function Panel({ children, className = "", ...props }: HTMLAttributes<HTMLElement>) {
  return (
    <section className={`overflow-hidden rounded-md border border-ops-border bg-ops-panel ${className}`} {...props}>
      {children}
    </section>
  );
}

const buttonVariants = {
  primary: "border-ops-action-border bg-ops-action text-ops-text hover:bg-ops-action-hover",
  secondary: "border-ops-border bg-ops-panel text-ops-text hover:bg-ops-shell-hover",
  quiet: "border-transparent bg-transparent text-ops-accent hover:bg-ops-accent-soft"
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
      className={`touch-target inline-flex h-9 items-center justify-center gap-2 rounded-md border px-3 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-55 ${buttonVariants[variant]} ${className}`}
      {...props}
    >
      {Icon ? <Icon aria-hidden="true" className="h-4 w-4 shrink-0" /> : null}
      {children}
    </button>
  );
}

export const actionLinkClass =
  "touch-target inline-flex h-9 items-center justify-center gap-2 rounded-md border border-ops-border bg-ops-panel px-3 text-sm font-medium text-ops-text transition-colors hover:bg-ops-shell-hover";

export function LoadState({ children = "Loading..." }: { children?: ReactNode }) {
  return (
    <Panel className="p-5" aria-busy="true">
      <span className="sr-only">{children}</span>
      <div className="space-y-3">
        <div className="h-4 w-40 animate-pulse rounded bg-ops-border-muted" />
        <div className="h-10 animate-pulse rounded bg-ops-panel-alt" />
        <div className="h-10 animate-pulse rounded bg-ops-panel-alt" />
      </div>
    </Panel>
  );
}

export function ErrorState({ children, action }: { children: ReactNode; action?: ReactNode }) {
  return (
    <Panel className="flex flex-wrap items-center justify-between gap-3 border-ops-bad-border bg-ops-bad-soft p-4 text-sm text-ops-bad">
      <span className="flex min-w-0 items-center gap-2">
        <AlertCircle aria-hidden="true" className="h-4 w-4 shrink-0" />
        <span className="break-words">{children}</span>
      </span>
      {action}
    </Panel>
  );
}

export function EmptyState({ children }: { children: ReactNode }) {
  return <p className="px-5 py-8 text-sm text-ops-muted">{children}</p>;
}
