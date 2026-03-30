import { Lock, Clock, AlertCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { SecretsStatsCardsProps } from "./types";

export function SecretsStatsCards({
  total,
  expiringSoon,
  expired,
}: SecretsStatsCardsProps) {
  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Secrets</CardTitle>
          <Lock className="text-muted-foreground h-4 w-4" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{total}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Expiring Soon</CardTitle>
          <Clock className="text-muted-foreground h-4 w-4" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-yellow-600">
            {expiringSoon}
          </div>
          <p className="text-muted-foreground text-xs">Within 7 days</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Expired</CardTitle>
          <AlertCircle className="text-muted-foreground h-4 w-4" />
        </CardHeader>
        <CardContent>
          <div className="text-destructive text-2xl font-bold">{expired}</div>
        </CardContent>
      </Card>
    </div>
  );
}
