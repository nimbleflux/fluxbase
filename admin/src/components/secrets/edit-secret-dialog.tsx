import { useState } from "react";
import { AlertCircle } from "lucide-react";
import type { Secret } from "@/lib/api";
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

interface EditSecretDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  secret: Secret | null;
  onSubmit: (data: { value?: string; description?: string }) => void;
  isPending: boolean;
}

export function EditSecretDialog({
  open,
  onOpenChange,
  secret,
  onSubmit,
  isPending,
}: EditSecretDialogProps) {
  const [value, setValue] = useState("");
  const [description, setDescription] = useState("");

  const handleOpen = (isOpen: boolean) => {
    if (isOpen && secret) {
      setDescription(secret.description || "");
      setValue("");
    }
    onOpenChange(isOpen);
  };

  const handleSubmit = () => {
    const data: { value?: string; description?: string } = {};
    if (value.trim()) {
      data.value = value;
    }
    if (description !== secret?.description) {
      data.description = description.trim() || undefined;
    }
    onSubmit(data);
  };

  const handleCancel = () => {
    onOpenChange(false);
    setValue("");
    setDescription("");
  };

  return (
    <Dialog open={open} onOpenChange={handleOpen}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Update Secret</DialogTitle>
          <DialogDescription>
            Update the value for {secret?.name}. This will create a new version.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="rounded-md bg-yellow-50 p-4 dark:bg-yellow-950">
            <div className="flex">
              <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
              <div className="ml-3">
                <p className="text-sm text-yellow-700 dark:text-yellow-300">
                  The current secret value cannot be viewed. Enter a new value
                  to update.
                </p>
              </div>
            </div>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="editValue">New Value</Label>
            <Textarea
              id="editValue"
              placeholder="Enter new secret value..."
              value={value}
              onChange={(e) => setValue(e.target.value)}
              className="font-mono"
            />
            <p className="text-muted-foreground text-xs">
              Leave empty to keep the current value
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="editDescription">Description</Label>
            <Input
              id="editDescription"
              placeholder="Optional description..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleCancel}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isPending}>
            {isPending ? "Updating..." : "Update Secret"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
