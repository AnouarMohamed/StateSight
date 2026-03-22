import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { Badge, recommendationTone, severityTone } from "../components/Badge";
import { getIncident } from "../lib/api";

export function IncidentDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id ?? "";
  const { data, isLoading, error } = useQuery({
    queryKey: ["incident", id],
    queryFn: () => getIncident(id),
    enabled: id.length > 0
  });

  if (isLoading) {
    return <p className="text-ops-muted">Loading incident...</p>;
  }
  if (error) {
    return <p className="text-ops-bad">Failed to load incident: {(error as Error).message}</p>;
  }
  if (!data) {
    return <p className="text-ops-muted">Incident not found.</p>;
  }

  return (
    <section className="space-y-6">
      <div className="rounded-xl border border-ops-border bg-ops-panel p-5 shadow-panel">
        <h1 className="text-2xl font-semibold">{data.incident.title}</h1>
        <div className="mt-3 flex flex-wrap gap-2">
          <Badge label={data.incident.category} />
          <Badge label={data.incident.severity} tone={severityTone(data.incident.severity)} />
          <Badge label={data.incident.recommended_action} tone={recommendationTone(data.incident.recommended_action)} />
        </div>
        <p className="mt-4 text-sm text-ops-muted">Confidence: {(data.incident.confidence * 100).toFixed(0)}%</p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <article className="rounded-xl border border-ops-border bg-ops-panel p-5 shadow-panel">
          <h2 className="text-lg font-semibold">Field Evidence</h2>
          <div className="mt-3 space-y-3">
            {data.fields.length === 0 ? (
              <p className="text-sm text-ops-muted">No field entries available.</p>
            ) : (
              data.fields.map((field) => (
                <div key={field.id} className="rounded-lg border border-ops-border bg-[#101825] p-3">
                  <p className="text-xs uppercase tracking-wide text-ops-muted">{field.resource_ref}</p>
                  <p className="mt-1 font-mono text-sm">{field.field_path}</p>
                  <p className="mt-2 text-xs text-ops-muted">
                    desired: <span className="text-ops-text">{field.desired_value}</span> | live:{" "}
                    <span className="text-ops-text">{field.live_value}</span>
                  </p>
                </div>
              ))
            )}
          </div>
        </article>

        <article className="rounded-xl border border-ops-border bg-ops-panel p-5 shadow-panel">
          <h2 className="text-lg font-semibold">Attribution Evidence</h2>
          <div className="mt-3 space-y-3">
            {data.evidence.length === 0 ? (
              <p className="text-sm text-ops-muted">No attribution evidence available.</p>
            ) : (
              data.evidence.map((record) => (
                <div key={record.id} className="rounded-lg border border-ops-border bg-[#101825] p-3">
                  <p className="text-xs uppercase tracking-wide text-ops-muted">{record.source}</p>
                  <p className="mt-1 text-sm">{record.detail}</p>
                  <p className="mt-2 text-xs text-ops-muted">
                    actor: <span className="text-ops-text">{record.actor}</span> | confidence:{" "}
                    <span className="text-ops-text">{(record.confidence * 100).toFixed(0)}%</span>
                  </p>
                </div>
              ))
            )}
          </div>
        </article>
      </div>

      <article className="rounded-xl border border-ops-border bg-ops-panel p-5 shadow-panel">
        <h2 className="text-lg font-semibold">Timeline</h2>
        <div className="mt-3 space-y-3">
          {data.timeline.length === 0 ? (
            <p className="text-sm text-ops-muted">No timeline events available.</p>
          ) : (
            data.timeline.map((event, index) => (
              <div key={`${event.at}-${event.type}-${index}`} className="rounded-lg border border-ops-border bg-[#101825] p-3">
                <p className="text-xs uppercase tracking-wide text-ops-muted">{event.type}</p>
                <p className="mt-1 text-sm">{event.summary}</p>
                <p className="mt-2 text-xs text-ops-muted">{new Date(event.at).toLocaleString()}</p>
              </div>
            ))
          )}
        </div>
      </article>
    </section>
  );
}
