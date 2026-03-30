import type {
  RLSPolicy,
  SecurityWarning,
  PolicyTemplate,
  CreatePolicyRequest,
} from "@/lib/api";

export type { RLSPolicy, SecurityWarning, PolicyTemplate, CreatePolicyRequest };

export interface TableWithRLS {
  schema: string;
  table: string;
  rls_enabled: boolean;
  rls_forced: boolean;
  policy_count: number;
}

export interface TableDetails {
  rls_enabled: boolean;
  rls_forced: boolean;
  policies: RLSPolicy[];
}
