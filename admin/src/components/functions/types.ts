import type { EdgeFunction, EdgeFunctionExecution } from "@/lib/api";

export interface FunctionFormData {
  name: string;
  description: string;
  code: string;
  timeout_seconds: number;
  memory_limit_mb: number;
  allow_net: boolean;
  allow_env: boolean;
  allow_read: boolean;
  allow_write: boolean;
  cron_schedule: string;
}

export interface InvokeResult {
  success: boolean;
  data: string;
  error?: string;
}

export type { EdgeFunction, EdgeFunctionExecution };

export const DEFAULT_FUNCTION_CODE = `interface Request {
  method: string;
  url: string;
  headers: Record<string, string>;
  body: string;
}

async function handler(req: Request) {
  // Your code here
  const data = JSON.parse(req.body || "{}");

  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ message: "Hello from edge function!" })
  };
}`;

export const DEFAULT_FORM_DATA: FunctionFormData = {
  name: "",
  description: "",
  code: DEFAULT_FUNCTION_CODE,
  timeout_seconds: 30,
  memory_limit_mb: 128,
  allow_net: true,
  allow_env: true,
  allow_read: false,
  allow_write: false,
  cron_schedule: "",
};
