import { AlertTriangle, ArrowRight, Boxes, GitBranch, Rows3 } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { ActionButton, ErrorState, LoadState, PageHeader, Panel, actionLinkClass } from "../components/Primitives";
import { getOverview } from "../lib/api";

export function OverviewPage() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["overview"],
    queryFn: getOverview
  });

  if (isLoading) {
    return <LoadState>Loading overview...</LoadState>;
  }
  if (error) {
    return (
      <ErrorState
        action={
          <ActionButton onClick={() => void refetch()} type="button">
            Retry
          </ActionButton>
        }
      >
        Failed to load overview: {(error as Error).message}
      </ErrorState>
    );
  }
  if (!data) {
    return <LoadState>No overview data available.</LoadState>;
  }

  const queueRows = [
    {
      label: "Drift incidents",
      count: data.incident_count,
      state: data.incident_count > 0 ? "Review required" : "Clear",
      icon: AlertTriangle,
      tone: data.incident_count > 0 ? "text-ops-bad" : "text-ops-good"
    },
    {
      label: "Analysis jobs",
      count: data.open_jobs_count,
      state: data.open_jobs_count > 0 ? "Running" : "Idle",
      icon: Rows3,
      tone: data.open_jobs_count > 0 ? "text-ops-warn" : "text-ops-muted"
    }
  ];

  return (
    <section className="space-y-4">
      <PageHeader
        label="Operations"
        title="Operational status"
        actions={
          <Link to="/applications" className={actionLinkClass}>
            Applications
            <ArrowRight aria-hidden="true" className="h-4 w-4" />
          </Link>
        }
      />

      <Panel aria-label="Inventory totals">
        <dl className="grid grid-cols-2 divide-x divide-y divide-ops-border sm:grid-cols-4 sm:divide-y-0">
          <SummaryTotal icon={GitBranch} label="Workspaces" value={data.workspace_count} />
          <SummaryTotal icon={Boxes} label="Applications" value={data.application_count} />
          <SummaryTotal icon={AlertTriangle} label="Incidents" value={data.incident_count} urgent={data.incident_count > 0} />
          <SummaryTotal icon={Rows3} label="Open jobs" value={data.open_jobs_count} />
        </dl>
      </Panel>

      <Panel>
        <div className="border-b border-ops-border-muted px-3 py-2">
          <h2 className="console-label text-ops-dim">Review queue</h2>
        </div>
        <div className="divide-y divide-ops-border-muted">
          {queueRows.map(({ label, count, state, icon: Icon, tone }) => (
            <div key={label} className="console-copy grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-4 px-3 py-2.5">
              <span className="flex min-w-0 items-center gap-3">
                <Icon aria-hidden="true" className={`h-4 w-4 shrink-0 ${tone}`} />
                <span className="truncate">{label}</span>
              </span>
              <span className="font-mono text-ops-muted">{count}</span>
              <span className={`console-label min-w-24 text-right font-medium ${tone}`}>{state}</span>
            </div>
          ))}
        </div>
      </Panel>
    </section>
  );
}

function SummaryTotal({
  icon: Icon,
  label,
  value,
  urgent = false
}: {
  icon: LucideIcon;
  label: string;
  value: number;
  urgent?: boolean;
}) {
  return (
    <div className="flex min-h-20 items-center gap-3 px-3 py-3">
      <Icon aria-hidden="true" className={`h-4 w-4 shrink-0 ${urgent ? "text-ops-bad" : "text-ops-muted"}`} />
      <div className="min-w-0">
        <dt className="console-label text-ops-dim">{label}</dt>
        <dd className={`mt-1 font-mono text-xl ${urgent ? "text-ops-accent" : "text-ops-text"}`}>{value}</dd>
      </div>
    </div>
  );
}
