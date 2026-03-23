import { create } from "zustand";
import { persist } from "zustand/middleware";

export type BranchStatus =
  | "creating"
  | "ready"
  | "migrating"
  | "error"
  | "deleting"
  | "deleted";
export type BranchType = "main" | "preview" | "production" | "persistent";

export interface Branch {
  id: string;
  name: string;
  slug: string;
  database_name: string;
  status: BranchStatus;
  type: BranchType;
  parent_branch_id?: string;
  data_clone_mode?: string;
  github_pr_number?: number;
  github_pr_url?: string;
  github_repo?: string;
  error_message?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
  expires_at?: string;
}

interface BranchState {
  currentBranch: Branch | null;
  branches: Branch[];
  isBranchingEnabled: boolean;
  isLoading: boolean;
  setCurrentBranch: (branch: Branch | null) => void;
  setBranches: (branches: Branch[]) => void;
  setIsBranchingEnabled: (enabled: boolean) => void;
  setIsLoading: (loading: boolean) => void;
  clearBranch: () => void;
}

export const useBranchStore = create<BranchState>()(
  persist(
    (set) => ({
      currentBranch: null,
      branches: [],
      isBranchingEnabled: false,
      isLoading: true,
      setCurrentBranch: (branch) =>
        set((state) => ({
          ...state,
          currentBranch: branch,
        })),
      setBranches: (branches) =>
        set((state) => {
          // If no persisted branch, default to main or first ready branch
          if (!state.currentBranch) {
            const mainBranch = branches.find(
              (b) => b.type === "main" && b.status === "ready",
            );
            const firstReady = branches.find((b) => b.status === "ready");
            return {
              ...state,
              branches,
              currentBranch: mainBranch || firstReady || null,
            };
          }
          // Restore previously selected branch if it still exists and is ready
          const existingBranch = branches.find(
            (b) => b.id === state.currentBranch?.id && b.status === "ready",
          );
          return {
            ...state,
            branches,
            currentBranch:
              existingBranch ||
              branches.find((b) => b.status === "ready") ||
              null,
          };
        }),
      setIsBranchingEnabled: (enabled) =>
        set((state) => ({
          ...state,
          isBranchingEnabled: enabled,
        })),
      setIsLoading: (loading) =>
        set((state) => ({
          ...state,
          isLoading: loading,
        })),
      clearBranch: () =>
        set((state) => ({
          ...state,
          currentBranch: null,
        })),
    }),
    {
      name: "fluxbase-branch-store",
      partialize: (state) => ({
        currentBranch: state.currentBranch,
      }),
    },
  ),
);
