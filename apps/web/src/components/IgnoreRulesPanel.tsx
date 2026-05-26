import { Fragment, useState } from "react";
import type { Dispatch, FormEvent, SetStateAction } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { UseMutationResult } from "@tanstack/react-query";
import { Check, Pencil, Plus, Trash2, X } from "lucide-react";
import { createIgnoreRule, deleteIgnoreRule, setIgnoreRuleActive, updateIgnoreRule } from "../lib/api";
import type { IgnoreRule, IgnoreRuleInput } from "../lib/api";
import { Badge } from "./Badge";
import { ActionButton, EmptyState } from "./Primitives";

const emptyDraft: IgnoreRuleInput = {
  name: "",
  match_expression: "",
  resource_ref: "",
  reason: ""
};

const inputClass =
  "console-copy mt-1.5 block h-8 w-full border border-ops-border bg-ops-bg px-2.5 text-ops-text placeholder:text-ops-dim focus:border-ops-accent";

export function IgnoreRulesPanel({ applicationId, rules }: { applicationId: string; rules: IgnoreRule[] }) {
  const queryClient = useQueryClient();
  const [draft, setDraft] = useState<IgnoreRuleInput>(emptyDraft);
  const [editingRuleId, setEditingRuleId] = useState<string | null>(null);
  const [editDraft, setEditDraft] = useState<IgnoreRuleInput>(emptyDraft);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  const createMutation = useMutation({
    mutationFn: (input: IgnoreRuleInput) => createIgnoreRule(applicationId, input),
    onSuccess: async () => {
      setDraft(emptyDraft);
      await refreshApplication();
    }
  });
  const statusMutation = useMutation({
    mutationFn: ({ ruleId, active }: { ruleId: string; active: boolean }) => setIgnoreRuleActive(applicationId, ruleId, active),
    onSuccess: refreshApplication
  });
  const updateMutation = useMutation({
    mutationFn: ({ ruleId, input }: { ruleId: string; input: IgnoreRuleInput }) => updateIgnoreRule(applicationId, ruleId, input),
    onSuccess: refreshApplication
  });
  const deleteMutation = useMutation({
    mutationFn: (ruleId: string) => deleteIgnoreRule(applicationId, ruleId),
    onSuccess: refreshApplication
  });
  const commandPending = statusMutation.isPending || updateMutation.isPending || deleteMutation.isPending;

  function refreshApplication() {
    return queryClient.invalidateQueries({ queryKey: ["application", applicationId] });
  }

  function beginEdit(rule: IgnoreRule) {
    setEditingRuleId(rule.id);
    setEditDraft({
      name: rule.name,
      match_expression: rule.match_expression,
      resource_ref: rule.resource_ref,
      reason: rule.reason
    });
    setConfirmDeleteId(null);
    statusMutation.reset();
    deleteMutation.reset();
    updateMutation.reset();
  }

  function cancelEdit() {
    setEditingRuleId(null);
    setEditDraft(emptyDraft);
    updateMutation.reset();
  }

  function submitEdit(event: FormEvent<HTMLFormElement>, ruleId: string) {
    event.preventDefault();
    updateMutation.mutate(
      { ruleId, input: editDraft },
      {
        onSuccess: () => {
          setEditingRuleId(null);
          setEditDraft(emptyDraft);
        }
      }
    );
  }

  function confirmDelete(ruleId: string) {
    deleteMutation.mutate(ruleId, {
      onSuccess: () => {
        setConfirmDeleteId(null);
        if (editingRuleId === ruleId) {
          setEditingRuleId(null);
        }
      }
    });
  }

  const actionProps = {
    applicationId,
    commandsDisabled: commandPending,
    confirmDeleteId,
    deleteMutation,
    onBeginEdit: beginEdit,
    onCancelDelete: () => setConfirmDeleteId(null),
    onConfirmDelete: confirmDelete,
    onRequestDelete: (ruleId: string) => {
      deleteMutation.reset();
      setConfirmDeleteId(ruleId);
    },
    statusMutation
  };

  return (
    <div>
      <form
        className="border-b border-ops-border-muted bg-ops-panel-alt px-3 py-3"
        onSubmit={(event) => {
          event.preventDefault();
          createMutation.reset();
          createMutation.mutate(draft);
        }}
      >
        <div className="mb-4 flex items-center gap-2">
          <Plus aria-hidden="true" className="h-4 w-4 text-ops-muted" />
          <h2 className="console-label text-ops-dim">New application rule</h2>
        </div>
        <RuleFormFields draft={draft} idPrefix="new-rule" setDraft={setDraft} />
        <div className="mt-4 flex flex-wrap items-center gap-3">
          <ActionButton aria-busy={createMutation.isPending} disabled={createMutation.isPending} icon={Plus} type="submit" variant="primary">
            {createMutation.isPending ? "Creating" : "Create rule"}
          </ActionButton>
          {createMutation.error ? <MutationError message={createMutation.error.message} /> : null}
          {createMutation.isSuccess ? <MutationStatus message="Application rule created." /> : null}
        </div>
      </form>

      {rules.length === 0 ? (
        <EmptyState>No ignore rules configured for this application.</EmptyState>
      ) : (
        <>
          <div className="hidden overflow-x-auto md:block">
            <table className="console-copy min-w-[1120px] table-fixed w-full">
              <colgroup>
                <col className="w-[24%]" />
                <col className="w-[17%]" />
                <col className="w-[25%]" />
                <col className="w-[10%]" />
                <col className="w-[24%]" />
              </colgroup>
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
                {rules.map((rule) =>
                  editingRuleId === rule.id ? (
                    <tr key={rule.id}>
                      <td className="bg-ops-panel-alt px-3 py-3" colSpan={5}>
                        <RuleEditor
                          draft={editDraft}
                          error={updateMutation.error?.message}
                          idPrefix={`desktop-edit-rule-${rule.id}`}
                          isPending={updateMutation.isPending}
                          onCancel={cancelEdit}
                          onSubmit={(event) => submitEdit(event, rule.id)}
                          setDraft={setEditDraft}
                        />
                      </td>
                    </tr>
                  ) : (
                    <Fragment key={rule.id}>
                      <tr className="align-top hover:bg-ops-panel-alt">
                        <td className="px-3 py-2.5">
                          <p className="break-words font-medium">{rule.name}</p>
                          <p className="mt-1 max-w-xs break-words text-ops-muted">{rule.reason}</p>
                        </td>
                        <td className="px-3 py-2.5 break-all font-mono text-ops-text">{rule.match_expression}</td>
                        <td className="px-3 py-2.5">
                          <p>{scopeLabel(rule)}</p>
                          {rule.resource_ref ? <p className="mt-1 max-w-sm break-all font-mono text-ops-muted">{rule.resource_ref}</p> : null}
                        </td>
                        <td className="px-3 py-2.5">
                          <RuleStatus rule={rule} />
                        </td>
                        <td className="px-3 py-2.5 text-right">
                          <RuleActions {...actionProps} inlineConfirmation={false} rule={rule} />
                        </td>
                      </tr>
                      {confirmDeleteId === rule.id ? (
                        <tr>
                          <td className="bg-ops-panel-alt px-3 py-2.5" colSpan={5}>
                            <DeleteConfirmation commandsDisabled={commandPending} deleteMutation={deleteMutation} onCancel={actionProps.onCancelDelete} onConfirm={confirmDelete} rule={rule} wide />
                          </td>
                        </tr>
                      ) : null}
                    </Fragment>
                  )
                )}
              </tbody>
            </table>
          </div>

          <div className="divide-y divide-ops-border-muted md:hidden">
            {rules.map((rule) =>
              editingRuleId === rule.id ? (
                <RuleEditor
                  key={rule.id}
                  className="bg-ops-panel-alt px-3 py-4"
                  draft={editDraft}
                  error={updateMutation.error?.message}
                  idPrefix={`mobile-edit-rule-${rule.id}`}
                  isPending={updateMutation.isPending}
                  onCancel={cancelEdit}
                  onSubmit={(event) => submitEdit(event, rule.id)}
                  setDraft={setEditDraft}
                />
              ) : (
                <article className="space-y-3 px-3 py-4" key={rule.id}>
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p className="console-copy break-words font-medium text-ops-text">{rule.name}</p>
                      <p className="console-copy mt-1 break-words text-ops-muted">{rule.reason}</p>
                    </div>
                    <RuleStatus rule={rule} />
                  </div>
                  <dl className="console-copy space-y-2">
                    <div>
                      <dt className="console-label text-ops-dim">Match</dt>
                      <dd className="mt-1 break-all font-mono text-ops-text">{rule.match_expression}</dd>
                    </div>
                    <div>
                      <dt className="console-label text-ops-dim">Scope</dt>
                      <dd className="mt-1 text-ops-muted">{scopeLabel(rule)}</dd>
                      {rule.resource_ref ? <dd className="mt-1 break-all font-mono text-ops-muted">{rule.resource_ref}</dd> : null}
                    </div>
                  </dl>
                  <RuleActions {...actionProps} rule={rule} />
                </article>
              )
            )}
          </div>
        </>
      )}

      {statusMutation.error ? <MutationError message={statusMutation.error.message} padded /> : null}
      {statusMutation.isSuccess ? <MutationStatus message="Rule status updated." padded /> : null}
      {updateMutation.isSuccess ? <MutationStatus message="Rule changes saved." padded /> : null}
      {deleteMutation.error ? <MutationError message={deleteMutation.error.message} padded /> : null}
      {deleteMutation.isSuccess ? <MutationStatus message="Rule deleted. Existing suppression audit records remain available." padded /> : null}
    </div>
  );
}

