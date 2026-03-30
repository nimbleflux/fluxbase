import { formatDistanceToNow } from "date-fns";
import { RotateCcw } from "lucide-react";
import type { Secret, SecretVersion } from "@/lib/api";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface VersionHistoryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  secret: Secret | null;
  versions: SecretVersion[] | undefined;
  onRollback: (id: string, version: number) => void;
  isRollbackPending: boolean;
}

export function VersionHistoryDialog({
  open,
  onOpenChange,
  secret,
  versions,
  onRollback,
  isRollbackPending,
}: VersionHistoryDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Version History</DialogTitle>
          <DialogDescription>
            Version history for {secret?.name}. Current version: v
            {secret?.version}
          </DialogDescription>
        </DialogHeader>
        <div className="py-4">
          {versions && versions.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Version</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {versions.map((version) => (
                  <TableRow key={version.id}>
                    <TableCell>
                      <Badge
                        variant={
                          version.version === secret?.version
                            ? "default"
                            : "outline"
                        }
                      >
                        v{version.version}
                        {version.version === secret?.version && " (current)"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {formatDistanceToNow(new Date(version.created_at), {
                        addSuffix: true,
                      })}
                    </TableCell>
                    <TableCell className="text-right">
                      {version.version !== secret?.version && (
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => {
                                if (secret) {
                                  onRollback(secret.id, version.version);
                                }
                              }}
                              disabled={isRollbackPending}
                            >
                              <RotateCcw className="h-4 w-4" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>
                            Rollback to v{version.version}
                          </TooltipContent>
                        </Tooltip>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <p className="text-muted-foreground text-center text-sm">
              No version history available
            </p>
          )}
        </div>
        <DialogFooter>
          <Button
            onClick={() => {
              onOpenChange(false);
            }}
          >
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
