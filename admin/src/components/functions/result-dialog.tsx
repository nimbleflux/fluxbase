import { Copy } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import type { InvokeResult } from "./types";

interface ResultDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  result: InvokeResult | null;
  wordWrap: boolean;
  onWordWrapChange: (wrap: boolean) => void;
}

export function ResultDialog({
  open,
  onOpenChange,
  result,
  wordWrap,
  onWordWrapChange,
}: ResultDialogProps) {
  const handleCopy = () => {
    if (result?.success) {
      navigator.clipboard.writeText(result.data);
      toast.success("Copied to clipboard");
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[95vh] w-[95vw] max-w-[95vw]">
        <DialogHeader>
          <DialogTitle>
            {result?.success ? "Function Result" : "Function Error"}
          </DialogTitle>
          <DialogDescription>
            {result?.success
              ? "Function executed successfully"
              : "Function execution failed"}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="flex items-center space-x-2">
            <Switch
              id="word-wrap"
              checked={wordWrap}
              onCheckedChange={onWordWrapChange}
            />
            <Label htmlFor="word-wrap" className="cursor-pointer">
              Word wrap
            </Label>
          </div>

          {result?.success ? (
            <div className="w-full overflow-hidden">
              <Label>Response</Label>
              <div className="bg-muted mt-2 h-[70vh] overflow-auto rounded-lg border">
                <pre
                  className={`p-4 text-xs ${
                    wordWrap
                      ? "break-words whitespace-pre-wrap"
                      : "whitespace-pre"
                  }`}
                >
                  {result.data}
                </pre>
              </div>
            </div>
          ) : (
            <div className="w-full overflow-hidden">
              <Label>Error</Label>
              <div className="bg-destructive/10 mt-2 h-[70vh] overflow-auto rounded-lg border">
                <pre
                  className={`text-destructive p-4 text-xs ${
                    wordWrap
                      ? "break-words whitespace-pre-wrap"
                      : "whitespace-pre"
                  }`}
                >
                  {result?.error}
                </pre>
              </div>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={handleCopy}
            disabled={!result?.success}
          >
            <Copy className="mr-2 h-4 w-4" />
            Copy
          </Button>
          <Button onClick={() => onOpenChange(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
