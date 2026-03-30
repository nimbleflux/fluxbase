import { formatDistanceToNow } from "date-fns";
import {
  Clock,
  ChevronRight,
  Download,
  Eye,
  FolderOpen,
  Info,
  Trash2,
  Upload,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import type { FileListProps } from "./types";

export function FileList({
  prefixes,
  objects,
  selectedFiles,
  currentPrefix,
  loading,
  dragActive,
  onNavigateToPrefix,
  onToggleFileSelection,
  onFileMetadata,
  onFilePreview,
  onFileDownload,
  onFileDelete,
  onDragEnter,
  onDragLeave,
  onDragOver,
  onDrop,
  formatBytes,
  getFileIcon,
}: FileListProps) {
  return (
    <div
      className="flex-1 overflow-auto p-4"
      onDragEnter={onDragEnter}
      onDragLeave={onDragLeave}
      onDragOver={onDragOver}
      onDrop={onDrop}
    >
      {dragActive && (
        <div className="bg-primary/10 border-primary fixed inset-0 z-50 flex items-center justify-center border-4 border-dashed">
          <div className="text-center">
            <Upload className="text-primary mx-auto mb-4 h-12 w-12" />
            <p className="text-lg font-semibold">Drop files to upload</p>
          </div>
        </div>
      )}

      <div className="space-y-2">
        {prefixes.map((prefix) => (
          <Card
            key={prefix}
            className="hover:bg-muted/50 cursor-pointer p-3 transition-colors"
            onClick={() => onNavigateToPrefix(prefix)}
          >
            <div className="flex items-center gap-3">
              <FolderOpen className="h-5 w-5 text-blue-500" />
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">
                  {prefix.replace(currentPrefix, "").replace("/", "")}
                </p>
              </div>
              <ChevronRight className="text-muted-foreground h-4 w-4" />
            </div>
          </Card>
        ))}

        {objects.map((obj) => (
          <Card
            key={obj.path}
            className="hover:bg-muted/50 p-3 transition-colors"
          >
            <div className="flex items-center gap-3">
              <Checkbox
                checked={selectedFiles.has(obj.path)}
                onCheckedChange={() => onToggleFileSelection(obj.path)}
              />
              {getFileIcon(obj.mime_type)}
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">
                  {obj.path.replace(currentPrefix, "")}
                </p>
                <div className="text-muted-foreground flex items-center gap-3 text-xs">
                  <span>{formatBytes(obj.size)}</span>
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    {formatDistanceToNow(new Date(obj.updated_at), {
                      addSuffix: true,
                    })}
                  </span>
                  {obj.mime_type && (
                    <Badge variant="outline" className="text-xs">
                      {obj.mime_type.split("/")[1]}
                    </Badge>
                  )}
                </div>
              </div>
              <div className="flex gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => onFileMetadata(obj)}
                  title="File info"
                >
                  <Info className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => onFilePreview(obj)}
                  title="Preview"
                >
                  <Eye className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => onFileDownload(obj.path)}
                  title="Download"
                >
                  <Download className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => onFileDelete(obj.path)}
                  title="Delete"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </Card>
        ))}

        {objects.length === 0 && prefixes.length === 0 && !loading && (
          <div className="text-muted-foreground py-12 text-center">
            <FolderOpen className="mx-auto mb-4 h-12 w-12 opacity-50" />
            <p>No files in this folder</p>
            <p className="mt-2 text-sm">
              Drag and drop files here or click Upload
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
