import { api } from "./client";
import type { User } from "./auth";
export interface TenantAssignment {
  tenant_id: string;
  tenant_name: string;
  tenant_slug: string;
}
export interface EnrichedUser {
  id: string;
  email: string;
  email_verified: boolean;
  role: string;
  provider: "email" | "invite_pending" | "magic_link";
  active_sessions: number;
  last_sign_in: string | null;
  is_locked: boolean;
  user_metadata: Record<string, unknown> | null;
  app_metadata: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
  tenant_assignments?: TenantAssignment[];
}
export interface InviteUserRequest {
  email: string;
  role: string;
  password?: string;
  skip_email?: boolean;
}
export interface InviteUserResponse {
  user: User;
  temporary_password?: string;
  email_sent: boolean;
  message: string;
}
const mapUserType = (userType: "app" | "dashboard"): "app" | "platform" => {
  return userType === "dashboard" ? "platform" : userType;
};
export const userManagementApi = {
  listUsers: async (
    userType: "app" | "dashboard" = "app",
  ): Promise<{ users: EnrichedUser[]; total: number }> => {
    const response = await api.get<{ users: EnrichedUser[]; total: number }>(
      "/api/v1/admin/users",
      {
        params: { type: mapUserType(userType) },
      },
    );
    return response.data;
  },
  inviteUser: async (
    data: InviteUserRequest,
    userType: "app" | "dashboard" = "app",
  ): Promise<InviteUserResponse> => {
    const response = await api.post<InviteUserResponse>(
      "/api/v1/admin/users/invite",
      data,
      { params: { type: mapUserType(userType) } },
    );
    return response.data;
  },
  deleteUser: async (
    userId: string,
    userType: "app" | "dashboard" = "app",
  ): Promise<{ message: string }> => {
    const response = await api.delete<{ message: string }>(
      `/api/v1/admin/users/${userId}`,
      { params: { type: mapUserType(userType) } },
    );
    return response.data;
  },
  updateUserRole: async (
    userId: string,
    role: string,
    userType: "app" | "dashboard" = "app",
  ): Promise<User> => {
    const response = await api.patch<User>(
      `/api/v1/admin/users/${userId}/role`,
      {
        role,
      },
      {
        params: { type: mapUserType(userType) },
      },
    );
    return response.data;
  },
  resetUserPassword: async (
    userId: string,
    userType: "app" | "dashboard" = "app",
  ): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>(
      `/api/v1/admin/users/${userId}/reset-password`,
      {},
      { params: { type: mapUserType(userType) } },
    );
    return response.data;
  },
  updateUser: async (
    userId: string,
    data: {
      email?: string;
      role?: string;
      password?: string;
      user_metadata?: Record<string, unknown>;
    },
    userType: "app" | "dashboard" = "app",
  ): Promise<EnrichedUser> => {
    const response = await api.patch<EnrichedUser>(
      `/api/v1/admin/users/${userId}`,
      data,
      {
        params: { type: mapUserType(userType) },
      },
    );
    return response.data;
  },
};
