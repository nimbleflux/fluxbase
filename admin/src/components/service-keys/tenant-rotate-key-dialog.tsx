import { RefreshCw } from "lucide-react";
import { Label } from "@/components/ui/label";
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

interface RotateKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  targetKey: ServiceKey | null;
  gracePeriod: string;
  onGracePeriodChange: (period: string) => void;
  onRotate: () => void;
  isPending: boolean;
}

export function RotateKeyDialog({
  open,
  onOpenChange,
  targetKey,
  gracePeriod,
  onGracePeriodChange,
  onRotate,
  isPending,
}: RotateKeyDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <RefreshCw className="h-5 w-5" />
            Rotate Service Key
          </DialogTitle>
          <DialogDescription>
            Create a new key to replace "{targetKey?.name}". The old key will be
            deprecated with a grace period for migration.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="gracePeriodRotate">Grace Period for Old Key</Label>
            <select
              id="gracePeriodRotate"
              value={gracePeriod}
              onChange={(e) => onGracePeriodChange(e.target.value)}
              className="border-input bg-background ring-offset-background flex h-10 w-full rounded-md border px-3 py-2 text-sm"
            >
              <option value="1h">1 hour</option>
              <option value="6h">6 hours</option>
              <option value="12h">12 hours</option>
              <option value="24h">24 hours</option>
              <option value="48h">48 hours</option>
              <option value="7d">7 days</option>
              <option value="14d">14 days</option>
              <option value="30d">30 days</option>
            </select>
            <p className="text-muted-foreground text-xs">
              The old key will continue working for this period.
            </p>
          </div>
          <div className="rounded-md bg-blue-50 p-4 dark:bg-blue-950">
            <div className="text-sm text-blue-700 dark:text-blue-300">
              <p className="font-medium">What happens on rotation:</p>
              <ul className="mt-2 list-disc space-y-1 pl-5">
                <li>A new key is created with the same configuration</li>
                <li>The old key is marked as deprecated</li>
                <li>The old key continues working during the grace period</li>
                <li>After the grace period, the old key stops working</li>
              </ul>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onRotate} disabled={isPending}>
            {isPending ? "Rotating..." : "Rotate Key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
