import { Badge } from "@/components/ui/badge";
import { Server } from "lucide-react";

export function InstanceLevelIndicator() {
  return (
    <Badge
      variant="outline"
      className="gap-1 border-purple-300 bg-purple-50 text-purple-700 dark:border-purple-700 dark:bg-purple-950 dark:text-purple-300"
    >
      <Server className="h-3 w-3" />
      Instance Level
    </Badge>
  );
}
