import { ConfirmDialog } from "@/components/confirm-dialog";
import type { DeleteBucketDialogProps } from "./types";

export function DeleteBucketDialog({
  isOpen,
  onOpenChange,
  bucketName,
  onDelete,
}: DeleteBucketDialogProps) {
  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={onOpenChange}
      title="Delete Bucket"
      desc={`Are you sure you want to delete the bucket "${bucketName}"? This action cannot be undone.`}
      confirmText="Delete"
      destructive
      handleConfirm={() => {
        onDelete();
        onOpenChange(false);
      }}
    />
  );
}
