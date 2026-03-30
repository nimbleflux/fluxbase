import { api } from "./client";

export interface MCPTool {
  id: string;
  name: string;
  namespace: string;
  description?: string;
  code: string;
  input_schema?: Record<string, unknown>;
  required_scopes: string[];
  timeout_seconds: number;
  memory_limit_mb: number;
  allow_net: boolean;
  allow_env: boolean;
  allow_read: boolean;
  allow_write: boolean;
  enabled: boolean;
  created_at: string;
  updated_at: string;
  created_by?: string;
}

export interface MCPResource {
  id: string;
  uri: string;
  name: string;
  namespace: string;
  description?: string;
  mime_type: string;
  code: string;
  required_scopes: string[];
  timeout_seconds: number;
  memory_limit_mb: number;
  allow_net: boolean;
  allow_env: boolean;
  enabled: boolean;
  is_template: boolean;
  created_at: string;
  updated_at: string;
  created_by?: string;
}

export interface CreateMCPToolRequest {
  name: string;
  namespace?: string;
  description?: string;
  code: string;
  input_schema?: Record<string, unknown>;
  required_scopes?: string[];
  timeout_seconds?: number;
  memory_limit_mb?: number;
  allow_net?: boolean;
  allow_env?: boolean;
  allow_read?: boolean;
  allow_write?: boolean;
  enabled?: boolean;
}

export interface UpdateMCPToolRequest {
  name?: string;
  namespace?: string;
  description?: string;
  code?: string;
  input_schema?: Record<string, unknown>;
  required_scopes?: string[];
  timeout_seconds?: number;
  memory_limit_mb?: number;
  allow_net?: boolean;
  allow_env?: boolean;
  allow_read?: boolean;
  allow_write?: boolean;
  enabled?: boolean;
}

export interface CreateMCPResourceRequest {
  uri: string;
  name: string;
  namespace?: string;
  description?: string;
  mime_type?: string;
  code: string;
  required_scopes?: string[];
  timeout_seconds?: number;
  memory_limit_mb?: number;
  allow_net?: boolean;
  allow_env?: boolean;
  enabled?: boolean;
}

export interface UpdateMCPResourceRequest {
  uri?: string;
  name?: string;
  namespace?: string;
  description?: string;
  mime_type?: string;
  code?: string;
  required_scopes?: string[];
  timeout_seconds?: number;
  memory_limit_mb?: number;
  allow_net?: boolean;
  allow_env?: boolean;
  enabled?: boolean;
}

export interface MCPToolTestResult {
  content: Array<{ type: string; text: string }>;
  isError?: boolean;
}

export interface MCPResourceTestResult {
  uri: string;
  mimeType?: string;
  text?: string;
  blob?: string;
}

export interface MCPConfig {
  enabled: boolean;
  base_path: string;
  tools_dir: string;
  auto_load_on_boot: boolean;
  rate_limit_per_min: number;
}

export const mcpConfigApi = {
  get: async (): Promise<MCPConfig> => {
    const response = await api.get<MCPConfig>("/api/v1/mcp/config");
    return response.data;
  },
};

export const mcpToolsApi = {
  list: async (namespace?: string): Promise<MCPTool[]> => {
    const params = namespace ? `?namespace=${namespace}` : "";
    const response = await api.get<{ tools: MCPTool[]; count: number }>(
      `/api/v1/mcp/tools${params}`,
    );
    return response.data.tools || [];
  },

  get: async (id: string): Promise<MCPTool> => {
    const response = await api.get<MCPTool>(`/api/v1/mcp/tools/${id}`);
    return response.data;
  },

  create: async (data: CreateMCPToolRequest): Promise<MCPTool> => {
    const response = await api.post<MCPTool>("/api/v1/mcp/tools", data);
    return response.data;
  },

  update: async (id: string, data: UpdateMCPToolRequest): Promise<MCPTool> => {
    const response = await api.put<MCPTool>(`/api/v1/mcp/tools/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/mcp/tools/${id}`);
  },

  sync: async (data: CreateMCPToolRequest): Promise<MCPTool> => {
    const response = await api.post<MCPTool>("/api/v1/mcp/tools/sync", {
      ...data,
      upsert: true,
    });
    return response.data;
  },

  test: async (
    id: string,
    args: Record<string, unknown>,
  ): Promise<{ success: boolean; result: MCPToolTestResult }> => {
    const response = await api.post<{
      success: boolean;
      result: MCPToolTestResult;
    }>(`/api/v1/mcp/tools/${id}/test`, { args });
    return response.data;
  },
};

export const mcpResourcesApi = {
  list: async (namespace?: string): Promise<MCPResource[]> => {
    const params = namespace ? `?namespace=${namespace}` : "";
    const response = await api.get<{ resources: MCPResource[]; count: number }>(
      `/api/v1/mcp/resources${params}`,
    );
    return response.data.resources || [];
  },

  get: async (id: string): Promise<MCPResource> => {
    const response = await api.get<MCPResource>(`/api/v1/mcp/resources/${id}`);
    return response.data;
  },

  create: async (data: CreateMCPResourceRequest): Promise<MCPResource> => {
    const response = await api.post<MCPResource>("/api/v1/mcp/resources", data);
    return response.data;
  },

  update: async (
    id: string,
    data: UpdateMCPResourceRequest,
  ): Promise<MCPResource> => {
    const response = await api.put<MCPResource>(
      `/api/v1/mcp/resources/${id}`,
      data,
    );
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/mcp/resources/${id}`);
  },

  sync: async (data: CreateMCPResourceRequest): Promise<MCPResource> => {
    const response = await api.post<MCPResource>("/api/v1/mcp/resources/sync", {
      ...data,
      upsert: true,
    });
    return response.data;
  },

  test: async (
    id: string,
    params: Record<string, string>,
  ): Promise<{ success: boolean; contents: MCPResourceTestResult[] }> => {
    const response = await api.post<{
      success: boolean;
      contents: MCPResourceTestResult[];
    }>(`/api/v1/mcp/resources/${id}/test`, { params });
    return response.data;
  },
};

export interface CaptchaSettingsResponse {
  enabled: boolean;
  provider: string;
  site_key: string;
  secret_key_set: boolean;
  score_threshold: number;
  endpoints: string[];
  cap_server_url: string;
  cap_api_key_set: boolean;
  _overrides: Record<
    string,
    {
      is_overridden: boolean;
      env_var?: string;
    }
  >;
}

export interface UpdateCaptchaSettingsRequest {
  enabled?: boolean;
  provider?: string;
  site_key?: string;
  secret_key?: string;
  score_threshold?: number;
  endpoints?: string[];
  cap_server_url?: string;
  cap_api_key?: string;
}

export const captchaSettingsApi = {
  get: async (): Promise<CaptchaSettingsResponse> => {
    const response = await api.get<CaptchaSettingsResponse>(
      "/api/v1/admin/settings/captcha",
    );
    return response.data;
  },

  update: async (
    request: UpdateCaptchaSettingsRequest,
  ): Promise<CaptchaSettingsResponse> => {
    const response = await api.put<CaptchaSettingsResponse>(
      "/api/v1/admin/settings/captcha",
      request,
    );
    return response.data;
  },
};
