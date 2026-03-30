import {
  BookOpen,
  FileText,
  Search,
  Settings,
  Trash2,
  Lock,
  Globe,
  Users,
} from "lucide-react";
import type { KnowledgeBaseCardProps } from "./types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

export function KnowledgeBaseCard({
  kb,
  onToggleEnabled,
  onDelete,
  onNavigate,
}: KnowledgeBaseCardProps) {
  return (
    <Card key={kb.id} className="relative">
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <BookOpen className="h-5 w-5" />
            <CardTitle className="text-lg">{kb.name}</CardTitle>
          </div>
          <div className="flex items-center gap-1">
            <Switch
              checked={kb.enabled}
              onCheckedChange={() => onToggleEnabled(kb)}
            />
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 w-8 p-0"
                  onClick={() => onNavigate(`/knowledge-bases/${kb.id}`)}
                >
                  <FileText className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>View Documents</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 w-8 p-0"
                  onClick={() => onNavigate(`/knowledge-bases/${kb.id}/search`)}
                >
                  <Search className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Search</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 w-8 p-0"
                  onClick={() =>
                    onNavigate(`/knowledge-bases/${kb.id}/settings`)
                  }
                >
                  <Settings className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Settings</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-destructive hover:text-destructive h-8 w-8 p-0"
                  onClick={() => onDelete(kb.id)}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Delete</TooltipContent>
            </Tooltip>
          </div>
        </div>
        {kb.namespace !== "default" && (
          <Badge variant="outline" className="w-fit text-[10px]">
            {kb.namespace}
          </Badge>
        )}
        <Badge
          variant="outline"
          className={`flex w-fit items-center gap-1 text-[10px] ${
            kb.visibility === "private"
              ? "bg-amber-500/10 text-amber-600 dark:text-amber-400"
              : kb.visibility === "public"
                ? "bg-blue-500/10 text-blue-600 dark:text-blue-400"
                : "bg-purple-500/10 text-purple-600 dark:text-purple-400"
          }`}
        >
          {kb.visibility === "private" && <Lock className="h-3 w-3" />}
          {kb.visibility === "public" && <Globe className="h-3 w-3" />}
          {kb.visibility === "shared" && <Users className="h-3 w-3" />}
          {kb.visibility || "private"}
        </Badge>
      </CardHeader>
      <CardContent>
        {kb.description && (
          <CardDescription className="mb-3 line-clamp-2">
            {kb.description}
          </CardDescription>
        )}
        <div className="flex flex-wrap gap-2 text-xs">
          <Badge variant="secondary">
            {kb.document_count}{" "}
            {kb.document_count === 1 ? "document" : "documents"}
          </Badge>
          <Badge variant="secondary">{kb.total_chunks} chunks</Badge>
          {kb.embedding_model && (
            <Badge variant="outline" className="text-[10px]">
              {kb.embedding_model}
            </Badge>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
