import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useFluxbaseClient } from "@nimbleflux/fluxbase-sdk-react";
import type { AIProvider } from "@nimbleflux/fluxbase-sdk";
import {
  Bot,
  Plus,
  Star,
  Sparkles,
  MoreHorizontal,
  Pencil,
  Trash2,
  Check,
} from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { CreateProviderDialog } from "./create-provider-dialog";
import { EditProviderDialog } from "./edit-provider-dialog";

export function AIProvidersTab() {
  const client = useFluxbaseClient();
  const queryClient = useQueryClient();

  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [editingProvider, setEditingProvider] = useState<AIProvider | null>(
    null,
  );
  const [deletingProvider, setDeletingProvider] = useState<AIProvider | null>(
    null,
  );

  const { data: providers = [], isLoading } = useQuery<AIProvider[]>({
    queryKey: ["ai-providers", client.admin.ai],
    queryFn: async () => {
      const { data, error } = await client.admin.ai.listProviders();
      if (error) throw error;
      return data ?? [];
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const { error } = await client.admin.ai.deleteProvider(id);
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ai-providers"] });
      setDeletingProvider(null);
      toast.success("Provider deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to delete provider");
    },
  });

  const setDefaultMutation = useMutation({
    mutationFn: async (id: string) => {
      const { error } = await client.admin.ai.setDefaultProvider(id);
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ai-providers"] });
      toast.success("Default provider updated");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to set default provider");
    },
  });

  const setEmbeddingMutation = useMutation({
    mutationFn: async (id: string) => {
      const { error } = await client.admin.ai.setEmbeddingProvider(id);
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ai-providers"] });
      toast.success("Embedding provider updated");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to set embedding provider");
    },
  });

  const clearEmbeddingMutation = useMutation({
    mutationFn: async (id: string) => {
      const { error } = await client.admin.ai.clearEmbeddingProvider(id);
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ai-providers"] });
      toast.success("Embedding provider cleared");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to clear embedding provider");
    },
  });

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Bot className="h-5 w-5" />
                AI Providers
              </CardTitle>
              <CardDescription>
                Manage AI providers for chatbot and embedding functionality.
                Providers can be configured here or via environment variables.
              </CardDescription>
            </div>
            <Button onClick={() => setShowCreateDialog(true)} size="sm">
              <Plus className="mr-1 h-4 w-4" />
              Add Provider
            </Button>
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
                Add an AI provider or configure one via environment variables to
                enable chatbot functionality
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
                  <TableHead className="w-[50px]" />
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
                          {provider.display_name || provider.name}
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
                        {provider.config?.model ||
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
                      <TableCell>
                        {!provider.from_config && (
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                              >
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem
                                onClick={() => setEditingProvider(provider)}
                              >
                                <Pencil className="mr-2 h-4 w-4" />
                                Edit
                              </DropdownMenuItem>
                              {!provider.is_default && (
                                <DropdownMenuItem
                                  onClick={() =>
                                    setDefaultMutation.mutate(provider.id)
                                  }
                                  disabled={setDefaultMutation.isPending}
                                >
                                  <Star className="mr-2 h-4 w-4" />
                                  Set as Default
                                </DropdownMenuItem>
                              )}
                              {!isEmbeddingProvider && (
                                <DropdownMenuItem
                                  onClick={() =>
                                    setEmbeddingMutation.mutate(provider.id)
                                  }
                                  disabled={setEmbeddingMutation.isPending}
                                >
                                  <Sparkles className="mr-2 h-4 w-4" />
                                  Set as Embedding Provider
                                </DropdownMenuItem>
                              )}
                              {isEmbeddingProvider && (
                                <DropdownMenuItem
                                  onClick={() =>
                                    clearEmbeddingMutation.mutate(provider.id)
                                  }
                                  disabled={clearEmbeddingMutation.isPending}
                                >
                                  <Check className="mr-2 h-4 w-4" />
                                  Clear Embedding Provider
                                </DropdownMenuItem>
                              )}
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="text-destructive"
                                onClick={() => setDeletingProvider(provider)}
                              >
                                <Trash2 className="mr-2 h-4 w-4" />
                                Delete
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
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

      <CreateProviderDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
      />

      <EditProviderDialog
        open={!!editingProvider}
        onOpenChange={(open) => {
          if (!open) setEditingProvider(null);
        }}
        provider={editingProvider}
      />

      <ConfirmDialog
        open={!!deletingProvider}
        onOpenChange={(open) => {
          if (!open) setDeletingProvider(null);
        }}
        title="Delete AI Provider"
        desc={`Are you sure you want to delete "${deletingProvider?.display_name || deletingProvider?.name}"? This action cannot be undone.`}
        confirmText="Delete"
        destructive
        handleConfirm={() => {
          if (deletingProvider) deleteMutation.mutate(deletingProvider.id);
        }}
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}
