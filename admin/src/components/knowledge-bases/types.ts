import type {
  KnowledgeBaseSummary,
  CreateKnowledgeBaseRequest,
  KBPermission,
  EnrichedUser,
  AIProvider,
} from "@/lib/api";

export type {
  KnowledgeBaseSummary,
  CreateKnowledgeBaseRequest,
  KBPermission,
  EnrichedUser,
  AIProvider,
};

export interface FeatureDisabledError {
  error: string;
  code: string;
  feature_key?: string;
}

export interface CreateKnowledgeBaseDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  newKB: CreateKnowledgeBaseRequest;
  onNewKBChange: (kb: CreateKnowledgeBaseRequest) => void;
  onCreate: () => void;
  users: EnrichedUser[];
  usersLoading: boolean;
  providers: AIProvider[];
  providersLoading: boolean;
}

export interface KnowledgeBaseCardProps {
  kb: KnowledgeBaseSummary;
  onToggleEnabled: (kb: KnowledgeBaseSummary) => void;
  onDelete: (id: string) => void;
  onNavigate: (path: string) => void;
}
