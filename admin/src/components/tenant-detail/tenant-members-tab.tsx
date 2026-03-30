import { useState } from "react";
import { format } from "date-fns";
import {
  Users,
  Plus,
  Trash2,
  Shield,
  ShieldCheck,
  Mail,
  RefreshCw,
} from "lucide-react";
import type {
  TenantMembership,
  EnrichedUser,
  AddMemberRequest,
  UpdateMemberRequest,
} from "@/lib/api";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
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
import { ScrollArea } from "@/components/ui/scroll-area";
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
import type { MemberRole } from "./types";

interface TenantMembersTabProps {
  tenant: {
    id: string;
    name: string;
    slug: string;
    is_default: boolean;
    created_at: string;
  };
  members: TenantMembership[] | undefined;
  membersLoading: boolean;
  users: EnrichedUser[];
  onAddMember: (data: AddMemberRequest) => void;
  onUpdateMemberRole: (userId: string, data: UpdateMemberRequest) => void;
  onRemoveMember: (userId: string) => void;
  isAddingMember: boolean;
}

export function TenantMembersTab({
  tenant,
  members,
  membersLoading,
  users,
  onAddMember,
  onUpdateMemberRole,
  onRemoveMember,
  isAddingMember,
}: TenantMembersTabProps) {
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [newMemberUserId, setNewMemberUserId] = useState("");
  const [newMemberRole, setNewMemberRole] =
    useState<MemberRole>("tenant_member");
  const [searchEmail, setSearchEmail] = useState("");

  const filteredUsers = users.filter(
    (user) =>
      !members?.some((m) => m.user_id === user.id) &&
      (searchEmail
        ? user.email.toLowerCase().includes(searchEmail.toLowerCase())
        : true),
  );

  const handleAddMember = () => {
    if (!newMemberUserId) return;
    onAddMember({
      user_id: newMemberUserId,
      role: newMemberRole,
    });
    setIsAddDialogOpen(false);
    setNewMemberUserId("");
    setNewMemberRole("tenant_member");
    setSearchEmail("");
  };

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5" />
                Members
              </CardTitle>
              <CardDescription>
                Users with access to this tenant
              </CardDescription>
            </div>
            <Button onClick={() => setIsAddDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Add Member
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {membersLoading ? (
            <div className="flex items-center justify-center py-8">
              <RefreshCw className="text-muted-foreground h-6 w-6 animate-spin" />
            </div>
          ) : members && members.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Email</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Added</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {members.map((member) => (
                  <TableRow key={member.id}>
                    <TableCell className="flex items-center gap-2">
                      <Mail className="text-muted-foreground h-4 w-4" />
                      {member.email || member.user_id}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {member.role === "tenant_admin" ? (
                          <ShieldCheck className="h-4 w-4 text-green-500" />
                        ) : (
                          <Shield className="text-muted-foreground h-4 w-4" />
                        )}
                        <Select
                          value={member.role}
                          onValueChange={(value) =>
                            onUpdateMemberRole(member.user_id, {
                              role: value as MemberRole,
                            })
                          }
                        >
                          <SelectTrigger className="h-8 w-[140px]">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="tenant_admin">Admin</SelectItem>
                            <SelectItem value="tenant_member">
                              Member
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {format(new Date(member.created_at), "MMM d, yyyy")}
                    </TableCell>
                    <TableCell className="text-right">
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-destructive hover:text-destructive hover:bg-destructive/10"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>Remove Member</AlertDialogTitle>
                            <AlertDialogDescription>
                              Are you sure you want to remove {member.email}{" "}
                              from this tenant?
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction
                              onClick={() => onRemoveMember(member.user_id)}
                              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            >
                              Remove
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Users className="text-muted-foreground mb-4 h-12 w-12" />
              <p className="mb-2 text-lg font-medium">No members yet</p>
              <p className="text-muted-foreground mb-4 text-sm">
                Add members to give them access to this tenant
              </p>
              <Button onClick={() => setIsAddDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Add Member
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Tenant Details</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-2 gap-4">
            <div>
              <dt className="text-muted-foreground text-sm">ID</dt>
              <dd className="font-mono text-sm">{tenant.id}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-sm">Slug</dt>
              <dd className="font-mono text-sm">{tenant.slug}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-sm">Created</dt>
              <dd className="text-sm">
                {format(new Date(tenant.created_at), "PPPpp")}
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-sm">Default</dt>
              <dd className="text-sm">{tenant.is_default ? "Yes" : "No"}</dd>
            </div>
          </dl>
        </CardContent>
      </Card>

      <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Member</DialogTitle>
            <DialogDescription>Add a user to this tenant</DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="search">Search Users</Label>
              <Input
                id="search"
                placeholder="Search by email..."
                value={searchEmail}
                onChange={(e) => setSearchEmail(e.target.value)}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="user">Select User</Label>
              <Select
                value={newMemberUserId}
                onValueChange={setNewMemberUserId}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select a user" />
                </SelectTrigger>
                <SelectContent>
                  <ScrollArea className="h-[200px]">
                    {filteredUsers.length > 0 ? (
                      filteredUsers.map((user) => (
                        <SelectItem key={user.id} value={user.id}>
                          {user.email}
                        </SelectItem>
                      ))
                    ) : (
                      <div className="text-muted-foreground p-2 text-center text-sm">
                        No users available
                      </div>
                    )}
                  </ScrollArea>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="role">Role</Label>
              <Select
                value={newMemberRole}
                onValueChange={(v) => setNewMemberRole(v as MemberRole)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="tenant_admin">Admin</SelectItem>
                  <SelectItem value="tenant_member">Member</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-muted-foreground text-xs">
                Admins can manage members. Members have read access.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleAddMember}
              disabled={isAddingMember || !newMemberUserId}
            >
              {isAddingMember ? "Adding..." : "Add Member"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
