import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowRight, Play, Search } from "lucide-react";
import { Link } from "react-router-dom";
import { Badge } from "../components/Badge";
import { ActionButton, EmptyState, ErrorState, LoadState, PageHeader, Panel } from "../components/Primitives";
import { analyzeApplication, getApplications } from "../lib/api";
import { severityTone } from "../lib/badgeTones";
import { formatDate } from "../lib/format";

export function ApplicationsPage() {
  const [query, setQuery] = useState("");
  const queryClient = useQueryClient();
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["applications"],
    queryFn: getApplications
  });

  const analyzeMutation = useMutation({
    mutationFn: (id: string) => analyzeApplication(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["overview"] });
    }
  });

  const applications = useMemo(() => {
    const search = query.trim().toLowerCase();
    if (!data || search.length === 0) {
      return data ?? [];
    }
    return data.filter((app) => [app.name, app.namespace, app.status].some((value) => value.toLowerCase().includes(search)));
  }, [data, query]);

  if (isLoading) {
    return <LoadState>Loading applications...</LoadState>;
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
        Failed to load applications: {(error as Error).message}
      </ErrorState>
    );
  }

  return (
    <section className="space-y-4">
      <PageHeader label="Inventory" title="Applications">
        <span>{data?.length ?? 0} tracked workloads</span>
      </PageHeader>

      {analyzeMutation.error ? <ErrorState>Unable to queue analysis: {(analyzeMutation.error as Error).message}</ErrorState> : null}

      <Panel>
        <div className="flex flex-wrap items-center justify-end gap-4 border-b border-ops-border-muted px-3 py-2">
          <label className="relative block w-full sm:w-72">
            <span className="sr-only">Filter applications</span>
            <Search aria-hidden="true" className="absolute left-2.5 top-2 h-3.5 w-3.5 text-ops-dim" />
            <input
              className="h-8 w-full border border-ops-border bg-ops-bg px-2.5 pl-8 text-xs text-ops-text placeholder:text-ops-dim"
              placeholder="Filter applications"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
          </label>
        </div>

        {applications.length === 0 ? (
          <EmptyState>{query.length > 0 ? "No applications match this filter." : "No applications are configured."}</EmptyState>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full text-xs">
                <thead className="bg-ops-panel-alt text-left text-[10px] uppercase tracking-[0.12em] text-ops-dim">
                  <tr>
                    <th className="px-3 py-2 font-normal">Application</th>
                    <th className="px-3 py-2 font-normal">Namespace</th>
                    <th className="px-3 py-2 font-normal">Status</th>
                    <th className="px-3 py-2 font-normal">Updated</th>
                    <th className="px-3 py-2 text-right font-normal">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-ops-border-muted">
                  {applications.map((app) => (
                    <tr key={app.id} className="transition-colors hover:bg-ops-panel-alt">
                      <td className="px-3 py-2.5">
                        <Link className="break-words font-medium text-ops-accent hover:underline" to={`/applications/${app.id}`}>
                          {app.name}
                        </Link>
                      </td>
                      <td className="px-3 py-2.5 font-mono text-ops-muted">{app.namespace}</td>
                      <td className="px-3 py-2.5">
                        <Badge label={app.status} tone={severityTone(app.status === "active" ? "low" : "medium")} />
                      </td>
                      <td className="whitespace-nowrap px-3 py-2.5 text-ops-muted">{formatDate(app.updated_at)}</td>
                      <td className="px-3 py-2.5">
                        <div className="flex justify-end gap-2">
                          <ActionButton
                            icon={Play}
                            onClick={() => analyzeMutation.mutate(app.id)}
                            disabled={analyzeMutation.isPending}
                            type="button"
                          >
                            {analyzeMutation.isPending && analyzeMutation.variables === app.id ? "Queueing" : "Analyze"}
                          </ActionButton>
                          <Link
                            aria-label={`Open ${app.name}`}
                            className="touch-target inline-flex h-8 w-8 items-center justify-center text-ops-muted hover:bg-ops-shell-hover hover:text-ops-text"
                            to={`/applications/${app.id}`}
                          >
                            <ArrowRight aria-hidden="true" className="h-4 w-4" />
                          </Link>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="divide-y divide-ops-border-muted md:hidden">
              {applications.map((app) => (
                <article key={app.id} className="space-y-3 px-4 py-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <Link className="break-words font-medium text-ops-accent" to={`/applications/${app.id}`}>
                        {app.name}
                      </Link>
                      <p className="mt-1 break-all font-mono text-xs text-ops-muted">{app.namespace}</p>
                    </div>
                    <Badge label={app.status} tone={severityTone(app.status === "active" ? "low" : "medium")} />
                  </div>
                  <p className="text-xs text-ops-muted">Updated {formatDate(app.updated_at)}</p>
                  <div className="flex gap-2">
                    <ActionButton icon={Play} onClick={() => analyzeMutation.mutate(app.id)} disabled={analyzeMutation.isPending} type="button">
                      Analyze
                    </ActionButton>
                    <Link className="touch-target inline-flex h-8 items-center gap-2 px-3 text-xs font-medium uppercase tracking-[0.08em] text-ops-accent" to={`/applications/${app.id}`}>
                      Open
                      <ArrowRight aria-hidden="true" className="h-4 w-4" />
                    </Link>
                  </div>
                </article>
              ))}
            </div>
          </>
        )}
      </Panel>
    </section>
  );
}
