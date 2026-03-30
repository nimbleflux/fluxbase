import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { TemplateEditorProps } from "./types";

export function TemplateEditor({
  selectedTemplate,
  editingTemplate,
  isSaving,
  onSave,
  onCancel,
  onUpdate,
}: TemplateEditorProps) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="capitalize">
              Edit {selectedTemplate.replace(/_/g, " ")} Template
            </CardTitle>
            <CardDescription>
              Customize the email template with variables like {"{{.AppName}}"},{" "}
              {"{{.MagicLink}}"}, etc.
            </CardDescription>
          </div>
          <Button variant="outline" onClick={onCancel}>
            Back to Templates
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="subject">Subject</Label>
          <Input
            id="subject"
            value={editingTemplate?.subject || ""}
            onChange={(e) =>
              onUpdate({
                ...editingTemplate,
                subject: e.target.value,
              })
            }
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="html_body">HTML Body</Label>
          <Textarea
            id="html_body"
            value={editingTemplate?.html_body || ""}
            onChange={(e) =>
              onUpdate({
                ...editingTemplate,
                html_body: e.target.value,
              })
            }
            rows={15}
            className="font-mono text-sm"
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="text_body">Text Body (Optional)</Label>
          <Textarea
            id="text_body"
            value={editingTemplate?.text_body || ""}
            onChange={(e) =>
              onUpdate({
                ...editingTemplate,
                text_body: e.target.value,
              })
            }
            rows={10}
            className="font-mono text-sm"
          />
        </div>

        <div className="flex gap-2">
          <Button onClick={onSave} disabled={isSaving}>
            {isSaving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Save Template
          </Button>
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
