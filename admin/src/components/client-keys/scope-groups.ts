import type { ScopeGroup } from "./types";

export const SCOPE_GROUPS: ScopeGroup[] = [
  {
    name: "Tables",
    description: "Database table access",
    scopes: [
      {
        id: "tables:read",
        label: "Read",
        description: "Query database tables",
      },
      {
        id: "tables:write",
        label: "Write",
        description: "Insert, update, delete records",
      },
    ],
  },
  {
    name: "Storage",
    description: "File storage access",
    scopes: [
      { id: "storage:read", label: "Read", description: "Download files" },
      {
        id: "storage:write",
        label: "Write",
        description: "Upload and delete files",
      },
    ],
  },
  {
    name: "Functions",
    description: "Edge Functions",
    scopes: [
      { id: "functions:read", label: "Read", description: "View functions" },
      {
        id: "functions:execute",
        label: "Execute",
        description: "Invoke functions",
      },
    ],
  },
  {
    name: "Auth",
    description: "Authentication",
    scopes: [
      { id: "auth:read", label: "Read", description: "View user profile" },
      {
        id: "auth:write",
        label: "Write",
        description: "Update profile, manage 2FA",
      },
    ],
  },
  {
    name: "Client Keys",
    description: "API key management",
    scopes: [
      { id: "clientkeys:read", label: "Read", description: "List client keys" },
      {
        id: "clientkeys:write",
        label: "Write",
        description: "Create, update, revoke",
      },
    ],
  },
  {
    name: "Webhooks",
    description: "Webhook management",
    scopes: [
      { id: "webhooks:read", label: "Read", description: "List webhooks" },
      {
        id: "webhooks:write",
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
        id: "monitoring:read",
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
      { id: "rpc:read", label: "Read", description: "List procedures" },
      { id: "rpc:execute", label: "Execute", description: "Invoke procedures" },
    ],
  },
  {
    name: "Jobs",
    description: "Background jobs",
    scopes: [
      { id: "jobs:read", label: "Read", description: "View job queues" },
      { id: "jobs:write", label: "Write", description: "Manage job entries" },
    ],
  },
  {
    name: "AI",
    description: "AI & chatbots",
    scopes: [
      { id: "ai:read", label: "Read", description: "View conversations" },
      { id: "ai:write", label: "Write", description: "Send messages" },
    ],
  },
  {
    name: "Secrets",
    description: "Secret management",
    scopes: [
      { id: "secrets:read", label: "Read", description: "View secret names" },
      {
        id: "secrets:write",
        label: "Write",
        description: "Create, update, delete",
      },
    ],
  },
];
