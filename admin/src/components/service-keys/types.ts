import type { ServiceKey } from "@/lib/api";

export interface ScopeGroup {
  name: string;
  description: string;
  scopes: Array<{
    id: string;
    label: string;
    description: string;
  }>;
}

export const SCOPE_GROUPS: ScopeGroup[] = [
  {
    name: "Tables",
    description: "Database table access",
    scopes: [
      {
        id: "read:tables",
        label: "Read",
        description: "Query database tables",
      },
      {
        id: "write:tables",
        label: "Write",
        description: "Insert, update, delete records",
      },
    ],
  },
  {
    name: "Storage",
    description: "File storage access",
    scopes: [
      { id: "read:storage", label: "Read", description: "Download files" },
      {
        id: "write:storage",
        label: "Write",
        description: "Upload and delete files",
      },
    ],
  },
  {
    name: "Functions",
    description: "Edge Functions",
    scopes: [
      { id: "read:functions", label: "Read", description: "View functions" },
      {
        id: "execute:functions",
        label: "Execute",
        description: "Invoke functions",
      },
    ],
  },
  {
    name: "Auth",
    description: "Authentication",
    scopes: [
      { id: "read:auth", label: "Read", description: "View user profile" },
      {
        id: "write:auth",
        label: "Write",
        description: "Update profile, manage 2FA",
      },
    ],
  },
  {
    name: "Client Keys",
    description: "API key management",
    scopes: [
      { id: "read:clientkeys", label: "Read", description: "List client keys" },
      {
        id: "write:clientkeys",
        label: "Write",
        description: "Create, update, revoke",
      },
    ],
  },
  {
    name: "Webhooks",
    description: "Webhook management",
    scopes: [
      { id: "read:webhooks", label: "Read", description: "List webhooks" },
      {
        id: "write:webhooks",
        label: "Write",
        description: "Create, update, delete",
      },
    ],
  },
  {
    name: "Monitoring",
    description: "System monitoring",
    scopes: [
      {
        id: "read:monitoring",
        label: "Read",
        description: "View metrics, health, logs",
      },
    ],
  },
  {
    name: "Realtime",
    description: "WebSocket channels",
    scopes: [
      {
        id: "realtime:connect",
        label: "Connect",
        description: "Connect to channels",
      },
      {
        id: "realtime:broadcast",
        label: "Broadcast",
        description: "Send messages",
      },
    ],
  },
  {
    name: "RPC",
    description: "Remote procedures",
    scopes: [
      { id: "read:rpc", label: "Read", description: "List procedures" },
      { id: "execute:rpc", label: "Execute", description: "Invoke procedures" },
    ],
  },
  {
    name: "Jobs",
    description: "Background jobs",
    scopes: [
      { id: "read:jobs", label: "Read", description: "View job queues" },
      { id: "write:jobs", label: "Write", description: "Manage job entries" },
    ],
  },
  {
    name: "AI",
    description: "AI & chatbots",
    scopes: [
      { id: "read:ai", label: "Read", description: "View conversations" },
      { id: "write:ai", label: "Write", description: "Send messages" },
    ],
  },
  {
    name: "Secrets",
    description: "Secret management",
    scopes: [
      { id: "read:secrets", label: "Read", description: "View secret names" },
      {
        id: "write:secrets",
        label: "Write",
        description: "Create, update, delete",
      },
    ],
  },
  {
    name: "Migrations",
    description: "Database migrations",
    scopes: [
      {
        id: "migrations:read",
        label: "Read",
        description: "View migration status",
      },
      {
        id: "migrations:execute",
        label: "Execute",
        description: "Apply migrations",
      },
    ],
  },
];

export const isExpired = (expiresAt?: string): boolean => {
  if (!expiresAt) return false;
  return new Date(expiresAt) < new Date();
};

export const getKeyStatus = (
  key: ServiceKey,
): {
  label: string;
  variant: "destructive" | "outline" | "secondary" | "default";
} => {
  if (key.revoked_at) return { label: "Revoked", variant: "destructive" };
  if (key.deprecated_at) {
    if (
      key.grace_period_ends_at &&
      new Date(key.grace_period_ends_at) > new Date()
    ) {
      return { label: "Deprecated", variant: "outline" };
    }
    return { label: "Expired", variant: "destructive" };
  }
  if (!key.enabled) return { label: "Disabled", variant: "secondary" };
  if (isExpired(key.expires_at))
    return { label: "Expired", variant: "destructive" };
  return { label: "Active", variant: "default" };
};

export const canModify = (key: ServiceKey): boolean => !key.revoked_at;

export const formatRateLimit = (key: ServiceKey): string => {
  const parts: string[] = [];
  if (key.rate_limit_per_minute) {
    parts.push(`${key.rate_limit_per_minute}/min`);
  }
  if (key.rate_limit_per_hour) {
    parts.push(`${key.rate_limit_per_hour}/hr`);
  }
  return parts.length > 0 ? parts.join(", ") : "Unlimited";
};
