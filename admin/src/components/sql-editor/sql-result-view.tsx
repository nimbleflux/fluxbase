import {
  AlertCircle,
  CheckCircle,
  Download,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { SQLResultViewProps } from "./types";

export function SQLResultView({
  result,
  currentPage,
  totalPages,
  paginatedRows,
  onExportCSV,
  onExportJSON,
  onPrevPage,
  onNextPage,
}: SQLResultViewProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {result.error ? (
            <AlertCircle className="text-destructive h-4 w-4" />
          ) : (
            <CheckCircle className="h-4 w-4 text-green-500" />
          )}
          <code className="text-muted-foreground text-xs">
            {result.statement.length > 60
              ? result.statement.substring(0, 60) + "..."
              : result.statement}
          </code>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline">
            {result.execution_time_ms.toFixed(2)}ms
          </Badge>
          {result.rows && result.rows.length > 0 && (
            <>
              <Button variant="ghost" size="sm" onClick={onExportCSV}>
                <Download className="mr-1 h-3 w-3" />
                CSV
              </Button>
              <Button variant="ghost" size="sm" onClick={onExportJSON}>
                <Download className="mr-1 h-3 w-3" />
                JSON
              </Button>
            </>
          )}
        </div>
      </div>

      {result.error && (
        <div className="bg-destructive/10 text-destructive rounded-md p-3 text-sm">
          {result.error}
        </div>
      )}

      {result.rows && result.rows.length > 0 && (
        <>
          <div className="max-w-full overflow-auto rounded-md border">
            <Table className="w-max min-w-full">
              <TableHeader>
                <TableRow>
                  {result.columns!.map((col) => (
                    <TableHead
                      key={col}
                      className="font-mono text-xs whitespace-nowrap"
                    >
                      {col}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {paginatedRows.map((row, rowIdx) => (
                  <TableRow key={rowIdx}>
                    {result.columns!.map((col) => (
                      <TableCell
                        key={col}
                        className="font-mono text-xs whitespace-nowrap"
                      >
                        {row[col] === null ? (
                          <span className="text-muted-foreground italic">
                            null
                          </span>
                        ) : typeof row[col] === "object" ? (
                          JSON.stringify(row[col])
                        ) : (
                          String(row[col])
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between">
              <p className="text-muted-foreground text-xs">
                Page {currentPage} of {totalPages} ({result.rows!.length} total
                rows)
              </p>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onPrevPage}
                  disabled={currentPage === 1}
                >
                  <ChevronLeft className="h-4 w-4" />
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onNextPage}
                  disabled={currentPage === totalPages}
                >
                  Next
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}

          {totalPages <= 1 && (
            <p className="text-muted-foreground text-xs">
              Showing {result.rows!.length} row(s)
            </p>
          )}
        </>
      )}

      {!result.rows && !result.error && (
        <div className="rounded-md bg-green-500/10 p-3 text-sm text-green-600">
          {result.affected_rows !== undefined
            ? `Success: ${result.affected_rows} row(s) affected`
            : "Query executed successfully"}
        </div>
      )}
    </div>
  );
}
