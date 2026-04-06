import {
  Terminal,
  Globe,
  Lock,
  Shield,
  Timer,
  Clock,
  Code,
  Copy,
  Loader2,
  Database,
} from "lucide-react";
import type { RPCProcedure } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";

interface ProcedureDetailsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  procedure: RPCProcedure | null;
  loading: boolean;
  onCopy: (text: string, label: string) => void;
}

export function ProcedureDetailsDialog({
  open,
  onOpenChange,
  procedure,
  loading,
  onCopy,
}: ProcedureDetailsDialogProps) {
  if (!procedure) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] w-[90vw] max-w-5xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Terminal className="h-5 w-5" />
            Procedure Details
          </DialogTitle>
          <DialogDescription>
            {procedure.namespace}/{procedure.name}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant={procedure.enabled ? "secondary" : "outline"}>
              {procedure.enabled ? "Enabled" : "Disabled"}
            </Badge>
            <Badge variant="outline">v{procedure.version}</Badge>
            {procedure.is_public ? (
              <Badge variant="outline">
                <Globe className="mr-1 h-3 w-3" />
                Public
              </Badge>
            ) : (
              <Badge variant="outline">
                <Lock className="mr-1 h-3 w-3" />
                Private
              </Badge>
            )}
            {procedure.require_role && (
              <Badge variant="outline">
                <Shield className="mr-1 h-3 w-3" />
                Role: {procedure.require_role}
              </Badge>
            )}
            <Badge variant="outline">
              <Timer className="mr-1 h-3 w-3" />
              {procedure.max_execution_time_seconds}s timeout
            </Badge>
            {procedure.schedule && (
              <Badge variant="outline">
                <Clock className="mr-1 h-3 w-3" />
                Schedule: {procedure.schedule}
              </Badge>
            )}
          </div>

          {procedure.description && (
            <div>
              <Label className="text-muted-foreground text-xs">
                Description
              </Label>
              <p className="mt-1 text-sm">{procedure.description}</p>
            </div>
          )}

          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="text-muted-foreground h-6 w-6 animate-spin" />
            </div>
          ) : procedure.original_code ? (
            <div>
              <div className="mb-2 flex items-center justify-between">
                <Label className="flex items-center gap-1">
                  <Code className="h-3 w-3" />
                  RPC Code
                </Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onCopy(procedure.original_code!, "RPC code")}
                >
                  <Copy className="h-3 w-3" />
                </Button>
              </div>
              <pre className="bg-muted max-h-64 overflow-x-auto rounded-md p-3 font-mono text-xs whitespace-pre-wrap">
                {procedure.original_code}
              </pre>
            </div>
          ) : null}

          <div>
            <div className="mb-2 flex items-center justify-between">
              <Label className="flex items-center gap-1">
                <Terminal className="h-3 w-3" />
                SDK Usage
              </Label>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  const params = procedure.input_schema
                    ? Object.keys(procedure.input_schema)
                        .map((k) => `${k}: /* ${procedure.input_schema![k]} */`)
                        .join(", ")
                    : "";
                  const code = `const { data, error } = await client.rpc.invoke('${procedure.name}'${params ? `, { ${params} }` : ""})`;
                  onCopy(code, "SDK usage");
                }}
              >
                <Copy className="h-3 w-3" />
              </Button>
            </div>
            <pre className="bg-muted overflow-x-auto rounded-md p-3 font-mono text-xs whitespace-pre-wrap">
              {`const { data, error } = await client.rpc.invoke('${procedure.name}'${
                procedure.input_schema &&
                Object.keys(procedure.input_schema).length > 0
                  ? `, {\n${Object.entries(procedure.input_schema)
                      .map(([k, v]) => `  ${k}: /* ${v} */`)
                      .join(",\n")}\n}`
                  : ""
              })`}
            </pre>
          </div>

          <div className="grid grid-cols-2 gap-4">
            {procedure.input_schema &&
              Object.keys(procedure.input_schema).length > 0 && (
                <div>
                  <Label className="text-muted-foreground text-xs">
                    Input Parameters
                  </Label>
                  <div className="mt-1 space-y-1">
                    {Object.entries(procedure.input_schema).map(
                      ([name, type]) => (
                        <div
                          key={name}
                          className="flex items-center gap-2 text-sm"
                        >
                          <code className="bg-muted rounded px-1 text-xs">
                            {name}
                          </code>
                          <span className="text-muted-foreground text-xs">
                            {type}
                          </span>
                        </div>
                      ),
                    )}
                  </div>
                </div>
              )}
            {procedure.output_schema &&
              Object.keys(procedure.output_schema).length > 0 && (
                <div>
                  <Label className="text-muted-foreground text-xs">
                    Output Columns
                  </Label>
                  <div className="mt-1 space-y-1">
                    {Object.entries(procedure.output_schema).map(
                      ([name, type]) => (
                        <div
                          key={name}
                          className="flex items-center gap-2 text-sm"
                        >
                          <code className="bg-muted rounded px-1 text-xs">
                            {name}
                          </code>
                          <span className="text-muted-foreground text-xs">
                            {type}
                          </span>
                        </div>
                      ),
                    )}
                  </div>
                </div>
              )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            {procedure.allowed_tables &&
              procedure.allowed_tables.length > 0 && (
                <div>
                  <Label className="text-muted-foreground flex items-center gap-1 text-xs">
                    <Database className="h-3 w-3" />
                    Allowed Tables
                  </Label>
                  <div className="mt-1 flex flex-wrap gap-1">
                    {procedure.allowed_tables.map((table) => (
                      <Badge key={table} variant="outline" className="text-xs">
                        {table}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            {procedure.allowed_schemas &&
              procedure.allowed_schemas.length > 0 && (
                <div>
                  <Label className="text-muted-foreground flex items-center gap-1 text-xs">
                    <Database className="h-3 w-3" />
                    Allowed Schemas
                  </Label>
                  <div className="mt-1 flex flex-wrap gap-1">
                    {procedure.allowed_schemas.map((schema) => (
                      <Badge key={schema} variant="outline" className="text-xs">
                        {schema}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
          </div>

          <div className="grid grid-cols-3 gap-4 border-t pt-2 text-sm">
            <div>
              <span className="text-muted-foreground text-xs">Created:</span>
              <p className="text-xs">
                {new Date(procedure.created_at).toLocaleString()}
              </p>
            </div>
            <div>
              <span className="text-muted-foreground text-xs">Updated:</span>
              <p className="text-xs">
                {new Date(procedure.updated_at).toLocaleString()}
              </p>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
