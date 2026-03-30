import { useRef, useMemo } from "react";
import {
  CheckCircle,
  AlertCircle,
  Loader2,
  Clock,
  XCircle,
  Activity,
  Copy,
  RefreshCw,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import {
  LOG_LEVEL_COLORS,
  LOG_LEVEL_BADGE_COLORS,
  LOG_LEVEL_PRIORITY_MAP,
  type Job,
  type LogLevel,
  type CollapsedLog,
  type ExecutionLogLevel,
} from "./types";
import type { ExecutionLog } from "@/hooks/use-execution-logs";

const getStatusIcon = (status: string) => {
  switch (status) {
    case "completed":
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    case "failed":
      return <AlertCircle className="h-4 w-4 text-red-500" />;
    case "running":
      return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
    case "pending":
      return <Clock className="h-4 w-4 text-yellow-500" />;
    case "cancelled":
      return <XCircle className="h-4 w-4 text-gray-500" />;
    default:
      return <Activity className="h-4 w-4" />;
  }
};

const getStatusBadgeVariant = (
  status: string,
): "default" | "secondary" | "destructive" | "outline" => {
  switch (status) {
    case "completed":
      return "default";
    case "failed":
      return "destructive";
    case "running":
      return "secondary";
    default:
      return "outline";
  }
};

const formatJsonValue = (value: unknown): string => {
  if (value === null || value === undefined) {
    return "";
  }
  if (typeof value === "string") {
    try {
      const parsed = JSON.parse(value);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return value;
    }
  }
  return JSON.stringify(value, null, 2);
};

const collapseConsecutiveLogs = (logs: ExecutionLog[]): CollapsedLog[] => {
  if (logs.length === 0) return [];

  const result: CollapsedLog[] = [];
  let currentLog = logs[0];
  let count = 1;

  for (let i = 1; i < logs.length; i++) {
    if (
      logs[i].message === currentLog.message &&
      logs[i].level === currentLog.level
    ) {
      count++;
    } else {
      result.push({
        id: `log-${currentLog.id}-${count}`,
        level: currentLog.level || "info",
        message: currentLog.message,
        count,
      });
      currentLog = logs[i];
      count = 1;
    }
  }
  result.push({
    id: `log-${currentLog.id}-${count}`,
    level: currentLog.level || "info",
    message: currentLog.message,
    count,
  });

  return result;
};

const filterLogsByLevel = (
  logs: ExecutionLog[],
  logLevelFilter: LogLevel | "all",
): ExecutionLog[] => {
  if (logLevelFilter === "all") return logs;
  const filterLevel = logLevelFilter === "warning" ? "warn" : logLevelFilter;
  const minPriority =
    LOG_LEVEL_PRIORITY_MAP[filterLevel as ExecutionLogLevel] ?? 0;
  return logs.filter(
    (log) => LOG_LEVEL_PRIORITY_MAP[log.level || "info"] >= minPriority,
  );
};

const formatLogsForClipboard = (logs: ExecutionLog[]): string => {
  return collapseConsecutiveLogs(logs)
    .map((log) => {
      const prefix = log.count > 1 ? `(${log.count}x) ` : "";
      return `[${log.level.toUpperCase()}] ${prefix}${log.message}`;
    })
    .join("\n");
};

const copyToClipboard = async (text: string, label: string) => {
  try {
    await navigator.clipboard.writeText(text);
    const { toast } = await import("sonner");
    toast.success(`${label} copied to clipboard`);
  } catch {
    const { toast } = await import("sonner");
    toast.error("Failed to copy to clipboard");
  }
};

interface JobDetailsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  job: Job | null;
  executionLogs: ExecutionLog[];
  loadingLogs: boolean;
  logLevelFilter: LogLevel | "all";
  onLogLevelFilterChange: (level: LogLevel | "all") => void;
  onCancelJob: (jobId: string) => void;
  onResubmitJob: (jobId: string) => void;
}

