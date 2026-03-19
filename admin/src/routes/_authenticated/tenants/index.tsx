import { useState } from 'react'
import { format } from 'date-fns'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { Building2, Plus, Trash2, Search, Users, CheckCircle, Pencil } from 'lucide-react'
import { toast } from 'sonner'
import { tenantsApi, type CreateTenantRequest,
 type Tenant,
 type UpdateTenantRequest as UpdateTenantReq } from '@/lib/api'
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
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

export const Route = createFileRoute('/_authenticated/tenants/')({
  component: TenantsPage,
})

function TenantsPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [searchQuery, setSearchQuery] = useState('')
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [editingTenant, setEditingTenant] = useState<Tenant | null>(null)
  const [newTenant, setNewTenant] = useState({ name: '', slug: '' })
  const [editTenant, setEditTenant] = useState({ name: '' })

  const { data: tenants, isLoading } = useQuery({
    queryKey: ['tenants'],
    queryFn: tenantsApi.list,
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateTenantRequest) => tenantsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      toast.success('Tenant created successfully')
      setCreateDialogOpen(false)
      setNewTenant({ name: '', slug: '' })
    },
    onError: (error: Error) => {
      toast.error(`Failed to create tenant: ${error.message}`)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => tenantsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      toast.success('Tenant deleted successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete tenant: ${error.message}`)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTenantReq }) =>
      tenantsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      toast.success('Tenant updated successfully')
      setEditDialogOpen(false)
      setEditingTenant(null)
      setEditTenant({ name: '' })
    },
    onError: (error: Error) => {
      toast.error(`Failed to update tenant: ${error.message}`)
    },
  })

  const handleEditTenant = (tenant: Tenant) => {
    setEditingTenant(tenant)
    setEditTenant({ name: tenant.name })
    setEditDialogOpen(true)
  }

  const handleUpdateTenant = () => {
    if (!editingTenant) return
    if (!editTenant.name.trim()) {
      toast.error('Name is required')
      return
    }
    updateMutation.mutate({
      id: editingTenant.id,
      data: {
        name: editTenant.name.trim(),
      },
    })
  }

  const handleCreateTenant = () => {
    if (!newTenant.name.trim()) {
      toast.error('Name is required')
      return
    }
    if (!newTenant.slug.trim()) {
      toast.error('Slug is required')
      return
    }
    if (!/^[a-z][a-z0-9-]*[a-z0-9]$/.test(newTenant.slug)) {
      toast.error(
        'Slug must start with a lowercase letter, contain only lowercase letters, numbers, and hyphens, and end with a letter or number'
      )
      return
    }
    createMutation.mutate({
      name: newTenant.name.trim(),
      slug: newTenant.slug.trim().toLowerCase(),
    })
  }

  const generateSlug = (name: string) => {
    return name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '')
  }

  const filteredTenants = tenants?.filter(
    (tenant) =>
      tenant.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      tenant.slug.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className='flex h-full flex-col'>
      <div className='bg-background flex items-center justify-between border-b px-6 py-4'>
        <div className='flex items-center gap-3'>
          <div className='bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg'>
            <Building2 className='text-primary h-5 w-5' />
          </div>
          <div>
            <h1 className='text-xl font-semibold'>Tenants</h1>
            <p className='text-muted-foreground text-sm'>
              Manage multi-tenant organizations
            </p>
          </div>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className='mr-2 h-4 w-4' />
          Create Tenant
        </Button>
      </div>

      <div className='flex-1 overflow-auto p-6'>
        <div className='flex flex-col gap-6'>
          <div className='grid gap-4 md:grid-cols-3'>
            <Card>
              <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                <CardTitle className='text-sm font-medium'>Total Tenants</CardTitle>
                <Building2 className='text-muted-foreground h-4 w-4' />
              </CardHeader>
              <CardContent>
                <div className='text-2xl font-bold'>{tenants?.length || 0}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                <CardTitle className='text-sm font-medium'>Default Tenant</CardTitle>
                <CheckCircle className='text-muted-foreground h-4 w-4' />
              </CardHeader>
              <CardContent>
                <div className='text-2xl font-bold'>
                  {tenants?.filter((t) => t.is_default).length || 0}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                <CardTitle className='text-sm font-medium'>Custom Tenants</CardTitle>
                <Users className='text-muted-foreground h-4 w-4' />
              </CardHeader>
              <CardContent>
                <div className='text-2xl font-bold'>
                  {tenants?.filter((t) => !t.is_default).length || 0}
                </div>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <div className='flex items-center justify-between'>
                <div>
                  <CardTitle>Tenants</CardTitle>
                  <CardDescription>
                    All tenants in the system
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className='mb-4'>
                <div className='relative'>
                  <Search className='text-muted-foreground absolute top-2.5 left-2 h-4 w-4' />
                  <Input
                    placeholder='Search by name or slug...'
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className='pl-8'
                  />
                </div>
              </div>

              {isLoading ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Slug</TableHead>
                      <TableHead>Default</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead className='text-right'>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {Array(3)
                      .fill(0)
                      .map((_, i) => (
                        <TableRow key={i}>
                          <TableCell>
                            <Skeleton className='h-4 w-32' />
                          </TableCell>
                          <TableCell>
                            <Skeleton className='h-4 w-24' />
                          </TableCell>
                          <TableCell>
                            <Skeleton className='h-5 w-16' />
                          </TableCell>
                          <TableCell>
                            <Skeleton className='h-4 w-24' />
                          </TableCell>
                          <TableCell className='text-right'>
                            <Skeleton className='ml-auto h-8 w-8' />
                          </TableCell>
                        </TableRow>
                      ))}
                  </TableBody>
                </Table>
              ) : filteredTenants && filteredTenants.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Slug</TableHead>
                      <TableHead>Default</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead className='text-right'>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredTenants.map((tenant) => (
                      <TableRow
                        key={tenant.id}
                        className='cursor-pointer'
                        onClick={() => navigate({ to: '/tenants/$tenantId', params: { tenantId: tenant.id } })}
                      >
                        <TableCell className='font-medium'>{tenant.name}</TableCell>
                        <TableCell>
                          <code className='text-xs'>{tenant.slug}</code>
                        </TableCell>
                        <TableCell>
                          {tenant.is_default ? (
                            <Badge variant='default'>Default</Badge>
                          ) : (
                            <Badge variant='outline'>Custom</Badge>
                          )}
                        </TableCell>
                        <TableCell className='text-muted-foreground text-sm'>
                          {format(new Date(tenant.created_at), 'MMM d, yyyy')}
                        </TableCell>
                         <TableCell className='text-right' onClick={(e) => e.stopPropagation()}>
                           <div className='flex items-center justify-end gap-1'>
                             <Button
                               variant='ghost'
                               size='sm'
                               onClick={() => handleEditTenant(tenant)}
                             >
                               <Pencil className='h-4 w-4' />
                             </Button>
                             {!tenant.is_default && (
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
                                     <AlertDialogTitle>Delete Tenant</AlertDialogTitle>
                                     <AlertDialogDescription>
                                       Are you sure you want to delete "{tenant.name}"? This
                                       action cannot be undone.
                                     </AlertDialogDescription>
                                   </AlertDialogHeader>
                                   <AlertDialogFooter>
                                     <AlertDialogCancel>Cancel</AlertDialogCancel>
                                     <AlertDialogAction
                                       onClick={() => deleteMutation.mutate(tenant.id)}
                                       className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
                                     >
                                       Delete
                                     </AlertDialogAction>
                                   </AlertDialogFooter>
                                 </AlertDialogContent>
                               </AlertDialog>
                             )}
                           </div>
                         </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className='flex flex-col items-center justify-center py-12 text-center'>
                  <Building2 className='text-muted-foreground mb-4 h-12 w-12' />
                  <p className='text-muted-foreground'>
                    {searchQuery ? 'No tenants match your search' : 'No tenants yet'}
                  </p>
                  {!searchQuery && (
                    <Button
                      onClick={() => setCreateDialogOpen(true)}
                      variant='outline'
                      className='mt-4'
                    >
                      Create Your First Tenant
                    </Button>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Tenant</DialogTitle>
            <DialogDescription>
              Create a new tenant organization. The slug is used as a unique identifier.
            </DialogDescription>
          </DialogHeader>
          <div className='grid gap-4 py-4'>
            <div className='grid gap-2'>
              <Label htmlFor='name'>
                Name <span className='text-destructive'>*</span>
              </Label>
              <Input
                id='name'
                placeholder='Acme Corporation'
                value={newTenant.name}
                onChange={(e) => {
                  const name = e.target.value
                  setNewTenant({
                    ...newTenant,
                    name,
                    slug: newTenant.slug || generateSlug(name),
                  })
                }}
              />
            </div>
            <div className='grid gap-2'>
              <Label htmlFor='slug'>
                Slug <span className='text-destructive'>*</span>
              </Label>
              <Input
                id='slug'
                placeholder='acme-corporation'
                value={newTenant.slug}
                onChange={(e) =>
                  setNewTenant({ ...newTenant, slug: e.target.value })
                }
              />
              <p className='text-muted-foreground text-xs'>
                Lowercase letters, numbers, and hyphens only. Must start with a letter.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant='outline'
              onClick={() => setCreateDialogOpen(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreateTenant}
              disabled={createMutation.isPending}
            >
              {createMutation.isPending ? 'Creating...' : 'Create Tenant'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Tenant</DialogTitle>
            <DialogDescription>
              Update tenant name.
            </DialogDescription>
          </DialogHeader>
          <div className='grid gap-4 py-4'>
            <div className='grid gap-2'>
              <Label htmlFor='edit-name'>
                Name <span className='text-destructive'>*</span>
              </Label>
              <Input
                id='edit-name'
                value={editTenant.name}
                onChange={(e) =>
                  setEditTenant({ ...editTenant, name: e.target.value })
                }
                placeholder='Acme Corporation'
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant='outline'
              onClick={() => setEditDialogOpen(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleUpdateTenant}
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? 'Updating...' : 'Update Tenant'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
