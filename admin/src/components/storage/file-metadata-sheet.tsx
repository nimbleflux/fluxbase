import { formatDistanceToNow } from "date-fns";
import {
  Calendar,
  Copy,
  Download,
  Eye,
  FileType,
  HardDrive,
  Link,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import type { FileMetadataSheetProps } from "./types";

export function FileMetadataSheet({
  isOpen,
  onOpenChange,
  file,
  currentPrefix,
  signedUrl,
  signedUrlExpiry,
  generatingUrl,
  onSignedUrlExpiryChange,
  onGenerateSignedUrl,
  onDownload,
  onPreview,
  getPublicUrl,
  formatBytes,
  getFileIcon,
  copyToClipboard,
}: FileMetadataSheetProps) {
  return (
    <Sheet open={isOpen} onOpenChange={onOpenChange}>
      <SheetContent className="w-full overflow-y-auto sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>File Details</SheetTitle>
          <SheetDescription>View and manage file metadata</SheetDescription>
        </SheetHeader>

        {file && (
          <div className="mt-6 space-y-6">
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                {getFileIcon(file.mime_type)}
                <div className="min-w-0 flex-1">
                  <h3 className="truncate font-medium">
                    {file.path.replace(currentPrefix, "")}
                  </h3>
                  <p className="text-muted-foreground truncate text-sm">
                    {file.path}
                  </p>
                </div>
              </div>

              <div className="grid gap-3">
                <div className="flex items-center justify-between border-b py-2">
                  <div className="text-muted-foreground flex items-center gap-2 text-sm">
                    <HardDrive className="h-4 w-4" />
                    <span>Size</span>
                  </div>
                  <span className="text-sm font-medium">
                    {formatBytes(file.size)}
                  </span>
                </div>

                <div className="flex items-center justify-between border-b py-2">
                  <div className="text-muted-foreground flex items-center gap-2 text-sm">
                    <FileType className="h-4 w-4" />
                    <span>Type</span>
                  </div>
                  <Badge variant="outline" className="text-xs">
                    {file.mime_type || "Unknown"}
                  </Badge>
                </div>

                <div className="flex items-center justify-between border-b py-2">
                  <div className="text-muted-foreground flex items-center gap-2 text-sm">
                    <Calendar className="h-4 w-4" />
                    <span>Modified</span>
                  </div>
                  <span className="text-sm font-medium">
                    {formatDistanceToNow(new Date(file.updated_at), {
                      addSuffix: true,
                    })}
                  </span>
                </div>
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">Public URL</label>
              <div className="flex gap-2">
                <Input
                  value={getPublicUrl(file.path)}
                  readOnly
                  className="flex-1 font-mono text-xs"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() =>
                    copyToClipboard(getPublicUrl(file.path), "URL")
                  }
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>

            <div className="space-y-3 border-t pt-4">
              <div className="flex items-center gap-2">
                <Link className="h-4 w-4" />
                <h4 className="font-medium">Generate Signed URL</h4>
              </div>
              <p className="text-muted-foreground text-sm">
                Create a temporary URL with an expiration time for secure file
                sharing.
              </p>

              <div className="space-y-2">
                <label className="text-sm font-medium">Expires In</label>
                <Select
                  value={signedUrlExpiry.toString()}
                  onValueChange={(val) =>
                    onSignedUrlExpiryChange(parseInt(val))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="900">15 minutes</SelectItem>
                    <SelectItem value="1800">30 minutes</SelectItem>
                    <SelectItem value="3600">1 hour</SelectItem>
                    <SelectItem value="7200">2 hours</SelectItem>
                    <SelectItem value="21600">6 hours</SelectItem>
                    <SelectItem value="86400">24 hours</SelectItem>
                    <SelectItem value="604800">7 days</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <Button
                onClick={onGenerateSignedUrl}
                disabled={generatingUrl}
                className="w-full"
              >
                {generatingUrl ? (
                  <>
                    <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                    Generating...
                  </>
                ) : (
                  <>
                    <Link className="mr-2 h-4 w-4" />
                    Generate Signed URL
                  </>
                )}
              </Button>

              {signedUrl && (
                <div className="bg-muted space-y-2 rounded-lg p-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Signed URL</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(signedUrl, "Signed URL")}
                    >
                      <Copy className="mr-1 h-3 w-3" />
                      Copy
                    </Button>
                  </div>
                  <p className="text-muted-foreground font-mono text-xs break-all">
                    {signedUrl}
                  </p>
                  <p className="text-muted-foreground text-xs">
                    Expires in{" "}
                    {signedUrlExpiry < 3600
                      ? `${signedUrlExpiry / 60} minutes`
                      : `${signedUrlExpiry / 3600} hours`}
                  </p>
                </div>
              )}
            </div>

            {file.metadata && Object.keys(file.metadata).length > 0 && (
              <div className="space-y-3 border-t pt-4">
                <h4 className="font-medium">Custom Metadata</h4>
                <div className="space-y-2">
                  {Object.entries(file.metadata).map(([key, value]) => (
                    <div
                      key={key}
                      className="flex items-center justify-between border-b py-2"
                    >
                      <span className="text-muted-foreground text-sm">
                        {key}
                      </span>
                      <span className="max-w-[200px] truncate text-sm font-medium">
                        {typeof value === "object"
                          ? JSON.stringify(value)
                          : String(value)}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="flex gap-2 border-t pt-4">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => onDownload()}
              >
                <Download className="mr-2 h-4 w-4" />
                Download
              </Button>
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => {
                  onOpenChange(false);
                  onPreview();
                }}
              >
                <Eye className="mr-2 h-4 w-4" />
                Preview
              </Button>
            </div>
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
