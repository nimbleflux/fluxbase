import type { PolicyTemplate } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface TemplateCardProps {
  template: PolicyTemplate;
  onUse: () => void;
}

export function TemplateCard({ template, onUse }: TemplateCardProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{template.name}</CardTitle>
        <CardDescription>{template.description}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <Badge variant="outline">{template.command}</Badge>
          <pre className="bg-muted overflow-auto rounded p-2 text-xs">
            {template.using}
          </pre>
          {template.with_check && (
            <pre className="bg-muted overflow-auto rounded p-2 text-xs">
              WITH CHECK: {template.with_check}
            </pre>
          )}
          <Button size="sm" onClick={onUse} className="w-full">
            Use Template
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
