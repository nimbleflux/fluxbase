import { Plus, Users, X } from "lucide-react";
import { toast } from "sonner";
import type { CreateKnowledgeBaseDialogProps, KBPermission } from "./types";
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

export function CreateKnowledgeBaseDialog({
  open,
  onOpenChange,
  newKB,
  onNewKBChange,
  onCreate,
  users,
  usersLoading,
  providers,
  providersLoading,
}: CreateKnowledgeBaseDialogProps) {
  const newPermission = { user_id: "", permission: "viewer" as KBPermission };

  const addPermission = () => {
    if (!newPermission.user_id) {
      toast.error("Please select a user");
      return;
    }
    if (
      newKB.initial_permissions?.some(
        (p) => p.user_id === newPermission.user_id,
      )
    ) {
      toast.error("User already has permission");
      return;
    }
    onNewKBChange({
      ...newKB,
      initial_permissions: [
        ...(newKB.initial_permissions || []),
        newPermission,
      ],
    });
  };

  const removePermission = (userId: string) => {
    onNewKBChange({
      ...newKB,
      initial_permissions:
        newKB.initial_permissions?.filter((p) => p.user_id !== userId) || [],
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Knowledge Base</DialogTitle>
          <DialogDescription>
            Create a new knowledge base to store documents for RAG-powered AI
            chatbots.
          </DialogDescription>
        </DialogHeader>
        <div className="grid max-h-[60vh] gap-4 overflow-y-auto py-4">
          <div className="grid gap-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={newKB.name}
              onChange={(e) =>
                onNewKBChange({ ...newKB, name: e.target.value })
              }
              placeholder="e.g., product-docs"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={newKB.description || ""}
              onChange={(e) =>
                onNewKBChange({ ...newKB, description: e.target.value })
              }
              placeholder="What kind of documents will this knowledge base contain?"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="visibility">Visibility</Label>
            <select
              id="visibility"
              value={newKB.visibility}
              onChange={(e) =>
                onNewKBChange({
                  ...newKB,
                  visibility: e.target.value as "private" | "shared" | "public",
                })
              }
              className="border-input focus-visible:ring-ring flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:ring-1 focus-visible:outline-none"
            >
              <option value="private">Private - Only you can access</option>
              <option value="shared">
                Shared - Specific users you grant access
              </option>
              <option value="public">
                Public - All authenticated users can contribute
              </option>
            </select>
            <p className="text-muted-foreground text-xs">
              {newKB.visibility === "private" &&
                "Only you will be able to access this knowledge base."}
              {newKB.visibility === "shared" &&
                "Grant access to specific users below."}
              {newKB.visibility === "public" &&
                "All authenticated users get read-only access. Grant explicit permissions for write access."}
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="embedding_model">Embedding Model</Label>
            <select
              id="embedding_model"
              value={newKB.embedding_model || ""}
              onChange={(e) =>
                onNewKBChange({
                  ...newKB,
                  embedding_model: e.target.value || undefined,
                })
              }
              disabled={providersLoading}
              className="border-input focus-visible:ring-ring flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:ring-1 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
            >
              <option value="">
                {providersLoading
                  ? "Loading providers..."
                  : providers.length === 0
                    ? "No AI providers configured"
                    : "Use default embedding model"}
              </option>
              {providers.map((p) => (
                <option key={p.id} value={p.embedding_model || p.id}>
                  {p.display_name}
                  {p.embedding_model
                    ? ` (${p.embedding_model})`
                    : p.use_for_embeddings
                      ? " (embedding provider)"
                      : ""}
                </option>
              ))}
            </select>
            <p className="text-muted-foreground text-xs">
              Select an AI provider to use for generating embeddings. Configure
              providers in the AI Providers settings.
            </p>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label htmlFor="chunk_size">Chunk Size</Label>
              <Input
                id="chunk_size"
                type="number"
                value={newKB.chunk_size}
                onChange={(e) =>
                  onNewKBChange({
                    ...newKB,
                    chunk_size: parseInt(e.target.value) || 512,
                  })
                }
              />
              <p className="text-muted-foreground text-xs">
                Characters per chunk
              </p>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="chunk_overlap">Chunk Overlap</Label>
              <Input
                id="chunk_overlap"
                type="number"
                value={newKB.chunk_overlap}
                onChange={(e) =>
                  onNewKBChange({
                    ...newKB,
                    chunk_overlap: parseInt(e.target.value) || 50,
                  })
                }
              />
              <p className="text-muted-foreground text-xs">
                Overlap between chunks
              </p>
            </div>
          </div>

          {newKB.visibility === "shared" && (
            <div className="grid gap-3 border-t pt-4">
              <Label className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                Share with Users
              </Label>
              <div className="flex gap-2">
                <select
                  value=""
                  onChange={(e) => {
                    const userId = e.target.value;
                    if (
                      userId &&
                      !newKB.initial_permissions?.some(
                        (p) => p.user_id === userId,
                      )
                    ) {
                      onNewKBChange({
                        ...newKB,
                        initial_permissions: [
                          ...(newKB.initial_permissions || []),
                          { user_id: userId, permission: "viewer" },
                        ],
                      });
                    }
                  }}
                  className="border-input focus-visible:ring-ring flex h-9 flex-1 rounded-md border bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:ring-1 focus-visible:outline-none"
                  disabled={usersLoading}
                >
                  <option value="">Select a user...</option>
                  {users.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.email || u.id}
                    </option>
                  ))}
                </select>
                <Button onClick={addPermission} size="sm" variant="outline">
                  <Plus className="h-4 w-4" />
                </Button>
              </div>
              {newKB.initial_permissions &&
                newKB.initial_permissions.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {newKB.initial_permissions.map((perm) => {
                      const user = users.find((u) => u.id === perm.user_id);
                      return (
                        <Badge
                          key={perm.user_id}
                          variant="secondary"
                          className="gap-1 pr-1"
                        >
                          <span>{user?.email || perm.user_id}</span>
                          <span className="text-muted-foreground">
                            ({perm.permission})
                          </span>
                          <button
                            onClick={() => removePermission(perm.user_id)}
                            className="hover:text-destructive ml-1"
                          >
                            <X className="h-3 w-3" />
                          </button>
                        </Badge>
                      );
                    })}
                  </div>
                )}
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onCreate}>Create</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
