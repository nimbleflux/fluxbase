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
import type { FunctionFormData } from "./types";

interface EditFunctionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  formData: FunctionFormData;
  onFormDataChange: (data: FunctionFormData) => void;
  fetching: boolean;
  onSubmit: () => void;
}

export function EditFunctionDialog({
  open,
  onOpenChange,
  formData,
  onFormDataChange,
  fetching,
  onSubmit,
}: EditFunctionDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-4xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Edge Function</DialogTitle>
          <DialogDescription>
            Update function code and settings
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
                  onFormDataChange({ ...formData, description: e.target.value })
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

            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="edit-timeout">Timeout (seconds)</Label>
                <Input
                  id="edit-timeout"
                  type="number"
                  min={1}
                  max={300}
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
                <Label htmlFor="edit-cron">Cron Schedule</Label>
                <Input
                  id="edit-cron"
                  placeholder="0 0 * * *"
                  value={formData.cron_schedule}
                  onChange={(e) =>
                    onFormDataChange({
                      ...formData,
                      cron_schedule: e.target.value,
                    })
                  }
                />
              </div>
            </div>

            <div>
              <Label>Permissions</Label>
              <div className="mt-2 grid grid-cols-2 gap-3">
                <label className="flex cursor-pointer items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.allow_net}
                    onChange={(e) =>
                      onFormDataChange({
                        ...formData,
                        allow_net: e.target.checked,
                      })
                    }
                  />
                  <span>Allow Network Access</span>
                </label>
                <label className="flex cursor-pointer items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.allow_env}
                    onChange={(e) =>
                      onFormDataChange({
                        ...formData,
                        allow_env: e.target.checked,
                      })
                    }
                  />
                  <span>Allow Environment Variables</span>
                </label>
                <label className="flex cursor-pointer items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.allow_read}
                    onChange={(e) =>
                      onFormDataChange({
                        ...formData,
                        allow_read: e.target.checked,
                      })
                    }
                  />
                  <span>Allow File Read</span>
                </label>
                <label className="flex cursor-pointer items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.allow_write}
                    onChange={(e) =>
                      onFormDataChange({
                        ...formData,
                        allow_write: e.target.checked,
                      })
                    }
                  />
                  <span>Allow File Write</span>
                </label>
              </div>
            </div>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={fetching}>
            Update Function
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
