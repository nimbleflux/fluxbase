import { Database, Braces, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { HistoryItemProps } from "./types";

export function HistoryItem({
  history,
  isSelected,
  onSelect,
  onRemove,
}: HistoryItemProps) {
  return (
    <div
      className={`hover:bg-accent flex cursor-pointer items-center justify-between rounded-md p-2 ${
        isSelected ? "bg-accent" : ""
      }`}
      onClick={onSelect}
    >
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          {history.mode === "sql" ? (
            <Database className="text-muted-foreground h-3 w-3 flex-shrink-0" />
          ) : (
            <Braces className="text-muted-foreground h-3 w-3 flex-shrink-0" />
          )}
          <Badge variant="outline" className="text-xs">
            {history.mode.toUpperCase()}
          </Badge>
          <span className="text-muted-foreground text-xs">
            {history.timestamp.toLocaleString()}
          </span>
          {history.mode === "sql" && history.results && (
            <Badge variant="secondary" className="text-xs">
              {history.results.length} result(s)
            </Badge>
          )}
          {history.executionTime && (
            <Badge variant="secondary" className="text-xs">
              {history.executionTime.toFixed(0)}ms
            </Badge>
          )}
        </div>
        <code className="text-muted-foreground mt-1 block truncate text-xs">
          {history.query
            .split("\n")
            .find(
              (l) =>
                l.trim() &&
                !l.trim().startsWith("--") &&
                !l.trim().startsWith("#"),
            )
            ?.substring(0, 80) || history.query.split("\n")[0].substring(0, 80)}
        </code>
      </div>
      <Button
        variant="ghost"
        size="sm"
        className="ml-2 h-6 w-6 p-0"
        onClick={(e) => {
          e.stopPropagation();
          onRemove();
        }}
      >
        <X className="h-3 w-3" />
      </Button>
    </div>
  );
}
