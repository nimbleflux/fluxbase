import { useState } from "react";
import { Loader2, Plus } from "lucide-react";
import type { PolicyTemplate, CreatePolicyRequest } from "@/lib/api";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";

interface CreatePolicyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  schema: string;
  table: string;
  templates: PolicyTemplate[];
  onSubmit: (data: CreatePolicyRequest) => void;
  isLoading: boolean;
}

export function CreatePolicyDialog({
  open,
  onOpenChange,
  schema,
  table,
  templates,
  onSubmit,
  isLoading,
}: CreatePolicyDialogProps) {
  const [formData, setFormData] = useState<CreatePolicyRequest>({
    schema,
    table,
    name: "",
    command: "ALL",
    roles: ["authenticated"],
    using: "",
    with_check: "",
    permissive: true,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      ...formData,
      schema,
      table,
    });
  };

  const handleTemplateSelect = (templateId: string) => {
    const template = templates.find((t) => t.id === templateId);
    if (template) {
      setFormData({
        ...formData,
        command: template.command,
        using: template.using,
        with_check: template.with_check || "",
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Create Policy</DialogTitle>
          <DialogDescription>
            Create a new RLS policy for {schema}.{table}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="name">Policy Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder="e.g., users_select_own"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="command">Command</Label>
              <Select
                value={formData.command}
                onValueChange={(value) =>
                  setFormData({ ...formData, command: value })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="ALL">ALL</SelectItem>
                  <SelectItem value="SELECT">SELECT</SelectItem>
                  <SelectItem value="INSERT">INSERT</SelectItem>
                  <SelectItem value="UPDATE">UPDATE</SelectItem>
                  <SelectItem value="DELETE">DELETE</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {templates.length > 0 && (
            <div className="space-y-2">
              <Label>Use Template</Label>
              <Select onValueChange={handleTemplateSelect}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a template..." />
                </SelectTrigger>
                <SelectContent>
                  {templates.map((t) => (
                    <SelectItem key={t.id} value={t.id}>
                      {t.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="using">USING Expression</Label>
            <Textarea
              id="using"
              value={formData.using || ""}
              onChange={(e) =>
                setFormData({ ...formData, using: e.target.value })
              }
              placeholder="e.g., auth.uid() = user_id"
              rows={3}
              className="font-mono text-sm"
            />
            <p className="text-muted-foreground text-xs">
              Expression that returns true for rows the user can access
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="check">WITH CHECK Expression (optional)</Label>
            <Textarea
              id="check"
              value={formData.with_check || ""}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  with_check: e.target.value,
                })
              }
              placeholder="e.g., auth.uid() = user_id"
              rows={3}
              className="font-mono text-sm"
            />
            <p className="text-muted-foreground text-xs">
              Expression that must be true for new/modified rows (INSERT/UPDATE)
            </p>
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Switch
                id="permissive"
                checked={formData.permissive}
                onCheckedChange={(checked) =>
                  setFormData({ ...formData, permissive: checked })
                }
              />
              <Label htmlFor="permissive">Permissive</Label>
            </div>
            <p className="text-muted-foreground text-xs">
              Permissive policies are combined with OR, restrictive with AND
            </p>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Plus className="mr-2 h-4 w-4" />
              )}
              Create Policy
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
