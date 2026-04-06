import type { Job, JobFunction, LogLevel } from "@/lib/api";
import type { ExecutionLogLevel } from "@/hooks/use-execution-logs";

export type { Job, JobFunction, LogLevel, ExecutionLogLevel };

export type CollapsedLog = {
  id: string;
  level: ExecutionLogLevel;
  message: string;
  count: number;
};

export type EditFormData = {
  description: string;
  code: string;
  timeout_seconds: number;
  max_retries: number;
  schedule: string;
};

export type DeleteConfirm = {
  namespace: string;
  name: string;
} | null;

export const LOG_LEVEL_COLORS: Record<ExecutionLogLevel, string> = {
  debug: "text-gray-400",
  info: "text-green-400",
  warn: "text-yellow-400",
  warning: "text-yellow-400",
  error: "text-red-400",
  fatal: "text-red-600 font-bold",
};

export const LOG_LEVEL_BADGE_COLORS: Record<ExecutionLogLevel, string> = {
  debug: "bg-gray-600",
  info: "bg-green-600",
  warn: "bg-yellow-600",
  warning: "bg-yellow-600",
  error: "bg-red-600",
  fatal: "bg-red-800",
};

export const LOG_LEVEL_PRIORITY_MAP: Record<ExecutionLogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  warning: 2,
  error: 3,
  fatal: 4,
};
