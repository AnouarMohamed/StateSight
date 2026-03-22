import { Link, NavLink, Outlet } from "react-router-dom";

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  `rounded-md px-3 py-2 text-sm font-medium transition ${
    isActive ? "bg-ops-border text-ops-text" : "text-ops-muted hover:bg-[#1a2533] hover:text-ops-text"
  }`;

export function Layout() {
  return (
    <div className="min-h-screen bg-ops-bg text-ops-text">
      <header className="border-b border-ops-border bg-gradient-to-r from-[#0f1824] to-[#0d1520]">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <Link to="/" className="text-lg font-semibold tracking-wide">
            StateSight
          </Link>
          <nav className="flex gap-2">
            <NavLink to="/" end className={navLinkClass}>
              Overview
            </NavLink>
            <NavLink to="/applications" className={navLinkClass}>
              Applications
            </NavLink>
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-8">
        <Outlet />
      </main>
    </div>
  );
}
