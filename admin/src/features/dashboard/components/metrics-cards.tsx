import { useQuery } from "@tanstack/react-query";
import { useFluxbaseClient } from "@nimbleflux/fluxbase-sdk-react";
import {
  ArrowUpRight,
  ArrowDownRight,
  Minus,
  Users,
  Activity,
  HardDrive,
  Zap,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

interface MetricCardProps {
  title: string;
  value: string | number;
  unit?: string;
  trend?: number;
  trendLabel?: string;
  icon: React.ReactNode;
  isLoading?: boolean;
}

function MetricCard({
  title,
  value,
  unit,
  trend,
  trendLabel,
  icon,
  isLoading,
}: MetricCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-8 w-20" />
        ) : (
          <>
            <div className="text-2xl font-bold">
              {typeof value === "number" ? value.toLocaleString() : value}
              {unit && (
                <span className="text-muted-foreground ml-1 text-sm font-normal">
                  {unit}
                </span>
              )}
            </div>
            {typeof trend === "number" && (
              <div className="flex items-center gap-1 text-xs">
                {trend > 0 ? (
                  <ArrowUpRight className="h-3 w-3 text-green-500" />
                ) : trend < 0 ? (
                  <ArrowDownRight className="h-3 w-3 text-red-500" />
                ) : (
                  <Minus className="text-muted-foreground h-3 w-3" />
                )}
                <span
                  className={cn(
                    "font-medium",
                    trend > 0
                      ? "text-green-500"
                      : trend < 0
                        ? "text-red-500"
                        : "text-muted-foreground",
                  )}
                >
                  {Math.abs(trend)}%
                </span>
                <span className="text-muted-foreground ml-1">{trendLabel}</span>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}

export function MetricsCards() {
  const client = useFluxbaseClient();

  // Fetch active users (signed in today)
  const { data: activeUsers, isLoading: isLoadingUsers } = useQuery({
    queryKey: ["dashboard", "active-users", client.admin],
    queryFn: async () => {
      try {
        const { data, error } = await client.admin.listUsers();
        if (error) return 0;

        const today = new Date();
        return (
          data?.users.filter((u) => {
            const lastLogin = u.last_login_at;
            if (!lastLogin) return false;
            const lastLoginDate = new Date(lastLogin);
            return (
              lastLoginDate.getDate() === today.getDate() &&
              lastLoginDate.getMonth() === today.getMonth() &&
              lastLoginDate.getFullYear() === today.getFullYear()
            );
          }).length ?? 0
        );
      } catch {
        return 0;
      }
    },
    refetchInterval: 60000, // Refresh every minute
  });

  // Fetch database size (from health endpoint if available)
  const { data: health, isLoading: isLoadingHealth } = useQuery({
    queryKey: ["health", client.admin],
    queryFn: async () => {
      const { data, error } = await client.admin.getHealth();
      if (error) throw error;
      return data;
    },
    refetchInterval: 30000,
  });

  // Get DB size from health endpoint, fallback to connection status
  const dbSize =
    health?.services?.database_size ??
    (health?.services?.database ? "Connected" : "Disconnected");

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <MetricCard
        title="Active Users"
        value={activeUsers ?? 0}
        trend={12}
        trendLabel="vs yesterday"
        icon={<Users className="text-muted-foreground h-4 w-4" />}
        isLoading={isLoadingUsers}
      />

      <MetricCard
        title="API Requests"
        value={0}
        unit="/min"
        trend={5}
        trendLabel="vs last hour"
        icon={<Activity className="text-muted-foreground h-4 w-4" />}
        isLoading={isLoadingHealth}
      />

      <MetricCard
        title="Avg Response"
        value={0}
        unit="ms"
        trend={-8}
        trendLabel="vs last hour"
        icon={<Zap className="text-muted-foreground h-4 w-4" />}
        isLoading={isLoadingHealth}
      />

      <MetricCard
        title="Database Size"
        value={dbSize}
        icon={<HardDrive className="text-muted-foreground h-4 w-4" />}
        isLoading={isLoadingHealth}
      />
    </div>
  );
}
