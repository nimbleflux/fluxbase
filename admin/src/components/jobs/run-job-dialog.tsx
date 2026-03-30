import { Play, Loader2 } from "lucide-react";
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
import { Textarea } from "@/components/ui/textarea";
import type { JobFunction } from "./types";

interface RunJobDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  jobFunction: JobFunction | null;
  namespace: string;
  payload: string;
  onPayloadChange: (payload: string) => void;
  submitting: boolean;
  onSubmit: () => void;
}

export function RunJobDialog({
  open,
  onOpenChange,
  jobFunction,
  namespace,
  payload,
  onPayloadChange,
  submitting,
  onSubmit,
}: RunJobDialogProps) {
  const handleClose = () => {
    onOpenChange(false);
    onPayloadChange("");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Play className="h-5 w-5" />
            Run Job
          </DialogTitle>
          <DialogDescription>
            Submit a new job for "{jobFunction?.name}" in the "{namespace}" '
            namespace
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {jobFunction && (
            <div className="bg-muted/50 rounded-lg border p-3">
              <div className="mb-2 flex items-center gap-2">
                <span className="font-medium">{jobFunction.name}</span>
                <Badge variant="outline">v{jobFunction.version}</Badge>
              </div>
              <p className="text-muted-foreground text-sm">
                {jobFunction.description || "No description"}
              </p>
              <div className="text-muted-foreground mt-2 flex items-center gap-4 text-xs">
                <span>Timeout: {jobFunction.timeout_seconds}s</span>
                <span>Max retries: {jobFunction.max_retries}</span>
              </div>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="job-payload">Payload (JSON)</Label>
            <Textarea
              id="job-payload"
              value={payload}
              onChange={(e) => onPayloadChange(e.target.value)}
              placeholder='{\n  "key": "value"\n}'
              className="min-h-[150px] font-mono text-sm"
            />
            <p className="text-muted-foreground text-xs">
              Enter the JSON payload to pass to the job's handler function. This
              will be available as{" "}
              <code className="bg-muted rounded px-1">request.payload</code> in
              your job code.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={submitting}>
            {submitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Submitting...
              </>
            ) : (
              <>
                <Play className="mr-2 h-4 w-4" />
                Run Job
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
