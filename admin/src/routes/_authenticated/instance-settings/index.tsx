import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  Settings,
  Lock,
  Shield,
  EyeOff,
  RefreshCw,
  Pencil,
  Save,
  X,
  CheckCircle,
  Cpu,
  Mail,
  HardDrive,
  ShieldCheck,
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
} from "@/components/ui/alert-dialog";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
} from "@/components/ui/dialog";
import { instanceSettingsApi } from "@/lib/api";

export const Route = createFileRoute("/_authenticated/instance-settings/")({
  component: InstanceSettingsPage,
});

// Resolved setting from API
interface ResolvedSetting {
  value: unknown;
  source: "config" | "instance" | "tenant" | "default";
  is_read_only?: boolean;
  is_overridable?: boolean;
  is_secret?: boolean;
  data_type?: "string" | "number" | "boolean" | "object" | "array";
}

// Setting template definition
interface SettingDefinition {
  path: string;
  name: string;
  description?: string;
}

// Settings category
interface SettingsCategory {
  id: string;
  name: string;
  description?: string;
  icon: React.ComponentType<{ className?: string }>;
  settings: SettingDefinition[];
}

// Default settings categories
const defaultCategories: SettingsCategory[] = [
  {
    id: "ai",
    name: "AI",
    description: "AI provider and model settings",
    icon: Cpu,
    settings: [
      {
        path: "ai.enabled",
        name: "Enable AI",
        description: "Enable or disable AI features",
      },
      {
        path: "ai.default_provider",
        name: "Default Provider",
        description: "Default AI provider to use",
      },
      {
        path: "ai.default_model",
        name: "Default Model",
        description: "Default model for AI operations",
      },
    ],
  },
  {
    id: "email",
    name: "Email",
    description: "Email provider settings",
    icon: Mail,
    settings: [
      {
        path: "email.provider",
        name: "Provider",
        description: "Email service provider",
      },
      {
        path: "email.from_address",
        name: "From Address",
        description: "Default sender email address",
      },
      {
        path: "email.from_name",
        name: "From Name",
        description: "Default sender name",
      },
    ],
  },
  {
    id: "storage",
    name: "Storage",
    description: "Storage limits and allowed file types",
    icon: HardDrive,
    settings: [
      {
        path: "storage.max_upload_size",
        name: "Max Upload Size",
        description: "Maximum file upload size in bytes",
      },
      {
        path: "storage.provider",
        name: "Provider",
        description: "Storage backend provider",
      },
    ],
  },
  {
    id: "auth",
    name: "Authentication",
    description: "OAuth and SAML settings",
    icon: ShieldCheck,
    settings: [
      {
        path: "auth.signup_enabled",
        name: "Enable Signup",
        description: "Allow new user registration",
      },
      {
        path: "auth.magic_link_enabled",
        name: "Enable Magic Link",
        description: "Allow passwordless login via magic link",
      },
    ],
  },
  {
    id: "security",
    name: "Security",
    description: "Instance-level security settings",
    icon: Shield,
    settings: [
      {
        path: "security.enable_global_rate_limit",
        name: "Global Rate Limit",
        description: "Apply rate limits to all API endpoints",
      },
    ],
  },
];

