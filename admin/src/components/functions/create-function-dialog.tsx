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

interface CreateFunctionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  formData: FunctionFormData;
  onFormDataChange: (data: FunctionFormData) => void;
  onSubmit: () => void;
  onReset: () => void;
}

export function CreateFunctionDialog({
  open,
  onOpenChange,
  formData,
  onFormDataChange,
  onSubmit,
  onReset,
}: CreateFunctionDialogProps) {
  const handleSubmit = () => {
    onSubmit();
  };

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      onReset();
    }
    onOpenChange(open);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-4xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Edge Function</DialogTitle>
          <DialogDescription>
            Deploy a new TypeScript/JavaScript function with Deno runtime
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <Label htmlFor="name">Function Name</Label>
            <Input
              id="name"
              placeholder="my_function"
              value={formData.name}
              onChange={(e) =>
                onFormDataChange({ ...formData, name: e.target.value })
              }
            />
          </div>

          <div>
            <Label htmlFor="description">Description (optional)</Label>
            <Input
              id="description"
              placeholder="What does this function do?"
              value={formData.description}
              onChange={(e) =>
                onFormDataChange({ ...formData, description: e.target.value })
              }
            />
          </div>

          <div>
            <Label htmlFor="code">Code (TypeScript)</Label>
            <Textarea
              id="code"
              className="min-h-[400px] font-mono text-sm"
              value={formData.code}
              onChange={(e) =>
                onFormDataChange({ ...formData, code: e.target.value })
              }
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="timeout">Timeout (seconds)</Label>
              <Input
                id="timeout"
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
              <Label htmlFor="cron">Cron Schedule (optional)</Label>
              <Input
                id="cron"
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

        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit}>Create Function</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
