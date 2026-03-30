import { api } from "./client";

export interface RLSPolicy {
  schema: string;
  table: string;
  policy_name: string;
  permissive: string;
  roles: string[];
  command: string;
  using?: string | null;
  with_check?: string | null;
}

export interface TableRLSStatus {
  schema: string;
  table: string;
  rls_enabled: boolean;
  rls_forced: boolean;
  policy_count: number;
  policies: RLSPolicy[];
  has_warnings: boolean;
}

export interface CreatePolicyRequest {
  schema: string;
  table: string;
  name: string;
  command: string;
  roles?: string[];
  using?: string;
  with_check?: string;
  permissive?: boolean;
}

export interface SecurityWarning {
  id: string;
  severity: "critical" | "high" | "medium" | "low";
  category: string;
  schema: string;
  table: string;
  policy_name?: string;
  message: string;
  suggestion: string;
  fix_sql?: string;
}

export interface SecurityWarningsResponse {
  warnings: SecurityWarning[];
  summary: {
    critical: number;
    high: number;
    medium: number;
    low: number;
    total: number;
  };
}

export interface PolicyTemplate {
  id: string;
  name: string;
  description: string;
  command: string;
  using: string;
  with_check: string;
}

export const policyApi = {
  list: async (schema?: string, table?: string): Promise<RLSPolicy[]> => {
    const params = new URLSearchParams();
    if (schema) params.set("schema", schema);
    if (table) params.set("table", table);
    const queryString = params.toString();
    const response = await api.get<RLSPolicy[]>(
      `/api/v1/admin/policies${queryString ? `?${queryString}` : ""}`,
    );
    return response.data || [];
  },

  getTablesWithRLS: async (schema?: string): Promise<TableRLSStatus[]> => {
    const params = schema ? `?schema=${schema}` : "";
    const response = await api.get<TableRLSStatus[]>(
      `/api/v1/admin/tables/rls${params}`,
    );
    return response.data || [];
  },

  getTableRLSStatus: async (
    schema: string,
    table: string,
  ): Promise<TableRLSStatus> => {
    const response = await api.get<TableRLSStatus>(
      `/api/v1/admin/tables/${schema}/${table}/rls`,
    );
    return response.data;
  },

  toggleTableRLS: async (
    schema: string,
    table: string,
    enable: boolean,
    forceRLS?: boolean,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.post<{ success: boolean; message: string }>(
      `/api/v1/admin/tables/${schema}/${table}/rls/toggle`,
      { enable, force_rls: forceRLS },
    );
    return response.data;
  },

  create: async (
    data: CreatePolicyRequest,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.post<{ success: boolean; message: string }>(
      "/api/v1/admin/policies",
      data,
    );
    return response.data;
  },

  delete: async (
    schema: string,
    table: string,
    policyName: string,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.delete<{ success: boolean; message: string }>(
      `/api/v1/admin/policies/${schema}/${table}/${policyName}`,
    );
    return response.data;
  },

  update: async (
    schema: string,
    table: string,
    policyName: string,
    data: {
      roles?: string[];
      using?: string | null;
      with_check?: string | null;
    },
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.put<{ success: boolean; message: string }>(
      `/api/v1/admin/policies/${schema}/${table}/${policyName}`,
      data,
    );
    return response.data;
  },

  getSecurityWarnings: async (): Promise<SecurityWarningsResponse> => {
    const response = await api.get<SecurityWarningsResponse>(
      "/api/v1/admin/security/warnings",
    );
    return response.data;
  },

  getTemplates: async (): Promise<PolicyTemplate[]> => {
    const response = await api.get<PolicyTemplate[]>(
      "/api/v1/admin/policies/templates",
    );
    return response.data || [];
  },
};
