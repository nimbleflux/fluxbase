import { api } from "./client";

export interface ClientKey {
  id: string;
  name: string;
  description?: string;
  key_prefix: string;
  scopes: string[];
  rate_limit_per_minute: number;
  last_used_at?: string;
  expires_at?: string;
  revoked_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateClientKeyRequest {
  name: string;
  description?: string;
  scopes: string[];
  rate_limit_per_minute: number;
  expires_at?: string;
}

export interface CreateClientKeyResponse {
  client_key: ClientKey;
  key: string;
}

export const clientKeysApi = {
  list: async (): Promise<ClientKey[]> => {
    const response = await api.get<ClientKey[]>("/api/v1/client-keys");
    return response.data;
  },

  create: async (
    request: CreateClientKeyRequest,
  ): Promise<CreateClientKeyResponse> => {
    const response = await api.post<CreateClientKeyResponse>(
      "/api/v1/client-keys",
      request,
    );
    return response.data;
  },

  revoke: async (id: string): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>(
      `/api/v1/client-keys/${id}/revoke`,
    );
    return response.data;
  },

  delete: async (id: string): Promise<{ message: string }> => {
    const response = await api.delete<{ message: string }>(
      `/api/v1/client-keys/${id}`,
    );
    return response.data;
  },
};

export interface ServiceKey {
  id: string;
  name: string;
  description?: string;
  key_prefix: string;
  scopes: string[];
  enabled: boolean;
  rate_limit_per_minute?: number;
  rate_limit_per_hour?: number;
  created_by?: string;
  created_at: string;
  last_used_at?: string;
  expires_at?: string;
  revoked_at?: string;
  revoked_by?: string;
  revocation_reason?: string;
  deprecated_at?: string;
  grace_period_ends_at?: string;
  replaced_by?: string;
}

export interface ServiceKeyRevocation {
  id: string;
  service_key_id: string;
  revocation_type: "emergency" | "rotation" | "expiration" | "deprecation";
  reason: string;
  revoked_by: string;
  created_at: string;
}

export interface RevokeServiceKeyRequest {
  reason: string;
}

export interface DeprecateServiceKeyRequest {
  grace_period: string;
  reason?: string;
}

export interface RotateServiceKeyRequest {
  grace_period: string;
}

export interface ServiceKeyWithPlaintext extends ServiceKey {
  key: string;
}

export interface RotateServiceKeyResponse extends ServiceKeyWithPlaintext {
  grace_period_ends_at: string;
}

export interface CreateServiceKeyRequest {
  name: string;
  description?: string;
  scopes?: string[];
  rate_limit_per_minute?: number;
  rate_limit_per_hour?: number;
  expires_at?: string;
}

export interface UpdateServiceKeyRequest {
  name?: string;
  description?: string;
  scopes?: string[];
  enabled?: boolean;
  rate_limit_per_minute?: number;
  rate_limit_per_hour?: number;
}

