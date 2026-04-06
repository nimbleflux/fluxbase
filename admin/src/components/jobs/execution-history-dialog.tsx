import {
  History,
  Loader2,
  Activity,
  CheckCircle,
  AlertCircle,
  Clock,
  XCircle,
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
import { ScrollArea } from "@/components/ui/scroll-area";
import type { Job, JobFunction } from "./types";

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

interface ExecutionHistoryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  jobFunction: JobFunction | null;
  historyJobs: Job[];
  loading: boolean;
  onViewJobDetails: (job: Job) => void;
}

export function ExecutionHistoryDialog({
  open,
  onOpenChange,
  jobFunction,
  historyJobs,
  loading,
  onViewJobDetails,
}: ExecutionHistoryDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-4xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <History className="h-5 w-5" />
            Execution History
          </DialogTitle>
          <DialogDescription>
            Recent executions for "{jobFunction?.name}"
          </DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
          </div>
        ) : historyJobs.length === 0 ? (
          <div className="py-12 text-center">
            <Activity className="text-muted-foreground mx-auto mb-4 h-12 w-12" />
            <p className="text-muted-foreground">No executions found</p>
          </div>
        ) : (
          <ScrollArea className="h-[400px]">
            <div className="space-y-2">
              {historyJobs.map((job) => (
                <div
                  key={job.id}
                  className="hover:bg-muted/50 flex cursor-pointer items-center justify-between rounded-lg border p-3"
                  onClick={() => {
                    onOpenChange(false);
                    onViewJobDetails(job);
                  }}
                >
                  <div className="flex items-center gap-3">
                    {getStatusIcon(job.status)}
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium">
                          {job.id.slice(0, 8)}...
                        </span>
                        <Badge variant={getStatusBadgeVariant(job.status)}>
                          {job.status}
                        </Badge>
                      </div>
                      <span className="text-muted-foreground text-xs">
                        {new Date(job.created_at).toLocaleString()}
                      </span>
                    </div>
                  </div>
                  <div className="text-right">
                    {job.started_at && job.completed_at && (
                      <span className="text-muted-foreground text-xs">
                        {new Date(job.completed_at).getTime() -
                          new Date(job.started_at).getTime()}
                        ms
                      </span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
