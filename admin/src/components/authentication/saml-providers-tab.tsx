import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Building2,
  Loader2,
  Check,
  Plus,
  Copy,
  Pencil,
  Trash2,
  FileText,
  Upload,
  Link,
} from "lucide-react";
import { toast } from "sonner";
import {
  samlProviderApi,
  type SAMLProviderConfig,
  type CreateSAMLProviderRequest,
  type UpdateSAMLProviderRequest,
} from "@/lib/api";
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
import { Switch } from "@/components/ui/switch";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { StringArrayEditor } from "@/components/string-array-editor";

export function SAMLProvidersTab() {
  const queryClient = useQueryClient();
  const baseUrl = window.location.origin;
  const [isAddProviderOpen, setIsAddProviderOpen] = useState(false);
  const [isEditProviderOpen, setIsEditProviderOpen] = useState(false);
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
  const [allowDashboardLogin, setAllowDashboardLogin] = useState(false);
  const [allowAppLogin, setAllowAppLogin] = useState(true);
  const [allowIdpInitiated, setAllowIdpInitiated] = useState(false);
  const [requiredGroups, setRequiredGroups] = useState<string[]>([]);
  const [requiredGroupsAll, setRequiredGroupsAll] = useState<string[]>([]);
  const [deniedGroups, setDeniedGroups] = useState<string[]>([]);
  const [groupAttribute, setGroupAttribute] = useState("groups");
  const [validatingMetadata, setValidatingMetadata] = useState(false);
  const [metadataValid, setMetadataValid] = useState<boolean | null>(null);
  const [metadataError, setMetadataError] = useState<string | null>(null);

  const { data: providers = [], isLoading } = useQuery({
    queryKey: ["samlProviders"],
    queryFn: samlProviderApi.list,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateSAMLProviderRequest) =>
      samlProviderApi.create(data),
    onSuccess: (data) => {
      toast.success(data.message);
      queryClient.invalidateQueries({ queryKey: ["samlProviders"] });
      setIsAddProviderOpen(false);
      resetForm();
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && "response" in error
          ? (error as { response?: { data?: { error?: string } } }).response
              ?.data?.error || "Failed to create SAML provider"
          : "Failed to create SAML provider";
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
    onSuccess: (data) => {
      toast.success(data.message);
      queryClient.invalidateQueries({ queryKey: ["samlProviders"] });
      setIsEditProviderOpen(false);
      setEditingProvider(null);
      resetForm();
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && "response" in error
          ? (error as { response?: { data?: { error?: string } } }).response
              ?.data?.error || "Failed to update SAML provider"
          : "Failed to update SAML provider";
      toast.error(errorMessage);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => samlProviderApi.delete(id),
    onSuccess: (data) => {
      toast.success(data.message);
      queryClient.invalidateQueries({ queryKey: ["samlProviders"] });
      setIsDeleteConfirmOpen(false);
      setDeletingProvider(null);
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && "response" in error
          ? (error as { response?: { data?: { error?: string } } }).response
              ?.data?.error || "Failed to delete SAML provider"
          : "Failed to delete SAML provider";
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
    setAllowDashboardLogin(false);
    setAllowAppLogin(true);
    setAllowIdpInitiated(false);
    setRequiredGroups([]);
    setRequiredGroupsAll([]);
    setDeniedGroups([]);
    setGroupAttribute("groups");
    setMetadataValid(null);
    setMetadataError(null);
  };

  const validateMetadata = async () => {
    setValidatingMetadata(true);
    setMetadataValid(null);
    setMetadataError(null);
    try {
      const result = await samlProviderApi.validateMetadata(
        metadataSource === "url" ? metadataUrl : undefined,
        metadataSource === "xml" ? metadataXml : undefined,
      );
      if (result.valid) {
        setMetadataValid(true);
        toast.success(`Metadata valid! IdP Entity ID: ${result.entity_id}`);
      } else {
        setMetadataValid(false);
        setMetadataError(result.error || "Invalid metadata");
      }
    } catch {
      setMetadataValid(false);
      setMetadataError("Failed to validate metadata");
    } finally {
      setValidatingMetadata(false);
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      const result = await samlProviderApi.uploadMetadata(file);
      if (result.valid && result.metadata) {
        setMetadataXml(result.metadata);
        setMetadataValid(true);
        toast.success(`Metadata uploaded! IdP Entity ID: ${result.entity_id}`);
      } else {
        setMetadataValid(false);
        setMetadataError(result.error || "Invalid metadata file");
      }
    } catch {
      setMetadataError("Failed to upload metadata file");
    }
  };

  const handleCreateProvider = () => {
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
      allow_dashboard_login: allowDashboardLogin,
      allow_app_login: allowAppLogin,
      allow_idp_initiated: allowIdpInitiated,
      ...(requiredGroups.length > 0 && { required_groups: requiredGroups }),
      ...(requiredGroupsAll.length > 0 && {
        required_groups_all: requiredGroupsAll,
      }),
      ...(deniedGroups.length > 0 && { denied_groups: deniedGroups }),
      group_attribute: groupAttribute || "groups",
    });
  };

  const handleEditProvider = (provider: SAMLProviderConfig) => {
    setEditingProvider(provider);
    setProviderName(provider.name);
    setDisplayName(provider.display_name);
    setMetadataUrl(provider.idp_metadata_url || "");
    setMetadataXml(provider.idp_metadata_xml || "");
    setMetadataSource(provider.idp_metadata_url ? "url" : "xml");
    setAutoCreateUsers(provider.auto_create_users);
    setDefaultRole(provider.default_role);
    setAllowDashboardLogin(provider.allow_dashboard_login);
    setAllowAppLogin(provider.allow_app_login);
    setAllowIdpInitiated(provider.allow_idp_initiated);
    setRequiredGroups(provider.required_groups || []);
    setRequiredGroupsAll(provider.required_groups_all || []);
    setDeniedGroups(provider.denied_groups || []);
    setGroupAttribute(provider.group_attribute || "groups");
    setIsEditProviderOpen(true);
  };

  const handleUpdateProvider = () => {
    if (!editingProvider) return;

    updateMutation.mutate({
      id: editingProvider.id,
      data: {
        display_name: displayName || undefined,
        idp_metadata_url: metadataSource === "url" ? metadataUrl : undefined,
        idp_metadata_xml: metadataSource === "xml" ? metadataXml : undefined,
        auto_create_users: autoCreateUsers,
        default_role: defaultRole,
        allow_dashboard_login: allowDashboardLogin,
        allow_app_login: allowAppLogin,
        allow_idp_initiated: allowIdpInitiated,
        ...(requiredGroups.length > 0 && { required_groups: requiredGroups }),
        ...(requiredGroupsAll.length > 0 && {
          required_groups_all: requiredGroupsAll,
        }),
        ...(deniedGroups.length > 0 && { denied_groups: deniedGroups }),
        group_attribute: groupAttribute || "groups",
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
              <CardTitle className="flex items-center gap-2">
                <Building2 className="h-5 w-5" />
                SAML SSO Providers
              </CardTitle>
              <CardDescription>
                Enterprise Single Sign-On via SAML 2.0. Configure Identity
                Providers like Okta, Azure AD, or OneLogin.
              </CardDescription>
            </div>
            <Button onClick={() => setIsAddProviderOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Add Provider
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {providers.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Building2 className="text-muted-foreground mb-4 h-12 w-12" />
              <p className="text-muted-foreground mb-2">
                No SAML providers configured
              </p>
              <Button
                variant="outline"
                onClick={() => setIsAddProviderOpen(true)}
              >
                <Plus className="mr-2 h-4 w-4" />
                Add your first SAML provider
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              {providers.map((provider) => (
                <Card
                  key={provider.id}
                  className={
                    provider.source === "config" ? "border-dashed" : ""
                  }
                >
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1 space-y-4">
                        <div className="flex flex-wrap items-center gap-2">
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
                          {provider.source === "config" && (
                            <Badge variant="outline">
                              <FileText className="mr-1 h-3 w-3" />
                              Config File
                            </Badge>
                          )}
                          {provider.tenant_id ? (
                            <Badge variant="outline">Tenant</Badge>
                          ) : (
                            <Badge variant="secondary">Instance</Badge>
                          )}
                          {provider.allow_dashboard_login && (
                            <Badge variant="outline">Dashboard Login</Badge>
                          )}
                          {provider.allow_app_login && (
                            <Badge variant="outline">App Login</Badge>
                          )}
                          {provider.auto_create_users && (
                            <Badge variant="outline">Auto-create Users</Badge>
                          )}
                        </div>

                        <div className="grid grid-cols-1 gap-4 text-sm md:grid-cols-2">
                          <div>
                            <Label className="text-muted-foreground">
                              Provider Name
                            </Label>
                            <p className="mt-1 font-mono text-xs">
                              {provider.name}
                            </p>
                          </div>
                          <div>
                            <Label className="text-muted-foreground">
                              Default Role
                            </Label>
                            <p className="mt-1 font-mono text-xs">
                              {provider.default_role}
                            </p>
                          </div>
                          <div>
                            <Label className="text-muted-foreground">
                              Entity ID (SP)
                            </Label>
                            <div className="mt-1 flex items-center gap-2">
                              <p className="flex-1 font-mono text-xs break-all">
                                {provider.entity_id}
                              </p>
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
                              <p className="flex-1 font-mono text-xs break-all">
                                {provider.acs_url}
                              </p>
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

                        <div className="mt-4 border-t pt-4">
                          <Label className="text-muted-foreground">
                            SP Metadata URL
                          </Label>
                          <div className="mt-1 flex items-center gap-2">
                            <code className="bg-muted flex-1 rounded px-2 py-1 text-xs">
                              {baseUrl}/api/v1/auth/saml/metadata/
                              {provider.name}
                            </code>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                copyToClipboard(
                                  `${baseUrl}/api/v1/auth/saml/metadata/${provider.name}`,
                                  "SP Metadata URL",
                                )
                              }
                            >
                              <Copy className="mr-1 h-3 w-3" />
                              Copy
                            </Button>
                          </div>
                        </div>

                        {(provider.required_groups ||
                          provider.required_groups_all ||
                          provider.denied_groups) && (
                          <div className="mt-4 border-t pt-4">
                            <Label className="text-muted-foreground mb-2 block">
                              RBAC Rules
                            </Label>
                            <div className="space-y-2">
                              {provider.required_groups &&
                                provider.required_groups.length > 0 && (
                                  <div>
                                    <span className="text-muted-foreground text-xs">
                                      Required Groups (OR):{" "}
                                    </span>
                                    <div className="mt-1 flex flex-wrap gap-1">
                                      {provider.required_groups.map((group) => (
                                        <Badge
                                          key={group}
                                          variant="outline"
                                          className="text-xs"
                                        >
                                          {group}
                                        </Badge>
                                      ))}
                                    </div>
                                  </div>
                                )}
                              {provider.required_groups_all &&
                                provider.required_groups_all.length > 0 && (
                                  <div>
                                    <span className="text-muted-foreground text-xs">
                                      Required Groups (AND):{" "}
                                    </span>
                                    <div className="mt-1 flex flex-wrap gap-1">
                                      {provider.required_groups_all.map(
                                        (group) => (
                                          <Badge
                                            key={group}
                                            variant="secondary"
                                            className="text-xs"
                                          >
                                            {group}
                                          </Badge>
                                        ),
                                      )}
                                    </div>
                                  </div>
                                )}
                              {provider.denied_groups &&
                                provider.denied_groups.length > 0 && (
                                  <div>
                                    <span className="text-muted-foreground text-xs">
                                      Denied Groups:{" "}
                                    </span>
                                    <div className="mt-1 flex flex-wrap gap-1">
                                      {provider.denied_groups.map((group) => (
                                        <Badge
                                          key={group}
                                          variant="destructive"
                                          className="text-xs"
                                        >
                                          {group}
                                        </Badge>
                                      ))}
                                    </div>
                                  </div>
                                )}
                              {provider.group_attribute &&
                                provider.group_attribute !== "groups" && (
                                  <div>
                                    <span className="text-muted-foreground text-xs">
                                      Group Attribute:{" "}
                                    </span>
                                    <Badge
                                      variant="outline"
                                      className="text-xs"
                                    >
                                      {provider.group_attribute}
                                    </Badge>
                                  </div>
                                )}
                            </div>
                          </div>
                        )}
                      </div>

                      {provider.source !== "config" && (
                        <div className="ml-4 flex gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleEditProvider(provider)}
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              setDeletingProvider(provider);
                              setIsDeleteConfirmOpen(true);
                            }}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={isAddProviderOpen} onOpenChange={setIsAddProviderOpen}>
        <DialogContent className="max-h-[90vh] max-w-2xl overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Add SAML Provider</DialogTitle>
            <DialogDescription>
              Configure a new SAML 2.0 Identity Provider for enterprise SSO.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Provider Name *</Label>
                <Input
                  placeholder="okta"
                  value={providerName}
                  onChange={(e) => setProviderName(e.target.value)}
                />
                <p className="text-muted-foreground text-xs">
                  Lowercase letters, numbers, underscores, hyphens only
                </p>
              </div>
              <div className="space-y-2">
                <Label>Display Name</Label>
                <Input
                  placeholder="Okta SSO"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label>IdP Metadata Source *</Label>
              <div className="flex gap-4">
                <Button
                  type="button"
                  variant={metadataSource === "url" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setMetadataSource("url")}
                >
                  <Link className="mr-2 h-4 w-4" />
                  URL
                </Button>
                <Button
                  type="button"
                  variant={metadataSource === "xml" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setMetadataSource("xml")}
                >
                  <Upload className="mr-2 h-4 w-4" />
                  Upload XML
                </Button>
              </div>
            </div>

            {metadataSource === "url" ? (
              <div className="space-y-2">
                <Label>IdP Metadata URL *</Label>
                <div className="flex gap-2">
                  <Input
                    placeholder="https://company.okta.com/app/xxx/sso/saml/metadata"
                    value={metadataUrl}
                    onChange={(e) => {
                      setMetadataUrl(e.target.value);
                      setMetadataValid(null);
                    }}
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    onClick={validateMetadata}
                    disabled={!metadataUrl || validatingMetadata}
                  >
                    {validatingMetadata ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : metadataValid ? (
                      <Check className="h-4 w-4 text-green-500" />
                    ) : metadataValid === false ? (
                      <X className="h-4 w-4 text-red-500" />
                    ) : (
                      "Validate"
                    )}
                  </Button>
                </div>
                {metadataError && (
                  <p className="text-xs text-red-500">{metadataError}</p>
                )}
              </div>
            ) : (
              <div className="space-y-2">
                <Label>IdP Metadata XML *</Label>
                <div className="space-y-2">
                  <Input
                    type="file"
                    accept=".xml,text/xml,application/xml"
                    onChange={handleFileUpload}
                  />
                  <textarea
                    className="h-32 w-full rounded-md border p-2 font-mono text-xs"
                    placeholder="Paste IdP metadata XML here..."
                    value={metadataXml}
                    onChange={(e) => {
                      setMetadataXml(e.target.value);
                      setMetadataValid(null);
                    }}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={validateMetadata}
                    disabled={!metadataXml || validatingMetadata}
                  >
                    {validatingMetadata ? (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : null}
                    Validate XML
                  </Button>
                  {metadataError && (
                    <p className="text-xs text-red-500">{metadataError}</p>
                  )}
                </div>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Default Role</Label>
                <Input
                  placeholder="authenticated"
                  value={defaultRole}
                  onChange={(e) => setDefaultRole(e.target.value)}
                />
              </div>
            </div>

            <div className="space-y-4 border-t pt-4">
              <Label className="text-base font-semibold">Options</Label>
              <div className="grid grid-cols-2 gap-4">
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Auto-create Users</Label>
                    <p className="text-muted-foreground text-xs">
                      Create user if not exists
                    </p>
                  </div>
                  <Switch
                    checked={autoCreateUsers}
                    onCheckedChange={setAutoCreateUsers}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Allow IdP-Initiated SSO</Label>
                    <p className="text-muted-foreground text-xs">Less secure</p>
                  </div>
                  <Switch
                    checked={allowIdpInitiated}
                    onCheckedChange={setAllowIdpInitiated}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Allow for App Users</Label>
                    <p className="text-muted-foreground text-xs">
                      End-user authentication
                    </p>
                  </div>
                  <Switch
                    checked={allowAppLogin}
                    onCheckedChange={setAllowAppLogin}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Allow for Dashboard</Label>
                    <p className="text-muted-foreground text-xs">Admin login</p>
                  </div>
                  <Switch
                    checked={allowDashboardLogin}
                    onCheckedChange={setAllowDashboardLogin}
                  />
                </div>
              </div>
            </div>

            <div className="space-y-4 border-t pt-4">
              <div>
                <Label className="text-sm font-semibold">
                  Role-Based Access Control (Optional)
                </Label>
                <p className="text-muted-foreground mt-1 text-xs">
                  Filter users based on SAML assertion groups/attributes
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="groupAttribute">Group Attribute Name</Label>
                <Input
                  id="groupAttribute"
                  value={groupAttribute}
                  onChange={(e) => setGroupAttribute(e.target.value)}
                  placeholder="groups"
                />
                <p className="text-muted-foreground text-xs">
                  SAML attribute containing group memberships (default:
                  "groups")
                </p>
              </div>

              <div className="space-y-2">
                <Label>Required Groups (OR logic)</Label>
                <p className="text-muted-foreground text-xs">
                  User must be in at least ONE of these groups
                </p>
                <StringArrayEditor
                  value={requiredGroups}
                  onChange={setRequiredGroups}
                  placeholder="FluxbaseAdmins"
                  addButtonText="Add Required Group"
                />
              </div>

              <div className="space-y-2">
                <Label>Required Groups (AND logic)</Label>
                <p className="text-muted-foreground text-xs">
                  User must be in ALL of these groups
                </p>
                <StringArrayEditor
                  value={requiredGroupsAll}
                  onChange={setRequiredGroupsAll}
                  placeholder="Verified"
                  addButtonText="Add Required Group"
                />
              </div>

              <div className="space-y-2">
                <Label>Denied Groups (Blocklist)</Label>
                <p className="text-muted-foreground text-xs">
                  Reject users in ANY of these groups
                </p>
                <StringArrayEditor
                  value={deniedGroups}
                  onChange={setDeniedGroups}
                  placeholder="Contractors"
                  addButtonText="Add Denied Group"
                />
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsAddProviderOpen(false);
                resetForm();
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreateProvider}
              disabled={createMutation.isPending}
            >
              {createMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : null}
              Create Provider
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={isEditProviderOpen} onOpenChange={setIsEditProviderOpen}>
        <DialogContent className="max-h-[90vh] max-w-2xl overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit SAML Provider</DialogTitle>
            <DialogDescription>
              Update the configuration for{" "}
              {editingProvider?.display_name || editingProvider?.name}.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Display Name</Label>
              <Input
                placeholder="Okta SSO"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <Label>IdP Metadata Source</Label>
              <div className="flex gap-4">
                <Button
                  type="button"
                  variant={metadataSource === "url" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setMetadataSource("url")}
                >
                  <Link className="mr-2 h-4 w-4" />
                  URL
                </Button>
                <Button
                  type="button"
                  variant={metadataSource === "xml" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setMetadataSource("xml")}
                >
                  <Upload className="mr-2 h-4 w-4" />
                  Upload XML
                </Button>
              </div>
            </div>

            {metadataSource === "url" ? (
              <div className="space-y-2">
                <Label>IdP Metadata URL</Label>
                <div className="flex gap-2">
                  <Input
                    placeholder="https://company.okta.com/app/xxx/sso/saml/metadata"
                    value={metadataUrl}
                    onChange={(e) => setMetadataUrl(e.target.value)}
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    onClick={validateMetadata}
                    disabled={!metadataUrl || validatingMetadata}
                  >
                    {validatingMetadata ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      "Validate"
                    )}
                  </Button>
                </div>
                {metadataError && (
                  <p className="text-xs text-red-500">{metadataError}</p>
                )}
              </div>
            ) : (
              <div className="space-y-2">
                <Label>IdP Metadata XML</Label>
                <Input
                  type="file"
                  accept=".xml,text/xml,application/xml"
                  onChange={handleFileUpload}
                />
                <textarea
                  className="h-32 w-full rounded-md border p-2 font-mono text-xs"
                  placeholder="Paste IdP metadata XML here..."
                  value={metadataXml}
                  onChange={(e) => {
                    setMetadataXml(e.target.value);
                    setMetadataValid(null);
                  }}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={validateMetadata}
                  disabled={!metadataXml || validatingMetadata}
                >
                  {validatingMetadata ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : null}
                  Validate XML
                </Button>
                {metadataError && (
                  <p className="text-xs text-red-500">{metadataError}</p>
                )}
              </div>
            )}

            <div className="space-y-2">
              <Label>Default Role</Label>
              <Input
                placeholder="authenticated"
                value={defaultRole}
                onChange={(e) => setDefaultRole(e.target.value)}
              />
            </div>

            <div className="space-y-4 border-t pt-4">
              <Label className="text-base font-semibold">Options</Label>
              <div className="grid grid-cols-2 gap-4">
                <div className="flex items-center justify-between">
                  <Label>Auto-create Users</Label>
                  <Switch
                    checked={autoCreateUsers}
                    onCheckedChange={setAutoCreateUsers}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <Label>Allow IdP-Initiated SSO</Label>
                  <Switch
                    checked={allowIdpInitiated}
                    onCheckedChange={setAllowIdpInitiated}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <Label>Allow for App Users</Label>
                  <Switch
                    checked={allowAppLogin}
                    onCheckedChange={setAllowAppLogin}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <Label>Allow for Dashboard</Label>
                  <Switch
                    checked={allowDashboardLogin}
                    onCheckedChange={setAllowDashboardLogin}
                  />
                </div>
              </div>
            </div>

            <div className="space-y-4 border-t pt-4">
              <div>
                <Label className="text-sm font-semibold">
                  Role-Based Access Control (Optional)
                </Label>
                <p className="text-muted-foreground mt-1 text-xs">
                  Filter users based on SAML assertion groups/attributes
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="editGroupAttribute">Group Attribute Name</Label>
                <Input
                  id="editGroupAttribute"
                  value={groupAttribute}
                  onChange={(e) => setGroupAttribute(e.target.value)}
                  placeholder="groups"
                />
                <p className="text-muted-foreground text-xs">
                  SAML attribute containing group memberships (default:
                  "groups")
                </p>
              </div>

              <div className="space-y-2">
                <Label>Required Groups (OR logic)</Label>
                <p className="text-muted-foreground text-xs">
                  User must be in at least ONE of these groups
                </p>
                <StringArrayEditor
                  value={requiredGroups}
                  onChange={setRequiredGroups}
                  placeholder="FluxbaseAdmins"
                  addButtonText="Add Required Group"
                />
              </div>

              <div className="space-y-2">
                <Label>Required Groups (AND logic)</Label>
                <p className="text-muted-foreground text-xs">
                  User must be in ALL of these groups
                </p>
                <StringArrayEditor
                  value={requiredGroupsAll}
                  onChange={setRequiredGroupsAll}
                  placeholder="Verified"
                  addButtonText="Add Required Group"
                />
              </div>

              <div className="space-y-2">
                <Label>Denied Groups (Blocklist)</Label>
                <p className="text-muted-foreground text-xs">
                  Reject users in ANY of these groups
                </p>
                <StringArrayEditor
                  value={deniedGroups}
                  onChange={setDeniedGroups}
                  placeholder="Contractors"
                  addButtonText="Add Denied Group"
                />
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsEditProviderOpen(false);
                resetForm();
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleUpdateProvider}
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : null}
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onOpenChange={setIsDeleteConfirmOpen}
        title="Delete SAML Provider"
        desc={`Are you sure you want to delete the SAML provider "${deletingProvider?.display_name || deletingProvider?.name}"? This action cannot be undone.`}
        confirmText="Delete"
        handleConfirm={() => {
          if (deletingProvider) {
            deleteMutation.mutate(deletingProvider.id);
          }
        }}
        isLoading={deleteMutation.isPending}
        destructive
      />
    </div>
  );
}

function X({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <path d="M18 6 6 18" />
      <path d="m6 6 12 12" />
    </svg>
  );
}
