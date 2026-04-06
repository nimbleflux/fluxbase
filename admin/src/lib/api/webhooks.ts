import { api } from "./client";

export interface EventConfig {
  table: string;
  operations: string[];
}
export interface WebhookType {
  id: string;
  name: string;
  description?: string;
  url: string;
  secret?: string;
  enabled: boolean;
  events: EventConfig[];
  max_retries: number;
  retry_backoff_seconds: number;
  timeout_seconds: number;
  headers: Record<string, string>;
  created_at: string;
  updated_at: string;
}
export interface WebhookDelivery {
  id: string;
  webhook_id: string;
  event_type: string;
  table_name: string;
  record_id?: string;
  payload: unknown;
  attempt_number: number;
  status: string;
  http_status_code?: number;
  response_body?: string;
  error_message?: string;
  created_at: string;
  delivered_at?: string;
}
export const webhooksApi = {
  list: async (): Promise<WebhookType[]> => {
    const response = await api.get<WebhookType[]>("/api/v1/webhooks");
    return response.data;
  },
  get: async (id: string): Promise<WebhookType> => {
    const response = await api.get<WebhookType>(`/api/v1/webhooks/${id}`);
    return response.data;
  },
  getDeliveries: async (
    webhookId: string,
    limit = 50,
  ): Promise<WebhookDelivery[]> => {
    const response = await api.get<WebhookDelivery[]>(
      `/api/v1/webhooks/${webhookId}/deliveries?limit=${limit}`,
    );
    return response.data;
  },
  create: async (webhook: Partial<WebhookType>): Promise<WebhookType> => {
    const response = await api.post<WebhookType>("/api/v1/webhooks", webhook);
    return response.data;
  },
  update: async (
    id: string,
    updates: Partial<WebhookType>,
  ): Promise<WebhookType> => {
    const response = await api.patch<WebhookType>(
      `/api/v1/webhooks/${id}`,
      updates,
    );
    return response.data;
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/webhooks/${id}`);
  },
  test: async (id: string): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>(
      `/api/v1/webhooks/${id}/test`,
    );
    return response.data;
  },
};
