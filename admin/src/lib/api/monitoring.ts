import { api } from "./client";

export interface SystemMetrics {
  uptime_seconds: number;
  go_version: string;
  num_goroutines: number;
  memory_alloc_mb: number;
  memory_total_alloc_mb: number;
  memory_sys_mb: number;
  num_gc: number;
  gc_pause_ms: number;
  database: {
    acquire_count: number;
    acquired_conns: number;
    canceled_acquire_count: number;
    constructing_conns: number;
    empty_acquire_count: number;
    idle_conns: number;
    max_conns: number;
    total_conns: number;
    new_conns_count: number;
    max_lifetime_destroy_count: number;
    max_idle_destroy_count: number;
    acquire_duration_ms: number;
  };
  realtime: {
    total_connections: number;
    active_channels: number;
    total_subscriptions: number;
  };
  storage?: {
    total_buckets: number;
    total_files: number;
    total_size_gb: number;
  };
}

export interface HealthStatus {
  status: string;
  message?: string;
  latency_ms?: number;
}

export interface SystemHealth {
  status: string;
  services: Record<string, HealthStatus>;
}

export const monitoringApi = {
  getMetrics: async (): Promise<SystemMetrics> => {
    const response = await api.get<SystemMetrics>("/api/v1/monitoring/metrics");
    return response.data;
  },

  getHealth: async (): Promise<SystemHealth> => {
    const response = await api.get<SystemHealth>("/api/v1/monitoring/health");
    return response.data;
  },
};

export interface ResolvedSetting {
  value: unknown;
  source: "config" | "instance" | "tenant" | "default";
  is_read_only?: boolean;
  is_overridable?: boolean;
  is_secret?: boolean;
  data_type?: "string" | "number" | "boolean" | "object" | "array";
}

export interface InstanceSettingsResponse {
  settings: Record<string, ResolvedSetting>;
  overridable_settings?: string[];
}

export interface UpdateInstanceSettingsRequest {
  settings: Record<string, unknown>;
}

export interface UpdateOverridableSettingsRequest {
  overridable_settings: string[];
}

export interface TenantSettingsResponse {
  tenant_id: string;
  tenant_name?: string;
  settings: Record<string, ResolvedSetting>;
  overridable_settings?: string[];
  created_at?: string;
  updated_at?: string;
}

export interface UpdateTenantSettingsRequest {
  settings?: Record<string, unknown>;
  secrets?: Record<string, string>;
}

export const instanceSettingsApi = {
  get: async (): Promise<InstanceSettingsResponse> => {
    const response = await api.get<InstanceSettingsResponse>(
      "/api/v1/admin/instance/settings",
    );
    return response.data;
  },

  getOverridable: async (): Promise<{ overridable_settings: string[] }> => {
    const response = await api.get<{ overridable_settings: string[] }>(
      "/api/v1/admin/instance/settings/overridable",
    );
    return response.data;
  },

  update: async (
    data: UpdateInstanceSettingsRequest,
  ): Promise<InstanceSettingsResponse> => {
    const response = await api.patch<InstanceSettingsResponse>(
      "/api/v1/admin/instance/settings",
      data,
    );
    return response.data;
  },

  updateOverridable: async (
    data: UpdateOverridableSettingsRequest,
  ): Promise<{ overridable_settings: string[] }> => {
    const response = await api.put<{ overridable_settings: string[] }>(
      "/api/v1/admin/instance/settings/overridable",
      data,
    );
    return response.data;
  },
};

export const tenantSettingsApi = {
  get: async (tenantId: string): Promise<TenantSettingsResponse> => {
    const response = await api.get<TenantSettingsResponse>(
      `/api/v1/admin/tenants/${tenantId}/settings`,
    );
    return response.data;
  },

  update: async (
    tenantId: string,
    data: UpdateTenantSettingsRequest,
  ): Promise<TenantSettingsResponse> => {
    const response = await api.patch<TenantSettingsResponse>(
      `/api/v1/admin/tenants/${tenantId}/settings`,
      data,
    );
    return response.data;
  },

  delete: async (tenantId: string, path: string): Promise<void> => {
    await api.delete(`/api/v1/admin/tenants/${tenantId}/settings/${path}`);
  },

  getSetting: async (
    tenantId: string,
    path: string,
  ): Promise<ResolvedSetting> => {
    const response = await api.get<ResolvedSetting>(
      `/api/v1/admin/tenants/${tenantId}/settings/${path}`,
    );
    return response.data;
  },
};
