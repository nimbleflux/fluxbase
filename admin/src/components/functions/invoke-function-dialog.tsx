import { Play, RefreshCw, Code, AlignLeft } from "lucide-react";
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
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { HeadersEditor, type HeaderEntry } from "./headers-editor";
import type { EdgeFunction } from "@/lib/api";

interface InvokeFunctionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedFunction: EdgeFunction | null;
  invokeMethod: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  onInvokeMethodChange: (
    method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH",
  ) => void;
  invokeBody: string;
  onInvokeBodyChange: (body: string) => void;
  invokeHeaders: HeaderEntry[];
  onInvokeHeadersChange: (headers: HeaderEntry[]) => void;
  invoking: boolean;
  onSubmit: () => void;
}

export function InvokeFunctionDialog({
  open,
  onOpenChange,
  selectedFunction,
  invokeMethod,
  onInvokeMethodChange,
  invokeBody,
  onInvokeBodyChange,
  invokeHeaders,
  onInvokeHeadersChange,
  invoking,
  onSubmit,
}: InvokeFunctionDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] w-[90vw] max-w-5xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Invoke Edge Function</DialogTitle>
          <DialogDescription>
            Test {selectedFunction?.name} with custom HTTP request
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <Label htmlFor="method">HTTP Method</Label>
            <Select
              value={invokeMethod}
              onValueChange={(value) =>
                onInvokeMethodChange(
                  value as "GET" | "POST" | "PUT" | "DELETE" | "PATCH",
                )
              }
            >
              <SelectTrigger className="w-[180px]" id="method">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="GET">GET</SelectItem>
                <SelectItem value="POST">POST</SelectItem>
                <SelectItem value="PUT">PUT</SelectItem>
                <SelectItem value="DELETE">DELETE</SelectItem>
                <SelectItem value="PATCH">PATCH</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <Tabs defaultValue="body" className="space-y-4">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="body" className="flex items-center gap-2">
                <Code className="h-4 w-4" />
                Body
              </TabsTrigger>
              <TabsTrigger value="headers" className="flex items-center gap-2">
                <AlignLeft className="h-4 w-4" />
                Headers
              </TabsTrigger>
            </TabsList>

            <TabsContent value="body" className="space-y-2">
              <Label htmlFor="invoke-body">Request Body (JSON)</Label>
              <Textarea
                id="invoke-body"
                className="min-h-[300px] font-mono text-sm"
                value={invokeBody}
                onChange={(e) => onInvokeBodyChange(e.target.value)}
                placeholder='{"key": "value"}'
              />
            </TabsContent>

            <TabsContent value="headers" className="space-y-2">
              <Label>Custom Headers</Label>
              <ScrollArea className="max-h-[350px]">
                <HeadersEditor
                  headers={invokeHeaders}
                  onChange={onInvokeHeadersChange}
                />
              </ScrollArea>
            </TabsContent>
          </Tabs>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={invoking}>
            {invoking ? (
              <>
                <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                Invoking...
              </>
            ) : (
              <>
                <Play className="mr-2 h-4 w-4" />
                Invoke
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
