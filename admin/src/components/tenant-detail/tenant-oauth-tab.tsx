import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, Check, AlertCircle, Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  oauthProviderApi,
  type OAuthProviderConfig,
  type CreateOAuthProviderRequest,
  type UpdateOAuthProviderRequest,
} from "@/lib/api";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

export function TenantOAuthProvidersTab({ tenantId }: { tenantId: string }) {
  const queryClient = useQueryClient();
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [editingProvider, setEditingProvider] =
    useState<OAuthProviderConfig | null>(null);
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [deletingProvider, setDeletingProvider] =
    useState<OAuthProviderConfig | null>(null);

  const [selectedProvider, setSelectedProvider] = useState("");
  const [customProviderName, setCustomProviderName] = useState("");
  const [clientId, setClientId] = useState("");
  const [clientSecret, setClientSecret] = useState("");
  const [customAuthUrl, setCustomAuthUrl] = useState("");
  const [customTokenUrl, setCustomTokenUrl] = useState("");
  const [customUserInfoUrl, setCustomUserInfoUrl] = useState("");
  const [allowAppLogin, setAllowAppLogin] = useState(true);

  const { data: providers = [], isLoading } = useQuery({
    queryKey: ["tenant-oauth-providers", tenantId],
    queryFn: oauthProviderApi.list,
  });

  const tenantProviders = providers.filter(
    (p) => !p.allow_dashboard_login || p.source === "database",
  );

  const createMutation = useMutation({
    mutationFn: (data: CreateOAuthProviderRequest) =>
      oauthProviderApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["tenant-oauth-providers", tenantId],
      });
      toast.success("OAuth provider created");
      setIsAddDialogOpen(false);
      resetForm();
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && "response" in error
          ? (error as { response?: { data?: { error?: string } } }).response
              ?.data?.error || "Failed to create provider"
          : "Failed to create provider";
      toast.error(errorMessage);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: UpdateOAuthProviderRequest;
    }) => oauthProviderApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["tenant-oauth-providers", tenantId],
      });
      toast.success("OAuth provider updated");
      setIsEditDialogOpen(false);
      setEditingProvider(null);
      resetForm();
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && "response" in error
          ? (error as { response?: { data?: { error?: string } } }).response
              ?.data?.error || "Failed to update provider"
          : "Failed to update provider";
      toast.error(errorMessage);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => oauthProviderApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["tenant-oauth-providers", tenantId],
      });
      toast.success("OAuth provider deleted");
      setIsDeleteConfirmOpen(false);
      setDeletingProvider(null);
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && "response" in error
          ? (error as { response?: { data?: { error?: string } } }).response
              ?.data?.error || "Failed to delete provider"
          : "Failed to delete provider";
      toast.error(errorMessage);
    },
  });

  const resetForm = () => {
    setSelectedProvider("");
    setCustomProviderName("");
    setClientId("");
    setClientSecret("");
    setCustomAuthUrl("");
    setCustomTokenUrl("");
    setCustomUserInfoUrl("");
    setAllowAppLogin(true);
  };

  const handleCreate = () => {
    if (!selectedProvider || !clientId || !clientSecret) {
      toast.error("Please fill in all required fields");
      return;
    }

    const isCustom = selectedProvider === "custom";
    createMutation.mutate({
      provider_name: isCustom
        ? customProviderName.toLowerCase().replace(/[^a-z0-9_-]/g, "_")
        : selectedProvider,
      display_name: isCustom
        ? customProviderName
        : selectedProvider.charAt(0).toUpperCase() + selectedProvider.slice(1),
      enabled: true,
      client_id: clientId,
      client_secret: clientSecret,
      redirect_url: `${window.location.origin}/api/v1/auth/oauth/callback`,
      scopes:
        selectedProvider === "google"
          ? ["openid", "email", "profile"]
          : selectedProvider === "github"
            ? ["read:user", "user:email"]
            : selectedProvider === "microsoft"
              ? ["openid", "email", "profile"]
              : ["openid", "email", "profile"],
      is_custom: isCustom,
      authorization_url: isCustom ? customAuthUrl : undefined,
      token_url: isCustom ? customTokenUrl : undefined,
      user_info_url: isCustom ? customUserInfoUrl : undefined,
      allow_dashboard_login: false,
      allow_app_login: allowAppLogin,
    });
  };

  const handleEdit = (provider: OAuthProviderConfig) => {
    setEditingProvider(provider);
    setClientId(provider.client_id);
    setClientSecret("");
    setAllowAppLogin(provider.allow_app_login);
    if (provider.is_custom) {
      setSelectedProvider("custom");
      setCustomProviderName(provider.display_name);
      setCustomAuthUrl(provider.authorization_url || "");
      setCustomTokenUrl(provider.token_url || "");
      setCustomUserInfoUrl(provider.user_info_url || "");
    } else {
      setSelectedProvider(provider.provider_name);
    }
    setIsEditDialogOpen(true);
  };

  const handleUpdate = () => {
    if (!editingProvider) return;
    updateMutation.mutate({
      id: editingProvider.id,
      data: {
        client_id: clientId || undefined,
        ...(clientSecret && { client_secret: clientSecret }),
        allow_app_login: allowAppLogin,
      },
    });
  };

  const availableProviders = [
    { id: "google", name: "Google" },
    { id: "github", name: "GitHub" },
    { id: "microsoft", name: "Microsoft" },
    { id: "custom", name: "Custom Provider" },
  ];

  if (isLoading) {
    return (
      <div className="flex justify-center p-8">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>OAuth Providers</CardTitle>
              <CardDescription>
                Configure OAuth providers for tenant authentication
              </CardDescription>
            </div>
            <Button onClick={() => setIsAddDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Add Provider
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {tenantProviders.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <AlertCircle className="text-muted-foreground mb-4 h-12 w-12" />
              <p className="text-muted-foreground mb-2">
                No OAuth providers configured
              </p>
              <Button
                variant="outline"
                onClick={() => setIsAddDialogOpen(true)}
              >
                Add Your First Provider
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              {tenantProviders.map((provider) => (
                <Card key={provider.id}>
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1 space-y-2">
                        <div className="flex items-center gap-2">
                          <h3 className="text-lg font-semibold">
                            {provider.display_name}
                          </h3>
                          {provider.enabled ? (
                            <Badge variant="default" className="gap-1">
                              <Check className="h-3 w-3" />
                              Enabled
                            </Badge>
                          ) : (
                            <Badge variant="secondary">Disabled</Badge>
                          )}
                          {provider.allow_app_login && (
                            <Badge variant="outline" className="text-xs">
                              App
                            </Badge>
                          )}
                        </div>
                        <div className="text-sm">
                          <span className="text-muted-foreground">
                            Client ID:{" "}
                          </span>
                          <code className="font-mono text-xs">
                            {provider.client_id}
                          </code>
                        </div>
                      </div>
                      <div className="flex gap-2">
                        {provider.source !== "config" && (
                          <>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleEdit(provider)}
                            >
                              Edit
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-destructive hover:text-destructive"
                              onClick={() => {
                                setDeletingProvider(provider);
                                setIsDeleteConfirmOpen(true);
                              }}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add OAuth Provider</DialogTitle>
            <DialogDescription>
              Configure a new OAuth provider for this tenant
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label>Provider</Label>
              <Select
                value={selectedProvider}
                onValueChange={setSelectedProvider}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select provider" />
                </SelectTrigger>
                <SelectContent>
                  {availableProviders.map((p) => (
                    <SelectItem key={p.id} value={p.id}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {selectedProvider === "custom" && (
              <div className="grid gap-2">
                <Label>Provider Name</Label>
                <Input
                  value={customProviderName}
                  onChange={(e) => setCustomProviderName(e.target.value)}
                  placeholder="my-provider"
                />
              </div>
            )}
            <div className="grid gap-2">
              <Label>Client ID</Label>
              <Input
                value={clientId}
                onChange={(e) => setClientId(e.target.value)}
                placeholder="Enter client ID"
              />
            </div>
            <div className="grid gap-2">
              <Label>Client Secret</Label>
              <Input
                type="password"
                value={clientSecret}
                onChange={(e) => setClientSecret(e.target.value)}
                placeholder="Enter client secret"
              />
            </div>
            {selectedProvider === "custom" && (
              <>
                <div className="grid gap-2">
                  <Label>Authorization URL</Label>
                  <Input
                    value={customAuthUrl}
                    onChange={(e) => setCustomAuthUrl(e.target.value)}
                    placeholder="https://provider.com/oauth/authorize"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>Token URL</Label>
                  <Input
                    value={customTokenUrl}
                    onChange={(e) => setCustomTokenUrl(e.target.value)}
                    placeholder="https://provider.com/oauth/token"
                  />
                </div>
                <div className="grid gap-2">
                  <Label>User Info URL</Label>
                  <Input
                    value={customUserInfoUrl}
                    onChange={(e) => setCustomUserInfoUrl(e.target.value)}
                    placeholder="https://provider.com/oauth/userinfo"
                  />
                </div>
              </>
            )}
            <div className="flex items-center gap-2">
              <Switch
                checked={allowAppLogin}
                onCheckedChange={setAllowAppLogin}
              />
              <Label>Allow App Login</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending}>
              {createMutation.isPending ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit OAuth Provider</DialogTitle>
            <DialogDescription>
              Update the OAuth provider configuration
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label>Client ID</Label>
              <Input
                value={clientId}
                onChange={(e) => setClientId(e.target.value)}
              />
            </div>
            <div className="grid gap-2">
              <Label>Client Secret (leave empty to keep current)</Label>
              <Input
                type="password"
                value={clientSecret}
                onChange={(e) => setClientSecret(e.target.value)}
                placeholder="Enter new secret or leave empty"
              />
            </div>
            <div className="flex items-center gap-2">
              <Switch
                checked={allowAppLogin}
                onCheckedChange={setAllowAppLogin}
              />
              <Label>Allow App Login</Label>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsEditDialogOpen(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleUpdate} disabled={updateMutation.isPending}>
              {updateMutation.isPending ? "Updating..." : "Update"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={isDeleteConfirmOpen}
        onOpenChange={setIsDeleteConfirmOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete OAuth Provider</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;
              {deletingProvider?.display_name}&quot;? This action cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() =>
                deletingProvider && deleteMutation.mutate(deletingProvider.id)
              }
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
