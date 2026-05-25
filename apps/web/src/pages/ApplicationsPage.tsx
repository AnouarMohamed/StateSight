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
    <section className="space-y-6">
      <PageHeader label="Inventory / Applications" title="Applications">
        <span>{data?.length ?? 0} tracked workloads</span>
      </PageHeader>

      {analyzeMutation.error ? <ErrorState>Unable to queue analysis: {(analyzeMutation.error as Error).message}</ErrorState> : null}

      <Panel>
        <div className="flex flex-wrap items-center justify-end gap-4 border-b border-ops-border px-4 py-3 sm:px-5">
          <label className="relative block w-full sm:w-72">
            <span className="sr-only">Filter applications</span>
            <Search aria-hidden="true" className="absolute left-3 top-2.5 h-4 w-4 text-ops-muted" />
            <input
              className="h-9 w-full rounded-md border border-ops-border bg-ops-bg px-3 pl-9 text-sm text-ops-text placeholder:text-ops-muted"
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
              <table className="w-full text-sm">
                <thead className="bg-ops-panel-alt text-left text-xs font-medium text-ops-muted">
                  <tr>
                    <th className="px-5 py-2.5">Application</th>
                    <th className="px-5 py-2.5">Namespace</th>
                    <th className="px-5 py-2.5">Status</th>
                    <th className="px-5 py-2.5">Updated</th>
                    <th className="px-5 py-2.5 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-ops-border-muted">
                  {applications.map((app) => (
                    <tr key={app.id} className="transition-colors hover:bg-ops-panel-alt">
                      <td className="px-5 py-3">
                        <Link className="break-words font-medium text-ops-accent hover:underline" to={`/applications/${app.id}`}>
                          {app.name}
                        </Link>
                      </td>
                      <td className="px-5 py-3 font-mono text-xs text-ops-muted">{app.namespace}</td>
                      <td className="px-5 py-3">
                        <Badge label={app.status} tone={severityTone(app.status === "active" ? "low" : "medium")} />
                      </td>
                      <td className="whitespace-nowrap px-5 py-3 text-ops-muted">{formatDate(app.updated_at)}</td>
                      <td className="px-5 py-3">
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
                            className="touch-target inline-flex h-9 w-9 items-center justify-center rounded-md text-ops-muted hover:bg-ops-shell-hover hover:text-ops-text"
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
                    <Link className="touch-target inline-flex h-9 items-center gap-2 rounded-md px-3 text-sm font-medium text-ops-accent" to={`/applications/${app.id}`}>
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
