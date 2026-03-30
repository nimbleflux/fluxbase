/**
 * Table Export Hooks
 *
 * React hooks for exporting database tables to knowledge bases and managing sync configurations.
 */

import { useState, useEffect, useCallback } from "react";
import { useFluxbaseClient } from "./context";
import type {
  ExportTableOptions,
  ExportTableResult,
  TableDetails,
  TableExportSyncConfig,
  CreateTableExportSyncConfig,
  UpdateTableExportSyncConfig,
} from "@nimbleflux/fluxbase-sdk";

// ============================================================================
// useTableDetails Hook
// ============================================================================

export interface UseTableDetailsOptions {
  schema?: string;
  table?: string;
  autoFetch?: boolean;
}

export interface UseTableDetailsReturn {
  data: TableDetails | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

/**
 * Hook for fetching detailed table information including columns
 *
 * @example
 * ```tsx
 * function TableColumnsList({ schema, table }: { schema: string; table: string }) {
 *   const { data, isLoading, error } = useTableDetails({ schema, table })
 *
 *   if (isLoading) return <div>Loading...</div>
 *   if (error) return <div>Error: {error.message}</div>
 *
 *   return (
 *     <ul>
 *       {data?.columns.map(col => (
 *         <li key={col.name}>
 *           {col.name} ({col.data_type})
 *           {col.is_primary_key && ' 🔑'}
 *         </li>
 *       ))}
 *     </ul>
 *   )
 * }
 * ```
 */
export function useTableDetails(
  options: UseTableDetailsOptions,
): UseTableDetailsReturn {
  const { schema, table, autoFetch = true } = options;
  const client = useFluxbaseClient();

  const [data, setData] = useState<TableDetails | null>(null);
  const [isLoading, setIsLoading] = useState(autoFetch && !!schema && !!table);
  const [error, setError] = useState<Error | null>(null);

  const fetchData = useCallback(async () => {
    if (!schema || !table) return;

    try {
      setIsLoading(true);
      setError(null);
      const result = await client.admin.ai.getTableDetails(schema, table);
      if (result.error) {
        throw result.error;
      }
      setData(result.data);
    } catch (err) {
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
  }, [client, schema, table]);

  useEffect(() => {
    if (autoFetch && schema && table) {
      fetchData();
    }
  }, [autoFetch, fetchData, schema, table]);

  return {
    data,
    isLoading,
    error,
    refetch: fetchData,
  };
}

// ============================================================================
// useExportTable Hook
// ============================================================================

export interface UseExportTableReturn {
  exportTable: (
    options: ExportTableOptions,
  ) => Promise<ExportTableResult | null>;
  isLoading: boolean;
  error: Error | null;
  reset: () => void;
}

/**
 * Hook for exporting a table to a knowledge base
 *
 * @example
 * ```tsx
 * function ExportTableButton({ kbId, schema, table }: Props) {
 *   const { exportTable, isLoading, error } = useExportTable(kbId)
 *
 *   const handleExport = async () => {
 *     const result = await exportTable({
 *       schema,
 *       table,
 *       columns: ['id', 'name', 'email'],
 *       include_foreign_keys: true,
 *     })
 *     if (result) {
 *       console.log('Exported document:', result.document_id)
 *     }
 *   }
 *
 *   return (
 *     <button onClick={handleExport} disabled={isLoading}>
 *       {isLoading ? 'Exporting...' : 'Export Table'}
 *     </button>
 *   )
 * }
 * ```
 */
export function useExportTable(knowledgeBaseId: string): UseExportTableReturn {
  const client = useFluxbaseClient();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const exportTable = useCallback(
    async (options: ExportTableOptions): Promise<ExportTableResult | null> => {
      try {
        setIsLoading(true);
        setError(null);
        const result = await client.admin.ai.exportTable(
          knowledgeBaseId,
          options,
        );
        if (result.error) {
          throw result.error;
        }
        return result.data;
      } catch (err) {
        setError(err as Error);
        return null;
      } finally {
        setIsLoading(false);
      }
    },
    [client, knowledgeBaseId],
  );

  const reset = useCallback(() => {
    setError(null);
    setIsLoading(false);
  }, []);

  return {
    exportTable,
    isLoading,
    error,
    reset,
  };
}

// ============================================================================
// useTableExportSyncs Hook
// ============================================================================

export interface UseTableExportSyncsOptions {
  autoFetch?: boolean;
}

export interface UseTableExportSyncsReturn {
  configs: TableExportSyncConfig[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

/**
 * Hook for listing table export sync configurations
 *
 * @example
 * ```tsx
 * function SyncConfigsList({ kbId }: { kbId: string }) {
 *   const { configs, isLoading, error } = useTableExportSyncs(kbId)
 *
 *   if (isLoading) return <div>Loading...</div>
 *
 *   return (
 *     <ul>
 *       {configs.map(config => (
 *         <li key={config.id}>
 *           {config.schema_name}.{config.table_name} ({config.sync_mode})
 *         </li>
 *       ))}
 *     </ul>
 *   )
 * }
 * ```
 */
export function useTableExportSyncs(
  knowledgeBaseId: string,
  options: UseTableExportSyncsOptions = {},
): UseTableExportSyncsReturn {
  const { autoFetch = true } = options;
  const client = useFluxbaseClient();

  const [configs, setConfigs] = useState<TableExportSyncConfig[]>([]);
  const [isLoading, setIsLoading] = useState(autoFetch);
  const [error, setError] = useState<Error | null>(null);

  const fetchData = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const result =
        await client.admin.ai.listTableExportSyncs(knowledgeBaseId);
      if (result.error) {
        throw result.error;
      }
      setConfigs(result.data || []);
    } catch (err) {
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
  }, [client, knowledgeBaseId]);

  useEffect(() => {
    if (autoFetch) {
      fetchData();
    }
  }, [autoFetch, fetchData]);

  return {
    configs,
    isLoading,
    error,
    refetch: fetchData,
  };
}

// ============================================================================
// useCreateTableExportSync Hook
// ============================================================================

export interface UseCreateTableExportSyncReturn {
  createSync: (
    config: CreateTableExportSyncConfig,
  ) => Promise<TableExportSyncConfig | null>;
  isLoading: boolean;
  error: Error | null;
}

/**
 * Hook for creating a table export sync configuration
 *
 * @example
 * ```tsx
 * function CreateSyncForm({ kbId }: { kbId: string }) {
 *   const { createSync, isLoading, error } = useCreateTableExportSync(kbId)
 *
 *   const handleSubmit = async (e: React.FormEvent) => {
 *     e.preventDefault()
 *     const config = await createSync({
 *       schema_name: 'public',
 *       table_name: 'users',
 *       columns: ['id', 'name', 'email'],
 *       sync_mode: 'automatic',
 *       sync_on_insert: true,
 *       sync_on_update: true,
 *     })
 *     if (config) {
 *       console.log('Created sync config:', config.id)
 *     }
 *   }
 *
 *   return <form onSubmit={handleSubmit}>...</form>
 * }
 * ```
 */
export function useCreateTableExportSync(
  knowledgeBaseId: string,
): UseCreateTableExportSyncReturn {
  const client = useFluxbaseClient();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const createSync = useCallback(
    async (
      config: CreateTableExportSyncConfig,
    ): Promise<TableExportSyncConfig | null> => {
      try {
        setIsLoading(true);
        setError(null);
        const result = await client.admin.ai.createTableExportSync(
          knowledgeBaseId,
          config,
        );
        if (result.error) {
          throw result.error;
        }
        return result.data;
      } catch (err) {
        setError(err as Error);
        return null;
      } finally {
        setIsLoading(false);
      }
    },
    [client, knowledgeBaseId],
  );

  return {
    createSync,
    isLoading,
    error,
  };
}

// ============================================================================
// useUpdateTableExportSync Hook
// ============================================================================

export interface UseUpdateTableExportSyncReturn {
  updateSync: (
    syncId: string,
    updates: UpdateTableExportSyncConfig,
  ) => Promise<TableExportSyncConfig | null>;
  isLoading: boolean;
  error: Error | null;
}

/**
 * Hook for updating a table export sync configuration
 */
export function useUpdateTableExportSync(
  knowledgeBaseId: string,
): UseUpdateTableExportSyncReturn {
  const client = useFluxbaseClient();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const updateSync = useCallback(
    async (
      syncId: string,
      updates: UpdateTableExportSyncConfig,
    ): Promise<TableExportSyncConfig | null> => {
      try {
        setIsLoading(true);
        setError(null);
        const result = await client.admin.ai.updateTableExportSync(
          knowledgeBaseId,
          syncId,
          updates,
        );
        if (result.error) {
          throw result.error;
        }
        return result.data;
      } catch (err) {
        setError(err as Error);
        return null;
      } finally {
        setIsLoading(false);
      }
    },
    [client, knowledgeBaseId],
  );

  return {
    updateSync,
    isLoading,
    error,
  };
}

// ============================================================================
// useDeleteTableExportSync Hook
// ============================================================================

export interface UseDeleteTableExportSyncReturn {
  deleteSync: (syncId: string) => Promise<boolean>;
  isLoading: boolean;
  error: Error | null;
}

/**
 * Hook for deleting a table export sync configuration
 */
export function useDeleteTableExportSync(
  knowledgeBaseId: string,
): UseDeleteTableExportSyncReturn {
  const client = useFluxbaseClient();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const deleteSync = useCallback(
    async (syncId: string): Promise<boolean> => {
      try {
        setIsLoading(true);
        setError(null);
        const result = await client.admin.ai.deleteTableExportSync(
          knowledgeBaseId,
          syncId,
        );
        if (result.error) {
          throw result.error;
        }
        return true;
      } catch (err) {
        setError(err as Error);
        return false;
      } finally {
        setIsLoading(false);
      }
    },
    [client, knowledgeBaseId],
  );

  return {
    deleteSync,
    isLoading,
    error,
  };
}

// ============================================================================
// useTriggerTableExportSync Hook
// ============================================================================

export interface UseTriggerTableExportSyncReturn {
  triggerSync: (syncId: string) => Promise<ExportTableResult | null>;
  isLoading: boolean;
  error: Error | null;
}

/**
 * Hook for manually triggering a table export sync
 *
 * @example
 * ```tsx
 * function TriggerSyncButton({ kbId, syncId }: Props) {
 *   const { triggerSync, isLoading, error } = useTriggerTableExportSync(kbId)
 *
 *   const handleTrigger = async () => {
 *     const result = await triggerSync(syncId)
 *     if (result) {
 *       console.log('Sync completed:', result.document_id)
 *     }
 *   }
 *
 *   return (
 *     <button onClick={handleTrigger} disabled={isLoading}>
 *       {isLoading ? 'Syncing...' : 'Sync Now'}
 *     </button>
 *   )
 * }
 * ```
 */
export function useTriggerTableExportSync(
  knowledgeBaseId: string,
): UseTriggerTableExportSyncReturn {
  const client = useFluxbaseClient();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const triggerSync = useCallback(
    async (syncId: string): Promise<ExportTableResult | null> => {
      try {
        setIsLoading(true);
        setError(null);
        const result = await client.admin.ai.triggerTableExportSync(
          knowledgeBaseId,
          syncId,
        );
        if (result.error) {
          throw result.error;
        }
        return result.data;
      } catch (err) {
        setError(err as Error);
        return null;
      } finally {
        setIsLoading(false);
      }
    },
    [client, knowledgeBaseId],
  );

  return {
    triggerSync,
    isLoading,
    error,
  };
}
