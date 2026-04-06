import { formatDistanceToNow } from "date-fns";
import { Globe, FolderOpen, Pencil, History, Trash2 } from "lucide-react";
import type { Secret } from "@/lib/api";
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

interface SecretTableRowProps {
  secret: Secret;
  onEdit: (secret: Secret) => void;
  onHistory: (secret: Secret) => void;
  onDelete: (id: string) => void;
  isDeleting: boolean;
}

export function SecretTableRow({
  secret,
  onEdit,
  onHistory,
  onDelete,
  isDeleting,
}: SecretTableRowProps) {
  return (
    <TableRow>
      <TableCell>
        <div>
          <div className="font-mono font-medium">
            FLUXBASE_SECRET_{secret.name}
          </div>
          {secret.description && (
            <div className="text-muted-foreground text-xs">
              {secret.description}
            </div>
          )}
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-1">
          {secret.scope === "global" ? (
            <Badge variant="default" className="gap-1">
              <Globe className="h-3 w-3" />
              Global
            </Badge>
          ) : (
            <Badge variant="secondary" className="gap-1">
              <FolderOpen className="h-3 w-3" />
              {secret.namespace}
            </Badge>
          )}
        </div>
      </TableCell>
      <TableCell>
        <Badge variant="outline">v{secret.version}</Badge>
      </TableCell>
      <TableCell>
        {secret.expires_at ? (
          <span className={secret.is_expired ? "text-destructive" : ""}>
            {secret.is_expired
              ? "Expired"
              : formatDistanceToNow(new Date(secret.expires_at), {
                  addSuffix: true,
                })}
          </span>
        ) : (
          <span className="text-muted-foreground">Never</span>
        )}
      </TableCell>
      <TableCell className="text-muted-foreground text-sm">
        {formatDistanceToNow(new Date(secret.updated_at), {
          addSuffix: true,
        })}
      </TableCell>
      <TableCell className="text-right">
        <div className="flex justify-end gap-1">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="sm" onClick={() => onEdit(secret)}>
                <Pencil className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Edit secret</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onHistory(secret)}
              >
                <History className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Version history</TooltipContent>
          </Tooltip>
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
              <TooltipContent>Delete secret</TooltipContent>
            </Tooltip>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete Secret</AlertDialogTitle>
                <AlertDialogDescription>
                  Are you sure you want to delete "{secret.name}"? This action
                  cannot be undone and any functions or jobs using this secret
                  will fail.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={() => onDelete(secret.id)}
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
