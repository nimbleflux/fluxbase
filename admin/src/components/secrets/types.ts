import type { Secret, SecretVersion } from "@/lib/api";

export interface SecretsStatsCardsProps {
  total: number;
  expiringSoon: number;
  expired: number;
}

export interface SecretTableRowProps {
  secret: Secret;
  onEdit: (secret: Secret) => void;
  onHistory: (secret: Secret) => void;
  onDelete: (id: string) => void;
  isDeleting: boolean;
}

export interface CreateSecretDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: {
    name: string;
    value: string;
    scope: "global" | "namespace";
    namespace?: string;
    description?: string;
    expires_at?: string;
  }) => void;
  isPending: boolean;
}

export interface EditSecretDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  secret: Secret | null;
  onSubmit: (data: { value?: string; description?: string }) => void;
  isPending: boolean;
}

export interface VersionHistoryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  secret: Secret | null;
  versions: SecretVersion[] | undefined;
  onRollback: (id: string, version: number) => void;
  isRollbackPending: boolean;
}
