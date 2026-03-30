import type { ClientKey } from "@/lib/api";

export interface ClientKeyWithPlaintext extends ClientKey {
  key: string;
}

export interface ScopeItem {
  id: string;
  label: string;
  description: string;
}

export interface ScopeGroup {
  name: string;
  description: string;
  scopes: ScopeItem[];
}

export interface ClientKeyStatsCardsProps {
  total: number;
  active: number;
  revoked: number;
}

export interface ClientKeyTableRowProps {
  clientKey: ClientKey;
  onRevoke: (id: string) => void;
  onDelete: (id: string) => void;
  isRevoking: boolean;
  isDeleting: boolean;
}

export interface CreateClientKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  scopeGroups: ScopeGroup[];
  selectedScopes: string[];
  onToggleScope: (scopeId: string) => void;
  onSubmit: () => void;
  isPending: boolean;
  name: string;
  onNameChange: (name: string) => void;
  description: string;
  onDescriptionChange: (desc: string) => void;
  rateLimit: number;
  onRateLimitChange: (limit: number) => void;
  expiresAt: string;
  onExpiresAtChange: (expires: string) => void;
}

export interface ShowCreatedKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  createdKey: ClientKeyWithPlaintext | null;
  onCopy: (text: string) => void;
}
