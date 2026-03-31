import { Loader2, StopCircle, Copy } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { type RPCExecution } from "@/lib/api";
import { type ExecutionLog } from "@/hooks/use-execution-logs";
import {
  getStatusIcon,
  getStatusVariant,
  canCancelExecution,
} from "./execution-utils";

interface ExecutionDetailsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  execution: RPCExecution | null;
  logs: ExecutionLog[];
  loadingLogs: boolean;
  cancellingExecutionId: string | null;
  onCancelExecution: (executionId: string) => void;
  onCopy: (text: string, label: string) => void;
}

export function ExecutionDetailsDialog({
  open,
  onOpenChange,
  execution,
  logs,
  loadingLogs,
  cancellingExecutionId,
  onCancelExecution,
  onCopy,
}: ExecutionDetailsDialogProps) {
  if (!execution) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-4xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {getStatusIcon(execution.status)}
            Execution Details
          </DialogTitle>
          <DialogDescription>
            {execution.procedure_name} - {execution.id}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant={getStatusVariant(execution.status)}>
              {execution.status}
            </Badge>
            {execution.user_email && (
              <Badge variant="outline">{execution.user_email}</Badge>
            )}
            {execution.user_role && (
              <Badge variant="outline">role: {execution.user_role}</Badge>
            )}
            {execution.is_async && <Badge variant="outline">async</Badge>}
            {canCancelExecution(execution.status) && (
              <Button
                variant="destructive"
                size="sm"
                className="ml-auto"
                onClick={() => onCancelExecution(execution.id)}
                disabled={cancellingExecutionId === execution.id}
              >
                {cancellingExecutionId === execution.id ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <StopCircle className="mr-2 h-4 w-4" />
                )}
                Cancel Execution
              </Button>
            )}
          </div>

          <div className="grid grid-cols-3 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">Created:</span>
              <p>{new Date(execution.created_at).toLocaleString()}</p>
            </div>
            {execution.started_at && (
              <div>
                <span className="text-muted-foreground">Started:</span>
                <p>{new Date(execution.started_at).toLocaleString()}</p>
              </div>
            )}
            {execution.completed_at && (
              <div>
                <span className="text-muted-foreground">Completed:</span>
                <p>{new Date(execution.completed_at).toLocaleString()}</p>
              </div>
            )}
          </div>

          <div className="flex gap-4 text-sm">
            {execution.duration_ms !== undefined && (
              <div>
                <span className="text-muted-foreground">Duration: </span>
                <span className="font-medium">{execution.duration_ms}ms</span>
              </div>
            )}
            {execution.rows_returned !== undefined && (
              <div>
                <span className="text-muted-foreground">Rows Returned: </span>
                <span className="font-medium">{execution.rows_returned}</span>
              </div>
            )}
          </div>

          {execution.input_params &&
            Object.keys(execution.input_params).length > 0 && (
              <div>
                <div className="mb-2 flex items-center justify-between">
                  <Label>Input Parameters</Label>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() =>
                      onCopy(
                        JSON.stringify(execution.input_params, null, 2),
                        "Input params",
                      )
                    }
                  >
                    <Copy className="h-3 w-3" />
                  </Button>
                </div>
                <pre className="bg-muted max-h-32 overflow-auto rounded-md p-3 text-xs">
                  {JSON.stringify(execution.input_params, null, 2)}
                </pre>
              </div>
            )}

          {execution.result !== undefined && execution.result !== null && (
            <div>
              <div className="mb-2 flex items-center justify-between">
                <Label>Result</Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() =>
                    onCopy(JSON.stringify(execution.result, null, 2), "Result")
                  }
                >
                  <Copy className="h-3 w-3" />
                </Button>
              </div>
              <pre className="bg-muted max-h-48 overflow-auto rounded-md p-3 text-xs">
                {typeof execution.result === "string"
                  ? execution.result
                  : JSON.stringify(execution.result, null, 2)}
              </pre>
            </div>
          )}

          {execution.error_message && (
            <div>
              <Label className="text-destructive">Error</Label>
              <div className="bg-destructive/10 mt-2 rounded-md p-3">
                <p className="text-destructive text-sm">
                  {execution.error_message}
                </p>
              </div>
            </div>
          )}

          <div>
            <div className="mb-2 flex items-center justify-between">
              <Label>Logs</Label>
              {logs.length > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() =>
                    onCopy(
                      logs.map((l) => `[${l.level}] ${l.message}`).join("\n"),
                      "Logs",
                    )
                  }
                >
                  <Copy className="h-3 w-3" />
                </Button>
              )}
            </div>
            <ScrollArea className="bg-muted h-48 rounded-md border p-3">
              {loadingLogs ? (
                <div className="flex h-full items-center justify-center">
                  <Loader2 className="text-muted-foreground h-6 w-6 animate-spin" />
                </div>
              ) : logs.length === 0 ? (
                <p className="text-muted-foreground text-center text-sm">
                  No logs available
                </p>
              ) : (
                <div className="space-y-1">
                  {logs.map((log) => (
                    <div key={log.id} className="font-mono text-xs">
                      <span
                        className={
                          log.level === "error"
                            ? "text-red-500"
                            : log.level === "warn"
                              ? "text-yellow-500"
                              : log.level === "info"
                                ? "text-blue-500"
                                : "text-muted-foreground"
                        }
                      >
                        [{log.level}]
                      </span>
                      <span className="ml-2">{log.message}</span>
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
