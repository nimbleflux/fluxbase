import { api } from "./client";

export interface JobFunction {
  id: string;
  name: string;
  namespace: string;
  description?: string;
  code?: string;
  original_code?: string;
  is_bundled: boolean;
  bundle_error?: string;
  enabled: boolean;
  schedule?: string;
  timeout_seconds: number;
  memory_limit_mb: number;
  max_retries: number;
  progress_timeout_seconds: number;
  allow_net: boolean;
  allow_env: boolean;
  allow_read: boolean;
  allow_write: boolean;
  require_role?: string;
  version: number;
  created_by?: string;
  source: string;
  created_at: string;
  updated_at: string;
}

export interface Job {
  id: string;
  namespace: string;
  job_function_id?: string;
  job_name: string;
  status: "pending" | "running" | "completed" | "failed" | "cancelled";
  payload?: unknown;
  result?: unknown;
  error_message?: string;
  priority: number;
  max_duration_seconds?: number;
  progress_timeout_seconds?: number;
  progress_percent?: number;
  progress_message?: string;
  progress_data?: unknown;
  max_retries: number;
  retry_count: number;
  worker_id?: string;
  created_by?: string;
  user_role?: string;
  user_email?: string;
  user_name?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  scheduled_at?: string;
  last_progress_at?: string;
  estimated_completion_at?: string;
  estimated_seconds_left?: number;
}
export type LogLevel = "debug" | "info" | "warning" | "error" | "fatal";
export interface JobStats {
  namespace?: string;
  pending: number;
  running: number;
  completed: number;
  failed: number;
  cancelled: number;
  total: number;
}
export interface JobWorker {
  id: string;
  hostname: string;
  status: "active" | "idle" | "dead";
  current_jobs: number;
  total_completed: number;
  started_at: string;
  last_heartbeat_at: string;
}
export interface CreateJobFunctionRequest {
  name: string;
  namespace?: string;
  description?: string;
  code: string;
  enabled?: boolean;
  schedule?: string;
  timeout_seconds?: number;
  memory_limit_mb?: number;
  max_retries?: number;
  progress_timeout_seconds?: number;
  allow_net?: boolean;
  allow_env?: boolean;
  allow_read?: boolean;
  allow_write?: boolean;
}
export interface UpdateJobFunctionRequest {
  description?: string;
  code?: string;
  enabled?: boolean;
  schedule?: string;
  timeout_seconds?: number;
  memory_limit_mb?: number;
  max_retries?: number;
  progress_timeout_seconds?: number;
  allow_net?: boolean;
  allow_env?: boolean;
  allow_read?: boolean;
  allow_write?: boolean;
}
export interface SubmitJobRequest {
  job_name: string;
  namespace?: string;
  payload?: unknown;
  priority?: number;
  scheduled?: string;
}
export interface JobSyncResult {
  message?: string;
  summary: {
    created: number;
    updated: number;
    deleted: number;
    unchanged: number;
    errors: number;
  };
  functions?: JobFunction[];
  errors?: Array<{ name: string; error: string }>;
}
export const jobsApi = {
  listNamespaces: async (): Promise<string[]> => {
    const response = await api.get<{ namespaces: string[] }>(
      "/api/v1/admin/jobs/namespaces",
    );
    return response.data.namespaces || ["default"];
  },

  listFunctions: async (namespace?: string): Promise<JobFunction[]> => {
    const params = namespace ? `?namespace=${namespace}` : "";
    const response = await api.get<JobFunction[]>(
      `/api/v1/admin/jobs/functions${params}`,
    );
    return response.data;
  },

  getFunction: async (
    namespace: string,
    name: string,
  ): Promise<JobFunction> => {
    const response = await api.get<JobFunction>(
      `/api/v1/admin/jobs/functions/${namespace}/${name}`,
    );
    return response.data;
  },

  createFunction: async (
    data: CreateJobFunctionRequest,
  ): Promise<JobFunction> => {
    const response = await api.post<JobFunction>(
      "/api/v1/admin/jobs/functions",
      data,
    );
    return response.data;
  },

  updateFunction: async (
    namespace: string,
    name: string,
    data: UpdateJobFunctionRequest,
  ): Promise<JobFunction> => {
    const response = await api.put<JobFunction>(
      `/api/v1/admin/jobs/functions/${namespace}/${name}`,
      data,
    );
    return response.data;
  },

  deleteFunction: async (namespace: string, name: string): Promise<void> => {
    await api.delete(`/api/v1/admin/jobs/functions/${namespace}/${name}`);
  },

  submitJob: async (
    data: SubmitJobRequest,
    config?: { headers?: Record<string, string> },
  ): Promise<Job> => {
    const response = await api.post<Job>("/api/v1/jobs/submit", data, config);
    return response.data;
  },

  listJobs: async (filters?: {
    status?: string;
    namespace?: string;
    limit?: number;
    offset?: number;
  }): Promise<Job[]> => {
    const params = new URLSearchParams();
    if (filters?.status) params.append("status", filters.status);
    if (filters?.namespace) params.append("namespace", filters.namespace);
    if (filters?.limit) params.append("limit", filters.limit.toString());
    if (filters?.offset) params.append("offset", filters.offset.toString());

    const queryString = params.toString();
    const response = await api.get<{
      jobs: Job[];
      limit: number;
      offset: number;
    }>(`/api/v1/admin/jobs/queue${queryString ? `?${queryString}` : ""}`);
    return response.data.jobs;
  },

  getJob: async (jobId: string): Promise<Job> => {
    const response = await api.get<Job>(`/api/v1/admin/jobs/queue/${jobId}`);
    return response.data;
  },

  cancelJob: async (jobId: string): Promise<void> => {
    await api.post(`/api/v1/admin/jobs/queue/${jobId}/cancel`, {});
  },

  terminateJob: async (jobId: string): Promise<void> => {
    await api.post(`/api/v1/admin/jobs/queue/${jobId}/terminate`, {});
  },

  retryJob: async (jobId: string): Promise<Job> => {
    const response = await api.post<Job>(
      `/api/v1/admin/jobs/queue/${jobId}/retry`,
      {},
    );
    return response.data;
  },

  resubmitJob: async (jobId: string): Promise<Job> => {
    const response = await api.post<Job>(
      `/api/v1/admin/jobs/queue/${jobId}/resubmit`,
      {},
    );
    return response.data;
  },

  getStats: async (namespace?: string): Promise<JobStats> => {
    const params = namespace ? `?namespace=${namespace}` : "";
    const response = await api.get<JobStats>(
      `/api/v1/admin/jobs/stats${params}`,
    );
    return response.data;
  },

  listWorkers: async (): Promise<JobWorker[]> => {
    const response = await api.get<JobWorker[]>("/api/v1/admin/jobs/workers");
    return response.data;
  },

  sync: async (namespace: string): Promise<JobSyncResult> => {
    const response = await api.post<JobSyncResult>("/api/v1/admin/jobs/sync", {
      namespace,
    });
    return response.data;
  },
};
