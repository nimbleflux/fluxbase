import { AlertCircle, Copy } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { ServiceKeyWithPlaintext } from "@/lib/api";

interface CreatedKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  createdKey: ServiceKeyWithPlaintext | null;
}

export function CreatedKeyDialog({
  open,
  onOpenChange,
  createdKey,
}: CreatedKeyDialogProps) {
  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Service Key Created</DialogTitle>
          <DialogDescription>
            Save this key now. You won't be able to see it again!
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="rounded-md bg-yellow-50 p-4 dark:bg-yellow-950">
            <div className="flex">
              <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
              <div className="ml-3">
                <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                  Important: Copy this key now
                </h3>
                <div className="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
                  <p>
                    This is the only time you'll see the full service key. Store
                    it securely.
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div className="grid gap-2">
            <Label>Service Key</Label>
            <div className="flex gap-2">
              <Input
                value={createdKey?.key || ""}
                readOnly
                className="font-mono text-xs"
              />
              <Button
                variant="outline"
                size="icon"
                onClick={() => copyToClipboard(createdKey?.key || "")}
              >
                <Copy className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <div className="grid gap-2">
            <Label>Name</Label>
            <Input value={createdKey?.name || ""} readOnly />
          </div>
          <div className="grid gap-2">
            <Label>Scopes</Label>
            <div className="flex flex-wrap gap-1">
              {createdKey?.scopes.map((scope) => (
                <Badge key={scope} variant="outline">
                  {scope}
                </Badge>
              ))}
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>
            I've Saved the Key
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
