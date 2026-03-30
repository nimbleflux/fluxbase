import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { getPageNumbers } from "@/lib/utils";

interface PaginationControlsProps {
  currentPage: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
}

export function PaginationControls({
  currentPage,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
}: PaginationControlsProps) {
  const totalPages = Math.ceil(total / pageSize) || 1;

  if (total === 0) return null;

  return (
    <div className="flex items-center justify-between border-t px-2 py-3">
      <div className="flex items-center gap-2">
        <span className="text-muted-foreground text-sm">Rows per page</span>
        <Select
          value={`${pageSize}`}
          onValueChange={(value) => {
            onPageSizeChange(Number(value));
          }}
        >
          <SelectTrigger className="h-8 w-[70px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent side="top">
            {[10, 25, 50, 100].map((size) => (
              <SelectItem key={size} value={`${size}`}>
                {size}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="flex items-center gap-2">
        <span className="text-muted-foreground text-sm">
          Page {currentPage + 1} of {totalPages} ({total} total)
        </span>

        <Button
          variant="outline"
          size="sm"
          className="h-8 w-8 p-0"
          onClick={() => onPageChange(0)}
          disabled={currentPage === 0}
        >
          <ChevronsLeft className="h-4 w-4" />
        </Button>

        <Button
          variant="outline"
          size="sm"
          className="h-8 w-8 p-0"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 0}
        >
          <ChevronLeft className="h-4 w-4" />
        </Button>

        {getPageNumbers(currentPage + 1, totalPages).map((pageNum, idx) =>
          pageNum === "..." ? (
            <span
              key={`ellipsis-${idx}`}
              className="text-muted-foreground px-1"
            >
              ...
            </span>
          ) : (
            <Button
              key={pageNum}
              variant={currentPage + 1 === pageNum ? "default" : "outline"}
              size="sm"
              className="h-8 min-w-8 px-2"
              onClick={() => onPageChange((pageNum as number) - 1)}
            >
              {pageNum}
            </Button>
          ),
        )}

        <Button
          variant="outline"
          size="sm"
          className="h-8 w-8 p-0"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage >= totalPages - 1}
        >
          <ChevronRight className="h-4 w-4" />
        </Button>

        <Button
          variant="outline"
          size="sm"
          className="h-8 w-8 p-0"
          onClick={() => onPageChange(totalPages - 1)}
          disabled={currentPage >= totalPages - 1}
        >
          <ChevronsRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
