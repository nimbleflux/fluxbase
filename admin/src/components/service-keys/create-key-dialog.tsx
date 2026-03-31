import { useState } from "react";
import { useForm } from "react-hook-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { platformServiceKeysApi, type PlatformServiceKey } from "@/lib/api";
import { useTenants } from "@nimbleflux/fluxbase-sdk-react";

const KEY_TYPES = [
  { value: "anon", label: "Anonymous", description: "Public anonymous access" },
  {
    value: "publishable",
    label: "Publishable",
    description: "Client-side SDK access",
  },
  {
    value: "tenant_service",
    label: "Tenant Service",
    description: "Server-side tenant access",
  },
  {
    value: "global_service",
    label: "Global Service",
    description: "Admin-level system access",
  },
] as const;

type KeyType = (typeof KEY_TYPES)[number]["value"];

interface CreateKeyFormData {
  name: string;
  description: string;
  key_type: KeyType;
  tenant_id: string;
  scopes: string[];
}

interface CreateKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: (key: PlatformServiceKey & { key?: string }) => void;
}

export function CreateKeyDialog({
  open,
  onOpenChange,
  onSuccess,
}: CreateKeyDialogProps) {
  const queryClient = useQueryClient();
  const { tenants, isLoading: tenantsLoading } = useTenants();
  const [selectedScopes, setSelectedScopes] = useState<string[]>(["*"]);
  const [selectedKeyType, setSelectedKeyType] =
    useState<KeyType>("publishable");

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<CreateKeyFormData>({
    defaultValues: {
      name: "",
      description: "",
      key_type: "publishable",
      tenant_id: "",
      scopes: ["*"],
    },
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateKeyFormData) => {
      return platformServiceKeysApi.create({
        name: data.name,
        description: data.description || undefined,
        key_type: data.key_type,
        tenant_id:
          data.key_type === "tenant_service" ? data.tenant_id : undefined,
        scopes: selectedScopes,
      });
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["platform-service-keys"] });
      toast.success("Service key created successfully");
      reset();
      setSelectedScopes(["*"]);
      setSelectedKeyType("publishable");
      onSuccess?.(data);
      onOpenChange(false);
    },
    onError: (error: Error) => {
      toast.error(`Failed to create service key: ${error.message}`);
    },
  });

  const onSubmit = (data: CreateKeyFormData) => {
    createMutation.mutate(data);
  };

  const handleKeyTypeChange = (value: KeyType) => {
    setSelectedKeyType(value);
    setValue("key_type", value);
  };

  const handleScopeToggle = (scope: string) => {
    setSelectedScopes((prev) => {
      if (scope === "*") {
        return ["*"];
      }
      const filtered = prev.filter((s) => s !== "*");
      if (filtered.includes(scope)) {
        return filtered.filter((s) => s !== scope);
      }
      return [...filtered, scope];
    });
  };

  const SCOPE_OPTIONS = [
    { id: "*", label: "All Scopes" },
    { id: "tables:read", label: "Read Tables" },
    { id: "tables:write", label: "Write Tables" },
    { id: "storage:read", label: "Read Storage" },
    { id: "storage:write", label: "Write Storage" },
    { id: "functions:execute", label: "Execute Functions" },
    { id: "auth:read", label: "Read Auth" },
    { id: "auth:write", label: "Write Auth" },
  ];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Create Service Key</DialogTitle>
          <DialogDescription>
            Create a new service key for API access. The key will be shown only
            once.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">
                Name <span className="text-destructive">*</span>
              </Label>
              <Input
                id="name"
                placeholder="My API Key"
                {...register("name", { required: "Name is required" })}
              />
              {errors.name && (
                <p className="text-destructive text-xs">
                  {errors.name.message}
                </p>
              )}
            </div>

            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                placeholder="Used for..."
                {...register("description")}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="key_type">
                Key Type <span className="text-destructive">*</span>
              </Label>
              <Select
                value={selectedKeyType}
                onValueChange={handleKeyTypeChange}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select key type" />
                </SelectTrigger>
                <SelectContent>
                  {KEY_TYPES.map((type) => (
                    <SelectItem key={type.value} value={type.value}>
                      <div className="flex flex-col">
                        <span>{type.label}</span>
                        <span className="text-muted-foreground text-xs">
                          {type.description}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {selectedKeyType === "tenant_service" && (
              <div className="grid gap-2">
                <Label htmlFor="tenant_id">
                  Tenant <span className="text-destructive">*</span>
                </Label>
                <Select
                  disabled={tenantsLoading}
                  onValueChange={(value) => setValue("tenant_id", value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select tenant" />
                  </SelectTrigger>
                  <SelectContent>
                    {tenants?.map((tenant) => (
                      <SelectItem key={tenant.id} value={tenant.id}>
                        {tenant.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {tenantsLoading && (
                  <p className="text-muted-foreground text-xs">
                    Loading tenants...
                  </p>
                )}
              </div>
            )}

            <div className="grid gap-2">
              <Label>Scopes</Label>
              <div className="flex flex-wrap gap-2 rounded-md border p-3">
                {SCOPE_OPTIONS.map((scope) => (
                  <label
                    key={scope.id}
                    className="flex items-center gap-2 text-sm cursor-pointer"
                  >
                    <Checkbox
                      checked={selectedScopes.includes(scope.id)}
                      onCheckedChange={() => handleScopeToggle(scope.id)}
                    />
                    {scope.label}
                  </label>
                ))}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              Create Key
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
