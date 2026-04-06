import { api } from "./client";

export interface LogEntryAPI {
  id: string;
  timestamp: string;
  category: string;
  level: string;
  message: string;
  custom_category?: string;
  request_id?: string;
  trace_id?: string;
  component?: string;
  user_id?: string;
  ip_address?: string;
  fields?: Record<string, unknown>;
  execution_id?: string;
  execution_type?: string;
  line_number?: number;
}

export interface LogQueryOptionsAPI {
  category?: string;
  custom_category?: string;
  levels?: string[];
  component?: string;
  request_id?: string;
  trace_id?: string;
  user_id?: string;
  execution_id?: string;
  search?: string;
  start_time?: string;
  end_time?: string;
  limit?: number;
  offset?: number;
  sort_asc?: boolean;
  hide_static_assets?: boolean;
}

export interface LogQueryResultAPI {
  entries: LogEntryAPI[];
  total_count: number;
  has_more: boolean;
}

export interface LogStatsAPI {
  total_entries: number;
  entries_by_category: Record<string, number>;
  entries_by_level: Record<string, number>;
  oldest_entry?: string;
  newest_entry?: string;
}

export const logsApi = {
  query: async (options: LogQueryOptionsAPI): Promise<LogQueryResultAPI> => {
    const params = new URLSearchParams();
    if (options.category) params.set("category", options.category);
    if (options.custom_category)
      params.set("custom_category", options.custom_category);
    if (options.levels?.length) params.set("level", options.levels.join(","));
    if (options.component) params.set("component", options.component);
    if (options.request_id) params.set("request_id", options.request_id);
    if (options.trace_id) params.set("trace_id", options.trace_id);
    if (options.user_id) params.set("user_id", options.user_id);
    if (options.execution_id) params.set("execution_id", options.execution_id);
    if (options.search) params.set("search", options.search);
    if (options.start_time) params.set("start_time", options.start_time);
    if (options.end_time) params.set("end_time", options.end_time);
    if (options.limit) params.set("limit", options.limit.toString());
    if (options.offset) params.set("offset", options.offset.toString());
    if (options.sort_asc) params.set("sort_asc", "true");
    if (options.hide_static_assets) params.set("hide_static_assets", "true");

    const response = await api.get<LogQueryResultAPI>(
      `/api/v1/admin/logs?${params.toString()}`,
    );
    return response.data;
  },

  getStats: async (): Promise<LogStatsAPI> => {
    const response = await api.get<LogStatsAPI>("/api/v1/admin/logs/stats");
    return response.data;
  },

  getExecutionLogs: async (
    executionId: string,
    afterLine?: number,
  ): Promise<{ entries: LogEntryAPI[]; count: number }> => {
    const params = afterLine ? `?after_line=${afterLine}` : "";
    const response = await api.get<{ entries: LogEntryAPI[]; count: number }>(
      `/api/v1/admin/logs/executions/${executionId}${params}`,
    );
    return response.data;
  },
};
