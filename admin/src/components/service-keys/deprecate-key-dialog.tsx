import { Clock } from "lucide-react";
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

interface DeprecateKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  targetKey: ServiceKey | null;
  gracePeriod: string;
  onGracePeriodChange: (period: string) => void;
  deprecateReason: string;
  onDeprecateReasonChange: (reason: string) => void;
  onDeprecate: () => void;
  isPending: boolean;
}

export function DeprecateKeyDialog({
  open,
  onOpenChange,
  targetKey,
  gracePeriod,
  onGracePeriodChange,
  deprecateReason,
  onDeprecateReasonChange,
  onDeprecate,
  isPending,
}: DeprecateKeyDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Clock className="h-5 w-5" />
            Deprecate Service Key
          </DialogTitle>
          <DialogDescription>
            Mark "{targetKey?.name}" as deprecated with a grace period. The key
            will continue working during the grace period, allowing time for
            migration.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="gracePeriodDeprecate">Grace Period</Label>
            <select
              id="gracePeriodDeprecate"
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
              The key will stop working after this period.
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="deprecateReason">Reason (optional)</Label>
            <Input
              id="deprecateReason"
              placeholder="e.g., Scheduled rotation, security policy"
              value={deprecateReason}
              onChange={(e) => onDeprecateReasonChange(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onDeprecate} disabled={isPending}>
            {isPending ? "Deprecating..." : "Deprecate Key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
