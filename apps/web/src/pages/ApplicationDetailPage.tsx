import { useRef, useState } from "react";
import type { Dispatch, KeyboardEvent, SetStateAction } from "react";
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
  "console-copy mt-1.5 block h-8 w-full border border-ops-border bg-ops-bg px-2.5 text-ops-text placeholder:text-ops-dim focus:border-ops-accent";

export function ApplicationDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id ?? "";
  const [activeTab, setActiveTab] = useState<ApplicationTab>("incidents");
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([]);
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

  function handleTabKeyDown(event: KeyboardEvent<HTMLButtonElement>, currentIndex: number) {
    let nextIndex: number | undefined;
    switch (event.key) {
      case "ArrowRight":
        nextIndex = (currentIndex + 1) % tabs.length;
        break;
      case "ArrowLeft":
        nextIndex = (currentIndex + tabs.length - 1) % tabs.length;
        break;
      case "Home":
        nextIndex = 0;
        break;
      case "End":
        nextIndex = tabs.length - 1;
        break;
      default:
        return;
    }
    event.preventDefault();
    setActiveTab(tabs[nextIndex].id);
    tabRefs.current[nextIndex]?.focus();
  }

  return (
    <section className="space-y-4">
      <Link to="/applications" className="console-action console-transition touch-target inline-flex items-center gap-2 text-ops-muted hover:text-ops-text active:text-ops-text">
        <ArrowLeft aria-hidden="true" className="h-4 w-4" />
        Applications
      </Link>

      <PageHeader
        label={`Application / ${data.application.namespace}`}
        title={data.application.name}
        actions={
          <ActionButton
            aria-busy={analyzeMutation.isPending}
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
      {analyzeMutation.isSuccess ? (
        <p className="console-action text-ops-good" role="status">
          Analysis job queued.
        </p>
      ) : null}

      <Panel>
        <div aria-label="Application views" className="grid grid-cols-3 gap-1 border-b border-ops-border-muted px-2 py-2 sm:flex" role="tablist">
          {tabs.map((tab, index) => (
            <button
              key={tab.id}
              ref={(element) => {
                tabRefs.current[index] = element;
              }}
              id={`application-tab-${tab.id}`}
              type="button"
              aria-label={tab.label}
              aria-controls={`application-panel-${tab.id}`}
              aria-selected={activeTab === tab.id}
              className={`console-label console-transition touch-target inline-flex min-h-14 min-w-0 flex-col items-center justify-center gap-1 border px-1 py-1 sm:h-8 sm:min-h-0 sm:shrink-0 sm:flex-row sm:gap-1.5 sm:px-2 sm:py-0 ${
                activeTab === tab.id
                  ? "border-ops-action bg-ops-action font-medium text-ops-bg hover:bg-ops-action-hover active:bg-ops-action-hover"
                  : "border-transparent text-ops-muted hover:bg-ops-panel-alt hover:text-ops-text active:bg-ops-panel-alt"
              }`}
              onClick={() => setActiveTab(tab.id)}
              onKeyDown={(event) => handleTabKeyDown(event, index)}
              role="tab"
              tabIndex={activeTab === tab.id ? 0 : -1}
            >
              <span>{tab.label}</span>
              <span className={`px-1.5 py-0.5 font-mono text-sm ${activeTab === tab.id ? "bg-ops-accent-soft text-ops-text" : "bg-ops-panel-alt text-ops-muted"}`}>
                {tabCounts[tab.id]}
              </span>
            </button>
          ))}
        </div>

        <div aria-labelledby={`application-tab-${activeTab}`} id={`application-panel-${activeTab}`} role="tabpanel">
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
        </div>
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
        <table className="console-copy w-full">
          <thead className="console-label bg-ops-panel-alt text-left text-ops-dim">
            <tr>
              <th className="px-3 py-2 font-normal">Incident</th>
              <th className="px-3 py-2 font-normal">Category</th>
              <th className="px-3 py-2 font-normal">Severity</th>
              <th className="px-3 py-2 font-normal">Recommendation</th>
              <th className="px-3 py-2 text-right font-normal">Open</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-ops-border-muted">
            {incidents.map((incident) => (
              <tr key={incident.id} className="hover:bg-ops-panel-alt">
                <td className="max-w-lg px-3 py-2.5">
                  <p className="break-words font-medium">{incident.title}</p>
                  <p className="mt-1 text-ops-muted">{formatDate(incident.created_at)}</p>
                </td>
                <td className="px-3 py-2.5 text-ops-muted">{incident.category}</td>
                <td className="px-3 py-2.5">
                  <Badge label={incident.severity} tone={severityTone(incident.severity)} />
                </td>
                <td className="px-3 py-2.5">
                  <Badge label={incident.recommended_action} tone={recommendationTone(incident.recommended_action)} />
                </td>
                <td className="px-3 py-2.5 text-right">
                  <Link aria-label={`Open ${incident.title}`} className="touch-target inline-flex h-8 w-8 items-center justify-center text-ops-muted hover:bg-ops-shell-hover hover:text-ops-text" to={`/incidents/${incident.id}`}>
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
          <article className="space-y-3 px-4 py-4 text-sm" key={incident.id}>
            <p className="break-words font-medium">{incident.title}</p>
            <div className="flex flex-wrap gap-2">
              <Badge label={incident.severity} tone={severityTone(incident.severity)} />
              <Badge label={incident.recommended_action} tone={recommendationTone(incident.recommended_action)} />
            </div>
            <Link className="console-action console-transition touch-target inline-flex items-center gap-2 text-ops-accent hover:bg-ops-accent-soft active:bg-ops-accent-soft" to={`/incidents/${incident.id}`}>
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
      <table className="console-copy min-w-[720px] w-full">
        <thead className="console-label bg-ops-panel-alt text-left text-ops-dim">
          <tr>
            <th className="px-3 py-2 font-normal">Field</th>
            <th className="px-3 py-2 font-normal">Resource</th>
            <th className="px-3 py-2 font-normal">Matched rule</th>
            <th className="px-3 py-2 font-normal">Suppressed</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-ops-border-muted">
          {suppressions.map((suppression) => (
            <tr key={suppression.id} className="align-top hover:bg-ops-panel-alt">
              <td className="px-3 py-2.5">
                <p className="break-all font-mono text-ops-text">{suppression.field_path}</p>
                <p className="mt-1 max-w-sm break-words text-ops-muted">{suppression.title}</p>
              </td>
              <td className="max-w-sm px-3 py-2.5 break-all font-mono text-ops-muted">{suppression.resource_ref}</td>
              <td className="px-3 py-2.5">
                <p>{suppression.ignore_rule_name}</p>
                <p className="mt-1 max-w-sm break-words text-ops-muted">{suppression.ignore_rule_reason}</p>
              </td>
              <td className="whitespace-nowrap px-3 py-2.5 text-ops-muted">{formatDate(suppression.suppressed_at)}</td>
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
        className="border-b border-ops-border-muted bg-ops-panel-alt px-3 py-3"
        onSubmit={(event) => {
          event.preventDefault();
          createRuleMutation.mutate(draft);
        }}
      >
        <div className="mb-4 flex items-center gap-2">
          <Plus aria-hidden="true" className="h-4 w-4 text-ops-muted" />
          <h2 className="console-label text-ops-dim">New application rule</h2>
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
          <ActionButton aria-busy={createRuleMutation.isPending} icon={Plus} variant="primary" disabled={createRuleMutation.isPending} type="submit">
            {createRuleMutation.isPending ? "Creating" : "Create rule"}
          </ActionButton>
          {createRuleMutation.error ? <p className="text-sm text-ops-bad">{createRuleMutation.error.message}</p> : null}
        </div>
      </form>

      {rules.length === 0 ? (
        <EmptyState>No ignore rules configured for this application.</EmptyState>
      ) : (
        <div className="overflow-x-auto">
          <table className="console-copy min-w-[760px] w-full">
            <thead className="console-label bg-ops-panel-alt text-left text-ops-dim">
              <tr>
                <th className="px-3 py-2 font-normal">Rule</th>
                <th className="px-3 py-2 font-normal">Match</th>
                <th className="px-3 py-2 font-normal">Scope</th>
                <th className="px-3 py-2 font-normal">Status</th>
                <th className="px-3 py-2 text-right font-normal">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-ops-border-muted">
              {rules.map((rule) => (
                <tr key={rule.id} className="align-top hover:bg-ops-panel-alt">
                  <td className="px-3 py-2.5">
                    <p className="break-words font-medium">{rule.name}</p>
                    <p className="mt-1 max-w-xs break-words text-ops-muted">{rule.reason}</p>
                  </td>
                  <td className="px-3 py-2.5 break-all font-mono text-ops-text">{rule.match_expression}</td>
                  <td className="px-3 py-2.5">
                    <p>{rule.application_id ? "Application" : "Workspace (inherited)"}</p>
                    {rule.resource_ref ? <p className="mt-1 max-w-sm break-all font-mono text-ops-muted">{rule.resource_ref}</p> : null}
                  </td>
                  <td className="px-3 py-2.5">
                    <Badge label={rule.active ? "Active" : "Inactive"} tone={rule.active ? "good" : "neutral"} />
                  </td>
                  <td className="px-3 py-2.5 text-right">
                    {rule.application_id === applicationId ? (
                      <ActionButton
                        variant="quiet"
                        aria-busy={ruleStatusMutation.isPending && ruleStatusMutation.variables?.ruleId === rule.id}
                        disabled={ruleStatusMutation.isPending}
                        onClick={() => ruleStatusMutation.mutate({ ruleId: rule.id, active: !rule.active })}
                        type="button"
                      >
                        {rule.active ? "Disable" : "Enable"}
                      </ActionButton>
                    ) : (
                      <span className="text-ops-muted">Read only</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {ruleStatusMutation.error ? <p className="console-copy px-3 py-2.5 text-ops-bad">{ruleStatusMutation.error.message}</p> : null}
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
    <label className="console-label min-w-0 text-ops-muted" htmlFor={id}>
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
