import { AlertCircle, Copy } from "lucide-react";
import { toast } from "sonner";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { RotateServiceKeyResponse } from "@/lib/api";

interface RotatedKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  rotatedKey: RotateServiceKeyResponse | null;
}

export function RotatedKeyDialog({
  open,
  onOpenChange,
  rotatedKey,
}: RotatedKeyDialogProps) {
  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Key Rotated Successfully</DialogTitle>
          <DialogDescription>
            Save the new key now. You won't be able to see it again!
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="rounded-md bg-yellow-50 p-4 dark:bg-yellow-950">
            <div className="flex">
              <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
              <div className="ml-3">
                <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                  Important: Copy the new key now
                </h3>
                <div className="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
                  <p>
                    This is the only time you'll see the new service key. The
                    old key will continue working during the grace period.
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div className="grid gap-2">
            <Label>New Service Key</Label>
            <div className="flex gap-2">
              <Input
                value={rotatedKey?.key || ""}
                readOnly
                className="font-mono text-xs"
              />
              <Button
                variant="outline"
                size="icon"
                onClick={() => copyToClipboard(rotatedKey?.key || "")}
              >
                <Copy className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <div className="grid gap-2">
            <Label>Name</Label>
            <Input value={rotatedKey?.name || ""} readOnly />
          </div>
          {rotatedKey?.grace_period_ends_at && (
            <div className="grid gap-2">
              <Label>Old Key Expires</Label>
              <Input
                value={new Date(
                  rotatedKey.grace_period_ends_at,
                ).toLocaleString()}
                readOnly
              />
            </div>
          )}
        </div>
        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>
            I've Saved the New Key
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
