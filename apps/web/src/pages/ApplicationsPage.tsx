import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { analyzeApplication, getApplications } from "../lib/api";
import { Badge, severityTone } from "../components/Badge";

export function ApplicationsPage() {
  const queryClient = useQueryClient();
  const { data, isLoading, error } = useQuery({
    queryKey: ["applications"],
    queryFn: getApplications
  });

  const analyzeMutation = useMutation({
    mutationFn: (id: string) => analyzeApplication(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["overview"] });
    }
  });

  if (isLoading) {
    return <p className="text-ops-muted">Loading applications...</p>;
  }
  if (error) {
    return <p className="text-ops-bad">Failed to load applications: {(error as Error).message}</p>;
  }

  return (
    <section className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Applications</h1>
        <p className="mt-1 text-sm text-ops-muted">Tracked workloads available for drift analysis.</p>
      </div>

      <div className="overflow-hidden rounded-xl border border-ops-border bg-ops-panel shadow-panel">
        <table className="min-w-full divide-y divide-ops-border text-sm">
          <thead className="bg-[#111a26] text-left text-xs uppercase tracking-wide text-ops-muted">
            <tr>
              <th className="px-4 py-3">Application</th>
              <th className="px-4 py-3">Namespace</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Updated</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-ops-border">
            {(data ?? []).map((app) => (
              <tr key={app.id} className="hover:bg-[#162132]">
                <td className="px-4 py-3">
                  <Link className="font-medium text-ops-accent hover:underline" to={`/applications/${app.id}`}>
                    {app.name}
                  </Link>
                </td>
                <td className="px-4 py-3 text-ops-muted">{app.namespace}</td>
                <td className="px-4 py-3">
                  <Badge label={app.status} tone={severityTone(app.status === "active" ? "low" : "medium")} />
                </td>
                <td className="px-4 py-3 text-ops-muted">{new Date(app.updated_at).toLocaleString()}</td>
                <td className="px-4 py-3 text-right">
                  <button
                    className="rounded-md border border-ops-border bg-[#162132] px-3 py-1.5 text-xs font-medium text-ops-text hover:bg-[#1b2b40] disabled:opacity-50"
                    onClick={() => analyzeMutation.mutate(app.id)}
                    disabled={analyzeMutation.isPending}
                    type="button"
                  >
                    {analyzeMutation.isPending ? "Queuing..." : "Analyze"}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
