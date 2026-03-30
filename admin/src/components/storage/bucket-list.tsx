import { HardDrive, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { BucketListProps } from "./types";

export function BucketList({
  buckets,
  selectedBucket,
  onSelectBucket,
  onDeleteBucket,
  objectCount,
  totalSize,
  formatBytes,
  onCreateBucket,
}: BucketListProps) {
  return (
    <div className="bg-muted/10 flex h-full w-64 flex-col border-r p-4">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="font-semibold">Buckets</h3>
        <Button variant="ghost" size="icon" onClick={onCreateBucket}>
          <Plus className="h-4 w-4" />
        </Button>
      </div>

      <ScrollArea className="min-h-0 flex-1">
        <div className="space-y-1">
          {buckets.map((bucket) => (
            <div
              key={bucket.id}
              className={`group hover:bg-muted/50 flex cursor-pointer items-center justify-between rounded p-2 ${
                selectedBucket === bucket.name ? "bg-muted" : ""
              }`}
              onClick={() => onSelectBucket(bucket.name)}
            >
              <div className="flex min-w-0 flex-1 items-center gap-2">
                <HardDrive className="h-4 w-4 flex-shrink-0" />
                <span className="truncate text-sm">{bucket.name}</span>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 opacity-0 group-hover:opacity-100"
                onClick={(e) => {
                  e.stopPropagation();
                  onDeleteBucket(bucket.name);
                }}
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            </div>
          ))}
          {buckets.length === 0 && (
            <p className="text-muted-foreground text-sm">No buckets</p>
          )}
        </div>
      </ScrollArea>

      {selectedBucket && (
        <div className="space-y-2 border-t pt-4">
          <div className="text-muted-foreground text-xs">
            <div className="flex justify-between">
              <span>Files:</span>
              <span>{objectCount}</span>
            </div>
            <div className="flex justify-between">
              <span>Total Size:</span>
              <span>{formatBytes(totalSize)}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
