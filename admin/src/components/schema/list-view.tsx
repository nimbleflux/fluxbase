import { Shield, ShieldOff, AlertTriangle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { ListViewProps, SchemaNode } from "./types";

function getFullName(node: SchemaNode) {
  return `${node.schema}.${node.name}`;
}

export function ListView({
  nodes,
  relationships,
  onSelectTable,
  getTableWarningCount,
  getTableWarningSeverity,
  getTableWarnings,
}: ListViewProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Tables and Views</CardTitle>
        <CardDescription>{nodes.length} items found</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Schema</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Columns</TableHead>
              <TableHead>RLS</TableHead>
              <TableHead>Warnings</TableHead>
              <TableHead>Relationships</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {nodes.map((node) => {
              const fullName = getFullName(node);
              const nodeRelationships = relationships.filter(
                (r) =>
                  (r.source_schema === node.schema &&
                    r.source_table === node.name) ||
                  (r.target_schema === node.schema &&
                    r.target_table === node.name),
              );
              return (
                <TableRow
                  key={fullName}
                  className="cursor-pointer"
                  onClick={() => onSelectTable(fullName)}
                >
                  <TableCell className="font-medium">{node.name}</TableCell>
                  <TableCell>
                    <Badge variant="outline">{node.schema}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant="default">table</Badge>
                  </TableCell>
                  <TableCell>{node.columns.length}</TableCell>
                  <TableCell>
                    {node.rls_enabled ? (
                      <Shield className="h-4 w-4 text-green-500" />
                    ) : (
                      <ShieldOff className="text-muted-foreground h-4 w-4" />
                    )}
                  </TableCell>
                  <TableCell>
                    {(() => {
                      const count = getTableWarningCount(
                        node.schema,
                        node.name,
                      );
                      const severity = getTableWarningSeverity(
                        node.schema,
                        node.name,
                      );
                      const warnings = getTableWarnings(node.schema, node.name);
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
                                onClick={(e) => e.stopPropagation()}
                              >
                                <AlertTriangle className="h-3 w-3" />
                                {count}
                              </Badge>
                            </PopoverTrigger>
                            <PopoverContent className="w-72" align="start">
                              <div className="space-y-2">
                                <h4 className="text-sm font-medium">
                                  Warnings for {node.name}
                                </h4>
                                <div className="max-h-48 space-y-2 overflow-auto">
                                  {warnings.map((w) => (
                                    <div
                                      key={w.id}
                                      className="border-l-2 border-l-orange-500 pl-2 text-xs"
                                    >
                                      <Badge
                                        variant="outline"
                                        className="mb-1 text-xs"
                                      >
                                        {w.severity}
                                      </Badge>
                                      <p className="text-muted-foreground">
                                        {w.message}
                                      </p>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            </PopoverContent>
                          </Popover>
                        );
                      }
                      return <span className="text-muted-foreground">-</span>;
                    })()}
                  </TableCell>
                  <TableCell>
                    {nodeRelationships?.length ? (
                      <Badge variant="outline">
                        {nodeRelationships.length}
                      </Badge>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
