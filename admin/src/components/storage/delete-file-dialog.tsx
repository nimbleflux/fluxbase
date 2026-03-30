import { ConfirmDialog } from "@/components/confirm-dialog";
import type { DeleteFileDialogProps } from "./types";

export function DeleteFileDialog({
  isOpen,
  onOpenChange,
  filePath,
  onDelete,
}: DeleteFileDialogProps) {
  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={onOpenChange}
      title="Delete File"
      desc={`Are you sure you want to delete "${filePath}"? This action cannot be undone.`}
      confirmText="Delete"
      destructive
      handleConfirm={() => {
        onDelete();
        onOpenChange(false);
      }}
    />
  );
}
