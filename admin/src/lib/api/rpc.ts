import { api } from "./client";

export interface RPCProcedure {
  id: string;
  name: string;
  namespace: string;
  description?: string;
  sql_query: string;
  original_code?: string;
  input_schema?: Record<string, string>;
  output_schema?: Record<string, string>;
  allowed_tables: string[];
  allowed_schemas: string[];
  max_execution_time_seconds: number;
  require_role?: string;
  is_public: boolean;
  schedule?: string;
  enabled: boolean;
  version: number;
  source: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export type RPCExecutionStatus =
  | "pending"
  | "running"
  | "completed"
  | "failed"
  | "cancelled"
  | "timeout";

export interface RPCExecution {
  id: string;
  procedure_id?: string;
  procedure_name: string;
  namespace: string;
  status: RPCExecutionStatus;
  input_params?: Record<string, unknown>;
  result?: unknown;
  error_message?: string;
  rows_returned?: number;
  duration_ms?: number;
  user_id?: string;
  user_role?: string;
  user_email?: string;
  is_async: boolean;
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface RPCSyncResult {
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
  errors: Array<{ procedure: string; error: string }>;
  dry_run: boolean;
}

export interface UpdateRPCProcedureRequest {
  description?: string;
  enabled?: boolean;
  is_public?: boolean;
  require_role?: string;
  max_execution_time_seconds?: number;
  allowed_tables?: string[];
  allowed_schemas?: string[];
  schedule?: string;
}

export const rpcApi = {
  listNamespaces: async (): Promise<string[]> => {
    const response = await api.get<{ namespaces: string[] }>(
      "/api/v1/admin/rpc/namespaces",
    );
    return response.data.namespaces || ["default"];
  },

  listProcedures: async (namespace?: string): Promise<RPCProcedure[]> => {
    const params = namespace ? `?namespace=${namespace}` : "";
    const response = await api.get<{
      procedures: RPCProcedure[];
      count: number;
    }>(`/api/v1/admin/rpc/procedures${params}`);
    return response.data.procedures || [];
  },

  getProcedure: async (
    namespace: string,
    name: string,
  ): Promise<RPCProcedure> => {
    const response = await api.get<RPCProcedure>(
      `/api/v1/admin/rpc/procedures/${namespace}/${name}`,
    );
    return response.data;
  },

  updateProcedure: async (
    namespace: string,
    name: string,
    data: UpdateRPCProcedureRequest,
  ): Promise<RPCProcedure> => {
    const response = await api.put<RPCProcedure>(
      `/api/v1/admin/rpc/procedures/${namespace}/${name}`,
      data,
    );
    return response.data;
  },

  deleteProcedure: async (namespace: string, name: string): Promise<void> => {
    await api.delete(`/api/v1/admin/rpc/procedures/${namespace}/${name}`);
  },

  sync: async (namespace: string): Promise<RPCSyncResult> => {
    const response = await api.post<RPCSyncResult>("/api/v1/admin/rpc/sync", {
      namespace,
    });
    return response.data;
  },

  listExecutions: async (filters?: {
    namespace?: string;
    procedure?: string;
    status?: RPCExecutionStatus;
    limit?: number;
    offset?: number;
  }): Promise<{ executions: RPCExecution[]; total: number }> => {
    const params = new URLSearchParams();
    if (filters?.namespace) params.set("namespace", filters.namespace);
    if (filters?.procedure) params.set("procedure", filters.procedure);
    if (filters?.status) params.set("status", filters.status);
    if (filters?.limit) params.set("limit", filters.limit.toString());
    if (filters?.offset) params.set("offset", filters.offset.toString());

    const queryString = params.toString();
    const response = await api.get<{
      executions: RPCExecution[];
      count: number;
    }>(`/api/v1/admin/rpc/executions${queryString ? `?${queryString}` : ""}`);
    return {
      executions: response.data.executions || [],
      total: response.data.count || 0,
    };
  },

  getExecution: async (executionId: string): Promise<RPCExecution> => {
    const response = await api.get<RPCExecution>(
      `/api/v1/admin/rpc/executions/${executionId}`,
    );
    return response.data;
  },

  cancelExecution: async (executionId: string): Promise<void> => {
    await api.post(`/api/v1/admin/rpc/executions/${executionId}/cancel`);
  },
};
