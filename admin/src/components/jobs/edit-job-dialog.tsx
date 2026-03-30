import { Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { JobFunction, EditFormData } from "./types";

interface EditJobDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  jobFunction: JobFunction | null;
  fetching: boolean;
  formData: EditFormData;
  onFormDataChange: (data: EditFormData) => void;
  onUpdate: () => void;
}

export function EditJobDialog({
  open,
  onOpenChange,
  jobFunction,
  fetching,
  formData,
  onFormDataChange,
  onUpdate,
}: EditJobDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-4xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Job Function</DialogTitle>
          <DialogDescription>
            Update job function code and settings for "{jobFunction?.name}"
          </DialogDescription>
        </DialogHeader>

        {fetching ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
          </div>
        ) : (
          <div className="space-y-4">
            <div>
              <Label htmlFor="edit-description">Description</Label>
              <Input
                id="edit-description"
                value={formData.description}
                onChange={(e) =>
                  onFormDataChange({
                    ...formData,
                    description: e.target.value,
                  })
                }
              />
            </div>

            <div>
              <Label htmlFor="edit-code">Code</Label>
              <Textarea
                id="edit-code"
                className="min-h-[400px] font-mono text-sm"
                value={formData.code}
                onChange={(e) =>
                  onFormDataChange({ ...formData, code: e.target.value })
                }
              />
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div>
                <Label htmlFor="edit-timeout">Timeout (seconds)</Label>
                <Input
                  id="edit-timeout"
                  type="number"
                  min={1}
                  max={3600}
                  value={formData.timeout_seconds}
                  onChange={(e) =>
                    onFormDataChange({
                      ...formData,
                      timeout_seconds: parseInt(e.target.value),
                    })
                  }
                />
              </div>

              <div>
                <Label htmlFor="edit-retries">Max Retries</Label>
                <Input
                  id="edit-retries"
                  type="number"
                  min={0}
                  max={10}
                  value={formData.max_retries}
                  onChange={(e) =>
                    onFormDataChange({
                      ...formData,
                      max_retries: parseInt(e.target.value),
                    })
                  }
                />
              </div>

              <div>
                <Label htmlFor="edit-schedule">Schedule (cron)</Label>
                <Input
                  id="edit-schedule"
                  placeholder="0 0 * * *"
                  value={formData.schedule}
                  onChange={(e) =>
                    onFormDataChange({
                      ...formData,
                      schedule: e.target.value,
                    })
                  }
                />
              </div>
            </div>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onUpdate} disabled={fetching}>
            Update Job Function
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
