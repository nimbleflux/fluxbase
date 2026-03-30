import * as React from "react";
import { Loader } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type SelectDropdownProps = {
  onValueChange?: (value: string) => void;
  defaultValue: string | undefined;
  placeholder?: string;
  isPending?: boolean;
  items: { label: string; value: string }[] | undefined;
  disabled?: boolean;
  className?: string;
  isControlled?: boolean;
  // Form control props (passed through from FormControl via Slot)
  id?: string;
  "aria-describedby"?: string;
  "aria-invalid"?: boolean;
};

export const SelectDropdown = React.forwardRef<
  HTMLButtonElement,
  SelectDropdownProps
>(function SelectDropdown(
  {
    defaultValue,
    onValueChange,
    isPending,
    items,
    placeholder,
    disabled,
    className = "",
    isControlled = false,
    id,
    "aria-describedby": ariaDescribedBy,
    "aria-invalid": ariaInvalid,
  },
  ref,
) {
  const defaultState = isControlled
    ? { value: defaultValue, onValueChange }
    : { defaultValue, onValueChange };
  return (
    <Select {...defaultState}>
      <SelectTrigger
        ref={ref}
        disabled={disabled}
        className={cn(className)}
        id={id}
        aria-describedby={ariaDescribedBy}
        aria-invalid={ariaInvalid}
      >
        <SelectValue placeholder={placeholder ?? "Select"} />
      </SelectTrigger>
      <SelectContent>
        {isPending ? (
          <SelectItem disabled value="loading" className="h-14">
            <div className="flex items-center justify-center gap-2">
              <Loader className="h-5 w-5 animate-spin" />
              {"  "}
              Loading...
            </div>
          </SelectItem>
        ) : (
          items?.map(({ label, value }) => (
            <SelectItem key={value} value={value}>
              {label}
            </SelectItem>
          ))
        )}
      </SelectContent>
    </Select>
  );
});
