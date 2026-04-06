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
import type { CreateClientKeyDialogProps } from "./types";

export function CreateClientKeyDialog({
  open,
  onOpenChange,
  scopeGroups,
  selectedScopes,
  onToggleScope,
  onSubmit,
  isPending,
  name,
  onNameChange,
  description,
  onDescriptionChange,
  rateLimit,
  onRateLimitChange,
  expiresAt,
  onExpiresAtChange,
}: CreateClientKeyDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-3xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Client Key</DialogTitle>
          <DialogDescription>
            Generate a new client key for programmatic access. The key will be
            shown only once.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label htmlFor="name">
                Name <span className="text-destructive">*</span>
              </Label>
              <Input
                id="name"
                placeholder="Production Client Key"
                value={name}
                onChange={(e) => onNameChange(e.target.value)}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                placeholder="Used by the main application"
                value={description}
                onChange={(e) => onDescriptionChange(e.target.value)}
              />
            </div>
          </div>
          <div className="grid gap-2">
            <Label>
              Scopes/Permissions <span className="text-destructive">*</span>
            </Label>
            <div className="grid grid-cols-2 gap-3 rounded-md border p-4">
              {scopeGroups.map((group) => (
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
                          id={`create-${scope.id}`}
                          checked={selectedScopes.includes(scope.id)}
                          onChange={() => onToggleScope(scope.id)}
                          className="h-3.5 w-3.5 rounded border-gray-300"
                        />
                        <label
                          htmlFor={`create-${scope.id}`}
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
              <Label htmlFor="rateLimit">
                Rate Limit (requests per minute)
              </Label>
              <Input
                id="rateLimit"
                type="number"
                min="1"
                max="10000"
                value={rateLimit}
                onChange={(e) =>
                  onRateLimitChange(parseInt(e.target.value) || 100)
                }
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="expiresAt">Expiration Date (optional)</Label>
              <Input
                id="expiresAt"
                type="datetime-local"
                value={expiresAt}
                onChange={(e) => onExpiresAtChange(e.target.value)}
              />
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={isPending}>
            {isPending ? "Creating..." : "Generate Client Key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
