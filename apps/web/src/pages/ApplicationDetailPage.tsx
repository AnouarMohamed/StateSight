import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { analyzeApplication, getApplication } from "../lib/api";
import { Badge, recommendationTone, severityTone } from "../components/Badge";

const tabs = [
  { id: "incidents", label: "Incidents" },
  { id: "suppressions", label: "Suppressed" }
] as const;

type ApplicationTab = (typeof tabs)[number]["id"];

export function ApplicationDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id ?? "";
  const [activeTab, setActiveTab] = useState<ApplicationTab>("incidents");
  const queryClient = useQueryClient();
  const { data, isLoading, error } = useQuery({
    queryKey: ["application", id],
    queryFn: () => getApplication(id),
    enabled: id.length > 0
  });

  const analyzeMutation = useMutation({
    mutationFn: () => analyzeApplication(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["application", id] });
      await queryClient.invalidateQueries({ queryKey: ["overview"] });
    }
  });

  if (isLoading) {
    return <p className="text-ops-muted">Loading application details...</p>;
  }
  if (error) {
    return <p className="text-ops-bad">Failed to load application details: {(error as Error).message}</p>;
  }
  if (!data) {
    return <p className="text-ops-muted">Application not found.</p>;
  }
  const suppressions = data.suppressions ?? [];

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">{data.application.name}</h1>
          <p className="mt-1 text-sm text-ops-muted">
            Namespace: <span className="text-ops-text">{data.application.namespace}</span>
          </p>
        </div>
        <button
          type="button"
          className="rounded-md bg-ops-accent px-4 py-2 text-sm font-semibold text-[#09111a] hover:opacity-90 disabled:opacity-60"
          onClick={() => analyzeMutation.mutate()}
          disabled={analyzeMutation.isPending}
        >
          {analyzeMutation.isPending ? "Queueing Analysis..." : "Run Analyze"}
        </button>
      </div>

      <div className="rounded-xl border border-ops-border bg-ops-panel p-4 shadow-panel">
        <div className="flex flex-wrap gap-2 border-b border-ops-border pb-3">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              className={`rounded-md px-3 py-1.5 text-sm ${
                activeTab === tab.id ? "bg-[#1a2636] text-ops-text" : "text-ops-muted hover:bg-[#1a2636]"
              }`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {activeTab === "incidents" ? (
          <div className="mt-4 overflow-x-auto rounded-lg border border-ops-border">
            <table className="min-w-[720px] w-full divide-y divide-ops-border text-sm">
              <thead className="bg-[#111a26] text-left text-xs uppercase tracking-wide text-ops-muted">
                <tr>
                  <th className="px-4 py-3">Incident</th>
                  <th className="px-4 py-3">Category</th>
                  <th className="px-4 py-3">Severity</th>
                  <th className="px-4 py-3">Recommendation</th>
                  <th className="px-4 py-3 text-right">Detail</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-ops-border">
                {data.incidents.length === 0 ? (
                  <tr>
                    <td className="px-4 py-6 text-ops-muted" colSpan={5}>
                      No incidents recorded.
                    </td>
                  </tr>
                ) : (
                  data.incidents.map((incident) => (
                    <tr key={incident.id} className="hover:bg-[#162132]">
                      <td className="px-4 py-3 font-medium">{incident.title}</td>
                      <td className="px-4 py-3 text-ops-muted">{incident.category}</td>
                      <td className="px-4 py-3">
                        <Badge label={incident.severity} tone={severityTone(incident.severity)} />
                      </td>
                      <td className="px-4 py-3">
                        <Badge label={incident.recommended_action} tone={recommendationTone(incident.recommended_action)} />
                      </td>
                      <td className="px-4 py-3 text-right">
                        <Link to={`/incidents/${incident.id}`} className="font-medium text-ops-accent hover:underline">
                          View
                        </Link>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="mt-4 overflow-x-auto rounded-lg border border-ops-border">
            <table className="min-w-[760px] w-full divide-y divide-ops-border text-sm">
              <thead className="bg-[#111a26] text-left text-xs uppercase tracking-wide text-ops-muted">
                <tr>
                  <th className="px-4 py-3">Field</th>
                  <th className="px-4 py-3">Resource</th>
                  <th className="px-4 py-3">Rule</th>
                  <th className="px-4 py-3">Suppressed</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-ops-border">
                {suppressions.length === 0 ? (
                  <tr>
                    <td className="px-4 py-6 text-ops-muted" colSpan={4}>
                      No suppressed findings recorded.
                    </td>
                  </tr>
                ) : (
                  suppressions.map((suppression) => (
                    <tr key={suppression.id} className="hover:bg-[#162132]">
                      <td className="px-4 py-3">
                        <p className="font-mono text-xs text-ops-text">{suppression.field_path}</p>
                        <p className="mt-1 text-xs text-ops-muted">{suppression.title}</p>
                      </td>
                      <td className="px-4 py-3 font-mono text-xs text-ops-muted">{suppression.resource_ref}</td>
                      <td className="px-4 py-3">
                        <p>{suppression.ignore_rule_name}</p>
                        <p className="mt-1 text-xs text-ops-muted">{suppression.ignore_rule_reason}</p>
                      </td>
                      <td className="whitespace-nowrap px-4 py-3 text-ops-muted">
                        {new Date(suppression.suppressed_at).toLocaleString()}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
}
