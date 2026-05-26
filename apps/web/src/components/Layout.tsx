import { AlertTriangle, Boxes, LayoutDashboard, PanelLeftClose, PanelLeftOpen, Rows3 } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, NavLink, Outlet } from "react-router-dom";
import { getOverview } from "../lib/api";
import { GitMark } from "./GitMark";

type NavigationItem = {
  label: string;
  path: string;
  end?: boolean;
  icon: LucideIcon;
};

const navigation: NavigationItem[] = [
  { label: "Overview", path: "/", end: true, icon: LayoutDashboard },
  { label: "Applications", path: "/applications", icon: Boxes }
];

export function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const { data } = useQuery({
    queryKey: ["overview"],
    queryFn: getOverview
  });

  return (
    <div className="flex min-h-screen bg-ops-bg text-ops-text">
      <a
        href="#main-content"
        className="fixed left-3 top-3 z-50 -translate-y-20 border border-ops-accent bg-ops-panel px-3 py-2 text-sm text-ops-text focus:translate-y-0"
      >
        Skip to content
      </a>
      <aside
        className={`sticky top-0 hidden h-screen shrink-0 flex-col border-r border-ops-border-muted bg-ops-shell transition-[width] duration-150 md:flex ${
          sidebarOpen ? "w-60" : "w-14"
        }`}
      >
        <div className="flex h-14 items-center justify-between border-b border-ops-border-muted px-3">
          <Link to="/" aria-label="StateSight home" className="flex min-w-0 items-center gap-2">
            <GitMark className="h-7 w-7 shrink-0" />
            {sidebarOpen ? (
              <span className="min-w-0">
                <span className="block text-sm font-semibold leading-4 text-ops-text">StateSight</span>
                <span className="block truncate text-[10px] uppercase text-ops-dim">GitOps drift forensics</span>
              </span>
            ) : null}
          </Link>
          {sidebarOpen ? (
            <button
              aria-label="Collapse navigation"
              className="touch-target inline-flex h-8 w-8 items-center justify-center text-ops-dim hover:bg-ops-shell-hover hover:text-ops-text"
              onClick={() => setSidebarOpen(false)}
              type="button"
            >
              <PanelLeftClose aria-hidden="true" className="h-4 w-4" />
            </button>
          ) : null}
        </div>
        {!sidebarOpen ? (
          <button
            aria-label="Expand navigation"
            className="touch-target mx-auto mt-2 inline-flex h-9 w-9 items-center justify-center text-ops-dim hover:bg-ops-shell-hover hover:text-ops-text"
            onClick={() => setSidebarOpen(true)}
            type="button"
          >
            <PanelLeftOpen aria-hidden="true" className="h-4 w-4" />
          </button>
        ) : null}
        <nav aria-label="Primary" className="mt-2 flex-1 space-y-1 px-2">
          {navigation.map(({ label, path, end, icon: Icon }) => (
            <NavItem key={path} icon={Icon} label={label} path={path} end={end} expanded={sidebarOpen} />
          ))}
        </nav>
        {sidebarOpen && data ? (
          <div className="border-t border-ops-border-muted px-3 py-3">
            <p className="mb-3 text-[10px] uppercase tracking-[0.12em] text-ops-dim">Signal queue</p>
            <StatRow label="Incidents" value={data.incident_count} urgent={data.incident_count > 0} icon={AlertTriangle} />
            <StatRow label="Open jobs" value={data.open_jobs_count} icon={Rows3} />
            <StatRow label="Applications" value={data.application_count} icon={Boxes} />
          </div>
        ) : null}
        {sidebarOpen ? (
          <div className="border-t border-ops-border-muted px-3 py-2 font-mono text-[10px] uppercase text-ops-dim">Evidence console</div>
        ) : null}
      </aside>
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="border-b border-ops-border-muted bg-ops-shell md:hidden">
          <div className="flex h-14 items-center px-4">
            <Link to="/" className="flex items-center gap-2 font-semibold text-ops-text">
              <GitMark className="h-7 w-7" />
              StateSight
            </Link>
          </div>
          <nav aria-label="Primary" className="flex gap-1 overflow-x-auto border-t border-ops-border-muted px-3 py-2">
            {navigation.map(({ label, path, end, icon: Icon }) => (
              <NavItem key={path} icon={Icon} label={label} path={path} end={end} expanded />
            ))}
          </nav>
        </header>
        <main id="main-content" className="min-w-0 flex-1 px-4 py-4 sm:px-6 sm:py-5">
          <Outlet />
        </main>
      </div>
    </div>
  );
}

function NavItem({
  icon: Icon,
  label,
  path,
  end,
  expanded
}: NavigationItem & {
  expanded: boolean;
}) {
  return (
    <NavLink
      aria-label={label}
      to={path}
      end={end}
      className={({ isActive }) =>
        [
          "touch-target flex h-9 items-center gap-3 px-3 text-xs uppercase transition-colors",
          expanded ? "" : "justify-center px-0",
          isActive ? "bg-ops-action font-medium text-ops-bg" : "text-ops-dim hover:bg-ops-shell-hover hover:text-ops-text"
        ].join(" ")
      }
    >
      <Icon aria-hidden="true" className="h-3.5 w-3.5 shrink-0" />
      {expanded ? <span>{label}</span> : null}
    </NavLink>
  );
}

function StatRow({ icon: Icon, label, value, urgent = false }: { icon: LucideIcon; label: string; value: number; urgent?: boolean }) {
  return (
    <div className="mb-2 flex items-center justify-between gap-2 text-xs last:mb-0">
      <span className="flex items-center gap-2 text-ops-muted">
        <Icon aria-hidden="true" className="h-3.5 w-3.5" />
        {label}
      </span>
      <span className={`font-mono ${urgent ? "text-ops-accent" : "text-ops-text"}`}>{value}</span>
    </div>
  );
}
