import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { CreateBucketDialogProps } from "./types";

export function CreateBucketDialog({
  isOpen,
  onOpenChange,
  bucketName,
  onBucketNameChange,
  onCreate,
  loading,
}: CreateBucketDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create New Bucket</DialogTitle>
          <DialogDescription>
            Enter a name for your new storage bucket
          </DialogDescription>
        </DialogHeader>
        <Input
          placeholder="my-bucket"
          value={bucketName}
          onChange={(e) => onBucketNameChange(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") onCreate();
          }}
        />
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onCreate} disabled={loading}>
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
