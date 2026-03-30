import { formatDistanceToNow } from "date-fns";
import { History } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { ServiceKey, ServiceKeyRevocation } from "@/lib/api";

interface HistoryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  targetKey: ServiceKey | null;
  revocationHistory: ServiceKeyRevocation[];
}

export function HistoryDialog({
  open,
  onOpenChange,
  targetKey,
  revocationHistory,
}: HistoryDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <History className="h-5 w-5" />
            Key History: {targetKey?.name}
          </DialogTitle>
          <DialogDescription>
            View the revocation and rotation history for this service key.
          </DialogDescription>
        </DialogHeader>
        <div className="py-4">
          {revocationHistory.length === 0 ? (
            <div className="text-muted-foreground py-8 text-center">
              <History className="mx-auto mb-4 h-12 w-12 opacity-50" />
              <p>No revocation history for this key.</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Reason</TableHead>
                  <TableHead>By</TableHead>
                  <TableHead>Date</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {revocationHistory.map((rev) => (
                  <TableRow key={rev.id}>
                    <TableCell>
                      <Badge
                        variant={
                          rev.revocation_type === "emergency"
                            ? "destructive"
                            : rev.revocation_type === "rotation"
                              ? "default"
                              : "secondary"
                        }
                      >
                        {rev.revocation_type}
                      </Badge>
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate">
                      {rev.reason || "-"}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {rev.revoked_by || "-"}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {formatDistanceToNow(new Date(rev.created_at), {
                        addSuffix: true,
                      })}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
