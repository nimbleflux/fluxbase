import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { UserCog, User, UserX, Shield, X } from "lucide-react";
import { toast } from "sonner";
import {
  useImpersonationStore,
  type ImpersonationType,
} from "@/stores/impersonation-store";
import { getAccessToken } from "@/lib/auth";
import { setAuthToken as setSDKAuthToken } from "@/lib/fluxbase-client";
import { impersonationApi } from "@/lib/impersonation-api";
import { useAuth } from "@/hooks/use-auth";
import { useTenantStore } from "@/stores/tenant-store";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { UserSearch } from "./user-search";

function checkIsAdmin(user: unknown): boolean {
  if (!user || typeof user !== "object" || !("role" in user)) return false;
  const u = user as { role: unknown };
  if (Array.isArray(u.role)) {
    return (
      u.role.includes("instance_admin") || u.role.includes("tenant_admin")
    );
  }
  return u.role === "instance_admin" || u.role === "tenant_admin";
}

function checkIsInstanceAdmin(user: unknown): boolean {
  if (!user || typeof user !== "object" || !("role" in user)) return false;
  const u = user as { role: unknown };
  if (Array.isArray(u.role)) {
    return u.role.includes("instance_admin");
  }
  return u.role === "instance_admin";
}

export function ImpersonationSelector() {
  const { user } = useAuth();
  const {
    isImpersonating,
    startImpersonation,
    stopImpersonation,
    impersonatedUser,
    impersonationType: activeImpersonationType,
  } = useImpersonationStore();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [stopping, setStopping] = useState(false);
  const [impersonationType, setImpersonationType] =
    useState<ImpersonationType>("user");
  const [selectedUserId, setSelectedUserId] = useState<string>("");
  const [reason, setReason] = useState("");

  const isInstanceAdmin = checkIsInstanceAdmin(user);
  const currentTenant = useTenantStore((state) => state.currentTenant);
  const tenants = useTenantStore((state) => state.tenants);
  const setCurrentTenant = useTenantStore((state) => state.setCurrentTenant);

  // For tenant admins (non-instance-admins), currentTenant should always be
  // set by the TenantSelector's auto-selection logic. But as a fallback in
  // case of a race condition (e.g. Zustand persist hydration), derive it from
  // the tenants list.
  const effectiveTenant =
    currentTenant ||
    (!isInstanceAdmin && tenants.length > 0 ? tenants[0] : null);

  if (!checkIsAdmin(user) || !effectiveTenant) {
    return null;
  }

  // Ensure the store's currentTenant is set before any API calls so the
  // interceptor includes the X-FB-Tenant header.
  const ensureTenantInStore = () => {
    if (!currentTenant && effectiveTenant) {
      setCurrentTenant(effectiveTenant);
    }
  };

  const handleStartImpersonation = async () => {
    if (!reason.trim()) {
      toast.error("Please provide a reason for impersonation");
      return;
    }

    if (impersonationType === "user" && !selectedUserId) {
      toast.error("Please select a user to impersonate");
      return;
    }

    try {
      setLoading(true);
      ensureTenantInStore();
      let response;

      switch (impersonationType) {
        case "user":
          response = await impersonationApi.startUserImpersonation(
            selectedUserId,
            reason,
          );
          break;
        case "anon":
          response = await impersonationApi.startAnonImpersonation(reason);
          break;
        case "service":
          response = await impersonationApi.startServiceImpersonation(reason);
          break;
      }

      startImpersonation(
        response.access_token,
        response.refresh_token,
        response.target_user,
        response.session,
        impersonationType,
      );

      setSDKAuthToken(response.access_token);

      toast.success(
        `Started impersonating ${
          impersonationType === "user"
            ? response.target_user.email
            : impersonationType === "anon"
              ? "anonymous user"
              : response.target_user.role === "tenant_service"
                ? "tenant service"
                : "service role"
        }`,
      );

      setOpen(false);
      setSelectedUserId("");
      setReason("");

      queryClient.invalidateQueries();
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error && "response" in error
          ? (error as { response?: { data?: { error?: string } } }).response
              ?.data?.error
          : undefined;
      toast.error(errorMessage || "Failed to start impersonation");
    } finally {
      setLoading(false);
    }
  };

  const handleUserSelect = (userId: string, _userEmail: string) => {
    setSelectedUserId(userId);
  };

  const handleStopImpersonation = async () => {
    try {
      setStopping(true);
      await impersonationApi.stopImpersonation();
      stopImpersonation();

      const adminToken = getAccessToken();
      if (adminToken) {
        setSDKAuthToken(adminToken);
      }

      toast.success("Impersonation stopped");

      queryClient.invalidateQueries();
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error && "response" in error
          ? (error as { response?: { data?: { error?: string } } }).response
              ?.data?.error
          : undefined;
      toast.error(errorMessage || "Failed to stop impersonation");
    } finally {
      setStopping(false);
    }
  };

  const getDisplayLabel = () => {
    switch (activeImpersonationType) {
      case "user":
        return impersonatedUser?.email || "User";
      case "anon":
        return "Anonymous";
      case "service":
        return impersonatedUser?.role === "tenant_service"
          ? "Tenant Service"
          : "Service Role";
      default:
        return "User";
    }
  };

  const getIcon = () => {
    switch (impersonationType) {
      case "user":
        return <User className="h-4 w-4" />;
      case "anon":
        return <UserX className="h-4 w-4" />;
      case "service":
        return <Shield className="h-4 w-4" />;
    }
  };

  if (isImpersonating) {
    return (
      <Button
        variant="outline"
        size="sm"
        onClick={handleStopImpersonation}
        disabled={stopping}
        className="gap-2 border-amber-300 bg-amber-50 text-amber-800 hover:bg-amber-100 dark:border-amber-700 dark:bg-amber-950 dark:text-amber-200 dark:hover:bg-amber-900"
      >
        <X className="h-4 w-4" />
        {stopping ? "Stopping..." : `Cancel: ${getDisplayLabel()}`}
      </Button>
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <UserCog className="h-4 w-4" />
          Impersonate User
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Start User Impersonation</DialogTitle>
          <DialogDescription>
            View data as it appears to a specific user, anonymous visitor, or
            with service-level permissions. All actions will be logged for audit
            purposes.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="impersonation-type">Impersonation Type</Label>
            <Select
              value={impersonationType}
              onValueChange={(value) =>
                setImpersonationType(value as ImpersonationType)
              }
            >
              <SelectTrigger id="impersonation-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="user">
                  <div className="flex items-center gap-2">
                    <User className="h-4 w-4" />
                    Specific User
                  </div>
                </SelectItem>
                <SelectItem value="anon">
                  <div className="flex items-center gap-2">
                    <UserX className="h-4 w-4" />
                    Anonymous (anon key)
                  </div>
                </SelectItem>
                <SelectItem value="service">
                  <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4" />
                    {isInstanceAdmin ? "Service Role" : "Tenant Service"}
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {impersonationType === "user" && (
            <div className="grid gap-2">
              <Label htmlFor="user-select">User</Label>
              <UserSearch
                value={selectedUserId}
                onSelect={handleUserSelect}
                disabled={loading}
              />
            </div>
          )}

          <div className="grid gap-2">
            <Label htmlFor="reason">Reason</Label>
            <Textarea
              id="reason"
              placeholder="e.g., Customer support ticket #1234, debugging user-reported issue"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              disabled={loading}
              rows={3}
            />
            <p className="text-muted-foreground text-xs">
              This reason will be logged in the audit trail
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setOpen(false)}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button onClick={handleStartImpersonation} disabled={loading}>
            {loading ? (
              "Starting..."
            ) : (
              <>
                {getIcon()}
                <span className="ml-2">Start Impersonation</span>
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