function InstanceSettingsPage() {
  const queryClient = useQueryClient();
  const [activeCategory, setActiveCategory] = useState("ai");
  const [editingPath, setEditingPath] = useState<string | null>(null);
  const [editValue, setEditValue] = useState<unknown>(null);
  const [showSecretDialog, setShowSecretDialog] = useState(false);
  const [secretPath, setSecretPath] = useState<string | null>(null);
  const [showResetDialog, setShowResetDialog] = useState(false);
  const [resetPath, setResetPath] = useState<string | null>(null);
  const [showOverridableDialog, setShowOverridableDialog] = useState(false);

  // Fetch instance settings
  const { data: settingsData, isLoading } = useQuery({
    queryKey: ["instance-settings"],
    queryFn: () => instanceSettingsApi.get(),
  });

  // Fetch overridable settings list
  const { data: overridableData } = useQuery({
    queryKey: ["instance-settings", "overridable"],
    queryFn: () => instanceSettingsApi.getOverridable(),
  });

  // Update instance settings mutation
  const updateMutation = useMutation({
    mutationFn: (data: { settings: Record<string, unknown> }) =>
      instanceSettingsApi.update(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["instance-settings"] });
      toast.success("Settings updated successfully");
      setEditingPath(null);
      setEditValue(null);
    },
    onError: (error: Error) => {
      toast.error(`Failed to update settings: ${error.message}`);
    },
  });

  // Update overridable settings mutation
  const updateOverridableMutation = useMutation({
    mutationFn: (data: { overridable_settings: string[] }) =>
      instanceSettingsApi.updateOverridable(data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["instance-settings", "overridable"],
      });
      toast.success("Overridable settings updated");
      setShowOverridableDialog(false);
    },
    onError: (error: Error) => {
      toast.error(`Failed to update overridable settings: ${error.message}`);
    },
  });

  // Get current category settings
  const currentCategory = defaultCategories.find(
    (c) => c.id === activeCategory,
  );
  const categorySettings = currentCategory?.settings || [];

  // Get resolved setting from API data
  const getResolvedSetting = (path: string): ResolvedSetting => {
    const setting = settingsData?.settings?.[path] as
      | ResolvedSetting
      | undefined;
    return (
      setting || {
        value: undefined,
        source: "default" as const,
        is_read_only: false,
        is_overridable: false,
        is_secret: false,
      }
    );
  };

  // Handle edit
  const handleEdit = (path: string) => {
    const setting = getResolvedSetting(path);
    setEditingPath(path);
    setEditValue(setting.value);
  };

  // Handle save
  const handleSave = () => {
    if (editingPath && editValue !== null) {
      updateMutation.mutate({
        settings: { [editingPath]: editValue },
      });
    }
  };

  // Handle cancel
  const handleCancel = () => {
    setEditingPath(null);
    setEditValue(null);
  };

  // Handle secret save
  const handleSecretSave = (path: string, value: string) => {
    updateMutation.mutate({
      settings: { [path]: value },
    });
  };

  // Handle reset to default
  const handleReset = (path: string) => {
    setResetPath(path);
    setShowResetDialog(true);
  };

  // Confirm reset
  const confirmReset = () => {
    if (!resetPath) return;

    // Create a patch with the setting removed (set to undefined)
    const updatedSettings = { ...settingsData?.settings };
    delete updatedSettings[resetPath];

    updateMutation.mutate({
      settings: updatedSettings,
    });
    setShowResetDialog(false);
    setResetPath(null);
  };

  // Check if setting is overridable
  const isOverridable = (path: string): boolean => {
    return (
      overridableData?.overridable_settings?.some(
        (p: string) => path === p || path.startsWith(`${p}.`),
      ) ?? false
    );
  };

  // Render setting value based on type
  const renderValue = (value: unknown, type?: string): string => {
    if (value === null || value === undefined) return "Not set";

    switch (type || typeof value) {
      case "boolean":
        return value ? "Yes" : "No";
      case "number":
        return value.toLocaleString();
      case "object":
        return JSON.stringify(value, null, 2);
      case "array":
        return `${(value as unknown[]).length} items`;
      default:
        return String(value);
    }
  };

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <RefreshCw className="text-muted-foreground h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="bg-background flex items-center justify-between border-b px-6 py-4">
        <div className="flex items-center gap-3">
          <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg">
            <Settings className="text-primary h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-semibold">Instance Settings</h1>
            <p className="text-muted-foreground text-sm">
              Manage instance-level configuration for all tenants
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowOverridableDialog(true)}
          >
            <Shield className="h-4 w-4 mr-2" />
            Manage Overridable
          </Button>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 overflow-auto p-6">
        <div className="flex gap-6">
          {/* Categories sidebar */}
          <div className="w-64 shrink-0">
            <Card className="h-fit">
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">Categories</CardTitle>
              </CardHeader>
              <CardContent className="p-2">
                <nav className="space-y-1">
                  {defaultCategories.map((category) => {
                    const IconComponent = category.icon;
                    return (
                      <button
                        key={category.id}
                        onClick={() => setActiveCategory(category.id)}
                        className={`flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm ${
                          activeCategory === category.id
                            ? "bg-primary/10 text-primary"
                            : "hover:bg-muted/50"
                        }`}
                      >
                        <IconComponent className="h-4 w-4" />
                        <span>{category.name}</span>
                      </button>
                    );
                  })}
                </nav>
              </CardContent>
            </Card>
          </div>

          {/* Settings panel */}
          <div className="flex-1">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>{currentCategory?.name}</CardTitle>
                    <CardDescription>
                      {currentCategory?.description}
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {categorySettings.map((settingDef) => {
                    const resolved = getResolvedSetting(settingDef.path);
                    const overridable = isOverridable(settingDef.path);
                    const isEditing = editingPath === settingDef.path;

                    return (
                      <div
                        key={settingDef.path}
                        className="flex items-center justify-between rounded-lg border p-4"
                      >
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <Label className="text-sm font-medium">
                              {settingDef.name}
                            </Label>
                            {settingDef.description && (
                              <p className="text-muted-foreground text-xs">
                                {settingDef.description}
                              </p>
                            )}
                          </div>
                          <div className="flex items-center gap-2 mt-1">
                            {/* Source badge */}
                            {resolved.source === "config" ? (
                              <Badge variant="secondary" className="text-xs">
                                <Lock className="h-3 w-3 mr-1" />
                                Config
                              </Badge>
                            ) : resolved.source === "instance" ? (
                              <Badge variant="default" className="text-xs">
                                Instance
                              </Badge>
                            ) : (
                              <Badge variant="outline" className="text-xs">
                                Default
                              </Badge>
                            )}
                            {/* Read-only badge */}
                            {resolved.is_read_only && (
                              <Badge variant="outline" className="text-xs">
                                Read-only
                              </Badge>
                            )}
                            {/* Overridable badge */}
                            {overridable && !resolved.is_read_only && (
                              <Badge variant="outline" className="text-xs">
                                <CheckCircle className="h-3 w-3 mr-1" />
                                Overridable
                              </Badge>
                            )}
                          </div>
                        </div>

                        {/* Value display */}
                        <div className="flex items-center gap-2">
                          <div className="flex-1">
                            {isEditing ? (
                              resolved.is_secret ? (
                                <Input
                                  type="password"
                                  value={editValue as string}
                                  onChange={(e) => setEditValue(e.target.value)}
                                  className="flex-1"
                                />
                              ) : (
                                <Input
                                  value={editValue as string}
                                  onChange={(e) => setEditValue(e.target.value)}
                                  className="flex-1"
                                />
                              )
                            ) : (
                              <div className="flex items-center gap-2">
                                <span className="text-sm">
                                  {renderValue(
                                    resolved.value,
                                    resolved.data_type,
                                  )}
                                </span>
                                {resolved.is_secret && (
                                  <Badge variant="outline" className="text-xs">
                                    <EyeOff className="h-3 w-3" />
                                  </Badge>
                                )}
                              </div>
                            )}
                          </div>

                          {/* Action buttons */}
                          <div className="flex items-center gap-2">
                            {isEditing ? (
                              <>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={handleCancel}
                                >
                                  <X className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={handleSave}
                                  disabled={updateMutation.isPending}
                                >
                                  <Save className="h-4 w-4" />
                                </Button>
                              </>
                            ) : (
                              !resolved.is_read_only &&
                              overridable && (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleEdit(settingDef.path)}
                                >
                                  <Pencil className="h-4 w-4" />
                                </Button>
                              )
                            )}
                            {resolved.source === "instance" && !isEditing && (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleReset(settingDef.path)}
                              >
                                <RefreshCw className="h-4 w-4" />
                              </Button>
                            )}
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>

                {/* Edit dialog */}
                {editingPath && (
                  <Dialog
                    open={!!editingPath}
                    onOpenChange={() => setEditingPath(null)}
                  >
                    <DialogContent>
                      <DialogHeader>Edit Setting</DialogHeader>
                      <DialogDescription>
                        Update the value for {editingPath}
                      </DialogDescription>
                      <div className="py-4">
                        <Label htmlFor="edit-value">Value</Label>
                        {getResolvedSetting(editingPath).is_secret ? (
                          <Input
                            id="edit-value"
                            type="password"
                            value={editValue as string}
                            onChange={(e) => setEditValue(e.target.value)}
                          />
                        ) : (
                          <Input
                            id="edit-value"
                            value={editValue as string}
                            onChange={(e) => setEditValue(e.target.value)}
                          />
                        )}
                      </div>
                      <DialogFooter>
                        <Button variant="outline" onClick={handleCancel}>
                          Cancel
                        </Button>
                        <Button
                          onClick={handleSave}
                          disabled={updateMutation.isPending}
                        >
                          {updateMutation.isPending ? "Saving..." : "Save"}
                        </Button>
                      </DialogFooter>
                    </DialogContent>
                  </Dialog>
                )}

                {/* Secret dialog */}
                <Dialog
                  open={showSecretDialog}
                  onOpenChange={() => setShowSecretDialog(false)}
                >
                  <DialogContent>
                    <DialogHeader>Add Secret Setting</DialogHeader>
                    <DialogDescription>
                      Add a new secret setting (e.g., API key, password)
                    </DialogDescription>
                    <div className="grid gap-4 py-4">
                      <div className="grid gap-2">
                        <Label htmlFor="secret-path">Setting Path</Label>
                        <Input
                          id="secret-path"
                          value={secretPath || ""}
                          onChange={(e) => setSecretPath(e.target.value)}
                          placeholder="e.g., ai.providers.openai.api_key"
                        />
                      </div>
                      <div className="grid gap-2">
                        <Label htmlFor="secret-value">Secret Value</Label>
                        <Input
                          id="secret-value"
                          type="password"
                          placeholder="Enter secret value"
                        />
                      </div>
                    </div>
                    <DialogFooter>
                      <Button
                        variant="outline"
                        onClick={() => setShowSecretDialog(false)}
                      >
                        Cancel
                      </Button>
                      <Button
                        onClick={() => {
                          if (secretPath) {
                            handleSecretSave(
                              secretPath,
                              (
                                document.getElementById(
                                  "secret-value",
                                ) as HTMLInputElement
                              ).value,
                            );
                            setShowSecretDialog(false);
                          }
                        }}
                      >
                        Add Secret
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>

                {/* Reset dialog */}
                <AlertDialog
                  open={showResetDialog}
                  onOpenChange={() => setShowResetDialog(false)}
                >
                  <AlertDialogContent>
                    <AlertDialogHeader>Reset to Default</AlertDialogHeader>
                    <AlertDialogDescription>
                      Are you sure you want to reset &quot;{resetPath}&quot; to
                      its default value? This action cannot be undone.
                    </AlertDialogDescription>
                    <AlertDialogFooter>
                      <AlertDialogCancel
                        onClick={() => setShowResetDialog(false)}
                      >
                        Cancel
                      </AlertDialogCancel>
                      <AlertDialogAction
                        onClick={confirmReset}
                        className="bg-destructive"
                      >
                        Reset
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>

                {/* Overridable settings dialog */}
                <Dialog
                  open={showOverridableDialog}
                  onOpenChange={() => setShowOverridableDialog(false)}
                >
                  <DialogContent className="max-w-2xl">
                    <DialogHeader>Manage Overridable Settings</DialogHeader>
                    <DialogDescription>
                      Select which settings tenants can override at the tenant
                      level.
                    </DialogDescription>
                    <div className="py-4">
                      <Label>Allowed Setting Paths</Label>
                      <p className="text-muted-foreground text-sm">
                        Enter setting paths that tenants will be allowed to
                        override (e.g., &quot;ai.providers.openai.api_key&quot;)
                      </p>
                      <div className="space-y-2 mt-4">
                        {["ai", "email", "storage", "auth"].map((category) => (
                          <div
                            key={category}
                            className="flex items-center gap-2"
                          >
                            <input
                              type="checkbox"
                              checked={overridableData?.overridable_settings?.some(
                                (s: string) =>
                                  s.startsWith(`${category}.`) ||
                                  s === category,
                              )}
                              onChange={(e) => {
                                const current =
                                  overridableData?.overridable_settings || [];
                                let newSettings: string[];
                                if (e.target.checked) {
                                  newSettings = [...current, category];
                                } else {
                                  newSettings = current.filter(
                                    (s: string) =>
                                      !s.startsWith(`${category}.`) &&
                                      s !== category,
                                  );
                                }
                                updateOverridableMutation.mutate({
                                  overridable_settings: newSettings,
                                });
                              }}
                              className="h-4 w-4"
                            />
                            <span className="font-medium capitalize">
                              {category}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                    <DialogFooter>
                      <Button
                        variant="outline"
                        onClick={() => setShowOverridableDialog(false)}
                      >
                        Cancel
                      </Button>
                      <Button
                        onClick={() => setShowOverridableDialog(false)}
                        disabled={updateOverridableMutation.isPending}
                      >
                        {updateOverridableMutation.isPending
                          ? "Saving..."
                          : "Save"}
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
