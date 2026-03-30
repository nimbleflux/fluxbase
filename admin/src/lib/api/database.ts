import { api } from "./client";

export interface TableInfo {
  schema: string;
  name: string;
  type: "table" | "view" | "materialized_view";
  rest_path?: string;
  columns: Array<{
    name: string;
    data_type: string;
    is_nullable: boolean;
    default_value: string | null;
    is_primary_key: boolean;
    is_foreign_key: boolean;
    is_unique: boolean;
    max_length: number | null;
    position: number;
  }>;
  primary_key: string[];
  foreign_keys: unknown;
  indexes: unknown;
  rls_enabled: boolean;
}

export const databaseApi = {
  getSchemas: async (): Promise<string[]> => {
    const response = await api.get<string[]>("/api/v1/admin/schemas");
    return response.data;
  },

  getTables: async (schema?: string): Promise<TableInfo[]> => {
    const url = schema
      ? `/api/v1/admin/tables?schema=${encodeURIComponent(schema)}`
      : "/api/v1/admin/tables";
    const response = await api.get<TableInfo[]>(url);
    return response.data;
  },

  createSchema: async (
    name: string,
  ): Promise<{ success: boolean; schema: string; message: string }> => {
    const response = await api.post("/api/v1/admin/schemas", { name });
    return response.data;
  },

  createTable: async (data: {
    schema: string;
    name: string;
    columns: Array<{
      name: string;
      type: string;
      nullable: boolean;
      primaryKey: boolean;
      defaultValue: string;
    }>;
  }): Promise<{
    success: boolean;
    schema: string;
    table: string;
    message: string;
  }> => {
    const response = await api.post("/api/v1/admin/tables", data);
    return response.data;
  },

  deleteTable: async (
    schema: string,
    table: string,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.delete(
      `/api/v1/admin/tables/${schema}/${table}`,
    );
    return response.data;
  },

  renameTable: async (
    schema: string,
    table: string,
    newName: string,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.patch(
      `/api/v1/admin/tables/${schema}/${table}`,
      { newName },
    );
    return response.data;
  },

  addColumn: async (
    schema: string,
    table: string,
    column: {
      name: string;
      type: string;
      nullable: boolean;
      defaultValue?: string;
    },
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.post(
      `/api/v1/admin/tables/${schema}/${table}/columns`,
      column,
    );
    return response.data;
  },

  dropColumn: async (
    schema: string,
    table: string,
    column: string,
  ): Promise<{ success: boolean; message: string }> => {
    const response = await api.delete(
      `/api/v1/admin/tables/${schema}/${table}/columns/${column}`,
    );
    return response.data;
  },

  getTableData: async <T = unknown>(
    table: string,
    params?: {
      limit?: number;
      offset?: number;
      order?: string;
      select?: string;
      filter?: Record<string, unknown>;
    },
  ): Promise<T[]> => {
    const response = await api.get<T[]>(`/api/v1/tables/${table}`, { params });
    return response.data;
  },

  createRecord: async <T = unknown>(
    table: string,
    data: Record<string, unknown>,
  ): Promise<T> => {
    const response = await api.post<T>(`/api/v1/tables/${table}`, data);
    return response.data;
  },

  updateRecord: async <T = unknown>(
    table: string,
    id: string | number,
    data: Record<string, unknown>,
  ): Promise<T> => {
    const response = await api.patch<T>(`/api/v1/tables/${table}/${id}`, data);
    return response.data;
  },

  deleteRecord: async (table: string, id: string | number): Promise<void> => {
    await api.delete(`/api/v1/tables/${table}/${id}`);
  },

  getTableSchema: async (
    table: string,
  ): Promise<{
    columns: Array<{
      name: string;
      type: string;
      nullable: boolean;
      default: string | null;
      primary_key: boolean;
    }>;
  }> => {
    const response = await api.get(`/api/admin/tables/${table}/schema`);
    return response.data;
  },
};

export interface SQLExecuteRequest {
  query: string;
}

export interface SQLResult {
  columns?: string[];
  rows?: Array<Record<string, unknown>>;
  row_count: number;
  affected_rows?: number;
  execution_time_ms: number;
  error?: string;
  statement: string;
}

export interface SQLExecuteResponse {
  results: SQLResult[];
}

export const sqlApi = {
  execute: async (query: string): Promise<SQLExecuteResponse> => {
    const response = await api.post<SQLExecuteResponse>(
      "/admin/api/sql/execute",
      { query },
    );
    return response.data;
  },
};