function RuleEditor({
  className = "",
  draft,
  error,
  idPrefix,
  isPending,
  onCancel,
  onSubmit,
  setDraft
}: {
  className?: string;
  draft: IgnoreRuleInput;
  error?: string;
  idPrefix: string;
  isPending: boolean;
  onCancel: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  setDraft: Dispatch<SetStateAction<IgnoreRuleInput>>;
}) {
  return (
    <form className={className} onSubmit={onSubmit}>
      <div className="mb-4 flex items-center gap-2">
        <Pencil aria-hidden="true" className="h-4 w-4 text-ops-muted" />
        <h3 className="console-label text-ops-dim">Edit application rule</h3>
      </div>
      <RuleFormFields draft={draft} idPrefix={idPrefix} setDraft={setDraft} />
      <div className="mt-4 flex flex-wrap items-center gap-2">
        <ActionButton aria-busy={isPending} disabled={isPending} icon={Check} type="submit" variant="primary">
          {isPending ? "Saving" : "Save changes"}
        </ActionButton>
        <ActionButton disabled={isPending} icon={X} onClick={onCancel} type="button">
          Cancel
        </ActionButton>
        {error ? <MutationError message={error} /> : null}
      </div>
    </form>
  );
}

function RuleActions({
  applicationId,
  commandsDisabled,
  confirmDeleteId,
  deleteMutation,
  inlineConfirmation = true,
  onBeginEdit,
  onCancelDelete,
  onConfirmDelete,
  onRequestDelete,
  rule,
  statusMutation
}: {
  applicationId: string;
  commandsDisabled: boolean;
  confirmDeleteId: string | null;
  deleteMutation: UseMutationResult<Awaited<ReturnType<typeof deleteIgnoreRule>>, Error, string>;
  inlineConfirmation?: boolean;
  onBeginEdit: (rule: IgnoreRule) => void;
  onCancelDelete: () => void;
  onConfirmDelete: (ruleId: string) => void;
  onRequestDelete: (ruleId: string) => void;
  rule: IgnoreRule;
  statusMutation: UseMutationResult<Awaited<ReturnType<typeof setIgnoreRuleActive>>, Error, { ruleId: string; active: boolean }>;
}) {
  if (rule.application_id !== applicationId) {
    return <span className="console-copy text-ops-muted">Read only</span>;
  }

  const statusPending = statusMutation.isPending && statusMutation.variables?.ruleId === rule.id;
  if (confirmDeleteId === rule.id && inlineConfirmation) {
    return <DeleteConfirmation commandsDisabled={commandsDisabled} deleteMutation={deleteMutation} onCancel={onCancelDelete} onConfirm={onConfirmDelete} rule={rule} />;
  }

  const controlsDisabled = commandsDisabled || confirmDeleteId === rule.id;
  return (
    <div className="flex flex-wrap justify-end gap-1">
      <ActionButton aria-label={`Edit ${rule.name}`} disabled={controlsDisabled} icon={Pencil} onClick={() => onBeginEdit(rule)} type="button" variant="quiet">
        Edit
      </ActionButton>
      <ActionButton
        aria-busy={statusPending}
        aria-label={`${rule.active ? "Disable" : "Enable"} ${rule.name}`}
        disabled={controlsDisabled}
        onClick={() => {
          statusMutation.reset();
          statusMutation.mutate({ ruleId: rule.id, active: !rule.active });
        }}
        type="button"
        variant="quiet"
      >
        {rule.active ? "Disable" : "Enable"}
      </ActionButton>
      <ActionButton aria-label={`Delete ${rule.name}`} disabled={controlsDisabled} icon={Trash2} onClick={() => onRequestDelete(rule.id)} type="button" variant="dangerQuiet">
        Delete
      </ActionButton>
    </div>
  );
}

