import { Boxes, LayoutDashboard, ShieldCheck } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Link, NavLink, Outlet } from "react-router-dom";

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
  return (
    <div className="min-h-screen bg-ops-bg text-ops-text">
      <a
        href="#main-content"
        className="fixed left-3 top-3 z-50 -translate-y-20 rounded-md border border-ops-accent bg-ops-panel px-3 py-2 text-sm text-ops-text focus:translate-y-0"
      >
        Skip to content
      </a>
      <header className="border-b border-ops-border bg-ops-shell">
        <div className="mx-auto flex max-w-[1400px] flex-wrap items-center gap-x-7 gap-y-3 px-4 py-3 sm:px-6">
          <Link to="/" className="touch-target flex items-center gap-2 text-[15px] font-semibold text-ops-text">
            <ShieldCheck aria-hidden="true" className="h-5 w-5 text-ops-good" />
            StateSight
          </Link>
          <nav aria-label="Primary" className="order-3 flex w-full gap-1 sm:order-none sm:w-auto">
            {navigation.map(({ label, path, end, icon: Icon }) => (
              <NavLink
                key={path}
                to={path}
                end={end}
                className={({ isActive }) =>
                  [
                    "touch-target inline-flex h-9 items-center gap-2 rounded-md px-3 text-sm transition-colors",
                    isActive
                      ? "bg-ops-shell-hover font-medium text-ops-text"
                      : "text-ops-muted hover:bg-ops-shell-hover hover:text-ops-text"
                  ].join(" ")
                }
              >
                <Icon aria-hidden="true" className="h-4 w-4" />
                {label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>
      <main id="main-content" className="mx-auto max-w-[1400px] px-4 py-6 sm:px-6 lg:py-8">
        <Outlet />
      </main>
    </div>
  );
}