export function JobDetailsDialog({
  open,
  onOpenChange,
  job,
  executionLogs,
  loadingLogs,
  logLevelFilter,
  onLogLevelFilterChange,
  onCancelJob,
  onResubmitJob,
}: JobDetailsDialogProps) {
  const logsContainerRef = useRef<HTMLDivElement>(null);
  const isAtBottomRef = useRef<boolean>(true);

  const checkIfAtBottom = () => {
    if (!logsContainerRef.current) return true;
    const { scrollTop, scrollHeight, clientHeight } = logsContainerRef.current;
    return scrollHeight - scrollTop - clientHeight < 20;
  };

  const filteredAndCollapsedLogs = useMemo(() => {
    const filtered = filterLogsByLevel(executionLogs, logLevelFilter);
    return collapseConsecutiveLogs(filtered);
  }, [executionLogs, logLevelFilter]);

  const copyAllJobDetails = () => {
    if (!job) return;

    const parts: string[] = [];

    parts.push(`=== Job Details ===`);
    parts.push(`Job: ${job.job_name}`);
    parts.push(`ID: ${job.id}`);
    parts.push(`Status: ${job.status}`);
    parts.push(`Created: ${new Date(job.created_at).toLocaleString()}`);
    if (job.started_at) {
      parts.push(`Started: ${new Date(job.started_at).toLocaleString()}`);
    }
    if (job.completed_at) {
      parts.push(`Completed: ${new Date(job.completed_at).toLocaleString()}`);
    }
    parts.push("");

    if (job.payload !== undefined && job.payload !== null) {
      parts.push(`=== Payload ===`);
      parts.push(formatJsonValue(job.payload));
      parts.push("");
    }

    if (executionLogs.length > 0) {
      parts.push(`=== Logs ===`);
      parts.push(formatLogsForClipboard(executionLogs));
      parts.push("");
    }

    if (job.result !== undefined && job.result !== null) {
      parts.push(`=== Result ===`);
      parts.push(formatJsonValue(job.result));
      parts.push("");
    }

    if (job.error_message) {
      parts.push(`=== Error ===`);
      parts.push(job.error_message);
    }

    copyToClipboard(parts.join("\n"), "All job details");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] w-[90vw] max-w-[1600px] overflow-y-auto sm:max-w-none">
        <DialogHeader className="flex flex-row items-start justify-between">
          <div>
            <DialogTitle className="flex items-center gap-2">
              {job && getStatusIcon(job.status)}
              Job Details
            </DialogTitle>
            <DialogDescription>
              {job?.job_name} - {job?.id}
            </DialogDescription>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={copyAllJobDetails}
            className="shrink-0"
          >
            <Copy className="mr-2 h-4 w-4" />
            Copy All
          </Button>
        </DialogHeader>

        {job && (
          <div className="space-y-4">
            <div className="flex flex-wrap gap-2">
              <Badge variant={getStatusBadgeVariant(job.status)}>
                {job.status}
              </Badge>
              {job.user_email && (
                <Badge
                  variant="outline"
                  title={
                    job.user_name
                      ? `${job.user_name} (${job.user_email})`
                      : job.user_email
                  }
                >
                  {job.user_email}
                </Badge>
              )}
              {job.user_role && (
                <Badge variant="outline">role: {job.user_role}</Badge>
              )}
            </div>

            <Separator />

            <div className="grid gap-3">
              <div>
                <Label className="text-muted-foreground text-xs">Created</Label>
                <p className="text-sm">
                  {new Date(job.created_at).toLocaleString()}
                </p>
              </div>
              {job.started_at && (
                <div>
                  <Label className="text-muted-foreground text-xs">
                    Started
                  </Label>
                  <p className="text-sm">
                    {new Date(job.started_at).toLocaleString()}
                  </p>
                </div>
              )}
              {job.completed_at && (
                <div>
                  <Label className="text-muted-foreground text-xs">
                    Completed
                  </Label>
                  <p className="text-sm">
                    {new Date(job.completed_at).toLocaleString()}
                  </p>
                </div>
              )}
              {job.progress_percent !== undefined && (
                <div className="space-y-2">
                  <Label className="text-muted-foreground text-xs">
                    Progress
                  </Label>
                  <div className="space-y-1">
                    <div className="flex items-center justify-between text-sm">
                      <span className="font-medium">
                        {job.progress_percent}%
                      </span>
                      {job.estimated_seconds_left !== undefined &&
                        job.estimated_seconds_left > 0 && (
                          <span className="text-muted-foreground">
                            ~
                            {job.estimated_seconds_left < 60
                              ? `${job.estimated_seconds_left}s`
                              : job.estimated_seconds_left < 3600
                                ? `${Math.round(job.estimated_seconds_left / 60)}m`
                                : `${Math.round(job.estimated_seconds_left / 3600)}h`}{" "}
                            remaining
                          </span>
                        )}
                    </div>
                    <div className="bg-secondary h-3 w-full overflow-hidden rounded-full">
                      <div
                        className={`h-full transition-all duration-300 ${
                          job.status === "running"
                            ? "bg-blue-500"
                            : job.status === "completed"
                              ? "bg-green-500"
                              : job.status === "failed"
                                ? "bg-red-500"
                                : "bg-primary"
                        }`}
                        style={{ width: `${job.progress_percent}%` }}
                      />
                    </div>
                    {job.progress_message && (
                      <p className="text-muted-foreground text-sm">
                        {job.progress_message}
                      </p>
                    )}
                    {job.last_progress_at && (
                      <p className="text-muted-foreground text-xs">
                        Last updated:{" "}
                        {new Date(job.last_progress_at).toLocaleString()}
                      </p>
                    )}
                  </div>
                </div>
              )}
            </div>

            <Separator />

            {job.payload !== undefined && job.payload !== null && (
              <div>
                <div className="mb-2 flex items-center justify-between">
                  <Label>Payload</Label>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2"
                    onClick={() =>
                      copyToClipboard(formatJsonValue(job.payload), "Payload")
                    }
                  >
                    <Copy className="h-3 w-3" />
                  </Button>
                </div>
                <div className="bg-muted max-h-48 overflow-auto rounded-lg border p-4">
                  <pre className="text-xs break-all whitespace-pre-wrap">
                    {formatJsonValue(job.payload)}
                  </pre>
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
              <div className="flex flex-col">
                <div className="mb-2 flex items-center justify-between">
                  <Label>Logs</Label>
                  <div className="flex items-center gap-2">
                    <Select
                      value={logLevelFilter}
                      onValueChange={(value) =>
                        onLogLevelFilterChange(value as LogLevel | "all")
                      }
                    >
                      <SelectTrigger className="h-6 w-24 text-xs">
                        <SelectValue placeholder="Level" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All</SelectItem>
                        <SelectItem value="debug">Debug</SelectItem>
                        <SelectItem value="info">Info</SelectItem>
                        <SelectItem value="warning">Warning</SelectItem>
                        <SelectItem value="error">Error</SelectItem>
                        <SelectItem value="fatal">Fatal</SelectItem>
                      </SelectContent>
                    </Select>
                    {executionLogs.length > 0 && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 px-2"
                        onClick={() =>
                          copyToClipboard(
                            formatLogsForClipboard(
                              filterLogsByLevel(executionLogs, logLevelFilter),
                            ),
                            "Logs",
                          )
                        }
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    )}
                  </div>
                </div>
                <div
                  ref={logsContainerRef}
                  className="h-[400px] overflow-y-auto rounded-lg border bg-black/90 p-4 font-mono"
                  onScroll={() => {
                    isAtBottomRef.current = checkIfAtBottom();
                  }}
                >
                  {loadingLogs ? (
                    <span className="text-muted-foreground text-xs italic">
                      Loading logs...
                    </span>
                  ) : executionLogs.length > 0 ? (
                    <div className="flex flex-col gap-0.5">
                      {filteredAndCollapsedLogs.map((log) => (
                        <div
                          key={log.id}
                          className="flex items-start gap-2 text-xs"
                        >
                          <span
                            className={`w-12 shrink-0 rounded px-1 py-0.5 text-center text-[10px] font-medium text-white uppercase ${LOG_LEVEL_BADGE_COLORS[log.level]}`}
                          >
                            {log.level}
                          </span>
                          <span
                            className={`break-words ${LOG_LEVEL_COLORS[log.level]}`}
                          >
                            {log.count > 1 && (
                              <span className="text-gray-500">
                                ({log.count}x){" "}
                              </span>
                            )}
                            {log.message}
                          </span>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <span className="text-muted-foreground text-xs italic">
                      No logs available
                    </span>
                  )}
                </div>
              </div>

              {(job.result !== undefined && job.result !== null) ||
              job.error_message ? (
                <div className="flex flex-col gap-4">
                  {job.result !== undefined && job.result !== null && (
                    <div className="flex flex-1 flex-col">
                      <div className="mb-2 flex items-center justify-between">
                        <Label>Result</Label>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 px-2"
                          onClick={() =>
                            copyToClipboard(
                              formatJsonValue(job.result),
                              "Result",
                            )
                          }
                        >
                          <Copy className="h-3 w-3" />
                        </Button>
                      </div>
                      <div className="bg-muted max-h-[200px] min-h-[100px] flex-1 overflow-auto rounded-lg border p-4">
                        <pre className="text-xs break-all whitespace-pre-wrap">
                          {formatJsonValue(job.result)}
                        </pre>
                      </div>
                    </div>
                  )}

                  {job.error_message && (
                    <div className="flex flex-1 flex-col">
                      <div className="mb-2 flex items-center justify-between">
                        <Label className="text-destructive">Error</Label>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 px-2"
                          onClick={() =>
                            copyToClipboard(job.error_message || "", "Error")
                          }
                        >
                          <Copy className="h-3 w-3" />
                        </Button>
                      </div>
                      <div className="bg-destructive/10 border-destructive/20 max-h-[200px] min-h-[100px] flex-1 overflow-auto rounded-lg border p-4">
                        <pre className="text-destructive text-xs break-all whitespace-pre-wrap">
                          {job.error_message}
                        </pre>
                      </div>
                    </div>
                  )}
                </div>
              ) : null}
            </div>
          </div>
        )}

        <DialogFooter className="flex gap-2">
          {job && (job.status === "pending" || job.status === "running") && (
            <Button
              variant="destructive"
              onClick={() => {
                onCancelJob(job.id);
                onOpenChange(false);
              }}
            >
              <XCircle className="mr-2 h-4 w-4" />
              Cancel Job
            </Button>
          )}
          {job &&
            (job.status === "completed" ||
              job.status === "cancelled" ||
              job.status === "failed") && (
              <Button variant="secondary" onClick={() => onResubmitJob(job.id)}>
                <RefreshCw className="mr-2 h-4 w-4" />
                Re-submit
              </Button>
            )}
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
