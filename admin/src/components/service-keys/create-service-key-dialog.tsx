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
import type { CreateServiceKeyRequest } from "@/lib/api";
import { SCOPE_GROUPS } from "./types";

interface CreateServiceKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  name: string;
  onNameChange: (name: string) => void;
  description: string;
  onDescriptionChange: (description: string) => void;
  selectedScopes: string[];
  onScopesChange: (scopes: string[]) => void;
  rateLimitPerMinute: number | undefined;
  onRateLimitPerMinuteChange: (value: number | undefined) => void;
  rateLimitPerHour: number | undefined;
  onRateLimitPerHourChange: (value: number | undefined) => void;
  expiresAt: string;
  onExpiresAtChange: (value: string) => void;
  onSubmit: (request: CreateServiceKeyRequest) => void;
  isPending: boolean;
}

export function CreateServiceKeyDialog({
  open,
  onOpenChange,
  name,
  onNameChange,
  description,
  onDescriptionChange,
  selectedScopes,
  onScopesChange,
  rateLimitPerMinute,
  onRateLimitPerMinuteChange,
  rateLimitPerHour,
  onRateLimitPerHourChange,
  expiresAt,
  onExpiresAtChange,
  onSubmit,
  isPending,
}: CreateServiceKeyDialogProps) {
  const handleSubmit = () => {
    onSubmit({
      name: name.trim(),
      description: description.trim() || undefined,
      scopes: selectedScopes.length > 0 ? selectedScopes : undefined,
      rate_limit_per_minute: rateLimitPerMinute,
      rate_limit_per_hour: rateLimitPerHour,
      expires_at: expiresAt || undefined,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-3xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Service Key</DialogTitle>
          <DialogDescription>
            Generate a new service key for server-to-server API access. The key
            will be shown only once.
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
                placeholder="Migrations Key"
                value={name}
                onChange={(e) => onNameChange(e.target.value)}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                placeholder="Used by CI/CD pipeline"
                value={description}
                onChange={(e) => onDescriptionChange(e.target.value)}
              />
            </div>
          </div>
          <div className="grid gap-2">
            <div className="flex items-center justify-between">
              <Label>Scopes</Label>
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="wildcard-scope"
                  checked={selectedScopes.includes("*")}
                  onChange={(e) => {
                    if (e.target.checked) {
                      onScopesChange(["*"]);
                    } else {
                      onScopesChange([]);
                    }
                  }}
                  className="h-4 w-4 rounded border-gray-300"
                />
                <label htmlFor="wildcard-scope" className="text-sm font-medium">
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
                          id={`create-${scope.id}`}
                          checked={
                            selectedScopes.includes(scope.id) ||
                            selectedScopes.includes("*")
                          }
                          disabled={selectedScopes.includes("*")}
                          onChange={(e) => {
                            if (e.target.checked) {
                              onScopesChange([...selectedScopes, scope.id]);
                            } else {
                              onScopesChange(
                                selectedScopes.filter((s) => s !== scope.id),
                              );
                            }
                          }}
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
              <Label htmlFor="rateLimitPerMinute">
                Rate Limit (per minute)
              </Label>
              <Input
                id="rateLimitPerMinute"
                type="number"
                min="0"
                placeholder="Unlimited"
                value={rateLimitPerMinute ?? ""}
                onChange={(e) =>
                  onRateLimitPerMinuteChange(
                    e.target.value ? parseInt(e.target.value) : undefined,
                  )
                }
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="rateLimitPerHour">Rate Limit (per hour)</Label>
              <Input
                id="rateLimitPerHour"
                type="number"
                min="0"
                placeholder="Unlimited"
                value={rateLimitPerHour ?? ""}
                onChange={(e) =>
                  onRateLimitPerHourChange(
                    e.target.value ? parseInt(e.target.value) : undefined,
                  )
                }
              />
            </div>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="expiresAt">Expiration Date (optional)</Label>
            <Input
              id="expiresAt"
              type="datetime-local"
              value={expiresAt}
              onChange={(e) => onExpiresAtChange(e.target.value)}
            />
            <p className="text-muted-foreground text-xs">
              Leave empty for no expiration
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isPending}>
            {isPending ? "Creating..." : "Generate Service Key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
