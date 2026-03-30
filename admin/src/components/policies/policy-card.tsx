import { useState } from "react";
import { ChevronDown, ChevronRight, Pencil, Trash2 } from "lucide-react";
import type { RLSPolicy } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Label } from "@/components/ui/label";

interface PolicyCardProps {
  policy: RLSPolicy;
  onEdit: () => void;
  onDelete: () => void;
}

export function PolicyCard({ policy, onEdit, onDelete }: PolicyCardProps) {
  const [expanded, setExpanded] = useState(false);
  const isPermissive = policy.permissive === "PERMISSIVE";

  return (
    <Collapsible open={expanded} onOpenChange={setExpanded}>
      <div className="rounded-lg border">
        <CollapsibleTrigger className="hover:bg-muted/50 flex w-full items-center justify-between p-3">
          <div className="flex items-center gap-3">
            {expanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
            <span className="font-medium">{policy.policy_name}</span>
            <Badge variant="outline">{policy.command}</Badge>
            <Badge variant={isPermissive ? "default" : "secondary"}>
              {policy.permissive}
            </Badge>
          </div>
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={(e) => {
                e.stopPropagation();
                onEdit();
              }}
            >
              <Pencil className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={(e) => {
                e.stopPropagation();
                onDelete();
              }}
            >
              <Trash2 className="text-destructive h-4 w-4" />
            </Button>
          </div>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <div className="space-y-3 border-t px-3 pt-0 pb-3">
            <div className="pt-3">
              <Label className="text-muted-foreground text-xs">Roles</Label>
              <div className="mt-1 flex gap-1">
                {policy.roles.map((role) => (
                  <Badge key={role} variant="secondary">
                    {role}
                  </Badge>
                ))}
              </div>
            </div>
            {policy.using && (
              <div>
                <Label className="text-muted-foreground text-xs">
                  USING Expression
                </Label>
                <pre className="bg-muted mt-1 overflow-auto rounded p-2 text-xs">
                  {policy.using}
                </pre>
              </div>
            )}
            {policy.with_check && (
              <div>
                <Label className="text-muted-foreground text-xs">
                  WITH CHECK Expression
                </Label>
                <pre className="bg-muted mt-1 overflow-auto rounded p-2 text-xs">
                  {policy.with_check}
                </pre>
              </div>
            )}
          </div>
        </CollapsibleContent>
      </div>
    </Collapsible>
  );
}
