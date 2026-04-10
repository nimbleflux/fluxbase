import { useQuery } from "@tanstack/react-query";
import type { AIProvider } from "@nimbleflux/fluxbase-sdk";
import { useFluxbaseClient } from "@nimbleflux/fluxbase-sdk-react";
import { Bot, Star, Sparkles } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export function AIProvidersTab() {
  const client = useFluxbaseClient();

  // Fetch providers using SDK
  const { data: providers = [], isLoading } = useQuery<AIProvider[]>({
    queryKey: ["ai-providers", client.admin.ai],
    queryFn: async () => {
      const { data, error } = await client.admin.ai.listProviders();
      if (error) throw error;
      return data ?? [];
    },
  });

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div>
            <CardTitle className="flex items-center gap-2">
              <Bot className="h-5 w-5" />
              AI Providers
            </CardTitle>
            <CardDescription>
              AI providers configured for chatbot and embedding functionality.
              Providers are managed via environment variables or instance settings.
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="py-8 text-center">
              <p className="text-muted-foreground">Loading providers...</p>
            </div>
          ) : providers.length === 0 ? (
            <div className="py-8 text-center">
              <Bot className="text-muted-foreground mx-auto mb-4 h-12 w-12" />
              <p className="mb-1 text-lg font-medium">
                No providers configured
              </p>
              <p className="text-muted-foreground text-sm">
                Configure an AI provider via environment variables to enable chatbot functionality
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Model</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {providers.map((provider) => {
                  const isEmbeddingProvider =
                    provider.use_for_embeddings === true;
                  const isAutoEmbedding =
                    provider.use_for_embeddings === null && provider.is_default;

                  return (
                    <TableRow key={provider.id}>
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          {provider.display_name}
                          {provider.is_default && (
                            <Badge variant="default" className="text-xs">
                              <Star className="mr-1 h-3 w-3" />
                              Default
                            </Badge>
                          )}
                          {(isEmbeddingProvider || isAutoEmbedding) && (
                            <Badge
                              variant={
                                isEmbeddingProvider ? "default" : "outline"
                              }
                              className="text-xs"
                            >
                              <Sparkles className="mr-1 h-3 w-3" />
                              Embeddings {isAutoEmbedding && "(auto)"}
                            </Badge>
                          )}
                          {provider.from_config && (
                            <Badge variant="secondary" className="text-xs">
                              Config
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {provider.provider_type}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {provider.config.model ||
                          (provider.provider_type === "openai"
                            ? "gpt-4-turbo"
                            : "-")}
                      </TableCell>
                      <TableCell>
                        {provider.enabled ? (
                          <Badge
                            variant="outline"
                            className="border-green-500 text-green-500"
                          >
                            Enabled
                          </Badge>
                        ) : (
                          <Badge
                            variant="outline"
                            className="border-gray-500 text-gray-500"
                          >
                            Disabled
                          </Badge>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
