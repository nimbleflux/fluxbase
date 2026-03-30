import { AlertCircle, CheckCircle, Download } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { GraphQLResultViewProps } from "./types";

export function GraphQLResultView({
  response,
  executionTime,
}: GraphQLResultViewProps) {
  const hasErrors = response.errors && response.errors.length > 0;
  const hasData = response.data !== undefined && response.data !== null;
  const hasNoResults =
    response.data === undefined &&
    (!response.errors || response.errors.length === 0);

  const handleExportJSON = () => {
    const json = JSON.stringify(response, null, 2);
    const blob = new Blob([json], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `graphql-result-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success("Exported as JSON");
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {hasErrors ? (
            <AlertCircle className="text-destructive h-4 w-4" />
          ) : (
            <CheckCircle className="h-4 w-4 text-green-500" />
          )}
          <span className="text-sm font-medium">GraphQL Response</span>
        </div>
        <div className="flex items-center gap-2">
          {executionTime && (
            <Badge variant="outline">{executionTime.toFixed(0)}ms</Badge>
          )}
          {hasData && (
            <Button variant="ghost" size="sm" onClick={handleExportJSON}>
              <Download className="mr-1 h-3 w-3" />
              JSON
            </Button>
          )}
        </div>
      </div>

      {hasErrors && (
        <div className="space-y-2">
          {response.errors!.map((error, idx) => (
            <div
              key={idx}
              className="bg-destructive/10 text-destructive rounded-md p-3 text-sm"
            >
              <div className="font-medium">{error.message}</div>
              {error.locations && error.locations.length > 0 && (
                <div className="text-destructive/80 mt-1 text-xs">
                  Location: Line {error.locations[0].line}, Column{" "}
                  {error.locations[0].column}
                </div>
              )}
              {error.path && error.path.length > 0 && (
                <div className="text-destructive/80 mt-1 text-xs">
                  Path: {error.path.join(" → ")}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {hasData && (
        <div className="bg-muted/30 max-h-[500px] overflow-auto rounded-md border p-4">
          <pre className="font-mono text-xs whitespace-pre-wrap">
            {JSON.stringify(response.data, null, 2)}
          </pre>
        </div>
      )}

      {hasNoResults && (
        <div className="rounded-md bg-green-500/10 p-3 text-sm text-green-600">
          GraphQL query executed successfully (no data returned)
        </div>
      )}
    </div>
  );
}
