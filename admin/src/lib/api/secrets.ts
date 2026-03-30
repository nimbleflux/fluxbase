import { api } from "./client";

export interface Secret {
  id: string;
  name: string;
  scope: "global" | "namespace";
  namespace?: string;
  description?: string;
  version: number;
  expires_at?: string;
  is_expired?: boolean;
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
}

export interface SecretVersion {
  id: string;
  secret_id: string;
  version: number;
  created_at: string;
  created_by?: string;
}

export interface CreateSecretRequest {
  name: string;
  value: string;
  scope: "global" | "namespace";
  namespace?: string;
  description?: string;
  expires_at?: string;
}

export interface UpdateSecretRequest {
  value?: string;
  description?: string;
  expires_at?: string;
}

export interface SecretsStats {
  total: number;
  expiring_soon: number;
  expired: number;
}

export const secretsApi = {
  list: async (scope?: string, namespace?: string): Promise<Secret[]> => {
    const params = new URLSearchParams();
    if (scope) params.set("scope", scope);
    if (namespace) params.set("namespace", namespace);
    const queryString = params.toString();
    const response = await api.get<Secret[]>(
      `/api/v1/secrets${queryString ? `?${queryString}` : ""}`,
    );
    return response.data;
  },

  get: async (id: string): Promise<Secret> => {
    const response = await api.get<Secret>(`/api/v1/secrets/${id}`);
    return response.data;
  },

  create: async (request: CreateSecretRequest): Promise<Secret> => {
    const response = await api.post<Secret>("/api/v1/secrets", request);
    return response.data;
  },

  update: async (id: string, request: UpdateSecretRequest): Promise<Secret> => {
    const response = await api.put<Secret>(`/api/v1/secrets/${id}`, request);
    return response.data;
  },

  delete: async (id: string): Promise<{ message: string }> => {
    const response = await api.delete<{ message: string }>(
      `/api/v1/secrets/${id}`,
    );
    return response.data;
  },

  getVersions: async (id: string): Promise<SecretVersion[]> => {
    const response = await api.get<SecretVersion[]>(
      `/api/v1/secrets/${id}/versions`,
    );
    return response.data;
  },

  rollback: async (
    id: string,
    version: number,
  ): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>(
      `/api/v1/secrets/${id}/rollback/${version}`,
    );
    return response.data;
  },

  getStats: async (): Promise<SecretsStats> => {
    const response = await api.get<SecretsStats>("/api/v1/secrets/stats");
    return response.data;
  },
};
