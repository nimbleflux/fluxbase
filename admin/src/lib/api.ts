import { api, API_BASE_URL } from "./api/client";

export { api, API_BASE_URL };
export { api as apiClient } from "./api/client";
export default api;

export * from "./api/auth";
export * from "./api/branches";
export * from "./api/database";
export * from "./api/functions";
export * from "./api/jobs";
export * from "./api/knowledge-bases";
export * from "./api/logs";
export * from "./api/mcp";
export * from "./api/monitoring";
export * from "./api/policy";
export * from "./api/rpc";
export * from "./api/secrets";
export * from "./api/service-keys";
export * from "./api/storage";
export * from "./api/tenants";
export * from "./api/users";
export * from "./api/webhooks";

import type { User as AuthUser } from "./api/auth";
export type { AuthUser };
