import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export interface HeaderEntry {
  key: string;
  value: string;
}

interface HeadersEditorProps {
  headers: HeaderEntry[];
  onChange: (headers: HeaderEntry[]) => void;
}

export function HeadersEditor({ headers, onChange }: HeadersEditorProps) {
  const addHeader = () => {
    onChange([...headers, { key: "", value: "" }]);
  };

  const updateHeader = (
    index: number,
    field: "key" | "value",
    value: string,
  ) => {
    const updated = [...headers];
    updated[index][field] = value;
    onChange(updated);
  };

  const removeHeader = (index: number) => {
    onChange(headers.filter((_, i) => i !== index));
  };

  return (
    <div className="space-y-2">
      {headers.map((header, index) => (
        <div key={index} className="flex gap-2">
          <Input
            placeholder="Header name"
            value={header.key}
            onChange={(e) => updateHeader(index, "key", e.target.value)}
            className="flex-1"
          />
          <Input
            placeholder="Header value"
            value={header.value}
            onChange={(e) => updateHeader(index, "value", e.target.value)}
            className="flex-1"
          />
          <Button variant="ghost" size="sm" onClick={() => removeHeader(index)}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ))}
      <Button variant="outline" size="sm" onClick={addHeader}>
        <Plus className="mr-2 h-4 w-4" />
        Add Header
      </Button>
    </div>
  );
}