export interface BulkActionRequest {
  action: "delete" | "export";
  targets: string[];
  table: string;
}

export interface BulkActionResponse {
  success: boolean;
  message?: string;
  rows_affected?: number;
  count?: number;
  records?: Array<Record<string, unknown>>;
}

export const bulkOperationsApi = {
  execute: async (request: BulkActionRequest): Promise<BulkActionResponse> => {
    const response = await api.post<BulkActionResponse>(
      "/api/v1/bulk",
      request,
    );
    return response.data;
  },

  delete: async (
    table: string,
    targetIds: string[],
  ): Promise<BulkActionResponse> => {
    return bulkOperationsApi.execute({
      action: "delete",
      targets: targetIds,
      table,
    });
  },

  export: async (
    table: string,
    targetIds: string[],
  ): Promise<BulkActionResponse> => {
    return bulkOperationsApi.execute({
      action: "export",
      targets: targetIds,
      table,
    });
  },
};

import { getAccessToken } from "../auth";

const API_BASE_URL =
  window.__FLUXBASE_CONFIG__?.publicBaseURL ||
  import.meta.env.VITE_API_URL ||
  "";

export const dataExportApi = {
  export: async (
    table: string,
    targetIds: string[],
    format: "csv" | "json" = "csv",
  ): Promise<BulkActionResponse | string> => {
    const items = JSON.stringify(targetIds);
    const url = `/api/v1/export?table=${table}&items=${encodeURIComponent(items)}&format=${format}`;

    if (format === "json") {
      const response = await api.get(url);
      return response.data as BulkActionResponse;
    }

    const response = await fetch(`${API_BASE_URL}${url}`, {
      headers: {
        Authorization: `Bearer ${getAccessToken()}`,
      },
    });

    if (!response.ok) {
      throw new Error("Export failed");
    }

    const blob = await response.blob();
    const downloadUrl = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = downloadUrl;
    a.download = `export_${Date.now()}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(downloadUrl);

    return "Download started";
  },
};

export interface JSONBProperty {
  type: string;
  description?: string;
  properties?: Record<string, JSONBProperty>;
  items?: JSONBProperty;
}

export interface JSONBSchema {
  properties?: Record<string, JSONBProperty>;
  required?: string[];
}

export interface SchemaNodeColumn {
  name: string;
  data_type: string;
  nullable: boolean;
  is_primary_key: boolean;
  is_foreign_key: boolean;
  fk_target?: string;
  default_value?: string;
  is_unique: boolean;
  is_indexed: boolean;
  comment?: string;
  description?: string;
  jsonb_schema?: JSONBSchema;
}

export interface SchemaNode {
  schema: string;
  name: string;
  columns: SchemaNodeColumn[];
  primary_key: string[];
  rls_enabled: boolean;
  force_rls: boolean;
  row_estimate?: number;
  comment?: string;
  incoming_rel_count: number;
  outgoing_rel_count: number;
}

export interface SchemaRelationship {
  id: string;
  source_schema: string;
  source_table: string;
  source_column: string;
  target_schema: string;
  target_table: string;
  target_column: string;
  constraint_name: string;
  on_delete: string;
  on_update: string;
  cardinality: "one-to-one" | "many-to-one" | "one-to-many";
}

export interface SchemaGraphResponse {
  nodes: SchemaNode[];
  edges: SchemaRelationship[];
  schemas: string[];
}

export interface TableRelationshipsResponse {
  schema: string;
  table: string;
  outgoing: Array<{
    direction: string;
    constraint_name: string;
    local_column: string;
    foreign_schema: string;
    foreign_table: string;
    foreign_column: string;
    delete_rule: string;
    update_rule: string;
  }>;
  incoming: Array<{
    direction: string;
    constraint_name: string;
    local_column: string;
    foreign_schema: string;
    foreign_table: string;
    foreign_column: string;
    delete_rule: string;
    update_rule: string;
  }>;
}

export const schemaApi = {
  getGraph: async (schemas?: string[]): Promise<SchemaGraphResponse> => {
    const params = schemas?.length ? `?schemas=${schemas.join(",")}` : "";
    const response = await api.get<SchemaGraphResponse>(
      `/api/v1/admin/schema/graph${params}`,
    );
    return response.data;
  },

  getTableRelationships: async (
    schema: string,
    table: string,
  ): Promise<TableRelationshipsResponse> => {
    const response = await api.get<TableRelationshipsResponse>(
      `/api/v1/admin/tables/${schema}/${table}/relationships`,
    );
    return response.data;
  },
};
