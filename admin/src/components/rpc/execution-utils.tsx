import {
  CheckCircle,
  XCircle,
  Loader2,
  Clock,
  AlertCircle,
} from "lucide-react";

export function getStatusIcon(status: string) {
  switch (status) {
    case "completed":
      return <CheckCircle className="h-4 w-4 shrink-0 text-green-500" />;
    case "running":
      return (
        <Loader2 className="h-4 w-4 shrink-0 animate-spin text-blue-500" />
      );
    case "pending":
      return <Clock className="h-4 w-4 shrink-0 text-yellow-500" />;
    case "failed":
    case "cancelled":
    case "timeout":
      return <XCircle className="h-4 w-4 shrink-0 text-red-500" />;
    default:
      return <AlertCircle className="text-muted-foreground h-4 w-4 shrink-0" />;
  }
}

export function getStatusVariant(
  status: string,
): "secondary" | "destructive" | "outline" {
  switch (status) {
    case "completed":
      return "secondary";
    case "failed":
    case "cancelled":
    case "timeout":
      return "destructive";
    default:
      return "outline";
  }
}

export function canCancelExecution(status: string) {
  return status === "pending" || status === "running";
}
