import { useState } from "react";
import type { Dispatch, SetStateAction } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { UseMutationResult } from "@tanstack/react-query";
import { ArrowLeft, ArrowRight, Play, Plus } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { ActionButton, EmptyState, ErrorState, LoadState, PageHeader, Panel } from "../components/Primitives";
import { analyzeApplication, createIgnoreRule, getApplication, setIgnoreRuleActive, type CreateIgnoreRuleInput } from "../lib/api";
import { recommendationTone, severityTone } from "../lib/badgeTones";
import { formatDate } from "../lib/format";

const tabs = [
  { id: "incidents", label: "Incidents" },
  { id: "suppressions", label: "Suppressed" },
  { id: "ignore-rules", label: "Ignore rules" }
] as const;

type ApplicationTab = (typeof tabs)[number]["id"];

const emptyRuleDraft: CreateIgnoreRuleInput = {
  name: "",
  match_expression: "",
  resource_ref: "",
  reason: ""
};

const inputClass =
  "mt-1.5 block h-9 w-full rounded-md border border-ops-border bg-ops-bg px-3 text-sm text-ops-text placeholder:text-ops-muted focus:border-ops-accent";

export function ApplicationDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id ?? "";
  const [activeTab, setActiveTab] = useState<ApplicationTab>("incidents");
  const [ruleDraft, setRuleDraft] = useState<CreateIgnoreRuleInput>(emptyRuleDraft);
  const queryClient = useQueryClient();
  const { data, isLoading, error, refetch } = useQuery({
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
    return <LoadState>Loading application details...</LoadState>;
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
        Failed to load application details: {(error as Error).message}
      </ErrorState>
    );
  }
  if (!data) {
    return <EmptyState>Application not found.</EmptyState>;
  }

  const suppressions = data.suppressions ?? [];
  const ignoreRules = data.ignore_rules ?? [];
  const tabCounts: Record<ApplicationTab, number> = {
    incidents: data.incidents.length,
    suppressions: suppressions.length,
    "ignore-rules": ignoreRules.length
  };

  return (
    <section className="space-y-6">
      <Link to="/applications" className="touch-target inline-flex items-center gap-2 text-sm text-ops-muted hover:text-ops-text">
        <ArrowLeft aria-hidden="true" className="h-4 w-4" />
        Applications
      </Link>

      <PageHeader
        label={`Application / ${data.application.namespace}`}
        title={data.application.name}
        actions={
          <ActionButton
            icon={Play}
            variant="primary"
            onClick={() => analyzeMutation.mutate()}
            disabled={analyzeMutation.isPending}
            type="button"
          >
            {analyzeMutation.isPending ? "Queueing" : "Analyze"}
          </ActionButton>
        }
      >
        <Badge label={data.application.status} tone={data.application.status === "active" ? "good" : "warn"} />
        <span>Updated {formatDate(data.application.updated_at)}</span>
      </PageHeader>

      {analyzeMutation.error ? <ErrorState>Unable to queue analysis: {(analyzeMutation.error as Error).message}</ErrorState> : null}
      {analyzeMutation.isSuccess ? <p className="text-sm text-ops-good">Analysis job queued.</p> : null}

      <Panel>
        <div aria-label="Application views" className="flex gap-1 overflow-x-auto border-b border-ops-border px-2 py-2">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              aria-pressed={activeTab === tab.id}
              className={`touch-target inline-flex h-10 shrink-0 items-center gap-2 rounded-md border px-3 text-sm transition-colors ${
                activeTab === tab.id
                  ? "border-ops-border bg-ops-bg font-medium text-ops-text"
                  : "border-transparent text-ops-muted hover:bg-ops-panel-alt hover:text-ops-text"
              }`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
              <span className="rounded-md bg-ops-bg px-2 py-0.5 font-mono text-xs text-ops-muted">{tabCounts[tab.id]}</span>
            </button>
          ))}
        </div>

        {activeTab === "incidents" ? <IncidentList incidents={data.incidents} /> : null}
        {activeTab === "suppressions" ? <SuppressionList suppressions={suppressions} /> : null}
        {activeTab === "ignore-rules" ? (
          <IgnoreRules
            applicationId={data.application.id}
            rules={ignoreRules}
            draft={ruleDraft}
            setDraft={setRuleDraft}
            createRuleMutation={createRuleMutation}
            ruleStatusMutation={ruleStatusMutation}
          />
        ) : null}
      </Panel>
    </section>
  );
}

