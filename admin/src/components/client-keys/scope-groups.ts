import type { ScopeGroup } from "./types";

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
];
