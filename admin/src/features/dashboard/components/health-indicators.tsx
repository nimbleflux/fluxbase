import { useQuery } from "@tanstack/react-query";
import { useFluxbaseClient } from "@nimbleflux/fluxbase-sdk-react";
import { CheckCircle2, XCircle, AlertTriangle, Loader2 } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

interface ServiceHealth {
  name: string;
  status: "healthy" | "degraded" | "down";
  details?: string;
}

export function HealthIndicators() {
  const client = useFluxbaseClient();

  const { data: health, isLoading } = useQuery({
    queryKey: ["health"],
    queryFn: async () => {
      const { data, error } = await client.admin.getHealth();
      if (error) throw error;
      return data;
    },
    refetchInterval: 15000, // Refresh every 15 seconds
  });

  const apiService: ServiceHealth = {
    name: "API",
    status: health?.status === "ok" ? "healthy" : "down",
    details:
      health?.status === "ok"
        ? "All endpoints operational"
        : "Service unavailable",
  };

  const additionalServices: ServiceHealth[] = health?.services
    ? [
        {
          name: "Database",
          status: health.services.database === true ? "healthy" : "down",
          details:
            health.services.database === true
              ? "Connected"
              : "Connection failed",
        },
        {
          name: "Realtime",
          status: health.services.realtime === true ? "healthy" : "down",
          details:
            health.services.realtime === true
              ? "WebSocket server running"
              : "Disabled",
        },
      ]
    : [];

  const services: ServiceHealth[] = [apiService, ...additionalServices];

  // Determine overall system health
  const getOverallStatus = () => {
    if (isLoading) return { icon: "loading", text: "Checking..." };
    if (!health) return { icon: "warning", text: "Unable to check status" };

    // Use backend's overall status
    const status = health.status;
    if (status === "ok")
      return { icon: "success", text: "All systems operational" };
    if (status === "degraded")
      return { icon: "warning", text: "Some systems degraded" };
    return { icon: "error", text: "Systems unavailable" };
  };

  const overallStatus = getOverallStatus();

  return (
    <Card>
      <CardContent className="px-4 py-1">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {overallStatus.icon === "loading" ? (
              <Loader2 className="text-muted-foreground h-4 w-4 animate-spin" />
            ) : overallStatus.icon === "success" ? (
              <CheckCircle2 className="h-4 w-4 text-green-500" />
            ) : overallStatus.icon === "warning" ? (
              <AlertTriangle className="h-4 w-4 text-yellow-500" />
            ) : (
              <XCircle className="h-4 w-4 text-red-500" />
            )}
            <span className="text-sm font-medium">{overallStatus.text}</span>
          </div>
          <div className="flex gap-3">
            {services.map((service) => (
              <div
                key={service.name}
                className="flex items-center gap-1.5"
                title={service.details}
              >
                {isLoading ? (
                  <Skeleton className="h-4 w-4" />
                ) : (
                  <>
                    {service.status === "healthy" && (
                      <CheckCircle2 className="h-3.5 w-3.5 text-green-500" />
                    )}
                    {service.status === "degraded" && (
                      <AlertTriangle className="h-3.5 w-3.5 text-yellow-500" />
                    )}
                    {service.status === "down" && (
                      <XCircle className="h-3.5 w-3.5 text-red-500" />
                    )}
                    <span className="text-muted-foreground text-xs">
                      {service.name}
                    </span>
                  </>
                )}
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
