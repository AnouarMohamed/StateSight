const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";
const USER_ID = import.meta.env.VITE_USER_ID ?? "";
const WORKSPACE_ID = import.meta.env.VITE_WORKSPACE_ID ?? "";
const USER_EMAIL = import.meta.env.VITE_USER_EMAIL ?? "";

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

export type SuppressedFinding = {
  id: string;
  application_id: string;
  desired_snapshot_id: string;
  live_snapshot_id: string;
  ignore_rule_id: string;
  ignore_rule_name: string;
  ignore_rule_reason: string;
  title: string;
  category: string;
  severity: string;
  resource_ref: string;
  field_path: string;
  desired_value: string;
  live_value: string;
  difference_type: string;
  suppressed_at: string;
};

export type IgnoreRule = {
  id: string;
  workspace_id: string;
  application_id: string;
  resource_ref: string;
  name: string;
  match_expression: string;
  reason: string;
  created_by: string;
  active: boolean;
  created_at: string;
};

export type CreateIgnoreRuleInput = {
  name: string;
  match_expression: string;
  resource_ref: string;
  reason: string;
};

export type ApplicationDetails = {
  application: Application;
  incidents: Incident[];
  suppressions: SuppressedFinding[];
  ignore_rules: IgnoreRule[];
};

export type IncidentDetails = {
  incident: Incident;
  fields: DriftField[];
  evidence: EvidenceRecord[];
  timeline: TimelineEvent[];
};

export type TimelineEvent = {
  at: string;
  type: string;
  summary: string;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const authHeaders: Record<string, string> = {};
  if (USER_ID.length > 0) {
    authHeaders["X-User-ID"] = USER_ID;
  }
  if (WORKSPACE_ID.length > 0) {
    authHeaders["X-Workspace-ID"] = WORKSPACE_ID;
  }
  if (USER_EMAIL.length > 0) {
    authHeaders["X-User-Email"] = USER_EMAIL;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...authHeaders,
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

export function createIgnoreRule(applicationId: string, input: CreateIgnoreRuleInput) {
  return request<IgnoreRule>(`/api/v1/applications/${applicationId}/ignore-rules`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function setIgnoreRuleActive(applicationId: string, ruleId: string, active: boolean) {
  return request<IgnoreRule>(`/api/v1/applications/${applicationId}/ignore-rules/${ruleId}`, {
    method: "PATCH",
    body: JSON.stringify({ active })
  });
}

export function getIncident(id: string) {
  return request<IncidentDetails>(`/api/v1/incidents/${id}`);
}

export function getIncidentTimeline(id: string) {
  return request<TimelineEvent[]>(`/api/v1/incidents/${id}/timeline`);
}
