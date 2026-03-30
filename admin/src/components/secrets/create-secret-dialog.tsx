import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { CreateSecretDialogProps } from "./types";

export function CreateSecretDialog({
  open,
  onOpenChange,
  onSubmit,
  isPending,
}: CreateSecretDialogProps) {
  const [name, setName] = useState("");
  const [value, setValue] = useState("");
  const [scope, setScope] = useState<"global" | "namespace">("global");
  const [namespace, setNamespace] = useState("");
  const [description, setDescription] = useState("");
  const [expiresAt, setExpiresAt] = useState("");

  const resetForm = () => {
    setName("");
    setValue("");
    setScope("global");
    setNamespace("");
    setDescription("");
    setExpiresAt("");
  };

  const handleSubmit = () => {
    onSubmit({
      name: name
        .trim()
        .toUpperCase()
        .replace(/[^A-Z0-9_]/g, "_"),
      value,
      scope,
      namespace: scope === "namespace" ? namespace.trim() : undefined,
      description: description.trim() || undefined,
      expires_at: expiresAt ? new Date(expiresAt).toISOString() : undefined,
    });
    resetForm();
  };

  const handleCancel = () => {
    onOpenChange(false);
    resetForm();
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Create Secret</DialogTitle>
          <DialogDescription>
            Create a new encrypted secret. The value will be securely stored and
            available to edge functions and background jobs.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="name">
              Name <span className="text-destructive">*</span>
            </Label>
            <Input
              id="name"
              placeholder="API_KEY"
              value={name}
              onChange={(e) => setName(e.target.value.toUpperCase())}
            />
            <p className="text-muted-foreground text-xs">
              Available as FLUXBASE_SECRET_{name || "NAME"}
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="value">
              Value <span className="text-destructive">*</span>
            </Label>
            <Textarea
              id="value"
              placeholder="Enter secret value..."
              value={value}
              onChange={(e) => setValue(e.target.value)}
              className="font-mono"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="scope">Scope</Label>
            <Select
              value={scope}
              onValueChange={(v) => setScope(v as "global" | "namespace")}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="global">Global (all functions)</SelectItem>
                <SelectItem value="namespace">
                  Namespace (specific namespace)
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          {scope === "namespace" && (
            <div className="grid gap-2">
              <Label htmlFor="namespace">
                Namespace <span className="text-destructive">*</span>
              </Label>
              <Input
                id="namespace"
                placeholder="my-namespace"
                value={namespace}
                onChange={(e) => setNamespace(e.target.value)}
              />
            </div>
          )}
          <div className="grid gap-2">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              placeholder="Optional description..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="expiresAt">Expiration Date (optional)</Label>
            <Input
              id="expiresAt"
              type="datetime-local"
              value={expiresAt}
              onChange={(e) => setExpiresAt(e.target.value)}
            />
            <p className="text-muted-foreground text-xs">
              Expired secrets are automatically excluded from function and job
              execution
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleCancel}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isPending}>
            {isPending ? "Creating..." : "Create Secret"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
