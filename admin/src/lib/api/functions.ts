import { api } from "./client";

export interface EdgeFunction {
  id: string;
  name: string;
  description?: string;
  code: string;
  version: number;
  cron_schedule?: string;
  enabled: boolean;
  timeout_seconds: number;
  memory_limit_mb: number;
  allow_net: boolean;
  allow_env: boolean;
  allow_read: boolean;
  allow_write: boolean;
  source: string;
  created_at: string;
  updated_at: string;
}

export interface CreateEdgeFunctionRequest {
  name: string;
  description?: string;
  code: string;
  timeout_seconds: number;
  memory_limit_mb: number;
  allow_net: boolean;
  allow_env: boolean;
  allow_read: boolean;
  allow_write: boolean;
  cron_schedule?: string | null;
}

export interface UpdateEdgeFunctionRequest {
  code?: string;
  description?: string;
  timeout_seconds?: number;
  allow_net?: boolean;
  allow_env?: boolean;
  allow_read?: boolean;
  allow_write?: boolean;
  cron_schedule?: string | null;
  enabled?: boolean;
}

export interface EdgeFunctionExecution {
  id: string;
  function_id: string;
  function_name?: string;
  namespace?: string;
  trigger_type: string;
  status: string;
  status_code?: number;
  duration_ms?: number;
  result?: string;
  logs?: string;
  error_message?: string;
  executed_at: string;
  completed_at?: string;
}

export interface FunctionReloadResult {
  message?: string;
  created?: string[];
  updated?: string[];
  deleted?: string[];
  errors?: string[];
  total?: number;
}

export interface FunctionSyncSpec {
  name: string;
  description?: string;
  code: string;
  enabled?: boolean;
  timeout_seconds?: number;
  memory_limit_mb?: number;
  allow_net?: boolean;
  allow_env?: boolean;
  allow_read?: boolean;
  allow_write?: boolean;
  allow_unauthenticated?: boolean;
  is_public?: boolean;
  cron_schedule?: string;
}

export interface FunctionSyncOptions {
  namespace?: string;
  functions: FunctionSyncSpec[];
  options?: {
    delete_missing?: boolean;
    dry_run?: boolean;
  };
}

export interface FunctionSyncError {
  function: string;
  error: string;
  action: "create" | "update" | "delete" | "bundle";
}

export interface FunctionSyncResult {
  message: string;
  namespace: string;
  summary: {
    created: number;
    updated: number;
    deleted: number;
    unchanged: number;
    errors: number;
  };
  details: {
    created: string[];
    updated: string[];
    deleted: string[];
    unchanged: string[];
  };
  errors: FunctionSyncError[];
  dry_run: boolean;
}

export interface EdgeFunctionInvokeOptions {
  method?: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  headers?: Record<string, string>;
  body?: string;
}

export const functionsApi = {
  listNamespaces: async (): Promise<string[]> => {
    const response = await api.get<{ namespaces: string[] }>(
      "/api/v1/admin/functions/namespaces",
    );
    return response.data.namespaces || ["default"];
  },

  list: async (namespace?: string): Promise<EdgeFunction[]> => {
    const params = namespace ? `?namespace=${namespace}` : "";
    const response = await api.get<EdgeFunction[]>(
      `/api/v1/functions${params}`,
    );
    return response.data;
  },

  get: async (name: string): Promise<EdgeFunction> => {
    const response = await api.get<EdgeFunction>(`/api/v1/functions/${name}`);
    return response.data;
  },

  create: async (data: CreateEdgeFunctionRequest): Promise<EdgeFunction> => {
    const response = await api.post<EdgeFunction>("/api/v1/functions", data);
    return response.data;
  },

  update: async (
    name: string,
    data: UpdateEdgeFunctionRequest,
  ): Promise<EdgeFunction> => {
    const response = await api.put<EdgeFunction>(
      `/api/v1/functions/${name}`,
      data,
    );
    return response.data;
  },

  delete: async (name: string): Promise<void> => {
    await api.delete(`/api/v1/functions/${name}`);
  },

  invoke: async (
    name: string,
    options: EdgeFunctionInvokeOptions = {},
    config?: { headers?: Record<string, string> },
  ): Promise<string> => {
    const { method = "POST", headers = {}, body = "" } = options;

    const response = await api.request({
      url: `/api/v1/functions/${name}/invoke`,
      method,
      data: body,
      headers: {
        "Content-Type": "application/json",
        ...headers,
        ...config?.headers,
      },
      transformResponse: [(data) => data],
    });
    return response.data;
  },

  getExecutions: async (
    name: string,
    limit = 20,
  ): Promise<EdgeFunctionExecution[]> => {
    const response = await api.get<EdgeFunctionExecution[]>(
      `/api/v1/functions/${name}/executions`,
      { params: { limit } },
    );
    return response.data;
  },

  reload: async (): Promise<FunctionReloadResult> => {
    const response = await api.post<FunctionReloadResult>(
      "/api/v1/admin/functions/reload",
    );
    return response.data;
  },

  sync: async (options: FunctionSyncOptions): Promise<FunctionSyncResult> => {
    const response = await api.post<FunctionSyncResult>(
      "/api/v1/admin/functions/sync",
      options,
    );
    return response.data;
  },

  listAllExecutions: async (filters?: {
    namespace?: string;
    function_name?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<{ executions: EdgeFunctionExecution[]; count: number }> => {
    const params = new URLSearchParams();
    if (filters?.namespace) params.set("namespace", filters.namespace);
    if (filters?.function_name)
      params.set("function_name", filters.function_name);
    if (filters?.status) params.set("status", filters.status);
    if (filters?.limit) params.set("limit", filters.limit.toString());
    if (filters?.offset) params.set("offset", filters.offset.toString());

    const queryString = params.toString();
    const response = await api.get<{
      executions: EdgeFunctionExecution[];
      count: number;
    }>(
      `/api/v1/admin/functions/executions${queryString ? `?${queryString}` : ""}`,
    );
    return response.data;
  },
};