function DeleteConfirmation({
  commandsDisabled,
  deleteMutation,
  onCancel,
  onConfirm,
  rule,
  wide = false
}: {
  commandsDisabled: boolean;
  deleteMutation: UseMutationResult<Awaited<ReturnType<typeof deleteIgnoreRule>>, Error, string>;
  onCancel: () => void;
  onConfirm: (ruleId: string) => void;
  rule: IgnoreRule;
  wide?: boolean;
}) {
  const deletionPending = deleteMutation.isPending && deleteMutation.variables === rule.id;
  return (
    <div className={`console-copy flex gap-2 ${wide ? "flex-wrap items-center justify-end" : "flex-col items-end"}`}>
      <p className="max-w-sm text-left text-ops-bad">Delete this rule? Recorded suppressions remain in audit history.</p>
      <div className="flex flex-wrap justify-end gap-1">
        <ActionButton aria-busy={deletionPending} disabled={commandsDisabled} icon={Trash2} onClick={() => onConfirm(rule.id)} type="button" variant="danger">
          {deletionPending ? "Deleting" : "Confirm delete"}
        </ActionButton>
        <ActionButton disabled={commandsDisabled} icon={X} onClick={onCancel} type="button">
          Cancel
        </ActionButton>
      </div>
    </div>
  );
}

