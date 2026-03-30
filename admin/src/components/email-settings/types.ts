import type { EmailProviderSettings } from "@nimbleflux/fluxbase-sdk";

export interface EmailTemplate {
  id: string;
  template_type: string;
  subject: string;
  html_body: string;
  text_body?: string;
  is_custom: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProviderFormState {
  from_address: string;
  from_name: string;
  smtp_host: string;
  smtp_port: string;
  smtp_username: string;
  smtp_password: string;
  smtp_tls: boolean;
  sendgrid_api_key: string;
  mailgun_api_key: string;
  mailgun_domain: string;
  ses_access_key: string;
  ses_secret_key: string;
  ses_region: string;
}

export type ProviderType = "smtp" | "sendgrid" | "mailgun" | "ses";

export interface ProviderFormProps {
  formState: ProviderFormState;
  settings: EmailProviderSettings | undefined;
  showPassword: boolean;
  onFormChange: (
    field: keyof ProviderFormState,
    value: string | boolean,
  ) => void;
  onTogglePassword: () => void;
}

export interface TemplateCardProps {
  template: EmailTemplate;
  onEdit: (template: EmailTemplate) => void;
  onReset: (type: string) => void;
  onTest: (type: string) => void;
  isResetting: boolean;
  isTesting: boolean;
}

export interface TemplateEditorProps {
  selectedTemplate: string;
  editingTemplate: Partial<EmailTemplate> | null;
  isSaving: boolean;
  onSave: () => void;
  onCancel: () => void;
  onUpdate: (template: Partial<EmailTemplate>) => void;
}