export const serviceKeysApi = {
  list: async (): Promise<ServiceKey[]> => {
    const response = await api.get<ServiceKey[]>("/api/v1/admin/service-keys");
    return response.data;
  },

  get: async (id: string): Promise<ServiceKey> => {
    const response = await api.get<ServiceKey>(
      `/api/v1/admin/service-keys/${id}`,
    );
    return response.data;
  },

  create: async (
    request: CreateServiceKeyRequest,
  ): Promise<ServiceKeyWithPlaintext> => {
    const response = await api.post<ServiceKeyWithPlaintext>(
      "/api/v1/admin/service-keys",
      request,
    );
    return response.data;
  },

  update: async (
    id: string,
    request: UpdateServiceKeyRequest,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.patch<{ success: boolean; message: string }>(
      `/api/v1/admin/service-keys/${id}`,
      request,
    );
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/admin/service-keys/${id}`);
  },

  disable: async (
    id: string,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.post<{ success: boolean; message: string }>(
      `/api/v1/admin/service-keys/${id}/disable`,
    );
    return response.data;
  },

  enable: async (
    id: string,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.post<{ success: boolean; message: string }>(
      `/api/v1/admin/service-keys/${id}/enable`,
    );
    return response.data;
  },

  revoke: async (
    id: string,
    request: RevokeServiceKeyRequest,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.post<{ success: boolean; message: string }>(
      `/api/v1/admin/service-keys/${id}/revoke`,
      request,
    );
    return response.data;
  },

  deprecate: async (
    id: string,
    request: DeprecateServiceKeyRequest,
  ): Promise<ServiceKey> => {
    const response = await api.post<ServiceKey>(
      `/api/v1/admin/service-keys/${id}/deprecate`,
      request,
    );
    return response.data;
  },

  rotate: async (
    id: string,
    request: RotateServiceKeyRequest,
  ): Promise<RotateServiceKeyResponse> => {
    const response = await api.post<RotateServiceKeyResponse>(
      `/api/v1/admin/service-keys/${id}/rotate`,
      request,
    );
    return response.data;
  },

  revocations: async (id: string): Promise<ServiceKeyRevocation[]> => {
    const response = await api.get<ServiceKeyRevocation[]>(
      `/api/v1/admin/service-keys/${id}/revocations`,
    );
    return response.data;
  },
};

export interface PlatformServiceKey {
  id: string;
  name: string;
  description?: string;
  key_type: "anon" | "publishable" | "tenant_service" | "global_service";
  tenant_id?: string;
  key_prefix: string;
  scopes: string[];
  allowed_namespaces?: string[];
  rate_limit_per_minute?: number;
  is_active: boolean;
  is_config_managed: boolean;
  revoked_at?: string;
  revoked_by?: string;
  revocation_reason?: string;
  deprecated_at?: string;
  grace_period_ends_at?: string;
  replaced_by?: string;
  created_at: string;
  created_by?: string;
  updated_at: string;
  last_used_at?: string;
  expires_at?: string;
}

export interface PlatformServiceKeyWithPlaintext extends PlatformServiceKey {
  key?: string;
}

export interface CreatePlatformServiceKeyRequest {
  name: string;
  description?: string;
  key_type: "anon" | "publishable" | "tenant_service" | "global_service";
  tenant_id?: string;
  scopes?: string[];
  allowed_namespaces?: string[];
  rate_limit_per_minute?: number;
  expires_at?: string;
}

export interface UpdatePlatformServiceKeyRequest {
  name?: string;
  description?: string;
  scopes?: string[];
  allowed_namespaces?: string[];
  is_active?: boolean;
  rate_limit_per_minute?: number;
}

export interface RotatePlatformServiceKeyRequest {
  grace_period_hours?: number;
  new_key_name?: string;
  new_scopes?: string[];
}

export const platformServiceKeysApi = {
  list: async (): Promise<PlatformServiceKey[]> => {
    const response = await api.get<PlatformServiceKey[]>(
      "/api/v1/admin/platform/service-keys",
    );
    return response.data;
  },

  get: async (id: string): Promise<PlatformServiceKey> => {
    const response = await api.get<PlatformServiceKey>(
      `/api/v1/admin/platform/service-keys/${id}`,
    );
    return response.data;
  },

  create: async (
    request: CreatePlatformServiceKeyRequest,
  ): Promise<PlatformServiceKeyWithPlaintext> => {
    const response = await api.post<PlatformServiceKeyWithPlaintext>(
      "/api/v1/admin/platform/service-keys",
      request,
    );
    return response.data;
  },

  update: async (
    id: string,
    request: UpdatePlatformServiceKeyRequest,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.patch<{ success: boolean; message: string }>(
      `/api/v1/admin/platform/service-keys/${id}`,
      request,
    );
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/admin/platform/service-keys/${id}`);
  },

  disable: async (
    id: string,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.post<{ success: boolean; message: string }>(
      `/api/v1/admin/platform/service-keys/${id}/disable`,
    );
    return response.data;
  },

  enable: async (
    id: string,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.post<{ success: boolean; message: string }>(
      `/api/v1/admin/platform/service-keys/${id}/enable`,
    );
    return response.data;
  },

  rotate: async (
    id: string,
    request: RotatePlatformServiceKeyRequest,
  ): Promise<
    PlatformServiceKeyWithPlaintext & { grace_period_ends_at: string }
  > => {
    const response = await api.post<
      PlatformServiceKeyWithPlaintext & { grace_period_ends_at: string }
    >(`/api/v1/admin/platform/service-keys/${id}/rotate`, request);
    return response.data;
  },
};

export type APIKey = ClientKey;
export type CreateAPIKeyRequest = CreateClientKeyRequest;
export type CreateAPIKeyResponse = CreateClientKeyResponse;
export const apiKeysApi = clientKeysApi;