function RuleFormFields({
  draft,
  idPrefix,
  setDraft
}: {
  draft: IgnoreRuleInput;
  idPrefix: string;
  setDraft: Dispatch<SetStateAction<IgnoreRuleInput>>;
}) {
  return (
    <div className="grid gap-4 lg:grid-cols-4">
      <RuleInput id={`${idPrefix}-name`} label="Name" required value={draft.name} onChange={(value) => setDraft((current) => ({ ...current, name: value }))} />
      <RuleInput
        id={`${idPrefix}-field`}
        label="Field path"
        placeholder="spec.replicas"
        required
        value={draft.match_expression}
        onChange={(value) => setDraft((current) => ({ ...current, match_expression: value }))}
      />
      <RuleInput
        id={`${idPrefix}-resource`}
        label="Resource reference"
        placeholder="apps/v1/Deployment:payments/ledger-api"
        value={draft.resource_ref}
        onChange={(value) => setDraft((current) => ({ ...current, resource_ref: value }))}
      />
      <RuleInput id={`${idPrefix}-reason`} label="Reason" required value={draft.reason} onChange={(value) => setDraft((current) => ({ ...current, reason: value }))} />
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
        className={inputClass}
        id={id}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        required={required}
        value={value}
      />
    </label>
  );
}

function RuleStatus({ rule }: { rule: IgnoreRule }) {
  return <Badge label={rule.active ? "Active" : "Inactive"} tone={rule.active ? "good" : "neutral"} />;
}

function scopeLabel(rule: IgnoreRule) {
  return rule.application_id ? "Application" : "Workspace (inherited)";
}

function MutationError({ message, padded = false }: { message: string; padded?: boolean }) {
  return (
    <p className={`console-copy text-ops-bad ${padded ? "px-3 py-2.5" : ""}`} role="alert">
      {message}
    </p>
  );
}

function MutationStatus({ message, padded = false }: { message: string; padded?: boolean }) {
  return (
    <p className={`console-copy text-ops-good ${padded ? "px-3 py-2.5" : ""}`} role="status">
      {message}
    </p>
  );
}
