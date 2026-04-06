import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, Shield, Check, Copy, Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  samlProviderApi,
  type SAMLProviderConfig,
  type CreateSAMLProviderRequest,
  type UpdateSAMLProviderRequest,
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
import { Textarea } from "@/components/ui/textarea";

export function TenantSAMLProvidersTab({ tenantId }: { tenantId: string }) {
  const queryClient = useQueryClient();
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [editingProvider, setEditingProvider] =
    useState<SAMLProviderConfig | null>(null);
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [deletingProvider, setDeletingProvider] =
    useState<SAMLProviderConfig | null>(null);

  const [providerName, setProviderName] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [metadataSource, setMetadataSource] = useState<"url" | "xml">("url");
  const [metadataUrl, setMetadataUrl] = useState("");
  const [metadataXml, setMetadataXml] = useState("");
  const [autoCreateUsers, setAutoCreateUsers] = useState(true);
  const [defaultRole, setDefaultRole] = useState("authenticated");
  const [allowAppLogin, setAllowAppLogin] = useState(true);

  const { data: providers = [], isLoading } = useQuery({
    queryKey: ["tenant-saml-providers", tenantId],
    queryFn: samlProviderApi.list,
  });

  const tenantProviders = providers.filter(
    (p) => !p.allow_dashboard_login || p.source === "database",
  );

  const createMutation = useMutation({
    mutationFn: (data: CreateSAMLProviderRequest) =>
      samlProviderApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["tenant-saml-providers", tenantId],
      });
      toast.success("SAML provider created");
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
      data: UpdateSAMLProviderRequest;
    }) => samlProviderApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["tenant-saml-providers", tenantId],
      });
      toast.success("SAML provider updated");
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
    mutationFn: (id: string) => samlProviderApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["tenant-saml-providers", tenantId],
      });
      toast.success("SAML provider deleted");
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
    setProviderName("");
    setDisplayName("");
    setMetadataSource("url");
    setMetadataUrl("");
    setMetadataXml("");
    setAutoCreateUsers(true);
    setDefaultRole("authenticated");
    setAllowAppLogin(true);
  };

  const handleCreate = () => {
    if (!providerName) {
      toast.error("Provider name is required");
      return;
    }
    if (metadataSource === "url" && !metadataUrl) {
      toast.error("Metadata URL is required");
      return;
    }
    if (metadataSource === "xml" && !metadataXml) {
      toast.error("Metadata XML is required");
      return;
    }

    createMutation.mutate({
      name: providerName.toLowerCase().replace(/[^a-z0-9_-]/g, "_"),
      display_name: displayName || providerName,
      enabled: true,
      idp_metadata_url: metadataSource === "url" ? metadataUrl : undefined,
      idp_metadata_xml: metadataSource === "xml" ? metadataXml : undefined,
      auto_create_users: autoCreateUsers,
      default_role: defaultRole,
      allow_dashboard_login: false,
      allow_app_login: allowAppLogin,
    });
  };

  const handleEdit = (provider: SAMLProviderConfig) => {
    setEditingProvider(provider);
    setProviderName(provider.name);
    setDisplayName(provider.display_name);
    setMetadataUrl(provider.idp_metadata_url || "");
    setMetadataXml(provider.idp_metadata_xml || "");
    setMetadataSource(provider.idp_metadata_url ? "url" : "xml");
    setAutoCreateUsers(provider.auto_create_users);
    setDefaultRole(provider.default_role);
    setAllowAppLogin(provider.allow_app_login);
    setIsEditDialogOpen(true);
  };

  const handleUpdate = () => {
    if (!editingProvider) return;
    updateMutation.mutate({
      id: editingProvider.id,
      data: {
        display_name: displayName || undefined,
        idp_metadata_url: metadataSource === "url" ? metadataUrl : undefined,
        idp_metadata_xml: metadataSource === "xml" ? metadataXml : undefined,
        auto_create_users: autoCreateUsers,
        default_role: defaultRole,
        allow_app_login: allowAppLogin,
      },
    });
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast.success(`${label} copied to clipboard`);
  };

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
              <CardTitle>SAML SSO Providers</CardTitle>
              <CardDescription>
                Configure SAML providers for enterprise single sign-on
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
              <Shield className="text-muted-foreground mb-4 h-12 w-12" />
              <p className="text-muted-foreground mb-2">
                No SAML providers configured
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
                      <div className="flex-1 space-y-4">
                        <div className="flex items-center gap-2">
                          <h3 className="text-lg font-semibold">
                            {provider.display_name || provider.name}
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
                            <Badge variant="outline">App Login</Badge>
                          )}
                        </div>
                        <div className="grid grid-cols-1 gap-4 text-sm md:grid-cols-2">
                          <div>
                            <Label className="text-muted-foreground">
                              Entity ID (SP)
                            </Label>
                            <div className="mt-1 flex items-center gap-2">
                              <code className="flex-1 rounded bg-muted px-2 py-1 text-xs break-all">
                                {provider.entity_id}
                              </code>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0"
                                onClick={() =>
                                  copyToClipboard(
                                    provider.entity_id,
                                    "Entity ID",
                                  )
                                }
                              >
                                <Copy className="h-3 w-3" />
                              </Button>
                            </div>
                          </div>
                          <div>
                            <Label className="text-muted-foreground">
                              ACS URL
                            </Label>
                            <div className="mt-1 flex items-center gap-2">
                              <code className="flex-1 rounded bg-muted px-2 py-1 text-xs break-all">
                                {provider.acs_url}
                              </code>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0"
                                onClick={() =>
                                  copyToClipboard(provider.acs_url, "ACS URL")
                                }
                              >
                                <Copy className="h-3 w-3" />
                              </Button>
                            </div>
                          </div>
                        </div>
                      </div>
                      <div className="ml-4 flex gap-2">
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
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Add SAML Provider</DialogTitle>
            <DialogDescription>
              Configure a new SAML provider for this tenant
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Provider Name</Label>
                <Input
                  value={providerName}
                  onChange={(e) => setProviderName(e.target.value)}
                  placeholder="okta"
                />
              </div>
              <div className="grid gap-2">
                <Label>Display Name</Label>
                <Input
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="Okta SSO"
                />
              </div>
            </div>
            <div className="grid gap-2">
              <Label>Metadata Source</Label>
              <Select
                value={metadataSource}
                onValueChange={(v) => setMetadataSource(v as "url" | "xml")}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="url">Metadata URL</SelectItem>
                  <SelectItem value="xml">Metadata XML</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {metadataSource === "url" ? (
              <div className="grid gap-2">
                <Label>IdP Metadata URL</Label>
                <Input
                  value={metadataUrl}
                  onChange={(e) => setMetadataUrl(e.target.value)}
                  placeholder="https://idp.example.com/metadata"
                />
              </div>
            ) : (
              <div className="grid gap-2">
                <Label>IdP Metadata XML</Label>
                <Textarea
                  value={metadataXml}
                  onChange={(e) => setMetadataXml(e.target.value)}
                  placeholder="Paste metadata XML here..."
                  rows={6}
                />
              </div>
            )}
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Default Role</Label>
                <Input
                  value={defaultRole}
                  onChange={(e) => setDefaultRole(e.target.value)}
                  placeholder="authenticated"
                />
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <Switch
                  checked={autoCreateUsers}
                  onCheckedChange={setAutoCreateUsers}
                />
                <Label>Auto-create Users</Label>
              </div>
              <div className="flex items-center gap-2">
                <Switch
                  checked={allowAppLogin}
                  onCheckedChange={setAllowAppLogin}
                />
                <Label>Allow App Login</Label>
              </div>
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
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit SAML Provider</DialogTitle>
            <DialogDescription>
              Update the SAML provider configuration
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Display Name</Label>
                <Input
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                />
              </div>
              <div className="grid gap-2">
                <Label>Default Role</Label>
                <Input
                  value={defaultRole}
                  onChange={(e) => setDefaultRole(e.target.value)}
                />
              </div>
            </div>
            <div className="grid gap-2">
              <Label>Metadata Source</Label>
              <Select
                value={metadataSource}
                onValueChange={(v) => setMetadataSource(v as "url" | "xml")}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="url">Metadata URL</SelectItem>
                  <SelectItem value="xml">Metadata XML</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {metadataSource === "url" ? (
              <div className="grid gap-2">
                <Label>IdP Metadata URL</Label>
                <Input
                  value={metadataUrl}
                  onChange={(e) => setMetadataUrl(e.target.value)}
                />
              </div>
            ) : (
              <div className="grid gap-2">
                <Label>IdP Metadata XML</Label>
                <Textarea
                  value={metadataXml}
                  onChange={(e) => setMetadataXml(e.target.value)}
                  rows={6}
                />
              </div>
            )}
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <Switch
                  checked={autoCreateUsers}
                  onCheckedChange={setAutoCreateUsers}
                />
                <Label>Auto-create Users</Label>
              </div>
              <div className="flex items-center gap-2">
                <Switch
                  checked={allowAppLogin}
                  onCheckedChange={setAllowAppLogin}
                />
                <Label>Allow App Login</Label>
              </div>
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
            <AlertDialogTitle>Delete SAML Provider</AlertDialogTitle>
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
