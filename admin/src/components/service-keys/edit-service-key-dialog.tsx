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
import type { UpdateServiceKeyRequest } from "@/lib/api";
import { SCOPE_GROUPS } from "./types";

interface EditServiceKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editName: string;
  onEditNameChange: (name: string) => void;
  editDescription: string;
  onEditDescriptionChange: (description: string) => void;
  editScopes: string[];
  onEditScopesChange: (scopes: string[]) => void;
  editRateLimitPerMinute: number | undefined;
  onEditRateLimitPerMinuteChange: (value: number | undefined) => void;
  editRateLimitPerHour: number | undefined;
  onEditRateLimitPerHourChange: (value: number | undefined) => void;
  onSubmit: (request: UpdateServiceKeyRequest) => void;
  isPending: boolean;
}

export function EditServiceKeyDialog({
  open,
  onOpenChange,
  editName,
  onEditNameChange,
  editDescription,
  onEditDescriptionChange,
  editScopes,
  onEditScopesChange,
  editRateLimitPerMinute,
  onEditRateLimitPerMinuteChange,
  editRateLimitPerHour,
  onEditRateLimitPerHourChange,
  onSubmit,
  isPending,
}: EditServiceKeyDialogProps) {
  const handleSubmit = () => {
    onSubmit({
      name: editName.trim() || undefined,
      description: editDescription.trim(),
      scopes: editScopes.length > 0 ? editScopes : undefined,
      rate_limit_per_minute: editRateLimitPerMinute,
      rate_limit_per_hour: editRateLimitPerHour,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-3xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Service Key</DialogTitle>
          <DialogDescription>
            Update service key properties. The key value cannot be changed.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label htmlFor="editName">Name</Label>
              <Input
                id="editName"
                value={editName}
                onChange={(e) => onEditNameChange(e.target.value)}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="editDescription">Description</Label>
              <Input
                id="editDescription"
                value={editDescription}
                onChange={(e) => onEditDescriptionChange(e.target.value)}
              />
            </div>
          </div>
          <div className="grid gap-2">
            <div className="flex items-center justify-between">
              <Label>Scopes</Label>
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="edit-wildcard-scope"
                  checked={editScopes.includes("*")}
                  onChange={(e) => {
                    if (e.target.checked) {
                      onEditScopesChange(["*"]);
                    } else {
                      onEditScopesChange([]);
                    }
                  }}
                  className="h-4 w-4 rounded border-gray-300"
                />
                <label
                  htmlFor="edit-wildcard-scope"
                  className="text-sm font-medium"
                >
                  All Scopes
                </label>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3 rounded-md border p-4">
              {SCOPE_GROUPS.map((group) => (
                <div key={group.name} className="space-y-1">
                  <div className="text-sm font-medium">{group.name}</div>
                  <div className="text-muted-foreground text-xs">
                    {group.description}
                  </div>
                  <div className="flex flex-wrap gap-3 pt-1">
                    {group.scopes.map((scope) => (
                      <div
                        key={scope.id}
                        className="flex items-center space-x-1.5"
                      >
                        <input
                          type="checkbox"
                          id={`edit-${scope.id}`}
                          checked={
                            editScopes.includes(scope.id) ||
                            editScopes.includes("*")
                          }
                          disabled={editScopes.includes("*")}
                          onChange={(e) => {
                            if (e.target.checked) {
                              onEditScopesChange([...editScopes, scope.id]);
                            } else {
                              onEditScopesChange(
                                editScopes.filter((s) => s !== scope.id),
                              );
                            }
                          }}
                          className="h-3.5 w-3.5 rounded border-gray-300"
                        />
                        <label
                          htmlFor={`edit-${scope.id}`}
                          className="text-xs"
                          title={scope.description}
                        >
                          {scope.label}
                        </label>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label htmlFor="editRateLimitPerMinute">
                Rate Limit (per minute)
              </Label>
              <Input
                id="editRateLimitPerMinute"
                type="number"
                min="0"
                placeholder="Unlimited"
                value={editRateLimitPerMinute ?? ""}
                onChange={(e) =>
                  onEditRateLimitPerMinuteChange(
                    e.target.value ? parseInt(e.target.value) : undefined,
                  )
                }
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="editRateLimitPerHour">
                Rate Limit (per hour)
              </Label>
              <Input
                id="editRateLimitPerHour"
                type="number"
                min="0"
                placeholder="Unlimited"
                value={editRateLimitPerHour ?? ""}
                onChange={(e) =>
                  onEditRateLimitPerHourChange(
                    e.target.value ? parseInt(e.target.value) : undefined,
                  )
                }
              />
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isPending}>
            {isPending ? "Saving..." : "Save Changes"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
