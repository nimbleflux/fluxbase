import type { TenantMembership } from "@/lib/api";

export type MemberRole = "tenant_admin" | "tenant_member";

export interface TenantMembersTabProps {
  tenantId: string;
  tenant: {
    id: string;
    name: string;
    slug: string;
    is_default: boolean;
    created_at: string;
  };
  members: TenantMembership[] | undefined;
  membersLoading: boolean;
  onAddMember: () => void;
  onUpdateMemberRole: (userId: string, role: MemberRole) => void;
  onRemoveMember: (userId: string) => void;
}
