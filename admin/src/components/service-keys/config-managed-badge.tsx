import { Lock } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface ConfigManagedBadgeProps {
  isConfigManaged: boolean;
}

export function ConfigManagedBadge({
  isConfigManaged,
}: ConfigManagedBadgeProps) {
  if (!isConfigManaged) {
    return null;
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge variant="outline" className="gap-1 text-xs">
          <Lock className="h-3 w-3" />
          Config Managed
        </Badge>
      </TooltipTrigger>
      <TooltipContent>
        This key is managed by configuration and cannot be modified through the
        API
      </TooltipContent>
    </Tooltip>
  );
}
