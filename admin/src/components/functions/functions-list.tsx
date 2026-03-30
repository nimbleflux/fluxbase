import {
  Clock,
  Edit,
  History,
  Play,
  Plus,
  RefreshCw,
  Trash2,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { EmptyState } from "@/components/empty-state";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { EdgeFunction } from "@/lib/api";

interface FunctionsListProps {
  edgeFunctions: EdgeFunction[];
  namespaces: string[];
  selectedNamespace: string;
  reloading: boolean;
  onNamespaceChange: (namespace: string) => void;
  onReload: () => void;
  onRefresh: () => void;
  onCreateFunction: () => void;
  onEditFunction: (fn: EdgeFunction) => void;
  onInvokeFunction: (fn: EdgeFunction) => void;
  onViewLogs: (fn: EdgeFunction) => void;
  onDeleteFunction: (name: string) => void;
  onToggleFunction: (fn: EdgeFunction) => void;
}

export function FunctionsList({
  edgeFunctions,
  namespaces,
  selectedNamespace,
  reloading,
  onNamespaceChange,
  onReload,
  onRefresh,
  onCreateFunction,
  onEditFunction,
  onInvokeFunction,
  onViewLogs,
  onDeleteFunction,
  onToggleFunction,
}: FunctionsListProps) {
  return (
    <>
      <div className="mb-4 flex items-center justify-end gap-2">
        <div className="flex items-center gap-2">
          <Label
            htmlFor="func-namespace-select"
            className="text-muted-foreground text-sm whitespace-nowrap"
          >
            Namespace:
          </Label>
          <Select value={selectedNamespace} onValueChange={onNamespaceChange}>
            <SelectTrigger id="func-namespace-select" className="w-[180px]">
              <SelectValue placeholder="Select namespace" />
            </SelectTrigger>
            <SelectContent>
              {namespaces.map((ns) => (
                <SelectItem key={ns} value={ns}>
                  {ns}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <Button
          onClick={onReload}
          variant="outline"
          size="sm"
          disabled={reloading}
        >
          {reloading ? (
            <>
              <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
              Reloading...
            </>
          ) : (
            <>
              <RefreshCw className="mr-2 h-4 w-4" />
              Reload from Filesystem
            </>
          )}
        </Button>
        <Button onClick={onRefresh} variant="outline" size="sm">
          <RefreshCw className="mr-2 h-4 w-4" />
          Refresh
        </Button>
        <Button onClick={onCreateFunction} size="sm">
          <Plus className="mr-2 h-4 w-4" />
          New Function
        </Button>
      </div>

      <div className="mb-4 flex gap-4 text-sm">
        <div className="flex items-center gap-1.5">
          <span className="text-muted-foreground">Total:</span>
          <Badge variant="secondary" className="h-5 px-2">
            {edgeFunctions.length}
          </Badge>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="text-muted-foreground">Active:</span>
          <Badge
            variant="secondary"
            className="h-5 bg-green-500/10 px-2 text-green-600 dark:text-green-400"
          >
            {edgeFunctions.filter((f) => f.enabled).length}
          </Badge>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="text-muted-foreground">Scheduled:</span>
          <Badge variant="secondary" className="h-5 px-2">
            {edgeFunctions.filter((f) => f.cron_schedule).length}
          </Badge>
        </div>
      </div>

      <ScrollArea className="h-[calc(100vh-20rem)]">
        <div className="grid gap-1">
          {edgeFunctions.length === 0 ? (
            <div className="p-2">
              <EmptyState
                icon={RefreshCw}
                title="No edge functions yet"
                description="Deploy serverless functions or start from a template"
                templates={[
                  { label: "HTTP Handler", onClick: onCreateFunction },
                  { label: "Webhook", onClick: onCreateFunction },
                  { label: "Scheduled Job", onClick: onCreateFunction },
                  { label: "Auth Middleware", onClick: onCreateFunction },
                ]}
                actions={[
                  {
                    label: "Create Function",
                    onClick: onCreateFunction,
                    icon: <Plus className="h-4 w-4" />,
                  },
                ]}
              />
            </div>
          ) : (
            edgeFunctions.map((fn) => (
              <div
                key={fn.id}
                className="hover:border-primary/50 bg-card flex items-center justify-between gap-2 rounded-md border px-3 py-1.5 transition-colors"
              >
                <div className="flex min-w-0 flex-1 items-center gap-2">
                  <span className="truncate text-sm font-medium">
                    {fn.name}
                  </span>
                  <Badge
                    variant="outline"
                    className="h-4 shrink-0 px-1 py-0 text-[10px]"
                  >
                    v{fn.version}
                  </Badge>
                  {fn.cron_schedule && (
                    <Badge
                      variant="outline"
                      className="h-4 shrink-0 px-1 py-0 text-[10px]"
                    >
                      <Clock className="mr-0.5 h-2.5 w-2.5" />
                      cron
                    </Badge>
                  )}
                  <Switch
                    checked={fn.enabled}
                    onCheckedChange={() => onToggleFunction(fn)}
                    className="scale-75"
                  />
                </div>
                <div className="flex shrink-0 items-center gap-0.5">
                  {fn.source === "filesystem" && fn.updated_at && (
                    <span
                      className="text-muted-foreground mr-1 text-[10px]"
                      title={`Last synced: ${new Date(fn.updated_at).toLocaleString()}`}
                    >
                      synced {new Date(fn.updated_at).toLocaleDateString()}
                    </span>
                  )}
                  <span className="text-muted-foreground mr-1 text-[10px]">
                    {fn.timeout_seconds}s
                  </span>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        onClick={() => onViewLogs(fn)}
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0"
                      >
                        <History className="h-3 w-3" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>View logs</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        onClick={() => onInvokeFunction(fn)}
                        size="sm"
                        variant="ghost"
                        className="h-6 w-6 p-0"
                        disabled={!fn.enabled}
                      >
                        <Play className="h-3 w-3" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Invoke function</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        onClick={() => onEditFunction(fn)}
                        size="sm"
                        variant="ghost"
                        className="h-6 w-6 p-0"
                      >
                        <Edit className="h-3 w-3" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Edit function</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        onClick={() => onDeleteFunction(fn.name)}
                        size="sm"
                        variant="ghost"
                        className="text-destructive hover:text-destructive hover:bg-destructive/10 h-6 w-6 p-0"
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Delete function</TooltipContent>
                  </Tooltip>
                </div>
              </div>
            ))
          )}
        </div>
      </ScrollArea>
    </>
  );
}
