import { AlertCircle, ShieldAlert } from "lucide-react";
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
import type { ServiceKey } from "@/lib/api";

interface RevokeKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  targetKey: ServiceKey | null;
  revokeReason: string;
  onRevokeReasonChange: (reason: string) => void;
  onRevoke: () => void;
  isPending: boolean;
}

export function RevokeKeyDialog({
  open,
  onOpenChange,
  targetKey,
  revokeReason,
  onRevokeReasonChange,
  onRevoke,
  isPending,
}: RevokeKeyDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="text-destructive flex items-center gap-2">
            <ShieldAlert className="h-5 w-5" />
            Emergency Revoke
          </DialogTitle>
          <DialogDescription>
            This action is irreversible. The key "{targetKey?.name}" will be
            immediately disabled and marked as revoked. Any applications using
            this key will lose access instantly.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="rounded-md bg-red-50 p-4 dark:bg-red-950">
            <div className="flex">
              <AlertCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800 dark:text-red-200">
                  Warning: This cannot be undone
                </h3>
                <div className="mt-2 text-sm text-red-700 dark:text-red-300">
                  <p>
                    Use this only for security incidents. For planned key
                    rotation, use the Rotate or Deprecate options instead.
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="revokeReason">
              Reason for revocation <span className="text-destructive">*</span>
            </Label>
            <Input
              id="revokeReason"
              placeholder="e.g., Key compromised, employee departure"
              value={revokeReason}
              onChange={(e) => onRevokeReasonChange(e.target.value)}
            />
            <p className="text-muted-foreground text-xs">
              This will be recorded in the audit log.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onRevoke}
            disabled={isPending || !revokeReason.trim()}
          >
            {isPending ? "Revoking..." : "Revoke Key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
