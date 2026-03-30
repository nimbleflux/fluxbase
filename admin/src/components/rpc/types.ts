import type { RPCProcedure, RPCExecution } from "@/lib/api";

export type { RPCProcedure, RPCExecution };

export type ExecutionStatus =
  | "pending"
  | "running"
  | "completed"
  | "failed"
  | "cancelled"
  | "timeout";

export interface ExecutionStats {
  pending: number;
  running: number;
  completed: number;
  failed: number;
  total: number;
  avgDuration: number;
}
