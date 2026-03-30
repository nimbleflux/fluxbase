import { useState } from "react";
import { formatDistanceToNow } from "date-fns";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Users, Loader2, ChevronLeft, ChevronRight } from "lucide-react";
import { toast } from "sonner";
import api from "@/lib/api";
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { Session } from "./types";

export function ActiveSessionsTab() {
  const queryClient = useQueryClient();

  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(25);

  const { data: sessionsData, isLoading } = useQuery({
    queryKey: ["sessions", page, pageSize],
    queryFn: async () => {
      const response = await api.get<{
        sessions: Session[];
        count: number;
        total_count: number;
      }>(
        `/api/v1/admin/auth/sessions?include_expired=true&limit=${pageSize}&offset=${page * pageSize}`,
      );
      return response.data;
    },
  });

  const sessions = sessionsData?.sessions || [];
  const totalCount = sessionsData?.total_count || 0;
  const totalPages = Math.ceil(totalCount / pageSize);

  const revokeSessionMutation = useMutation({
    mutationFn: async (sessionId: string) => {
      await api.delete(`/api/v1/admin/auth/sessions/${sessionId}`);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["sessions", page, pageSize],
      });
      toast.success("Session revoked successfully");
    },
    onError: () => {
      toast.error("Failed to revoke session");
    },
  });

  const revokeAllUserSessionsMutation = useMutation({
    mutationFn: async (userId: string) => {
      await api.delete(`/api/v1/admin/auth/sessions/user/${userId}`);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["sessions", page, pageSize],
      });
      toast.success("All user sessions revoked successfully");
    },
    onError: () => {
      toast.error("Failed to revoke user sessions");
    },
  });

  const isExpired = (expiresAt: string) => new Date(expiresAt) < new Date();

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5" />
                Active Sessions
              </CardTitle>
              <CardDescription>
                Monitor and manage active user sessions across the platform
              </CardDescription>
            </div>
            <div className="flex gap-2">
              <Badge variant="outline" className="text-sm">
                {totalCount} Total
              </Badge>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
            </div>
          ) : sessions && sessions.length > 0 ? (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Session ID</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Expires</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sessions.map((session) => (
                    <TableRow key={session.id}>
                      <TableCell className="font-medium">
                        {session.user_email || "Unknown"}
                      </TableCell>
                      <TableCell className="font-mono text-xs">
                        {session.id.substring(0, 8)}...
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {formatDistanceToNow(new Date(session.created_at), {
                          addSuffix: true,
                        })}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {formatDistanceToNow(new Date(session.expires_at), {
                          addSuffix: true,
                        })}
                      </TableCell>
                      <TableCell>
                        {isExpired(session.expires_at) ? (
                          <Badge variant="secondary">Expired</Badge>
                        ) : (
                          <Badge variant="default">Active</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() =>
                              revokeSessionMutation.mutate(session.id)
                            }
                            disabled={revokeSessionMutation.isPending}
                          >
                            Revoke
                          </Button>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() =>
                              revokeAllUserSessionsMutation.mutate(
                                session.user_id,
                              )
                            }
                            disabled={revokeAllUserSessionsMutation.isPending}
                          >
                            Revoke All
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              <div className="mt-4 flex items-center justify-between border-t pt-4">
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground text-sm">
                    Rows per page
                  </span>
                  <Select
                    value={`${pageSize}`}
                    onValueChange={(value) => {
                      setPageSize(Number(value));
                      setPage(0);
                    }}
                  >
                    <SelectTrigger className="h-8 w-[70px]">
                      <SelectValue placeholder={pageSize} />
                    </SelectTrigger>
                    <SelectContent side="top">
                      {[10, 25, 50, 100].map((size) => (
                        <SelectItem key={size} value={`${size}`}>
                          {size}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground text-sm">
                    Page {page + 1} of {totalPages || 1} ({totalCount} total)
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => Math.max(0, p - 1))}
                    disabled={page === 0}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      setPage((p) => Math.min(totalPages - 1, p + 1))
                    }
                    disabled={page >= totalPages - 1}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Users className="text-muted-foreground mb-4 h-12 w-12" />
              <p className="text-muted-foreground">No active sessions found</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
