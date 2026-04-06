import {
  Loader2,
  Plus,
  Shield,
  ShieldCheck,
  ShieldOff,
  AlertTriangle,
} from "lucide-react";
import type { RLSPolicy, SecurityWarning } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Switch } from "@/components/ui/switch";
import { PolicyCard } from "./policy-card";

interface PolicyManagementModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  schema: string;
  table: string;
  warning?: SecurityWarning | null;
  tableDetails:
    | { rls_enabled: boolean; rls_forced: boolean; policies: RLSPolicy[] }
    | null
    | undefined;
  detailsLoading: boolean;
  onToggleRLS: (enable: boolean) => void;
  onEditPolicy: (policy: RLSPolicy) => void;
  onDeletePolicy: (policy: RLSPolicy) => void;
  onCreatePolicy: () => void;
}

export function PolicyManagementModal({
  open,
  onOpenChange,
  schema,
  table,
  warning,
  tableDetails,
  detailsLoading,
  onToggleRLS,
  onEditPolicy,
  onDeletePolicy,
  onCreatePolicy,
}: PolicyManagementModalProps) {
  if (!schema || !table) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] w-full overflow-y-auto sm:max-w-5xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Manage Policies: {table}
          </DialogTitle>
          <DialogDescription>
            {schema}.{table}
          </DialogDescription>
        </DialogHeader>

        {warning && (
          <div className="rounded-lg border border-orange-500/50 bg-orange-500/10 p-4">
            <div className="flex items-start gap-3">
              <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-orange-500" />
              <div className="flex-1">
                <div className="mb-1 flex items-center gap-2">
                  <Badge
                    variant={
                      warning.severity === "critical" ||
                      warning.severity === "high"
                        ? "destructive"
                        : "secondary"
                    }
                  >
                    {warning.severity}
                  </Badge>
                  <Badge variant="outline">{warning.category}</Badge>
                </div>
                <p className="font-medium">{warning.message}</p>
                {warning.suggestion && (
                  <p className="text-muted-foreground mt-1 text-sm">
                    {warning.suggestion}
                  </p>
                )}
                {warning.fix_sql && (
                  <pre className="bg-muted mt-2 overflow-auto rounded p-2 text-xs">
                    {warning.fix_sql}
                  </pre>
                )}
              </div>
            </div>
          </div>
        )}

        {detailsLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
          </div>
        ) : tableDetails ? (
          <div className="space-y-6">
            <div className="bg-muted/50 flex items-center justify-between rounded-lg p-4">
              <div className="flex items-center gap-3">
                {tableDetails.rls_enabled ? (
                  <ShieldCheck className="h-6 w-6 text-green-500" />
                ) : (
                  <ShieldOff className="text-muted-foreground h-6 w-6" />
                )}
                <div>
                  <div className="text-lg font-medium">
                    RLS {tableDetails.rls_enabled ? "Enabled" : "Disabled"}
                  </div>
                  <div className="text-muted-foreground text-sm">
                    Force RLS: {tableDetails.rls_forced ? "Yes" : "No"}
                  </div>
                </div>
              </div>
              <Switch
                checked={tableDetails.rls_enabled}
                onCheckedChange={onToggleRLS}
              />
            </div>

            <div>
              <div className="mb-4 flex items-center justify-between">
                <h3 className="text-lg font-semibold">
                  Policies ({tableDetails.policies.length})
                </h3>
                <Button onClick={onCreatePolicy}>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Policy
                </Button>
              </div>

              {tableDetails.policies.length === 0 ? (
                <div className="rounded-lg border py-12 text-center">
                  <ShieldOff className="text-muted-foreground mx-auto mb-3 h-12 w-12" />
                  <h4 className="text-lg font-medium">No policies defined</h4>
                  {tableDetails.rls_enabled && (
                    <p className="text-muted-foreground mt-1 text-sm">
                      All access will be denied by default when RLS is enabled
                    </p>
                  )}
                  <Button onClick={onCreatePolicy} className="mt-4">
                    <Plus className="mr-2 h-4 w-4" />
                    Create First Policy
                  </Button>
                </div>
              ) : (
                <div className="space-y-3">
                  {tableDetails.policies.map((policy) => (
                    <PolicyCard
                      key={policy.policy_name}
                      policy={policy}
                      onEdit={() => onEditPolicy(policy)}
                      onDelete={() => onDeletePolicy(policy)}
                    />
                  ))}
                </div>
              )}
            </div>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
