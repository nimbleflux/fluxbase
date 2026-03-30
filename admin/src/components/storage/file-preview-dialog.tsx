import { formatDistanceToNow } from "date-fns";
import { Copy, Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { toast } from "sonner";
import type { FilePreviewDialogProps } from "./types";

export function FilePreviewDialog({
  isOpen,
  onOpenChange,
  file,
  previewUrl,
  onDownload,
  formatBytes,
  formatJson,
  formatJsonWithHighlighting,
  isJsonFile,
}: FilePreviewDialogProps) {
  const handleCopy = () => {
    if (!file) return;
    const textToCopy = isJsonFile(file.mime_type, file.path)
      ? formatJson(previewUrl)
      : previewUrl;
    navigator.clipboard.writeText(textToCopy);
    toast.success("Copied to clipboard");
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-4xl">
        <DialogHeader>
          <DialogTitle>{file?.path}</DialogTitle>
          <DialogDescription>
            {file && (
              <div className="flex items-center gap-4 text-sm">
                <span>{formatBytes(file.size)}</span>
                <span>{file.mime_type}</span>
                <span>
                  {formatDistanceToNow(new Date(file.updated_at), {
                    addSuffix: true,
                  })}
                </span>
              </div>
            )}
          </DialogDescription>
        </DialogHeader>
        <ScrollArea className="max-h-[60vh]">
          {file?.mime_type?.startsWith("image/") ? (
            <img src={previewUrl} alt={file.path} className="w-full" />
          ) : isJsonFile(file?.mime_type, file?.path) ? (
            <div className="rounded-lg bg-slate-950 p-4">
              <pre className="font-mono text-sm">
                <code
                  className="language-json text-slate-100"
                  dangerouslySetInnerHTML={{
                    __html: formatJsonWithHighlighting(previewUrl),
                  }}
                />
              </pre>
            </div>
          ) : (
            <pre className="bg-muted/50 rounded p-4 font-mono text-sm">
              {previewUrl}
            </pre>
          )}
        </ScrollArea>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          {file && !file.mime_type?.startsWith("image/") && (
            <Button variant="outline" onClick={handleCopy}>
              <Copy className="mr-2 h-4 w-4" />
              Copy
            </Button>
          )}
          {file && (
            <Button onClick={onDownload}>
              <Download className="mr-2 h-4 w-4" />
              Download
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
