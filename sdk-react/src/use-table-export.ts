/**
 * Table Export Hooks
 *
 * React hooks for exporting database tables to knowledge bases.
 */

import { useState, useEffect, useCallback } from "react";
import { useFluxbaseClient } from "./context";
import type { TableDetails } from "@nimbleflux/fluxbase-sdk";

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
