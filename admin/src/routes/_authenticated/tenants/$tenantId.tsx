import { useState } from 'react'
import { format } from 'date-fns'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Building2,
  Users,
  Plus,
  Trash2,
  Shield,
  ShieldCheck,
  Mail,
  RefreshCw,
  Key,
  Check,
  AlertCircle,
  Copy,
  Loader2,
  Settings,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  tenantsApi,
  type AddMemberRequest,
  type UpdateMemberRequest,
  userManagementApi,
  oauthProviderApi,
  samlProviderApi,
  type OAuthProviderConfig,
  type SAMLProviderConfig,
  type CreateOAuthProviderRequest,
  type CreateSAMLProviderRequest,
  type UpdateOAuthProviderRequest,
  type UpdateSAMLProviderRequest,
} from '@/lib/api'
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
} from '@/components/ui/alert-dialog'
import { TenantSettingsTab } from './-TenantSettingsTab'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

export const Route = createFileRoute('/_authenticated/tenants/$tenantId')({
  component: TenantDetailPage,
})

function TenantDetailPage() {
  const { tenantId } = Route.useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState('members')
  const [addMemberDialogOpen, setAddMemberDialogOpen] = useState(false)
  const [newMemberUserId, setNewMemberUserId] = useState('')
  const [newMemberRole, setNewMemberRole] = useState<'tenant_admin' | 'tenant_member'>('tenant_member')
  const [searchEmail, setSearchEmail] = useState('')

  const { data: tenant, isLoading: tenantLoading } = useQuery({
    queryKey: ['tenant', tenantId],
    queryFn: () => tenantsApi.get(tenantId),
  })

  const { data: members, isLoading: membersLoading } = useQuery({
    queryKey: ['tenant-members', tenantId],
    queryFn: () => tenantsApi.listMembers(tenantId),
  })

  const { data: usersResponse } = useQuery({
    queryKey: ['users', 'dashboard'],
    queryFn: () => userManagementApi.listUsers('dashboard'),
  })

  const users = usersResponse?.users || []

  const addMemberMutation = useMutation({
    mutationFn: (data: AddMemberRequest) => tenantsApi.addMember(tenantId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant-members', tenantId] })
      toast.success('Member added successfully')
      setAddMemberDialogOpen(false)
      setNewMemberUserId('')
      setNewMemberRole('tenant_member')
      setSearchEmail('')
    },
    onError: (error: Error) => {
      toast.error(`Failed to add member: ${error.message}`)
    },
  })

  const updateMemberMutation = useMutation({
    mutationFn: ({ userId, data }: { userId: string; data: UpdateMemberRequest }) =>
      tenantsApi.updateMemberRole(tenantId, userId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant-members', tenantId] })
      toast.success('Member role updated')
    },
    onError: (error: Error) => {
      toast.error(`Failed to update member: ${error.message}`)
    },
  })

  const removeMemberMutation = useMutation({
    mutationFn: (userId: string) => tenantsApi.removeMember(tenantId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant-members', tenantId] })
      toast.success('Member removed')
    },
    onError: (error: Error) => {
      toast.error(`Failed to remove member: ${error.message}`)
    },
  })

  const handleAddMember = () => {
    if (!newMemberUserId) {
      toast.error('Please select a user')
      return
    }
    addMemberMutation.mutate({
      user_id: newMemberUserId,
      role: newMemberRole,
    })
  }

  const filteredUsers = users.filter(
    (user) =>
      !members?.some((m) => m.user_id === user.id) &&
      (searchEmail
        ? user.email.toLowerCase().includes(searchEmail.toLowerCase())
        : true)
  )

  if (tenantLoading) {
    return (
      <div className='flex h-96 items-center justify-center'>
        <RefreshCw className='text-muted-foreground h-8 w-8 animate-spin' />
      </div>
    )
  }

  if (!tenant) {
    return (
      <div className='flex h-96 flex-col items-center justify-center gap-4'>
        <p className='text-muted-foreground'>Tenant not found</p>
        <Button variant='outline' onClick={() => navigate({ to: '/tenants' })}>
          <ArrowLeft className='mr-2 h-4 w-4' />
          Back to Tenants
        </Button>
      </div>
    )
  }

  return (
    <div className='flex h-full flex-col'>
      <div className='bg-background flex items-center justify-between border-b px-6 py-4'>
        <div className='flex items-center gap-3'>
          <Button
            variant='ghost'
            size='sm'
            onClick={() => navigate({ to: '/tenants' })}
          >
            <ArrowLeft className='mr-2 h-4 w-4' />
            Back
          </Button>
          <div className='bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg'>
            <Building2 className='text-primary h-5 w-5' />
          </div>
          <div>
            <h1 className='text-xl font-semibold'>{tenant.name}</h1>
            <p className='text-muted-foreground text-sm'>
              <code className='text-xs'>{tenant.slug}</code>
              {tenant.is_default && (
                <Badge variant='default' className='ml-2'>
                  Default
                </Badge>
              )}
            </p>
          </div>
        </div>
      </div>

      <div className='flex-1 overflow-auto p-6'>
        <Tabs value={activeTab} onValueChange={setActiveTab} className='w-full'>
          <TabsList className='grid w-full max-w-lg grid-cols-4'>
            <TabsTrigger value='members'>
              <Users className='mr-2 h-4 w-4' />
              Members
            </TabsTrigger>
            <TabsTrigger value='oauth'>
              <Key className='mr-2 h-4 w-4' />
              OAuth
            </TabsTrigger>
            <TabsTrigger value='saml'>
              <Shield className='mr-2 h-4 w-4' />
              SAML
            </TabsTrigger>
            <TabsTrigger value='settings'>
              <Settings className='mr-2 h-4 w-4' />
              Settings
            </TabsTrigger>
          </TabsList>

          <TabsContent value='members' className='mt-6 space-y-6'>
            <Card>
              <CardHeader>
                <div className='flex items-center justify-between'>
                  <div>
                    <CardTitle className='flex items-center gap-2'>
                      <Users className='h-5 w-5' />
                      Members
                    </CardTitle>
                    <CardDescription>
                      Users with access to this tenant
                    </CardDescription>
                  </div>
                  <Button onClick={() => setAddMemberDialogOpen(true)}>
                    <Plus className='mr-2 h-4 w-4' />
                    Add Member
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                {membersLoading ? (
                  <div className='flex items-center justify-center py-8'>
                    <RefreshCw className='text-muted-foreground h-6 w-6 animate-spin' />
                  </div>
                ) : members && members.length > 0 ? (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Email</TableHead>
                        <TableHead>Role</TableHead>
                        <TableHead>Added</TableHead>
                        <TableHead className='text-right'>Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {members.map((member) => (
                        <TableRow key={member.id}>
                          <TableCell className='flex items-center gap-2'>
                            <Mail className='text-muted-foreground h-4 w-4' />
                            {member.email || member.user_id}
                          </TableCell>
                          <TableCell>
                            <div className='flex items-center gap-2'>
                              {member.role === 'tenant_admin' ? (
                                <ShieldCheck className='h-4 w-4 text-green-500' />
                              ) : (
                                <Shield className='text-muted-foreground h-4 w-4' />
                              )}
                              <Select
                                value={member.role}
                                onValueChange={(value) =>
                                  updateMemberMutation.mutate({
                                    userId: member.user_id,
                                    data: { role: value as 'tenant_admin' | 'tenant_member' },
                                  })
                                }
                              >
                                <SelectTrigger className='h-8 w-[140px]'>
                                  <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                  <SelectItem value='tenant_admin'>Admin</SelectItem>
                                  <SelectItem value='tenant_member'>Member</SelectItem>
                                </SelectContent>
                              </Select>
                            </div>
                          </TableCell>
                          <TableCell className='text-muted-foreground text-sm'>
                            {format(new Date(member.created_at), 'MMM d, yyyy')}
                          </TableCell>
                          <TableCell className='text-right'>
                            <AlertDialog>
                              <AlertDialogTrigger asChild>
                                <Button
                                  variant='ghost'
                                  size='sm'
                                  className='text-destructive hover:text-destructive hover:bg-destructive/10'
                                >
                                  <Trash2 className='h-4 w-4' />
                                </Button>
                              </AlertDialogTrigger>
                              <AlertDialogContent>
                                <AlertDialogHeader>
                                  <AlertDialogTitle>Remove Member</AlertDialogTitle>
                                  <AlertDialogDescription>
                                    Are you sure you want to remove {member.email} from
                                    this tenant?
                                  </AlertDialogDescription>
                                </AlertDialogHeader>
                                <AlertDialogFooter>
                                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                                  <AlertDialogAction
                                    onClick={() =>
                                      removeMemberMutation.mutate(member.user_id)
                                    }
                                    className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
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
                  <div className='flex flex-col items-center justify-center py-12 text-center'>
                    <Users className='text-muted-foreground mb-4 h-12 w-12' />
                    <p className='mb-2 text-lg font-medium'>No members yet</p>
                    <p className='text-muted-foreground mb-4 text-sm'>
                      Add members to give them access to this tenant
                    </p>
                    <Button onClick={() => setAddMemberDialogOpen(true)}>
                      <Plus className='mr-2 h-4 w-4' />
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
                <dl className='grid grid-cols-2 gap-4'>
                  <div>
                    <dt className='text-muted-foreground text-sm'>ID</dt>
                    <dd className='font-mono text-sm'>{tenant.id}</dd>
                  </div>
                  <div>
                    <dt className='text-muted-foreground text-sm'>Slug</dt>
                    <dd className='font-mono text-sm'>{tenant.slug}</dd>
                  </div>
                  <div>
                    <dt className='text-muted-foreground text-sm'>Created</dt>
                    <dd className='text-sm'>
                      {format(new Date(tenant.created_at), 'PPPpp')}
                    </dd>
                  </div>
                  <div>
                    <dt className='text-muted-foreground text-sm'>Default</dt>
                    <dd className='text-sm'>
                      {tenant.is_default ? 'Yes' : 'No'}
                    </dd>
                  </div>
                </dl>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='oauth' className='mt-6'>
            <TenantOAuthProvidersTab tenantId={tenantId} />
          </TabsContent>

          <TabsContent value='saml' className='mt-6'>
            <TenantSAMLProvidersTab tenantId={tenantId} />
          </TabsContent>

          <TabsContent value='settings' className='mt-6'>
            <TenantSettingsTab tenantId={tenantId} />
          </TabsContent>
        </Tabs>
      </div>

      <Dialog open={addMemberDialogOpen} onOpenChange={setAddMemberDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Member</DialogTitle>
            <DialogDescription>
              Add a user to this tenant
            </DialogDescription>
          </DialogHeader>
          <div className='grid gap-4 py-4'>
            <div className='grid gap-2'>
              <Label htmlFor='search'>Search Users</Label>
              <Input
                id='search'
                placeholder='Search by email...'
                value={searchEmail}
                onChange={(e) => setSearchEmail(e.target.value)}
              />
            </div>
            <div className='grid gap-2'>
              <Label htmlFor='user'>Select User</Label>
              <Select value={newMemberUserId} onValueChange={setNewMemberUserId}>
                <SelectTrigger>
                  <SelectValue placeholder='Select a user' />
                </SelectTrigger>
                <SelectContent>
                  <ScrollArea className='h-[200px]'>
                    {filteredUsers.length > 0 ? (
                      filteredUsers.map((user) => (
                        <SelectItem key={user.id} value={user.id}>
                          {user.email}
                        </SelectItem>
                      ))
                    ) : (
                      <div className='text-muted-foreground p-2 text-center text-sm'>
                        No users available
                      </div>
                    )}
                  </ScrollArea>
                </SelectContent>
              </Select>
            </div>
            <div className='grid gap-2'>
              <Label htmlFor='role'>Role</Label>
              <Select
                value={newMemberRole}
                onValueChange={(v) =>
                  setNewMemberRole(v as 'tenant_admin' | 'tenant_member')
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='tenant_admin'>Admin</SelectItem>
                  <SelectItem value='tenant_member'>Member</SelectItem>
                </SelectContent>
              </Select>
              <p className='text-muted-foreground text-xs'>
                Admins can manage members. Members have read access.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant='outline'
              onClick={() => setAddMemberDialogOpen(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleAddMember}
              disabled={addMemberMutation.isPending || !newMemberUserId}
            >
              {addMemberMutation.isPending ? 'Adding...' : 'Add Member'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function TenantOAuthProvidersTab({ tenantId }: { tenantId: string }) {
  const queryClient = useQueryClient()
  const [showAddDialog, setShowAddDialog] = useState(false)
  const [showEditDialog, setShowEditDialog] = useState(false)
  const [editingProvider, setEditingProvider] = useState<OAuthProviderConfig | null>(null)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deletingProvider, setDeletingProvider] = useState<OAuthProviderConfig | null>(null)

  const [selectedProvider, setSelectedProvider] = useState('')
  const [customProviderName, setCustomProviderName] = useState('')
  const [clientId, setClientId] = useState('')
  const [clientSecret, setClientSecret] = useState('')
  const [customAuthUrl, setCustomAuthUrl] = useState('')
  const [customTokenUrl, setCustomTokenUrl] = useState('')
  const [customUserInfoUrl, setCustomUserInfoUrl] = useState('')
  const [allowAppLogin, setAllowAppLogin] = useState(true)

  const { data: providers = [], isLoading } = useQuery({
    queryKey: ['tenant-oauth-providers', tenantId],
    queryFn: oauthProviderApi.list,
  })

  const tenantProviders = providers.filter(p => !p.allow_dashboard_login || p.source === 'database')

  const createMutation = useMutation({
    mutationFn: (data: CreateOAuthProviderRequest) => oauthProviderApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant-oauth-providers', tenantId] })
      toast.success('OAuth provider created')
      setShowAddDialog(false)
      resetForm()
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && 'response' in error
          ? (error as { response?: { data?: { error?: string } } }).response?.data?.error || 'Failed to create provider'
          : 'Failed to create provider'
      toast.error(errorMessage)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateOAuthProviderRequest }) =>
      oauthProviderApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant-oauth-providers', tenantId] })
      toast.success('OAuth provider updated')
      setShowEditDialog(false)
      setEditingProvider(null)
      resetForm()
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && 'response' in error
          ? (error as { response?: { data?: { error?: string } } }).response?.data?.error || 'Failed to update provider'
          : 'Failed to update provider'
      toast.error(errorMessage)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => oauthProviderApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant-oauth-providers', tenantId] })
      toast.success('OAuth provider deleted')
      setShowDeleteConfirm(false)
      setDeletingProvider(null)
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && 'response' in error
          ? (error as { response?: { data?: { error?: string } } }).response?.data?.error || 'Failed to delete provider'
          : 'Failed to delete provider'
      toast.error(errorMessage)
    },
  })

  const resetForm = () => {
    setSelectedProvider('')
    setCustomProviderName('')
    setClientId('')
    setClientSecret('')
    setCustomAuthUrl('')
    setCustomTokenUrl('')
    setCustomUserInfoUrl('')
    setAllowAppLogin(true)
  }

  const handleCreate = () => {
    if (!selectedProvider || !clientId || !clientSecret) {
      toast.error('Please fill in all required fields')
      return
    }

    const isCustom = selectedProvider === 'custom'
    createMutation.mutate({
      provider_name: isCustom ? customProviderName.toLowerCase().replace(/[^a-z0-9_-]/g, '_') : selectedProvider,
      display_name: isCustom ? customProviderName : selectedProvider.charAt(0).toUpperCase() + selectedProvider.slice(1),
      enabled: true,
      client_id: clientId,
      client_secret: clientSecret,
      redirect_url: `${window.location.origin}/api/v1/auth/oauth/callback`,
      scopes: selectedProvider === 'google' ? ['openid', 'email', 'profile'] :
              selectedProvider === 'github' ? ['read:user', 'user:email'] :
              selectedProvider === 'microsoft' ? ['openid', 'email', 'profile'] :
              ['openid', 'email', 'profile'],
      is_custom: isCustom,
      authorization_url: isCustom ? customAuthUrl : undefined,
      token_url: isCustom ? customTokenUrl : undefined,
      user_info_url: isCustom ? customUserInfoUrl : undefined,
      allow_dashboard_login: false,
      allow_app_login: allowAppLogin,
    })
  }

  const handleEdit = (provider: OAuthProviderConfig) => {
    setEditingProvider(provider)
    setClientId(provider.client_id)
    setClientSecret('')
    setAllowAppLogin(provider.allow_app_login)
    if (provider.is_custom) {
      setSelectedProvider('custom')
      setCustomProviderName(provider.display_name)
      setCustomAuthUrl(provider.authorization_url || '')
      setCustomTokenUrl(provider.token_url || '')
      setCustomUserInfoUrl(provider.user_info_url || '')
    } else {
      setSelectedProvider(provider.provider_name)
    }
    setShowEditDialog(true)
  }

  const handleUpdate = () => {
    if (!editingProvider) return
    updateMutation.mutate({
      id: editingProvider.id,
      data: {
        client_id: clientId || undefined,
        ...(clientSecret && { client_secret: clientSecret }),
        allow_app_login: allowAppLogin,
      },
    })
  }

  const availableProviders = [
    { id: 'google', name: 'Google' },
    { id: 'github', name: 'GitHub' },
    { id: 'microsoft', name: 'Microsoft' },
    { id: 'custom', name: 'Custom Provider' },
  ]

  if (isLoading) {
    return (
      <div className='flex justify-center p-8'>
        <Loader2 className='h-6 w-6 animate-spin' />
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <Card>
        <CardHeader>
          <div className='flex items-center justify-between'>
            <div>
              <CardTitle>OAuth Providers</CardTitle>
              <CardDescription>
                Configure OAuth providers for tenant authentication
              </CardDescription>
            </div>
            <Button onClick={() => setShowAddDialog(true)}>
              <Plus className='mr-2 h-4 w-4' />
              Add Provider
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {tenantProviders.length === 0 ? (
            <div className='flex flex-col items-center justify-center py-12 text-center'>
              <AlertCircle className='text-muted-foreground mb-4 h-12 w-12' />
              <p className='text-muted-foreground mb-2'>No OAuth providers configured</p>
              <Button variant='outline' onClick={() => setShowAddDialog(true)}>
                Add Your First Provider
              </Button>
            </div>
          ) : (
            <div className='space-y-4'>
              {tenantProviders.map((provider) => (
                <Card key={provider.id}>
                  <CardContent className='pt-6'>
                    <div className='flex items-start justify-between'>
                      <div className='flex-1 space-y-2'>
                        <div className='flex items-center gap-2'>
                          <h3 className='text-lg font-semibold'>{provider.display_name}</h3>
                          {provider.enabled ? (
                            <Badge variant='default' className='gap-1'>
                              <Check className='h-3 w-3' />
                              Enabled
                            </Badge>
                          ) : (
                            <Badge variant='secondary'>Disabled</Badge>
                          )}
                          {provider.allow_app_login && (
                            <Badge variant='outline' className='text-xs'>App</Badge>
                          )}
                        </div>
                        <div className='text-sm'>
                          <span className='text-muted-foreground'>Client ID: </span>
                          <code className='font-mono text-xs'>{provider.client_id}</code>
                        </div>
                      </div>
                      <div className='flex gap-2'>
                        {provider.source !== 'config' && (
                          <>
                            <Button variant='outline' size='sm' onClick={() => handleEdit(provider)}>
                              Edit
                            </Button>
                            <Button
                              variant='ghost'
                              size='sm'
                              className='text-destructive hover:text-destructive'
                              onClick={() => {
                                setDeletingProvider(provider)
                                setShowDeleteConfirm(true)
                              }}
                            >
                              <Trash2 className='h-4 w-4' />
                            </Button>
                          </>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add OAuth Provider</DialogTitle>
            <DialogDescription>
              Configure a new OAuth provider for this tenant
            </DialogDescription>
          </DialogHeader>
          <div className='grid gap-4 py-4'>
            <div className='grid gap-2'>
              <Label>Provider</Label>
              <Select value={selectedProvider} onValueChange={setSelectedProvider}>
                <SelectTrigger>
                  <SelectValue placeholder='Select provider' />
                </SelectTrigger>
                <SelectContent>
                  {availableProviders.map((p) => (
                    <SelectItem key={p.id} value={p.id}>{p.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {selectedProvider === 'custom' && (
              <div className='grid gap-2'>
                <Label>Provider Name</Label>
                <Input value={customProviderName} onChange={(e) => setCustomProviderName(e.target.value)} placeholder='my-provider' />
              </div>
            )}
            <div className='grid gap-2'>
              <Label>Client ID</Label>
              <Input value={clientId} onChange={(e) => setClientId(e.target.value)} placeholder='Enter client ID' />
            </div>
            <div className='grid gap-2'>
              <Label>Client Secret</Label>
              <Input type='password' value={clientSecret} onChange={(e) => setClientSecret(e.target.value)} placeholder='Enter client secret' />
            </div>
            {selectedProvider === 'custom' && (
              <>
                <div className='grid gap-2'>
                  <Label>Authorization URL</Label>
                  <Input value={customAuthUrl} onChange={(e) => setCustomAuthUrl(e.target.value)} placeholder='https://provider.com/oauth/authorize' />
                </div>
                <div className='grid gap-2'>
                  <Label>Token URL</Label>
                  <Input value={customTokenUrl} onChange={(e) => setCustomTokenUrl(e.target.value)} placeholder='https://provider.com/oauth/token' />
                </div>
                <div className='grid gap-2'>
                  <Label>User Info URL</Label>
                  <Input value={customUserInfoUrl} onChange={(e) => setCustomUserInfoUrl(e.target.value)} placeholder='https://provider.com/oauth/userinfo' />
                </div>
              </>
            )}
            <div className='flex items-center gap-2'>
              <Switch checked={allowAppLogin} onCheckedChange={setAllowAppLogin} />
              <Label>Allow App Login</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setShowAddDialog(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit OAuth Provider</DialogTitle>
            <DialogDescription>
              Update the OAuth provider configuration
            </DialogDescription>
          </DialogHeader>
          <div className='grid gap-4 py-4'>
            <div className='grid gap-2'>
              <Label>Client ID</Label>
              <Input value={clientId} onChange={(e) => setClientId(e.target.value)} />
            </div>
            <div className='grid gap-2'>
              <Label>Client Secret (leave empty to keep current)</Label>
              <Input type='password' value={clientSecret} onChange={(e) => setClientSecret(e.target.value)} placeholder='Enter new secret or leave empty' />
            </div>
            <div className='flex items-center gap-2'>
              <Switch checked={allowAppLogin} onCheckedChange={setAllowAppLogin} />
              <Label>Allow App Login</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setShowEditDialog(false)}>Cancel</Button>
            <Button onClick={handleUpdate} disabled={updateMutation.isPending}>
              {updateMutation.isPending ? 'Updating...' : 'Update'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete OAuth Provider</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{deletingProvider?.display_name}"? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deletingProvider && deleteMutation.mutate(deletingProvider.id)}
              className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

function TenantSAMLProvidersTab({ tenantId }: { tenantId: string }) {
  const queryClient = useQueryClient()
  const [showAddDialog, setShowAddDialog] = useState(false)
  const [showEditDialog, setShowEditDialog] = useState(false)
  const [editingProvider, setEditingProvider] = useState<SAMLProviderConfig | null>(null)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deletingProvider, setDeletingProvider] = useState<SAMLProviderConfig | null>(null)

  const [providerName, setProviderName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [metadataSource, setMetadataSource] = useState<'url' | 'xml'>('url')
  const [metadataUrl, setMetadataUrl] = useState('')
  const [metadataXml, setMetadataXml] = useState('')
  const [autoCreateUsers, setAutoCreateUsers] = useState(true)
  const [defaultRole, setDefaultRole] = useState('authenticated')
  const [allowAppLogin, setAllowAppLogin] = useState(true)

  const { data: providers = [], isLoading } = useQuery({
    queryKey: ['tenant-saml-providers', tenantId],
    queryFn: samlProviderApi.list,
  })

  const tenantProviders = providers.filter(p => !p.allow_dashboard_login || p.source === 'database')

  const createMutation = useMutation({
    mutationFn: (data: CreateSAMLProviderRequest) => samlProviderApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant-saml-providers', tenantId] })
      toast.success('SAML provider created')
      setShowAddDialog(false)
      resetForm()
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && 'response' in error
          ? (error as { response?: { data?: { error?: string } } }).response?.data?.error || 'Failed to create provider'
          : 'Failed to create provider'
      toast.error(errorMessage)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateSAMLProviderRequest }) =>
      samlProviderApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant-saml-providers', tenantId] })
      toast.success('SAML provider updated')
      setShowEditDialog(false)
      setEditingProvider(null)
      resetForm()
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && 'response' in error
          ? (error as { response?: { data?: { error?: string } } }).response?.data?.error || 'Failed to update provider'
          : 'Failed to update provider'
      toast.error(errorMessage)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => samlProviderApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant-saml-providers', tenantId] })
      toast.success('SAML provider deleted')
      setShowDeleteConfirm(false)
      setDeletingProvider(null)
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error && 'response' in error
          ? (error as { response?: { data?: { error?: string } } }).response?.data?.error || 'Failed to delete provider'
          : 'Failed to delete provider'
      toast.error(errorMessage)
    },
  })

  const resetForm = () => {
    setProviderName('')
    setDisplayName('')
    setMetadataSource('url')
    setMetadataUrl('')
    setMetadataXml('')
    setAutoCreateUsers(true)
    setDefaultRole('authenticated')
    setAllowAppLogin(true)
  }

  const handleCreate = () => {
    if (!providerName) {
      toast.error('Provider name is required')
      return
    }
    if (metadataSource === 'url' && !metadataUrl) {
      toast.error('Metadata URL is required')
      return
    }
    if (metadataSource === 'xml' && !metadataXml) {
      toast.error('Metadata XML is required')
      return
    }

    createMutation.mutate({
      name: providerName.toLowerCase().replace(/[^a-z0-9_-]/g, '_'),
      display_name: displayName || providerName,
      enabled: true,
      idp_metadata_url: metadataSource === 'url' ? metadataUrl : undefined,
      idp_metadata_xml: metadataSource === 'xml' ? metadataXml : undefined,
      auto_create_users: autoCreateUsers,
      default_role: defaultRole,
      allow_dashboard_login: false,
      allow_app_login: allowAppLogin,
    })
  }

  const handleEdit = (provider: SAMLProviderConfig) => {
    setEditingProvider(provider)
    setProviderName(provider.name)
    setDisplayName(provider.display_name)
    setMetadataUrl(provider.idp_metadata_url || '')
    setMetadataXml(provider.idp_metadata_xml || '')
    setMetadataSource(provider.idp_metadata_url ? 'url' : 'xml')
    setAutoCreateUsers(provider.auto_create_users)
    setDefaultRole(provider.default_role)
    setAllowAppLogin(provider.allow_app_login)
    setShowEditDialog(true)
  }

  const handleUpdate = () => {
    if (!editingProvider) return
    updateMutation.mutate({
      id: editingProvider.id,
      data: {
        display_name: displayName || undefined,
        idp_metadata_url: metadataSource === 'url' ? metadataUrl : undefined,
        idp_metadata_xml: metadataSource === 'xml' ? metadataXml : undefined,
        auto_create_users: autoCreateUsers,
        default_role: defaultRole,
        allow_app_login: allowAppLogin,
      },
    })
  }

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} copied to clipboard`)
  }

  if (isLoading) {
    return (
      <div className='flex justify-center p-8'>
        <Loader2 className='h-6 w-6 animate-spin' />
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <Card>
        <CardHeader>
          <div className='flex items-center justify-between'>
            <div>
              <CardTitle>SAML SSO Providers</CardTitle>
              <CardDescription>
                Configure SAML providers for enterprise single sign-on
              </CardDescription>
            </div>
            <Button onClick={() => setShowAddDialog(true)}>
              <Plus className='mr-2 h-4 w-4' />
              Add Provider
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {tenantProviders.length === 0 ? (
            <div className='flex flex-col items-center justify-center py-12 text-center'>
              <Shield className='text-muted-foreground mb-4 h-12 w-12' />
              <p className='text-muted-foreground mb-2'>No SAML providers configured</p>
              <Button variant='outline' onClick={() => setShowAddDialog(true)}>
                Add Your First Provider
              </Button>
            </div>
          ) : (
            <div className='space-y-4'>
              {tenantProviders.map((provider) => (
                <Card key={provider.id}>
                  <CardContent className='pt-6'>
                    <div className='flex items-start justify-between'>
                      <div className='flex-1 space-y-4'>
                        <div className='flex items-center gap-2'>
                          <h3 className='text-lg font-semibold'>{provider.display_name || provider.name}</h3>
                          {provider.enabled ? (
                            <Badge variant='default' className='gap-1'>
                              <Check className='h-3 w-3' />
                              Enabled
                            </Badge>
                          ) : (
                            <Badge variant='secondary'>Disabled</Badge>
                          )}
                          {provider.allow_app_login && (
                            <Badge variant='outline'>App Login</Badge>
                          )}
                        </div>
                        <div className='grid grid-cols-1 gap-4 text-sm md:grid-cols-2'>
                          <div>
                            <Label className='text-muted-foreground'>Entity ID (SP)</Label>
                            <div className='mt-1 flex items-center gap-2'>
                              <code className='flex-1 rounded bg-muted px-2 py-1 text-xs break-all'>{provider.entity_id}</code>
                              <Button variant='ghost' size='sm' className='h-6 w-6 p-0' onClick={() => copyToClipboard(provider.entity_id, 'Entity ID')}>
                                <Copy className='h-3 w-3' />
                              </Button>
                            </div>
                          </div>
                          <div>
                            <Label className='text-muted-foreground'>ACS URL</Label>
                            <div className='mt-1 flex items-center gap-2'>
                              <code className='flex-1 rounded bg-muted px-2 py-1 text-xs break-all'>{provider.acs_url}</code>
                              <Button variant='ghost' size='sm' className='h-6 w-6 p-0' onClick={() => copyToClipboard(provider.acs_url, 'ACS URL')}>
                                <Copy className='h-3 w-3' />
                              </Button>
                            </div>
                          </div>
                        </div>
                      </div>
                      <div className='ml-4 flex gap-2'>
                        {provider.source !== 'config' && (
                          <>
                            <Button variant='outline' size='sm' onClick={() => handleEdit(provider)}>
                              Edit
                            </Button>
                            <Button
                              variant='ghost'
                              size='sm'
                              className='text-destructive hover:text-destructive'
                              onClick={() => {
                                setDeletingProvider(provider)
                                setShowDeleteConfirm(true)
                              }}
                            >
                              <Trash2 className='h-4 w-4' />
                            </Button>
                          </>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
        <DialogContent className='max-w-2xl'>
          <DialogHeader>
            <DialogTitle>Add SAML Provider</DialogTitle>
            <DialogDescription>
              Configure a new SAML provider for this tenant
            </DialogDescription>
          </DialogHeader>
          <div className='grid gap-4 py-4'>
            <div className='grid grid-cols-2 gap-4'>
              <div className='grid gap-2'>
                <Label>Provider Name</Label>
                <Input value={providerName} onChange={(e) => setProviderName(e.target.value)} placeholder='okta' />
              </div>
              <div className='grid gap-2'>
                <Label>Display Name</Label>
                <Input value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder='Okta SSO' />
              </div>
            </div>
            <div className='grid gap-2'>
              <Label>Metadata Source</Label>
              <Select value={metadataSource} onValueChange={(v) => setMetadataSource(v as 'url' | 'xml')}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='url'>Metadata URL</SelectItem>
                  <SelectItem value='xml'>Metadata XML</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {metadataSource === 'url' ? (
              <div className='grid gap-2'>
                <Label>IdP Metadata URL</Label>
                <Input value={metadataUrl} onChange={(e) => setMetadataUrl(e.target.value)} placeholder='https://idp.example.com/metadata' />
              </div>
            ) : (
              <div className='grid gap-2'>
                <Label>IdP Metadata XML</Label>
                <Textarea value={metadataXml} onChange={(e) => setMetadataXml(e.target.value)} placeholder='Paste metadata XML here...' rows={6} />
              </div>
            )}
            <div className='grid grid-cols-2 gap-4'>
              <div className='grid gap-2'>
                <Label>Default Role</Label>
                <Input value={defaultRole} onChange={(e) => setDefaultRole(e.target.value)} placeholder='authenticated' />
              </div>
            </div>
            <div className='flex items-center gap-4'>
              <div className='flex items-center gap-2'>
                <Switch checked={autoCreateUsers} onCheckedChange={setAutoCreateUsers} />
                <Label>Auto-create Users</Label>
              </div>
              <div className='flex items-center gap-2'>
                <Switch checked={allowAppLogin} onCheckedChange={setAllowAppLogin} />
                <Label>Allow App Login</Label>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setShowAddDialog(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent className='max-w-2xl'>
          <DialogHeader>
            <DialogTitle>Edit SAML Provider</DialogTitle>
            <DialogDescription>
              Update the SAML provider configuration
            </DialogDescription>
          </DialogHeader>
          <div className='grid gap-4 py-4'>
            <div className='grid grid-cols-2 gap-4'>
              <div className='grid gap-2'>
                <Label>Display Name</Label>
                <Input value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
              </div>
              <div className='grid gap-2'>
                <Label>Default Role</Label>
                <Input value={defaultRole} onChange={(e) => setDefaultRole(e.target.value)} />
              </div>
            </div>
            <div className='grid gap-2'>
              <Label>Metadata Source</Label>
              <Select value={metadataSource} onValueChange={(v) => setMetadataSource(v as 'url' | 'xml')}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='url'>Metadata URL</SelectItem>
                  <SelectItem value='xml'>Metadata XML</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {metadataSource === 'url' ? (
              <div className='grid gap-2'>
                <Label>IdP Metadata URL</Label>
                <Input value={metadataUrl} onChange={(e) => setMetadataUrl(e.target.value)} />
              </div>
            ) : (
              <div className='grid gap-2'>
                <Label>IdP Metadata XML</Label>
                <Textarea value={metadataXml} onChange={(e) => setMetadataXml(e.target.value)} rows={6} />
              </div>
            )}
            <div className='flex items-center gap-4'>
              <div className='flex items-center gap-2'>
                <Switch checked={autoCreateUsers} onCheckedChange={setAutoCreateUsers} />
                <Label>Auto-create Users</Label>
              </div>
              <div className='flex items-center gap-2'>
                <Switch checked={allowAppLogin} onCheckedChange={setAllowAppLogin} />
                <Label>Allow App Login</Label>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setShowEditDialog(false)}>Cancel</Button>
            <Button onClick={handleUpdate} disabled={updateMutation.isPending}>
              {updateMutation.isPending ? 'Updating...' : 'Update'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete SAML Provider</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{deletingProvider?.display_name}"? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deletingProvider && deleteMutation.mutate(deletingProvider.id)}
              className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
