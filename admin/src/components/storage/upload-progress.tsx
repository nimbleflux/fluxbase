import type { UploadProgressProps } from "./types";

export function UploadProgress({ uploadProgress }: UploadProgressProps) {
  const entries = Object.entries(uploadProgress);
  if (entries.length === 0) return null;

  return (
    <div className="bg-muted/40 space-y-3 border-b p-4">
      <div className="text-sm font-medium">Uploading files...</div>
      {entries.map(([filename, progress]) => (
        <div key={filename} className="space-y-1.5">
          <div className="flex items-center justify-between text-xs">
            <span className="text-muted-foreground flex-1 truncate">
              {filename}
            </span>
            <span className="ml-2 font-medium">{progress}%</span>
          </div>
          <div className="bg-muted relative h-2 w-full overflow-hidden rounded-full">
            <div
              className="bg-primary h-full transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}
