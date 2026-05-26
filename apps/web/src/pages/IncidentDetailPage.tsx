import { ArrowLeft, CircleCheck, Clock3, GitBranch, Server, ShieldAlert, UserRound } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { ActionButton, EmptyState, ErrorState, LoadState, PageHeader, Panel } from "../components/Primitives";
import { recommendationTone, severityTone } from "../lib/badgeTones";
import { type EvidenceRecord, getIncident } from "../lib/api";
import { displayValue, formatDate, formatPercent } from "../lib/format";
import { actorLabel, evidenceSourceLabel, metadataBoolean, metadataText, parseEvidenceMetadata } from "../lib/provenance";

export function IncidentDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id ?? "";
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["incident", id],
    queryFn: () => getIncident(id),
    enabled: id.length > 0
  });

  if (isLoading) {
    return <LoadState>Loading incident...</LoadState>;
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
        Failed to load incident: {(error as Error).message}
      </ErrorState>
    );
  }
  if (!data) {
    return <EmptyState>Incident not found.</EmptyState>;
  }

  return (
    <section className="space-y-6">
      <Link
        to={`/applications/${data.incident.application_id}`}
        className="touch-target inline-flex items-center gap-2 text-sm text-ops-muted hover:text-ops-text"
      >
        <ArrowLeft aria-hidden="true" className="h-4 w-4" />
        Application
      </Link>

      <PageHeader label="Incident" title={data.incident.title}>
        <Badge label={data.incident.status} />
        <Badge label={data.incident.category} />
        <Badge label={data.incident.severity} tone={severityTone(data.incident.severity)} />
        <Badge label={data.incident.recommended_action} tone={recommendationTone(data.incident.recommended_action)} />
        <span className="font-mono text-xs">confidence {formatPercent(data.incident.confidence)}</span>
      </PageHeader>

      <div className="grid min-w-0 gap-5 xl:grid-cols-[minmax(0,1fr)_21rem]">
        <div className="min-w-0 space-y-5">
          <Panel>
            <PanelHeading title="Observed differences" count={data.fields.length} />
            {data.fields.length === 0 ? (
              <EmptyState>No compared fields available.</EmptyState>
            ) : (
              <div className="divide-y divide-ops-border-muted">
                {data.fields.map((field) => (
                  <article key={field.id} className="min-w-0 px-4 py-4 sm:px-5">
                    <div className="flex flex-wrap items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="break-all font-mono text-xs text-ops-muted">{field.resource_ref}</p>
                        <p className="mt-2 break-all font-mono text-sm text-ops-text">{field.field_path}</p>
                      </div>
                      <Badge label={field.difference_type} tone="warn" />
                    </div>
                    <dl className="mt-4 grid gap-3 sm:grid-cols-2">
                      <ValueCell label="Desired" value={displayValue(field.desired_value)} />
                      <ValueCell label="Live" value={displayValue(field.live_value)} changed />
                    </dl>
                  </article>
                ))}
              </div>
            )}
          </Panel>

          <Panel>
            <PanelHeading title="Provenance" count={data.evidence.length} />
            {data.evidence.length === 0 ? (
              <EmptyState>No provenance records available.</EmptyState>
            ) : (
              <div className="divide-y divide-ops-border-muted">
                {data.evidence.map((record) => (
                  <ProvenanceRecord key={record.id} record={record} />
                ))}
              </div>
            )}
          </Panel>
        </div>

        <Panel className="h-fit">
          <PanelHeading title="Timeline" count={data.timeline.length} />
          {data.timeline.length === 0 ? (
            <EmptyState>No timeline events available.</EmptyState>
          ) : (
            <ol className="divide-y divide-ops-border-muted">
              {data.timeline.map((event, index) => (
                <li key={`${event.at}-${event.type}-${index}`} className="px-4 py-3">
                  <p className="flex items-center gap-2 text-xs text-ops-muted">
                    <Clock3 aria-hidden="true" className="h-3.5 w-3.5" />
                    {formatDate(event.at)}
                  </p>
                  <p className="mt-2 break-words text-sm">{event.summary}</p>
                  <p className="mt-1 font-mono text-xs text-ops-muted">{event.type}</p>
                </li>
              ))}
            </ol>
          )}
        </Panel>
      </div>
    </section>
  );
}

