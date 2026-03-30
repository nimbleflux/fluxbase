import { FolderPlus, Home, RefreshCw, Upload, Trash2 } from "lucide-react";
import { ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ImpersonationPopover } from "@/features/impersonation/components/impersonation-popover";
import { Badge } from "@/components/ui/badge";
import { Image as ImageIcon, FileCode, FileText, FileCog } from "lucide-react";
import type { ToolbarProps, FileTypeFilterChipsProps } from "./types";

function FileTypeFilterChips({
  fileTypeFilter,
  onFilterChange,
}: FileTypeFilterChipsProps) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <Badge
        variant={fileTypeFilter === "all" ? "default" : "outline"}
        className="cursor-pointer"
        onClick={() => onFilterChange("all")}
      >
        All Files
      </Badge>
      <Badge
        variant={fileTypeFilter === "image" ? "default" : "outline"}
        className="cursor-pointer"
        onClick={() => onFilterChange("image")}
      >
        <ImageIcon className="mr-1 h-3 w-3" />
        Images
      </Badge>
      <Badge
        variant={fileTypeFilter === "video" ? "default" : "outline"}
        className="cursor-pointer"
        onClick={() => onFilterChange("video")}
      >
        <FileCode className="mr-1 h-3 w-3" />
        Videos
      </Badge>
      <Badge
        variant={fileTypeFilter === "audio" ? "default" : "outline"}
        className="cursor-pointer"
        onClick={() => onFilterChange("audio")}
      >
        <FileText className="mr-1 h-3 w-3" />
        Audio
      </Badge>
      <Badge
        variant={fileTypeFilter === "document" ? "default" : "outline"}
        className="cursor-pointer"
        onClick={() => onFilterChange("document")}
      >
        <FileText className="mr-1 h-3 w-3" />
        Documents
      </Badge>
      <Badge
        variant={fileTypeFilter === "code" ? "default" : "outline"}
        className="cursor-pointer"
        onClick={() => onFilterChange("code")}
      >
        <FileCode className="mr-1 h-3 w-3" />
        Code
      </Badge>
      <Badge
        variant={fileTypeFilter === "archive" ? "default" : "outline"}
        className="cursor-pointer"
        onClick={() => onFilterChange("archive")}
      >
        <FileCog className="mr-1 h-3 w-3" />
        Archives
      </Badge>
    </div>
  );
}

export function Toolbar({
  breadcrumbs,
  searchQuery,
  sortBy,
  fileTypeFilter,
  selectedCount,
  filteredCount,
  uploading,
  loading,
  onNavigateToPrefix,
  onSearchChange,
  onSortChange,
  onFileTypeFilterChange,
  onSelectAll,
  onDeselectAll,
  onRefresh,
  onCreateFolder,
  onUpload,
  onDeleteSelected,
  fileInputRef,
  onFileInputChange,
}: ToolbarProps) {
  return (
    <div className="space-y-4 border-b p-4">
      <div className="flex items-center gap-2 text-sm">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onNavigateToPrefix("")}
          className="h-7 px-2"
        >
          <Home className="h-3 w-3" />
        </Button>
        {breadcrumbs.map((crumb, i) => (
          <div key={i} className="flex items-center gap-2">
            <ChevronRight className="text-muted-foreground h-3 w-3" />
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                const prefix = breadcrumbs.slice(0, i + 1).join("/") + "/";
                onNavigateToPrefix(prefix);
              }}
              className="h-7 px-2"
            >
              {crumb}
            </Button>
          </div>
        ))}
      </div>

      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <div className="flex flex-1 items-center gap-2">
            <div className="relative max-w-sm flex-1">
              <Input
                placeholder="Search files..."
                value={searchQuery}
                onChange={(e) => onSearchChange(e.target.value)}
              />
            </div>
            <Select
              value={sortBy}
              onValueChange={(v) => onSortChange(v as "name" | "size" | "date")}
            >
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="name">Name</SelectItem>
                <SelectItem value="size">Size</SelectItem>
                <SelectItem value="date">Date</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={loading}
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </Button>

          <input
            ref={fileInputRef}
            type="file"
            multiple
            className="hidden"
            onChange={(e) => {
              if (e.target.files) {
                onFileInputChange(e.target.files);
                e.target.value = "";
              }
            }}
          />

          <Button variant="outline" onClick={onCreateFolder} size="sm">
            <FolderPlus className="mr-2 h-4 w-4" />
            New Folder
          </Button>

          <Button onClick={onUpload} disabled={uploading} size="sm">
            <Upload className="mr-2 h-4 w-4" />
            {uploading ? "Uploading..." : "Upload"}
          </Button>

          <ImpersonationPopover
            contextLabel="Browsing as"
            defaultReason="Storage browser testing"
          />
        </div>

        <FileTypeFilterChips
          fileTypeFilter={fileTypeFilter}
          onFilterChange={onFileTypeFilterChange}
        />

        {filteredCount > 0 && (
          <div className="flex items-center gap-2">
            <Checkbox
              checked={selectedCount === filteredCount && filteredCount > 0}
              onCheckedChange={(checked) => {
                if (checked) {
                  onSelectAll();
                } else {
                  onDeselectAll();
                }
              }}
            />
            <span className="text-muted-foreground text-sm">
              {selectedCount === 0 ? "Select All" : `${selectedCount} selected`}
            </span>
          </div>
        )}

        {selectedCount > 0 && (
          <Button variant="destructive" size="sm" onClick={onDeleteSelected}>
            <Trash2 className="mr-2 h-4 w-4" />
            Delete ({selectedCount})
          </Button>
        )}
      </div>
    </div>
  );
}
