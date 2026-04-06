import type { AdminStorageObject } from "@nimbleflux/fluxbase-sdk";

export interface StorageStats {
  files: number;
  totalSize: number;
}

export interface CreateBucketDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  bucketName: string;
  onBucketNameChange: (name: string) => void;
  onCreate: () => void;
  loading: boolean;
}

export interface CreateFolderDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  folderName: string;
  onFolderNameChange: (name: string) => void;
  onCreate: () => void;
  loading: boolean;
}

export interface DeleteConfirmDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  selectedCount: number;
  onDelete: () => void;
  loading: boolean;
}

export interface FilePreviewDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  file: AdminStorageObject | null;
  previewUrl: string;
  onDownload: () => void;
  formatBytes: (bytes: number) => string;
  formatJson: (text: string) => string;
  formatJsonWithHighlighting: (text: string) => string;
  isJsonFile: (contentType?: string, fileName?: string) => boolean;
}

export interface DeleteBucketDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  bucketName: string | null;
  onDelete: () => void;
}

export interface DeleteFileDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  filePath: string | null;
  onDelete: () => void;
}

export interface FileMetadataSheetProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  file: AdminStorageObject | null;
  currentPrefix: string;
  signedUrl: string;
  signedUrlExpiry: number;
  generatingUrl: boolean;
  onSignedUrlExpiryChange: (expiry: number) => void;
  onGenerateSignedUrl: () => void;
  onDownload: () => void;
  onPreview: () => void;
  getPublicUrl: (key: string) => string;
  formatBytes: (bytes: number) => string;
  getFileIcon: (contentType?: string) => React.ReactNode;
  copyToClipboard: (text: string, label: string) => void;
}

export interface BucketListProps {
  buckets: { id: string; name: string }[];
  selectedBucket: string;
  onSelectBucket: (name: string) => void;
  onDeleteBucket: (name: string) => void;
  objectCount: number;
  totalSize: number;
  formatBytes: (bytes: number) => string;
  onCreateBucket: () => void;
}

export interface FileListProps {
  prefixes: string[];
  objects: AdminStorageObject[];
  selectedFiles: Set<string>;
  currentPrefix: string;
  searchQuery: string;
  loading: boolean;
  dragActive: boolean;
  onNavigateToPrefix: (prefix: string) => void;
  onToggleFileSelection: (key: string) => void;
  onFileMetadata: (file: AdminStorageObject) => void;
  onFilePreview: (file: AdminStorageObject) => void;
  onFileDownload: (key: string) => void;
  onFileDelete: (key: string) => void;
  onDragEnter: (e: React.DragEvent) => void;
  onDragLeave: (e: React.DragEvent) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDrop: (e: React.DragEvent) => void;
  formatBytes: (bytes: number) => string;
  getFileIcon: (contentType?: string) => React.ReactNode;
}

export interface UploadProgressProps {
  uploadProgress: Record<string, number>;
}

export interface ToolbarProps {
  breadcrumbs: string[];
  searchQuery: string;
  sortBy: "name" | "size" | "date";
  fileTypeFilter: string;
  selectedCount: number;
  filteredCount: number;
  uploading: boolean;
  loading: boolean;
  onNavigateToPrefix: (prefix: string) => void;
  onSearchChange: (query: string) => void;
  onSortChange: (sort: "name" | "size" | "date") => void;
  onFileTypeFilterChange: (filter: string) => void;
  onSelectAll: () => void;
  onDeselectAll: () => void;
  onRefresh: () => void;
  onCreateFolder: () => void;
  onUpload: () => void;
  onDeleteSelected: () => void;
  fileInputRef: React.RefObject<HTMLInputElement | null>;
  onFileInputChange: (files: FileList | File[]) => void;
}

export interface FileTypeFilterChipsProps {
  fileTypeFilter: string;
  onFilterChange: (filter: string) => void;
}
