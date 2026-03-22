const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
  error?: {
    code: string;
    message: string;
  };
  meta?: Record<string, unknown>;
};

export type Overview = {
  workspace_count: number;
  application_count: number;
  incident_count: number;
  open_jobs_count: number;
};

export type Application = {
  id: string;
  workspace_id: string;
  cluster_id: string;
  source_definition_id: string;
  name: string;
  namespace: string;
  status: string;
  created_at: string;
  updated_at: string;
};

export type Incident = {
  id: string;
  application_id: string;
  desired_snapshot_id: string;
  live_snapshot_id: string;
  title: string;
  category: string;
  severity: string;
  confidence: number;
  recommended_action: string;
  status: string;
  created_at: string;
  updated_at: string;
};

export type DriftField = {
  id: string;
  incident_id: string;
  resource_ref: string;
  field_path: string;
  desired_value: string;
  live_value: string;
  difference_type: string;
  created_at: string;
};

export type EvidenceRecord = {
  id: string;
  incident_id: string;
  source: string;
  detail: string;
  actor: string;
  confidence: number;
  metadata: string;
  created_at: string;
};

export type ApplicationDetails = {
  application: Application;
  incidents: Incident[];
};

export type IncidentDetails = {
  incident: Incident;
  fields: DriftField[];
  evidence: EvidenceRecord[];
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    }
  });

  const data = (await response.json()) as ApiEnvelope<T>;
  if (!response.ok || !data.success) {
    const message = data.error?.message ?? `Request failed with status ${response.status}`;
    throw new Error(message);
  }
  return data.data;
}

export function getOverview() {
  return request<Overview>("/api/v1/overview");
}

export function getApplications() {
  return request<Application[]>("/api/v1/applications");
}

export function getApplication(id: string) {
  return request<ApplicationDetails>(`/api/v1/applications/${id}`);
}

export function analyzeApplication(id: string) {
  return request<{ job_id: string; status: string }>(`/api/v1/applications/${id}/analyze`, {
    method: "POST"
  });
}

export function getIncident(id: string) {
  return request<IncidentDetails>(`/api/v1/incidents/${id}`);
}
