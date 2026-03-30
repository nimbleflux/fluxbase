import { Link } from "@tanstack/react-router";
import {
  Database,
  Key,
  Link as LinkIcon,
  Shield,
  ShieldOff,
  AlertTriangle,
  ArrowRight,
  ArrowLeft,
  Columns,
  GitFork,
  Fingerprint,
  Hash,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { TableDetailsPanelProps } from "./types";

function getFullName(node: { schema: string; name: string }) {
  return `${node.schema}.${node.name}`;
}

export function TableDetailsPanel({
  selectedTableData,
  selectedTableRelationships,
  onSelectTable,
  getTableWarningCount,
  getTableWarningSeverity,
  getTableWarnings,
}: TableDetailsPanelProps) {
  return (
    <Card className="w-[420px] shrink-0">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Database className="h-4 w-4" />
            {selectedTableData.name}
          </CardTitle>
          <div className="flex items-center gap-2">
            {(() => {
              const count = getTableWarningCount(
                selectedTableData.schema,
                selectedTableData.name,
              );
              const severity = getTableWarningSeverity(
                selectedTableData.schema,
                selectedTableData.name,
              );
              const warnings = getTableWarnings(
                selectedTableData.schema,
                selectedTableData.name,
              );
              if (count > 0) {
                return (
                  <Popover>
                    <PopoverTrigger asChild>
                      <Badge
                        variant={
                          severity === "critical" || severity === "high"
                            ? "destructive"
                            : "secondary"
                        }
                        className="cursor-pointer gap-1"
                      >
                        <AlertTriangle className="h-3 w-3" />
                        {count}
                      </Badge>
                    </PopoverTrigger>
                    <PopoverContent className="w-80" align="end">
                      <div className="space-y-3">
                        <h4 className="text-sm font-medium">
                          Security Warnings
                        </h4>
                        <div className="max-h-60 space-y-2 overflow-auto">
                          {warnings.map((w) => (
                            <div
                              key={w.id}
                              className="border-l-2 border-l-orange-500 pl-2 text-sm"
                            >
                              <div className="flex items-center gap-1">
                                <Badge variant="outline" className="text-xs">
                                  {w.severity}
                                </Badge>
                              </div>
                              <p className="text-muted-foreground mt-1">
                                {w.message}
                              </p>
                              {w.suggestion && (
                                <p className="text-muted-foreground mt-1 text-xs italic">
                                  {w.suggestion}
                                </p>
                              )}
                            </div>
                          ))}
                        </div>
                        <Button
                          variant="outline"
                          size="sm"
                          className="w-full"
                          asChild
                        >
                          <Link to="/policies">Manage Policies</Link>
                        </Button>
                      </div>
                    </PopoverContent>
                  </Popover>
                );
              }
              return null;
            })()}
            {selectedTableData.rls_enabled ? (
              <Badge variant="default" className="gap-1">
                <Shield className="h-3 w-3" />
                RLS
              </Badge>
            ) : (
              <Badge variant="secondary" className="gap-1">
                <ShieldOff className="h-3 w-3" />
                No RLS
              </Badge>
            )}
          </div>
        </div>
        <CardDescription>{getFullName(selectedTableData)}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <h4 className="mb-2 flex items-center gap-2 text-sm font-medium">
            <Columns className="h-4 w-4" />
            Columns ({selectedTableData.columns.length})
          </h4>
          <div className="max-h-96 space-y-1 overflow-auto">
            {selectedTableData.columns.map((col) => (
              <TooltipProvider key={col.name}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div className="hover:bg-muted flex cursor-default items-start gap-1.5 rounded px-2 py-1 text-sm">
                      <div className="mt-0.5 flex shrink-0 items-center gap-1">
                        {col.is_primary_key && (
                          <Key className="h-3 w-3 text-yellow-500" />
                        )}
                        {col.is_foreign_key && (
                          <LinkIcon className="h-3 w-3 text-blue-500" />
                        )}
                        {col.is_unique && !col.is_primary_key && (
                          <Fingerprint className="h-3 w-3 text-purple-500" />
                        )}
                        {col.is_indexed &&
                          !col.is_primary_key &&
                          !col.is_unique && (
                            <Hash className="h-3 w-3 text-gray-500" />
                          )}
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center justify-between gap-2">
                          <span
                            className={cn(
                              "truncate",
                              col.is_primary_key && "font-medium",
                            )}
                          >
                            {col.name}
                            {!col.nullable && (
                              <span className="ml-0.5 text-xs text-red-500">
                                *
                              </span>
                            )}
                          </span>
                          <span className="text-muted-foreground shrink-0 text-xs">
                            {col.data_type}
                          </span>
                        </div>
                        {col.comment && (
                          <p className="text-muted-foreground truncate text-xs">
                            {col.comment}
                          </p>
                        )}
                      </div>
                    </div>
                  </TooltipTrigger>
                  <TooltipContent side="left" className="max-w-xs">
                    <div className="space-y-1 text-xs">
                      <div className="font-medium">{col.name}</div>
                      <div>Type: {col.data_type}</div>
                      <div>Nullable: {col.nullable ? "Yes" : "No"}</div>
                      {col.default_value && (
                        <div>Default: {col.default_value}</div>
                      )}
                      {col.is_primary_key && (
                        <div className="text-yellow-500">Primary Key</div>
                      )}
                      {col.is_foreign_key && col.fk_target && (
                        <div className="text-blue-500">
                          FK → {col.fk_target}
                        </div>
                      )}
                      {col.is_unique && (
                        <div className="text-purple-500">Unique</div>
                      )}
                      {col.is_indexed && (
                        <div className="text-gray-500">Indexed</div>
                      )}
                      {col.description && (
                        <div className="mt-1 italic">{col.description}</div>
                      )}
                      {col.jsonb_schema && (
                        <div className="mt-2 space-y-1 border-t border-border pt-2">
                          <div className="font-medium text-blue-500">
                            JSONB Schema
                          </div>
                          <div className="space-y-0.5">
                            {Object.entries(
                              col.jsonb_schema.properties || {},
                            ).map(([key, prop]) => (
                              <div key={key} className="flex gap-2">
                                <span className="font-mono">{key}</span>
                                <span className="text-muted-foreground">
                                  {prop.type}
                                </span>
                                {col.jsonb_schema?.required?.includes(key) && (
                                  <span className="text-red-500">*</span>
                                )}
                                {prop.description && (
                                  <span className="text-muted-foreground italic">
                                    {prop.description}
                                  </span>
                                )}
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            ))}
          </div>
        </div>

        {(selectedTableRelationships.incoming.length > 0 ||
          selectedTableRelationships.outgoing.length > 0) && (
          <div>
            <h4 className="mb-2 flex items-center gap-2 text-sm font-medium">
              <GitFork className="h-4 w-4" />
              Relationships (
              {selectedTableRelationships.outgoing.length +
                selectedTableRelationships.incoming.length}
              )
            </h4>
            <div className="max-h-48 space-y-2 overflow-auto">
              {selectedTableRelationships.outgoing.map((rel) => (
                <TooltipProvider key={rel.id}>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div
                        className="bg-muted/50 hover:bg-muted flex cursor-pointer items-center gap-2 rounded p-2 text-sm"
                        onClick={() =>
                          onSelectTable(
                            `${rel.target_schema}.${rel.target_table}`,
                          )
                        }
                      >
                        <ArrowRight className="h-4 w-4 shrink-0 text-blue-500" />
                        <span className="text-muted-foreground truncate">
                          {rel.source_column}
                        </span>
                        <Badge
                          variant="outline"
                          className="shrink-0 px-1 text-[10px]"
                        >
                          {rel.cardinality === "one-to-one" ? "1:1" : "N:1"}
                        </Badge>
                        <span className="truncate font-medium">
                          {rel.target_table}
                        </span>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent side="left" className="max-w-xs">
                      <div className="space-y-1 text-xs">
                        <div className="font-medium">Outgoing FK</div>
                        <div>
                          {rel.source_column} → {rel.target_schema}.
                          {rel.target_table}.{rel.target_column}
                        </div>
                        <div>Cardinality: {rel.cardinality}</div>
                        <div>ON DELETE: {rel.on_delete}</div>
                        <div>ON UPDATE: {rel.on_update}</div>
                        <div className="text-muted-foreground">
                          {rel.constraint_name}
                        </div>
                      </div>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              ))}
              {selectedTableRelationships.incoming.map((rel) => (
                <TooltipProvider key={rel.id}>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div
                        className="bg-muted/50 hover:bg-muted flex cursor-pointer items-center gap-2 rounded p-2 text-sm"
                        onClick={() =>
                          onSelectTable(
                            `${rel.source_schema}.${rel.source_table}`,
                          )
                        }
                      >
                        <ArrowLeft className="h-4 w-4 shrink-0 text-green-500" />
                        <span className="truncate font-medium">
                          {rel.source_table}
                        </span>
                        <Badge
                          variant="outline"
                          className="shrink-0 px-1 text-[10px]"
                        >
                          {rel.cardinality === "one-to-one" ? "1:1" : "1:N"}
                        </Badge>
                        <span className="text-muted-foreground truncate">
                          {rel.target_column}
                        </span>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent side="left" className="max-w-xs">
                      <div className="space-y-1 text-xs">
                        <div className="font-medium">Incoming FK</div>
                        <div>
                          {rel.source_schema}.{rel.source_table}.
                          {rel.source_column} → {rel.target_column}
                        </div>
                        <div>
                          Cardinality:{" "}
                          {rel.cardinality === "many-to-one"
                            ? "one-to-many"
                            : rel.cardinality}
                        </div>
                        <div>ON DELETE: {rel.on_delete}</div>
                        <div>ON UPDATE: {rel.on_update}</div>
                        <div className="text-muted-foreground">
                          {rel.constraint_name}
                        </div>
                      </div>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              ))}
            </div>
          </div>
        )}

        {selectedTableData.row_estimate !== undefined && (
          <div className="text-muted-foreground text-sm">
            ~{selectedTableData.row_estimate.toLocaleString()} rows
          </div>
        )}
      </CardContent>
    </Card>
  );
}
