import { AlertCircle, AlertTriangle, Info } from "lucide-react";
import type { SecurityWarning } from "@/lib/api";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";

interface WarningCardProps {
  warning: SecurityWarning;
  onNavigate: () => void;
}

export function WarningCard({ warning, onNavigate }: WarningCardProps) {
  const severityColors = {
    critical: "border-red-500/50 bg-red-500/5",
    high: "border-orange-500/50 bg-orange-500/5",
    medium: "border-yellow-500/50 bg-yellow-500/5",
    low: "border-blue-500/50 bg-blue-500/5",
  };

  const severityIcons = {
    critical: <AlertCircle className="h-5 w-5 text-red-500" />,
    high: <AlertTriangle className="h-5 w-5 text-orange-500" />,
    medium: <Info className="h-5 w-5 text-yellow-500" />,
    low: <Info className="h-5 w-5 text-blue-500" />,
  };

  return (
    <div
      className={cn(
        "cursor-pointer rounded-lg border p-4 transition-shadow hover:shadow-md",
        severityColors[warning.severity],
      )}
      onClick={onNavigate}
    >
      <div className="flex items-start gap-3">
        {severityIcons[warning.severity]}
        <div className="flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <Badge
              variant={
                warning.severity === "critical" || warning.severity === "high"
                  ? "destructive"
                  : "secondary"
              }
            >
              {warning.severity}
            </Badge>
            <Badge variant="outline">{warning.category}</Badge>
          </div>
          <p className="mt-2 text-sm">{warning.message}</p>
          <div className="mt-2 flex items-center gap-2">
            <Badge variant="outline">
              {warning.schema}.{warning.table}
            </Badge>
            {warning.policy_name && (
              <Badge variant="secondary">{warning.policy_name}</Badge>
            )}
          </div>
          <p className="bg-muted mt-2 rounded p-2 text-sm">
            <strong>Suggestion:</strong> {warning.suggestion}
          </p>
          {warning.fix_sql && (
            <pre className="bg-muted mt-2 overflow-auto rounded p-2 text-xs">
              {warning.fix_sql}
            </pre>
          )}
        </div>
      </div>
    </div>
  );
}
