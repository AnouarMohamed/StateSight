import { useQuery } from "@tanstack/react-query";
import { getOverview } from "../lib/api";

export function OverviewPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["overview"],
    queryFn: getOverview
  });

  if (isLoading) {
    return <p className="text-ops-muted">Loading overview...</p>;
  }
  if (error) {
    return <p className="text-ops-bad">Failed to load overview: {(error as Error).message}</p>;
  }
  if (!data) {
    return <p className="text-ops-muted">No overview data yet.</p>;
  }

  const cards = [
    { label: "Workspaces", value: data.workspace_count },
    { label: "Applications", value: data.application_count },
    { label: "Incidents", value: data.incident_count },
    { label: "Open Jobs", value: data.open_jobs_count }
  ];

  return (
    <section className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Platform Overview</h1>
        <p className="mt-1 text-sm text-ops-muted">Quick visibility into tracked applications, drift incidents, and queue load.</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map((card) => (
          <article key={card.label} className="rounded-xl border border-ops-border bg-ops-panel p-4 shadow-panel">
            <p className="text-xs uppercase tracking-wide text-ops-muted">{card.label}</p>
            <p className="mt-2 text-3xl font-semibold">{card.value}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
