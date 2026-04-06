import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useBranchStore } from "@/stores/branch-store";

/**
 * Hook that invalidates all queries when the branch context changes.
 * This ensures that data is refreshed when switching between database branches.
 */
export function useBranchQueryRefresh() {
  const queryClient = useQueryClient();
  const currentBranch = useBranchStore((state) => state.currentBranch);

  useEffect(() => {
    // Invalidate all queries when branch changes
    // This will cause them to refetch with the new branch context
    // (X-Fluxbase-Branch header is automatically added by api.ts interceptor)
    queryClient.invalidateQueries();
  }, [currentBranch?.id, queryClient]);
}
