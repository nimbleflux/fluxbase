import { api } from "./client";

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

export interface ListBranchesResponse {
  branches: Branch[];
  total: number;
  limit: number;
  offset: number;
}

export const branchesApi = {
  list: async (params?: {
    status?: BranchStatus;
    type?: BranchType;
    mine?: boolean;
  }): Promise<ListBranchesResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.set("status", params.status);
    if (params?.type) searchParams.set("type", params.type);
    if (params?.mine) searchParams.set("mine", "true");

    const queryString = searchParams.toString();
    const url = queryString
      ? `/api/v1/admin/branches?${queryString}`
      : "/api/v1/admin/branches";

    const response = await api.get<ListBranchesResponse>(url);
    return response.data;
  },

  get: async (id: string): Promise<Branch> => {
    const response = await api.get<Branch>(`/api/v1/admin/branches/${id}`);
    return response.data;
  },

  getActive: async (): Promise<{ branch: string; source: string }> => {
    const response = await api.get<{ branch: string; source: string }>(
      "/api/v1/admin/branches/active",
    );
    return response.data;
  },
};