function PanelHeading({ title, count }: { title: string; count: number }) {
  return (
    <header className="flex items-center justify-between border-b border-ops-border px-4 py-3 sm:px-5">
      <h2 className="text-sm font-semibold">{title}</h2>
      <span className="rounded-md bg-ops-bg px-2 py-0.5 font-mono text-xs text-ops-muted">{count}</span>
    </header>
  );
}

function ValueCell({ label, value, changed = false }: { label: string; value: string; changed?: boolean }) {
  return (
    <div className={`min-w-0 rounded-md border px-3 py-2 ${changed ? "border-ops-warn-border bg-ops-warn-soft" : "border-ops-border-muted bg-ops-panel-alt"}`}>
      <dt className="text-xs text-ops-muted">{label}</dt>
      <dd className="mt-1 break-all font-mono text-sm text-ops-text">{value}</dd>
    </div>
  );
}

function ProvenanceRecord({ record }: { record: EvidenceRecord }) {
  const metadata = parseEvidenceMetadata(record);
  const isSynthetic = record.source === "synthetic" || metadataBoolean(metadata, "trusted_observation") === false;
  const isOwnership = record.source === "managedFields";
  const trust = isSynthetic
    ? { label: "Untrusted", tone: "warn" as const }
    : isOwnership
      ? { label: "Ownership signal", tone: "neutral" as const }
      : { label: "Captured", tone: "good" as const };
  const SourceIcon: LucideIcon = record.source === "git" ? GitBranch : record.source === "managedFields" ? ShieldAlert : Server;

  const facts = compactFacts([
    ["Revision", metadataText(metadata, "revision")],
    ["Path", metadataText(metadata, "path")],
    ["Collector", metadataText(metadata, "collection_source")],
    ["Cluster", metadataText(metadata, "cluster_name")],
    ["Field", metadataText(metadata, "field_path")],
    ["Observed", metadataText(metadata, "observed_at")]
  ]);

  return (
    <article className="px-4 py-4 sm:px-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2">
          <SourceIcon aria-hidden="true" className="h-4 w-4 shrink-0 text-ops-muted" />
          <h3 className="text-sm font-medium">{evidenceSourceLabel(record.source)}</h3>
        </div>
        <Badge label={trust.label} tone={trust.tone} />
      </div>
      <p className="mt-3 max-w-[72ch] break-words text-sm leading-6 text-ops-text">{record.detail}</p>
      <dl className="mt-3 grid gap-x-6 gap-y-2 text-xs sm:grid-cols-2">
        <div className="flex min-w-0 items-center gap-2">
          <UserRound aria-hidden="true" className="h-3.5 w-3.5 shrink-0 text-ops-muted" />
          <dt className="text-ops-muted">Actor</dt>
          <dd className="break-all font-mono text-ops-text">{actorLabel(record.actor)}</dd>
        </div>
        <div className="flex items-center gap-2">
          <CircleCheck aria-hidden="true" className="h-3.5 w-3.5 shrink-0 text-ops-muted" />
          <dt className="text-ops-muted">Confidence</dt>
          <dd className="font-mono text-ops-text">{formatPercent(record.confidence)}</dd>
        </div>
        {facts.map(([label, value]) => (
          <div className="min-w-0" key={label}>
            <dt className="text-ops-muted">{label}</dt>
            <dd className="mt-1 break-all font-mono text-ops-text">{value}</dd>
          </div>
        ))}
      </dl>
      {isOwnership ? (
        <p className="mt-3 rounded-md border border-ops-border-muted bg-ops-panel-alt px-3 py-2 text-xs leading-5 text-ops-muted">
          Field ownership identifies a reported manager; it does not prove who introduced the drift.
        </p>
      ) : null}
    </article>
  );
}

function compactFacts(facts: Array<[string, string | undefined]>) {
  return facts.filter((fact): fact is [string, string] => fact[1] !== undefined);
}
