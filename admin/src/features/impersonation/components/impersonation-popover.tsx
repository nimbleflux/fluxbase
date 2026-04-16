import { useState } from "react";
import { UserCog, User, UserX, Shield, X } from "lucide-react";
import type { ImpersonationType } from "@/stores/impersonation-store";
import { cn } from "@/lib/utils";
import { useAuth } from "@/hooks/use-auth";
import { useTenantStore } from "@/stores/tenant-store";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useImpersonation } from "../hooks/use-impersonation";
import { UserSearch } from "./user-search";

export interface ImpersonationPopoverProps {
  contextLabel?: string;
  requireReason?: boolean;
  defaultReason?: string;
  onImpersonationStart?: () => void;
  onImpersonationStop?: () => void;
  className?: string;
  size?: "sm" | "default";
}

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

export function ImpersonationPopover({
  contextLabel = "Impersonating",
  requireReason = false,
  defaultReason = "Admin impersonation",
  onImpersonationStart,
  onImpersonationStop,
  className,
  size = "sm",
}: ImpersonationPopoverProps) {
  const { user } = useAuth();
  const [open, setOpen] = useState(false);
  const [impersonationType, setImpersonationType] =
    useState<ImpersonationType>("user");
  const [selectedUserId, setSelectedUserId] = useState("");
  const [selectedUserEmail, setSelectedUserEmail] = useState("");
  const [reason, setReason] = useState("");

  const isInstanceAdmin = checkIsInstanceAdmin(user);
  const currentTenant = useTenantStore((state) => state.currentTenant);

  const {
    isImpersonating,
    impersonationType: activeType,
    impersonatedUser,
    isLoading,
    startUserImpersonation,
    startAnonImpersonation,
    startServiceImpersonation,
    stopImpersonation,
    getDisplayLabel,
  } = useImpersonation({
    defaultReason,
    onStart: onImpersonationStart,
    onStop: onImpersonationStop,
  });

  if (!checkIsAdmin(user) || !currentTenant) {
    return null;
  }

  const handleStartImpersonation = async () => {
    const reasonToUse = requireReason ? reason : defaultReason;

    if (requireReason && !reason.trim()) {
      return;
    }

    if (impersonationType === "user" && !selectedUserId) {
      return;
    }

    switch (impersonationType) {
      case "user":
        await startUserImpersonation(
          selectedUserId,
          selectedUserEmail,
          reasonToUse,
        );
        break;
      case "anon":
        await startAnonImpersonation(reasonToUse);
        break;
      case "service":
        await startServiceImpersonation(reasonToUse);
        break;
    }

    setOpen(false);
    setSelectedUserId("");
    setSelectedUserEmail("");
    setReason("");
  };

  const handleUserSelect = (userId: string, userEmail: string) => {
    setSelectedUserId(userId);
    setSelectedUserEmail(userEmail);
  };

  const handleStopImpersonation = async () => {
    await stopImpersonation();
  };

  const getTypeIcon = (type: ImpersonationType) => {
    switch (type) {
      case "user":
        return <User className="h-3.5 w-3.5" />;
      case "anon":
        return <UserX className="h-3.5 w-3.5" />;
      case "service":
        return <Shield className="h-3.5 w-3.5" />;
    }
  };

  const getBadgeColors = () => {
    switch (activeType) {
      case "anon":
        return "border-orange-500 text-orange-600 dark:text-orange-400 bg-orange-50 dark:bg-orange-950";
      case "service":
        return "border-purple-500 text-purple-600 dark:text-purple-400 bg-purple-50 dark:bg-purple-950";
      case "user":
      default:
        return "border-blue-500 text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-950";
    }
  };

  const resolveDisplayLabel = () => {
    if (activeType === "service" && impersonatedUser?.role === "tenant_service") {
      return "Tenant Service";
    }
    return getDisplayLabel();
  };

  if (isImpersonating) {
    return (
      <div className={cn("flex items-center gap-2", className)}>
        <Badge
          variant="outline"
          className={cn("gap-1.5 px-3 py-1.5", getBadgeColors())}
        >
          {getTypeIcon(activeType!)}
          <span className="max-w-[200px] truncate">
            {contextLabel}: {resolveDisplayLabel()}
          </span>
        </Badge>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleStopImpersonation}
          className="h-7 w-7 p-0"
          title="Stop impersonation"
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
    );
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size={size}
          disabled={isLoading}
          className={cn("gap-2", className)}
        >
          <UserCog className="h-4 w-4" />
          Impersonate
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="end">
        <div className="grid gap-4">
          <div className="space-y-2">
            <h4 className="leading-none font-medium">Impersonate User</h4>
            <p className="text-muted-foreground text-sm">
              Execute operations as a different user or role
            </p>
          </div>

          <div className="grid gap-3">
            <div className="grid gap-1.5">
              <Label htmlFor="impersonation-type">Type</Label>
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
              <div className="grid gap-1.5">
                <Label htmlFor="user-select">User</Label>
                <UserSearch
                  value={selectedUserId}
                  onSelect={handleUserSelect}
                  disabled={isLoading}
                />
              </div>
            )}

            {requireReason && (
              <div className="grid gap-1.5">
                <Label htmlFor="reason">Reason</Label>
                <Textarea
                  id="reason"
                  placeholder="e.g., Testing RLS policies..."
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  disabled={isLoading}
                  rows={2}
                  className="resize-none"
                />
                <p className="text-muted-foreground text-xs">
                  Logged for audit trail
                </p>
              </div>
            )}

            <Button
              onClick={handleStartImpersonation}
              disabled={
                isLoading ||
                (impersonationType === "user" && !selectedUserId) ||
                (requireReason && !reason.trim())
              }
              className="w-full"
            >
              {isLoading ? "Starting..." : "Start Impersonation"}
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}
