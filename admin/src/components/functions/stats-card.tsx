import { Card, CardContent } from "@/components/ui/card";

interface ExecutionStats {
  success: number;
  failed: number;
  total: number;
  avgDuration: number;
}

interface StatsCardProps {
  stats: ExecutionStats;
}

export function StatsCard({ stats }: StatsCardProps) {
  const total = stats.success + stats.failed;
  const successRate =
    total > 0 ? ((stats.success / total) * 100).toFixed(0) : "0";

  return (
    <Card className="!gap-0 !py-0 mb-6">
      <CardContent className="px-4 py-2">
        <div className="flex items-center gap-4">
          <span className="text-muted-foreground text-xs">(Past 24 hours)</span>
          <div className="flex items-center gap-1">
            <span className="text-muted-foreground text-xs">Success:</span>
            <span className="text-sm font-semibold">{stats.success}</span>
          </div>
          <div className="flex items-center gap-1">
            <span className="text-muted-foreground text-xs">Failed:</span>
            <span className="text-sm font-semibold">{stats.failed}</span>
          </div>
          <div className="flex items-center gap-1">
            <span className="text-muted-foreground text-xs">Total:</span>
            <span className="text-sm font-semibold">{stats.total}</span>
          </div>
          <div className="flex items-center gap-1">
            <span className="text-muted-foreground text-xs">Success Rate:</span>
            <span className="text-sm font-semibold">{successRate}%</span>
          </div>
          <div className="flex items-center gap-1">
            <span className="text-muted-foreground text-xs">
              Avg. Duration:
            </span>
            <span className="text-sm font-semibold">{stats.avgDuration}ms</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
