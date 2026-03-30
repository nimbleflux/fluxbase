import { formatDistanceToNow } from "date-fns";
import { Trash2, X } from "lucide-react";
import type { ClientKey } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { TableCell, TableRow } from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface ClientKeyTableRowProps {
  clientKey: ClientKey;
  onRevoke: (id: string) => void;
  onDelete: (id: string) => void;
  isRevoking: boolean;
  isDeleting: boolean;
}

const isExpired = (expiresAt?: string) => {
  if (!expiresAt) return false;
  return new Date(expiresAt) < new Date();
};

const isRevoked = (revokedAt?: string) => !!revokedAt;

const getKeyStatus = (key: ClientKey) => {
  if (isRevoked(key.revoked_at))
    return { label: "Revoked", variant: "secondary" as const };
  if (isExpired(key.expires_at))
    return { label: "Expired", variant: "destructive" as const };
  return { label: "Active", variant: "default" as const };
};

export function ClientKeyTableRow({
  clientKey,
  onRevoke,
  onDelete,
  isRevoking,
  isDeleting,
}: ClientKeyTableRowProps) {
  const status = getKeyStatus(clientKey);

  return (
    <TableRow>
      <TableCell>
        <div>
          <div className="font-medium">{clientKey.name}</div>
          {clientKey.description && (
            <div className="text-muted-foreground text-xs">
              {clientKey.description}
            </div>
          )}
        </div>
      </TableCell>
      <TableCell>
        <code className="text-xs">{clientKey.key_prefix}...</code>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {clientKey.scopes.slice(0, 2).map((scope) => (
            <Badge key={scope} variant="outline" className="text-xs">
              {scope}
            </Badge>
          ))}
          {clientKey.scopes.length > 2 && (
            <Badge variant="outline" className="text-xs">
              +{clientKey.scopes.length - 2}
            </Badge>
          )}
        </div>
      </TableCell>
      <TableCell className="text-sm">
        {clientKey.rate_limit_per_minute}/min
      </TableCell>
      <TableCell className="text-muted-foreground text-sm">
        {clientKey.last_used_at
          ? formatDistanceToNow(new Date(clientKey.last_used_at), {
              addSuffix: true,
            })
          : "Never"}
      </TableCell>
      <TableCell>
        <Badge variant={status.variant}>{status.label}</Badge>
      </TableCell>
      <TableCell className="text-right">
        <div className="flex justify-end gap-1">
          {!isRevoked(clientKey.revoked_at) && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onRevoke(clientKey.id)}
                  disabled={isRevoking}
                >
                  <X className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Revoke client key</TooltipContent>
            </Tooltip>
          )}
          <AlertDialog>
            <Tooltip>
              <TooltipTrigger asChild>
                <AlertDialogTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={isDeleting}
                    className="text-destructive hover:text-destructive hover:bg-destructive/10"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </AlertDialogTrigger>
              </TooltipTrigger>
              <TooltipContent>Delete client key</TooltipContent>
            </Tooltip>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete Client Key</AlertDialogTitle>
                <AlertDialogDescription>
                  Are you sure you want to delete "{clientKey.name}"? Any
                  applications using this key will lose access immediately.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={() => onDelete(clientKey.id)}
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                >
                  Delete
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </TableCell>
    </TableRow>
  );
}
