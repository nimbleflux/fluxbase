import type { PolicyTemplate } from "@/lib/api";
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
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface TemplateApplicationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  template: PolicyTemplate | null;
  tables: { schema: string; table: string }[];
  selectedTable: string;
  onTableSelect: (table: string) => void;
  onApply: () => void;
}

export function TemplateApplicationDialog({
  open,
  onOpenChange,
  template,
  tables,
  selectedTable,
  onTableSelect,
  onApply,
}: TemplateApplicationDialogProps) {
  if (!template) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-full sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Apply Template: {template.name}</DialogTitle>
          <DialogDescription>
            Select the table to apply this policy template to
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="bg-muted/30 rounded-lg border p-4">
            <div className="mb-2 flex items-center gap-2">
              <Badge variant="outline">{template.command}</Badge>
              <span className="text-muted-foreground text-sm">
                {template.description}
              </span>
            </div>
            <div className="space-y-2">
              <div>
                <Label className="text-muted-foreground text-xs">
                  USING Expression
                </Label>
                <pre className="bg-muted mt-1 overflow-auto rounded p-2 text-xs">
                  {template.using}
                </pre>
              </div>
              {template.with_check && (
                <div>
                  <Label className="text-muted-foreground text-xs">
                    WITH CHECK Expression
                  </Label>
                  <pre className="bg-muted mt-1 overflow-auto rounded p-2 text-xs">
                    {template.with_check}
                  </pre>
                </div>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <Label>Select Table</Label>
            <Select value={selectedTable} onValueChange={onTableSelect}>
              <SelectTrigger>
                <SelectValue placeholder="Choose a table..." />
              </SelectTrigger>
              <SelectContent>
                {tables.map((t) => (
                  <SelectItem
                    key={`${t.schema}.${t.table}`}
                    value={`${t.schema}.${t.table}`}
                  >
                    {t.schema}.{t.table}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onApply} disabled={!selectedTable}>
            Continue
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
