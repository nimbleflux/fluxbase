import { FileText, RotateCcw, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { TemplateCardProps } from "./types";

export function TemplateCard({
  template,
  onEdit,
  onReset,
  onTest,
  isResetting,
  isTesting,
}: TemplateCardProps) {
  return (
    <Card className="relative">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span className="capitalize">
            {template.template_type.replace(/_/g, " ")}
          </span>
          {template.is_custom && <Badge variant="secondary">Custom</Badge>}
        </CardTitle>
        <CardDescription className="line-clamp-2">
          {template.subject}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        <Button
          variant="outline"
          className="w-full"
          onClick={() => onEdit(template)}
        >
          <FileText className="mr-2 h-4 w-4" />
          Edit Template
        </Button>
        {template.is_custom && (
          <Button
            variant="outline"
            className="w-full"
            onClick={() => onReset(template.template_type)}
            disabled={isResetting}
          >
            <RotateCcw className="mr-2 h-4 w-4" />
            Reset to Default
          </Button>
        )}
        <Button
          variant="outline"
          className="w-full"
          onClick={() => onTest(template.template_type)}
          disabled={isTesting}
        >
          <Send className="mr-2 h-4 w-4" />
          Send Test Email
        </Button>
      </CardContent>
    </Card>
  );
}