function IncidentList({ incidents }: { incidents: Awaited<ReturnType<typeof getApplication>>["incidents"] }) {
  if (incidents.length === 0) {
    return <EmptyState>No incidents recorded for this application.</EmptyState>;
  }

  return (
    <>
      <div className="hidden overflow-x-auto md:block">
        <table className="w-full text-sm">
          <thead className="bg-ops-panel-alt text-left text-xs font-medium text-ops-muted">
            <tr>
              <th className="px-5 py-2.5">Incident</th>
              <th className="px-5 py-2.5">Category</th>
              <th className="px-5 py-2.5">Severity</th>
              <th className="px-5 py-2.5">Recommendation</th>
              <th className="px-5 py-2.5 text-right">Open</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-ops-border-muted">
            {incidents.map((incident) => (
              <tr key={incident.id} className="hover:bg-ops-panel-alt">
                <td className="max-w-lg px-5 py-3">
                  <p className="break-words font-medium">{incident.title}</p>
                  <p className="mt-1 text-xs text-ops-muted">{formatDate(incident.created_at)}</p>
                </td>
                <td className="px-5 py-3 text-ops-muted">{incident.category}</td>
                <td className="px-5 py-3">
                  <Badge label={incident.severity} tone={severityTone(incident.severity)} />
                </td>
                <td className="px-5 py-3">
                  <Badge label={incident.recommended_action} tone={recommendationTone(incident.recommended_action)} />
                </td>
                <td className="px-5 py-3 text-right">
                  <Link aria-label={`Open ${incident.title}`} className="touch-target inline-flex h-9 w-9 items-center justify-center rounded-md text-ops-muted hover:bg-ops-shell-hover hover:text-ops-text" to={`/incidents/${incident.id}`}>
                    <ArrowRight aria-hidden="true" className="h-4 w-4" />
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="divide-y divide-ops-border-muted md:hidden">
        {incidents.map((incident) => (
          <article className="space-y-3 px-4 py-4" key={incident.id}>
            <p className="break-words text-sm font-medium">{incident.title}</p>
            <div className="flex flex-wrap gap-2">
              <Badge label={incident.severity} tone={severityTone(incident.severity)} />
              <Badge label={incident.recommended_action} tone={recommendationTone(incident.recommended_action)} />
            </div>
            <Link className="touch-target inline-flex items-center gap-2 text-sm font-medium text-ops-accent" to={`/incidents/${incident.id}`}>
              Open incident
              <ArrowRight aria-hidden="true" className="h-4 w-4" />
            </Link>
          </article>
        ))}
      </div>
    </>
  );
}

function SuppressionList({ suppressions }: { suppressions: Awaited<ReturnType<typeof getApplication>>["suppressions"] }) {
  if (suppressions.length === 0) {
    return <EmptyState>No suppressed findings recorded for this application.</EmptyState>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="min-w-[720px] w-full text-sm">
        <thead className="bg-ops-panel-alt text-left text-xs font-medium text-ops-muted">
          <tr>
            <th className="px-5 py-2.5">Field</th>
            <th className="px-5 py-2.5">Resource</th>
            <th className="px-5 py-2.5">Matched rule</th>
            <th className="px-5 py-2.5">Suppressed</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-ops-border-muted">
          {suppressions.map((suppression) => (
            <tr key={suppression.id} className="align-top hover:bg-ops-panel-alt">
              <td className="px-5 py-3">
                <p className="break-all font-mono text-xs text-ops-text">{suppression.field_path}</p>
                <p className="mt-1 max-w-sm break-words text-xs text-ops-muted">{suppression.title}</p>
              </td>
              <td className="max-w-sm px-5 py-3 break-all font-mono text-xs text-ops-muted">{suppression.resource_ref}</td>
              <td className="px-5 py-3">
                <p>{suppression.ignore_rule_name}</p>
                <p className="mt-1 max-w-sm break-words text-xs text-ops-muted">{suppression.ignore_rule_reason}</p>
              </td>
              <td className="whitespace-nowrap px-5 py-3 text-ops-muted">{formatDate(suppression.suppressed_at)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function IgnoreRules({
  applicationId,
  rules,
  draft,
  setDraft,
  createRuleMutation,
  ruleStatusMutation
}: {
  applicationId: string;
  rules: Awaited<ReturnType<typeof getApplication>>["ignore_rules"];
  draft: CreateIgnoreRuleInput;
  setDraft: Dispatch<SetStateAction<CreateIgnoreRuleInput>>;
  createRuleMutation: UseMutationResult<Awaited<ReturnType<typeof createIgnoreRule>>, Error, CreateIgnoreRuleInput>;
  ruleStatusMutation: UseMutationResult<Awaited<ReturnType<typeof setIgnoreRuleActive>>, Error, { ruleId: string; active: boolean }>;
}) {
  return (
    <div>
      <form
        className="border-b border-ops-border bg-ops-panel-alt px-4 py-4 sm:px-5"
        onSubmit={(event) => {
          event.preventDefault();
          createRuleMutation.mutate(draft);
        }}
      >
        <div className="mb-4 flex items-center gap-2">
          <Plus aria-hidden="true" className="h-4 w-4 text-ops-muted" />
          <h2 className="text-sm font-semibold">New application rule</h2>
        </div>
        <div className="grid gap-4 lg:grid-cols-4">
          <RuleInput
            id="rule-name"
            label="Name"
            required
            value={draft.name}
            onChange={(value) => setDraft((current) => ({ ...current, name: value }))}
          />
          <RuleInput
            id="rule-field"
            label="Field path"
            placeholder="spec.replicas"
            required
            value={draft.match_expression}
            onChange={(value) => setDraft((current) => ({ ...current, match_expression: value }))}
          />
          <RuleInput
            id="rule-resource"
            label="Resource reference"
            placeholder="apps/v1/Deployment:payments/ledger-api"
            value={draft.resource_ref}
            onChange={(value) => setDraft((current) => ({ ...current, resource_ref: value }))}
          />
          <RuleInput
            id="rule-reason"
            label="Reason"
            required
            value={draft.reason}
            onChange={(value) => setDraft((current) => ({ ...current, reason: value }))}
          />
        </div>
        <div className="mt-4 flex flex-wrap items-center gap-3">
          <ActionButton icon={Plus} variant="primary" disabled={createRuleMutation.isPending} type="submit">
            {createRuleMutation.isPending ? "Creating" : "Create rule"}
          </ActionButton>
          {createRuleMutation.error ? <p className="text-sm text-ops-bad">{createRuleMutation.error.message}</p> : null}
        </div>
      </form>

      {rules.length === 0 ? (
        <EmptyState>No ignore rules configured for this application.</EmptyState>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-[760px] w-full text-sm">
            <thead className="bg-ops-panel-alt text-left text-xs font-medium text-ops-muted">
              <tr>
                <th className="px-5 py-2.5">Rule</th>
                <th className="px-5 py-2.5">Match</th>
                <th className="px-5 py-2.5">Scope</th>
                <th className="px-5 py-2.5">Status</th>
                <th className="px-5 py-2.5 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-ops-border-muted">
              {rules.map((rule) => (
                <tr key={rule.id} className="align-top hover:bg-ops-panel-alt">
                  <td className="px-5 py-3">
                    <p className="break-words font-medium">{rule.name}</p>
                    <p className="mt-1 max-w-xs break-words text-xs text-ops-muted">{rule.reason}</p>
                  </td>
                  <td className="px-5 py-3 break-all font-mono text-xs text-ops-text">{rule.match_expression}</td>
                  <td className="px-5 py-3">
                    <p>{rule.application_id ? "Application" : "Workspace (inherited)"}</p>
                    {rule.resource_ref ? <p className="mt-1 max-w-sm break-all font-mono text-xs text-ops-muted">{rule.resource_ref}</p> : null}
                  </td>
                  <td className="px-5 py-3">
                    <Badge label={rule.active ? "Active" : "Inactive"} tone={rule.active ? "good" : "neutral"} />
                  </td>
                  <td className="px-5 py-3 text-right">
                    {rule.application_id === applicationId ? (
                      <ActionButton
                        variant="quiet"
                        disabled={ruleStatusMutation.isPending}
                        onClick={() => ruleStatusMutation.mutate({ ruleId: rule.id, active: !rule.active })}
                        type="button"
                      >
                        {rule.active ? "Disable" : "Enable"}
                      </ActionButton>
                    ) : (
                      <span className="text-xs text-ops-muted">Read only</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {ruleStatusMutation.error ? <p className="px-5 py-3 text-sm text-ops-bad">{ruleStatusMutation.error.message}</p> : null}
    </div>
  );
}

function RuleInput({
  id,
  label,
  value,
  placeholder,
  required = false,
  onChange
}: {
  id: string;
  label: string;
  value: string;
  placeholder?: string;
  required?: boolean;
  onChange: (value: string) => void;
}) {
  return (
    <label className="min-w-0 text-sm text-ops-muted" htmlFor={id}>
      {label}
      <input
        id={id}
        className={inputClass}
        placeholder={placeholder}
        required={required}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </label>
  );
}
