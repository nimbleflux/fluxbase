import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Building2,
  Users,
  Key,
  Shield,
  Settings,
  RefreshCw,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  tenantsApi,
  type AddMemberRequest,
  type UpdateMemberRequest,
  userManagementApi,
} from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { TenantSettingsTab } from './-TenantSettingsTab'
import { TenantMembersTab, TenantOAuthProvidersTab, TenantSAMLProvidersTab } from '@/components/tenant-detail'

export const Route = createFileRoute('/_authenticated/tenants/$tenantId')({
  component: TenantDetailPage,
})

function TenantDetailPage() {
  const { tenantId } = Route.useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState('members')

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

  const handleAddMember = (data: AddMemberRequest) => {
    addMemberMutation.mutate(data)
  }

  const handleUpdateMemberRole = (userId: string, data: UpdateMemberRequest) => {
    updateMemberMutation.mutate({ userId, data })
  }

  const handleRemoveMember = (userId: string) => {
    removeMemberMutation.mutate(userId)
  }

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
            <TenantMembersTab
              tenant={tenant}
              members={members}
              membersLoading={membersLoading}
              users={users}
              onAddMember={handleAddMember}
              onUpdateMemberRole={handleUpdateMemberRole}
              onRemoveMember={handleRemoveMember}
              isAddingMember={addMemberMutation.isPending}
            />
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
    </div>
  )
}
