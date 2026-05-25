import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { analyzeApplication, createIgnoreRule, getApplication, setIgnoreRuleActive, type CreateIgnoreRuleInput } from "../lib/api";
import { Badge } from "../components/Badge";
import { recommendationTone, severityTone } from "../lib/badgeTones";

const tabs = [
  { id: "incidents", label: "Incidents" },
  { id: "suppressions", label: "Suppressed" },
  { id: "ignore-rules", label: "Ignore Rules" }
] as const;

type ApplicationTab = (typeof tabs)[number]["id"];

const emptyRuleDraft: CreateIgnoreRuleInput = {
  name: "",
  match_expression: "",
  resource_ref: "",
  reason: ""
};

const inputClass =
  "mt-1 block w-full rounded-md border border-ops-border bg-[#0e1621] px-3 py-2 text-sm text-ops-text outline-none focus:border-ops-accent";

export function ApplicationDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id ?? "";
  const [activeTab, setActiveTab] = useState<ApplicationTab>("incidents");
  const [ruleDraft, setRuleDraft] = useState<CreateIgnoreRuleInput>(emptyRuleDraft);
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

  const createRuleMutation = useMutation({
    mutationFn: (input: CreateIgnoreRuleInput) => createIgnoreRule(id, input),
    onSuccess: async () => {
      setRuleDraft(emptyRuleDraft);
      await queryClient.invalidateQueries({ queryKey: ["application", id] });
    }
  });

  const ruleStatusMutation = useMutation({
    mutationFn: ({ ruleId, active }: { ruleId: string; active: boolean }) => setIgnoreRuleActive(id, ruleId, active),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["application", id] });
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
  const ignoreRules = data.ignore_rules ?? [];

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
        ) : activeTab === "suppressions" ? (
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
        ) : (
          <div className="mt-4">
            <form
              className="grid gap-4 border-b border-ops-border pb-5 md:grid-cols-2"
              onSubmit={(event) => {
                event.preventDefault();
                createRuleMutation.mutate(ruleDraft);
              }}
            >
              <div>
                <label className="text-sm text-ops-muted" htmlFor="rule-name">
                  Name
                </label>
                <input
                  id="rule-name"
                  className={inputClass}
                  required
                  value={ruleDraft.name}
                  onChange={(event) => setRuleDraft((draft) => ({ ...draft, name: event.target.value }))}
                />
              </div>
              <div>
                <label className="text-sm text-ops-muted" htmlFor="rule-field">
                  Field path
                </label>
                <input
                  id="rule-field"
                  className={inputClass}
                  required
                  placeholder="spec.replicas"
                  value={ruleDraft.match_expression}
                  onChange={(event) => setRuleDraft((draft) => ({ ...draft, match_expression: event.target.value }))}
                />
              </div>
              <div>
                <label className="text-sm text-ops-muted" htmlFor="rule-resource">
                  Resource reference
                </label>
                <input
                  id="rule-resource"
                  className={inputClass}
                  placeholder="apps/v1/Deployment:payments/ledger-api"
                  value={ruleDraft.resource_ref}
                  onChange={(event) => setRuleDraft((draft) => ({ ...draft, resource_ref: event.target.value }))}
                />
              </div>
              <div>
                <label className="text-sm text-ops-muted" htmlFor="rule-reason">
                  Reason
                </label>
                <input
                  id="rule-reason"
                  className={inputClass}
                  required
                  value={ruleDraft.reason}
                  onChange={(event) => setRuleDraft((draft) => ({ ...draft, reason: event.target.value }))}
                />
              </div>
              <div className="flex items-center gap-3 md:col-span-2">
                <button
                  type="submit"
                  className="rounded-md bg-ops-accent px-4 py-2 text-sm font-semibold text-[#09111a] hover:opacity-90 disabled:opacity-60"
                  disabled={createRuleMutation.isPending}
                >
                  {createRuleMutation.isPending ? "Creating..." : "Create Rule"}
                </button>
                {createRuleMutation.error ? (
                  <p className="text-sm text-ops-bad">{(createRuleMutation.error as Error).message}</p>
                ) : null}
              </div>
            </form>

            <div className="mt-5 overflow-x-auto rounded-lg border border-ops-border">
              <table className="min-w-[800px] w-full divide-y divide-ops-border text-sm">
                <thead className="bg-[#111a26] text-left text-xs uppercase tracking-wide text-ops-muted">
                  <tr>
                    <th className="px-4 py-3">Rule</th>
                    <th className="px-4 py-3">Match</th>
                    <th className="px-4 py-3">Scope</th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-ops-border">
                  {ignoreRules.length === 0 ? (
                    <tr>
                      <td className="px-4 py-6 text-ops-muted" colSpan={5}>
                        No ignore rules configured.
                      </td>
                    </tr>
                  ) : (
                    ignoreRules.map((rule) => (
                      <tr key={rule.id} className="hover:bg-[#162132]">
                        <td className="px-4 py-3">
                          <p className="font-medium">{rule.name}</p>
                          <p className="mt-1 text-xs text-ops-muted">{rule.reason}</p>
                        </td>
                        <td className="px-4 py-3 font-mono text-xs text-ops-text">{rule.match_expression}</td>
                        <td className="px-4 py-3">
                          <p>{rule.application_id ? "Application" : "Workspace (inherited)"}</p>
                          {rule.resource_ref ? <p className="mt-1 font-mono text-xs text-ops-muted">{rule.resource_ref}</p> : null}
                        </td>
                        <td className="px-4 py-3">
                          <Badge label={rule.active ? "Active" : "Inactive"} tone={rule.active ? "good" : "neutral"} />
                        </td>
                        <td className="px-4 py-3 text-right">
                          {rule.application_id === data.application.id ? (
                            <button
                              type="button"
                              className="font-medium text-ops-accent hover:underline disabled:text-ops-muted"
                              disabled={ruleStatusMutation.isPending}
                              onClick={() => ruleStatusMutation.mutate({ ruleId: rule.id, active: !rule.active })}
                            >
                              {rule.active ? "Disable" : "Enable"}
                            </button>
                          ) : (
                            <span className="text-ops-muted">Read only</span>
                          )}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
            {ruleStatusMutation.error ? (
              <p className="mt-3 text-sm text-ops-bad">{(ruleStatusMutation.error as Error).message}</p>
            ) : null}
          </div>
        )}
      </div>
    </section>
  );
}
