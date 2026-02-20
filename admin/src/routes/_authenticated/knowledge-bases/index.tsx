import { useState, useEffect, useCallback, useRef } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {
  BookOpen,
  Plus,
  RefreshCw,
  Trash2,
  Settings,
  Search,
  FileText,
  Users,
  Lock,
  Globe,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  knowledgeBasesApi,
  userKnowledgeBasesApi,
  userManagementApi,
  type KnowledgeBaseSummary,
  type CreateKnowledgeBaseRequest,
  type KBVisibility,
  type KBPermission,
  type EnrichedUser,
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
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

export const Route = createFileRoute('/_authenticated/knowledge-bases/')({
  component: KnowledgeBasesPage,
})

function KnowledgeBasesPage() {
  const navigate = useNavigate()
  const [knowledgeBases, setKnowledgeBases] = useState<KnowledgeBaseSummary[]>(
    []
  )
  const [loading, setLoading] = useState(true)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [users, setUsers] = useState<EnrichedUser[]>([])
  const [usersLoading, setUsersLoading] = useState(false)
  const [newKB, setNewKB] = useState<CreateKnowledgeBaseRequest>({
    name: '',
    description: '',
    visibility: 'private',
    default_user_permission: 'viewer',
    chunk_size: 512,
    chunk_overlap: 50,
    chunk_strategy: 'recursive',
    initial_permissions: [],
  })
  const [newPermission, setNewPermission] = useState<{
    user_id: string
    permission: KBPermission
  }>({ user_id: '', permission: 'viewer' })
  const usersLoadedRef = useRef(false)

  const fetchKnowledgeBases = async () => {
    setLoading(true)
    try {
      const data = await knowledgeBasesApi.list()
      setKnowledgeBases(data || [])
    } catch {
      toast.error('Failed to fetch knowledge bases')
    } finally {
      setLoading(false)
    }
  }

  const fetchUsers = useCallback(async () => {
    if (usersLoadedRef.current) return // Already loaded
    setUsersLoading(true)
    try {
      const { users: data } = await userManagementApi.listUsers('app')
      setUsers(data || [])
      usersLoadedRef.current = true
    } catch {
      toast.error('Failed to fetch users')
    } finally {
      setUsersLoading(false)
    }
  }, [])

  const addPermission = () => {
    if (!newPermission.user_id) {
      toast.error('Please select a user')
      return
    }
    if (
      newKB.initial_permissions?.some(
        (p) => p.user_id === newPermission.user_id
      )
    ) {
      toast.error('User already has permission')
      return
    }
    setNewKB({
      ...newKB,
      initial_permissions: [
        ...(newKB.initial_permissions || []),
        {
          user_id: newPermission.user_id,
          permission: newPermission.permission,
        },
      ],
    })
    setNewPermission({ user_id: '', permission: 'viewer' })
  }

  const removePermission = (userId: string) => {
    setNewKB({
      ...newKB,
      initial_permissions:
        newKB.initial_permissions?.filter((p) => p.user_id !== userId) || [],
    })
  }

  const handleCreate = async () => {
    if (!newKB.name.trim()) {
      toast.error('Name is required')
      return
    }

    try {
      await userKnowledgeBasesApi.create(newKB)
      toast.success('Knowledge base created')
      setCreateDialogOpen(false)
      setNewKB({
        name: '',
        description: '',
        visibility: 'private',
        default_user_permission: 'viewer',
        chunk_size: 512,
        chunk_overlap: 50,
        chunk_strategy: 'recursive',
        initial_permissions: [],
      })
      setNewPermission({ user_id: '', permission: 'viewer' })
      await fetchKnowledgeBases()
    } catch (error) {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data
          ?.error || 'Failed to create knowledge base'
      toast.error(message)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await knowledgeBasesApi.delete(id)
      toast.success('Knowledge base deleted')
      await fetchKnowledgeBases()
    } catch {
      toast.error('Failed to delete knowledge base')
    } finally {
      setDeleteConfirm(null)
    }
  }

  const toggleEnabled = async (kb: KnowledgeBaseSummary) => {
    try {
      await knowledgeBasesApi.update(kb.id, { enabled: !kb.enabled })
      toast.success(`Knowledge base ${kb.enabled ? 'disabled' : 'enabled'}`)
      await fetchKnowledgeBases()
    } catch {
      toast.error('Failed to update knowledge base')
    }
  }

  useEffect(() => {
    fetchKnowledgeBases()
  }, [])

  useEffect(() => {
    if (createDialogOpen) {
      fetchUsers()
    }
  }, [createDialogOpen, fetchUsers])

  if (loading) {
    return (
      <div className='flex h-96 items-center justify-center'>
        <RefreshCw className='text-muted-foreground h-8 w-8 animate-spin' />
      </div>
    )
  }

  return (
    <div className='flex h-full flex-col'>
      {/* Header */}
      <div className='bg-background flex items-center justify-between border-b px-6 py-4'>
        <div className='flex items-center gap-3'>
          <div className='bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg'>
            <BookOpen className='text-primary h-5 w-5' />
          </div>
          <div>
            <h1 className='text-xl font-semibold'>Knowledge Bases</h1>
            <p className='text-muted-foreground text-sm'>
              Manage knowledge bases for RAG-powered AI chatbots
            </p>
          </div>
        </div>
      </div>

      <div className='flex-1 overflow-auto p-6'>
        <div className='flex flex-col gap-6'>
          <div className='flex items-center justify-between'>
            <div className='flex gap-4 text-sm'>
              <div className='flex items-center gap-1.5'>
                <span className='text-muted-foreground'>Total:</span>
                <Badge variant='secondary' className='h-5 px-2'>
                  {knowledgeBases.length}
                </Badge>
              </div>
              <div className='flex items-center gap-1.5'>
                <span className='text-muted-foreground'>Active:</span>
                <Badge
                  variant='secondary'
                  className='h-5 bg-green-500/10 px-2 text-green-600 dark:text-green-400'
                >
                  {knowledgeBases.filter((kb) => kb.enabled).length}
                </Badge>
              </div>
              <div className='flex items-center gap-1.5'>
                <span className='text-muted-foreground'>Documents:</span>
                <Badge variant='secondary' className='h-5 px-2'>
                  {knowledgeBases.reduce(
                    (sum, kb) => sum + kb.document_count,
                    0
                  )}
                </Badge>
              </div>
            </div>
            <div className='flex items-center gap-2'>
              <Button
                onClick={() => fetchKnowledgeBases()}
                variant='outline'
                size='sm'
              >
                <RefreshCw className='mr-2 h-4 w-4' />
                Refresh
              </Button>
              <Dialog
                open={createDialogOpen}
                onOpenChange={setCreateDialogOpen}
              >
                <DialogTrigger asChild>
                  <Button size='sm'>
                    <Plus className='mr-2 h-4 w-4' />
                    Create Knowledge Base
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Create Knowledge Base</DialogTitle>
                    <DialogDescription>
                      Create a new knowledge base to store documents for
                      RAG-powered AI chatbots.
                    </DialogDescription>
                  </DialogHeader>
                  <div className='grid max-h-[60vh] gap-4 overflow-y-auto py-4'>
                    <div className='grid gap-2'>
                      <Label htmlFor='name'>Name</Label>
                      <Input
                        id='name'
                        value={newKB.name}
                        onChange={(e) =>
                          setNewKB({ ...newKB, name: e.target.value })
                        }
                        placeholder='e.g., product-docs'
                      />
                    </div>
                    <div className='grid gap-2'>
                      <Label htmlFor='description'>Description</Label>
                      <Textarea
                        id='description'
                        value={newKB.description || ''}
                        onChange={(e) =>
                          setNewKB({ ...newKB, description: e.target.value })
                        }
                        placeholder='What kind of documents will this knowledge base contain?'
                      />
                    </div>
                    <div className='grid gap-2'>
                      <Label htmlFor='visibility'>Visibility</Label>
                      <select
                        id='visibility'
                        value={newKB.visibility}
                        onChange={(e) =>
                          setNewKB({
                            ...newKB,
                            visibility: e.target.value as KBVisibility,
                          })
                        }
                        className='border-input focus-visible:ring-ring flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:ring-1 focus-visible:outline-none'
                      >
                        <option value='private'>
                          Private - Only you can access
                        </option>
                        <option value='shared'>
                          Shared - Specific users you grant access
                        </option>
                        <option value='public'>
                          Public - All authenticated users can contribute
                        </option>
                      </select>
                      <p className='text-muted-foreground text-xs'>
                        {newKB.visibility === 'private' &&
                          'Only you will be able to access this knowledge base.'}
                        {newKB.visibility === 'shared' &&
                          'Grant access to specific users below.'}
                        {newKB.visibility === 'public' &&
                          'All authenticated users can add documents. Each user can only see their own documents unless explicitly shared.'}
                      </p>
                    </div>
                    <div className='grid gap-2'>
                      <Label htmlFor='default_user_permission'>
                        Default User Permission
                      </Label>
                      <select
                        id='default_user_permission'
                        value={newKB.default_user_permission}
                        onChange={(e) =>
                          setNewKB({
                            ...newKB,
                            default_user_permission: e.target
                              .value as KBPermission,
                          })
                        }
                        className='border-input focus-visible:ring-ring flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:ring-1 focus-visible:outline-none'
                      >
                        <option value='viewer'>
                          Viewer - Users can add and view their own documents
                        </option>
                        <option value='editor'>
                          Editor - Users can add and edit their own documents
                        </option>
                      </select>
                      <p className='text-muted-foreground text-xs'>
                        Default permission level for all authenticated users.
                        Users can always see their own documents plus documents
                        shared with them.
                      </p>
                    </div>
                    <div className='grid gap-2'>
                      <Label htmlFor='description'>Description</Label>
                      <Textarea
                        id='description'
                        value={newKB.description || ''}
                        onChange={(e) =>
                          setNewKB({ ...newKB, description: e.target.value })
                        }
                        placeholder='What kind of documents will this knowledge base contain?'
                      />
                    </div>
                    <div className='grid grid-cols-2 gap-4'>
                      <div className='grid gap-2'>
                        <Label htmlFor='chunk_size'>Chunk Size</Label>
                        <Input
                          id='chunk_size'
                          type='number'
                          value={newKB.chunk_size}
                          onChange={(e) =>
                            setNewKB({
                              ...newKB,
                              chunk_size: parseInt(e.target.value) || 512,
                            })
                          }
                        />
                        <p className='text-muted-foreground text-xs'>
                          Characters per chunk
                        </p>
                      </div>
                      <div className='grid gap-2'>
                        <Label htmlFor='chunk_overlap'>Chunk Overlap</Label>
                        <Input
                          id='chunk_overlap'
                          type='number'
                          value={newKB.chunk_overlap}
                          onChange={(e) =>
                            setNewKB({
                              ...newKB,
                              chunk_overlap: parseInt(e.target.value) || 50,
                            })
                          }
                        />
                        <p className='text-muted-foreground text-xs'>
                          Overlap between chunks
                        </p>
                      </div>
                    </div>

                    {/* Permissions Section */}
                    {newKB.visibility === 'shared' && (
                      <div className='grid gap-3 border-t pt-4'>
                        <Label className='flex items-center gap-2'>
                          <Users className='h-4 w-4' />
                          Share with Users
                        </Label>
                        <div className='flex gap-2'>
                          <select
                            value={newPermission.user_id}
                            onChange={(e) =>
                              setNewPermission({
                                ...newPermission,
                                user_id: e.target.value,
                              })
                            }
                            className='border-input focus-visible:ring-ring flex h-9 flex-1 rounded-md border bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:ring-1 focus-visible:outline-none'
                            disabled={usersLoading}
                          >
                            <option value=''>Select a user...</option>
                            {users.map((u) => (
                              <option key={u.id} value={u.id}>
                                {u.email || u.id}
                              </option>
                            ))}
                          </select>
                          <select
                            value={newPermission.permission}
                            onChange={(e) =>
                              setNewPermission({
                                ...newPermission,
                                permission: e.target.value as KBPermission,
                              })
                            }
                            className='border-input focus-visible:ring-ring flex h-9 w-32 rounded-md border bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:ring-1 focus-visible:outline-none'
                          >
                            <option value='viewer'>Viewer</option>
                            <option value='editor'>Editor</option>
                          </select>
                          <Button
                            onClick={addPermission}
                            size='sm'
                            variant='outline'
                          >
                            <Plus className='h-4 w-4' />
                          </Button>
                        </div>
                        {newKB.initial_permissions &&
                          newKB.initial_permissions.length > 0 && (
                            <div className='flex flex-wrap gap-2'>
                              {newKB.initial_permissions.map((perm) => {
                                const user = users.find(
                                  (u) => u.id === perm.user_id
                                )
                                return (
                                  <Badge
                                    key={perm.user_id}
                                    variant='secondary'
                                    className='gap-1 pr-1'
                                  >
                                    <span>{user?.email || perm.user_id}</span>
                                    <span className='text-muted-foreground'>
                                      ({perm.permission})
                                    </span>
                                    <button
                                      onClick={() =>
                                        removePermission(perm.user_id)
                                      }
                                      className='hover:text-destructive ml-1'
                                    >
                                      <X className='h-3 w-3' />
                                    </button>
                                  </Badge>
                                )
                              })}
                            </div>
                          )}
                      </div>
                    )}
                  </div>
                  <DialogFooter>
                    <Button
                      variant='outline'
                      onClick={() => setCreateDialogOpen(false)}
                    >
                      Cancel
                    </Button>
                    <Button onClick={handleCreate}>Create</Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </div>
          </div>

          <ScrollArea className='h-[calc(100vh-16rem)]'>
            {knowledgeBases.length === 0 ? (
              <Card>
                <CardContent className='p-12 text-center'>
                  <BookOpen className='text-muted-foreground mx-auto mb-4 h-12 w-12' />
                  <p className='mb-2 text-lg font-medium'>
                    No knowledge bases yet
                  </p>
                  <p className='text-muted-foreground mb-4 text-sm'>
                    Create a knowledge base to store documents for RAG-powered
                    AI chatbots
                  </p>
                  <Button onClick={() => setCreateDialogOpen(true)}>
                    <Plus className='mr-2 h-4 w-4' />
                    Create Knowledge Base
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-3'>
                {knowledgeBases.map((kb) => (
                  <Card key={kb.id} className='relative'>
                    <CardHeader className='pb-2'>
                      <div className='flex items-start justify-between'>
                        <div className='flex items-center gap-2'>
                          <BookOpen className='h-5 w-5' />
                          <CardTitle className='text-lg'>{kb.name}</CardTitle>
                        </div>
                        <div className='flex items-center gap-1'>
                          <Switch
                            checked={kb.enabled}
                            onCheckedChange={() => toggleEnabled(kb)}
                          />
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant='ghost'
                                size='sm'
                                className='h-8 w-8 p-0'
                                onClick={() =>
                                  navigate({
                                    to: `/knowledge-bases/$id`,
                                    params: { id: kb.id },
                                  })
                                }
                              >
                                <FileText className='h-4 w-4' />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>View Documents</TooltipContent>
                          </Tooltip>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant='ghost'
                                size='sm'
                                className='h-8 w-8 p-0'
                                onClick={() =>
                                  navigate({
                                    to: `/knowledge-bases/$id/search`,
                                    params: { id: kb.id },
                                  })
                                }
                              >
                                <Search className='h-4 w-4' />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Search</TooltipContent>
                          </Tooltip>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant='ghost'
                                size='sm'
                                className='h-8 w-8 p-0'
                                onClick={() =>
                                  navigate({
                                    to: `/knowledge-bases/$id/settings`,
                                    params: { id: kb.id },
                                  })
                                }
                              >
                                <Settings className='h-4 w-4' />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Settings</TooltipContent>
                          </Tooltip>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant='ghost'
                                size='sm'
                                className='text-destructive hover:text-destructive h-8 w-8 p-0'
                                onClick={() => setDeleteConfirm(kb.id)}
                              >
                                <Trash2 className='h-4 w-4' />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Delete</TooltipContent>
                          </Tooltip>
                        </div>
                      </div>
                      {kb.namespace !== 'default' && (
                        <Badge variant='outline' className='w-fit text-[10px]'>
                          {kb.namespace}
                        </Badge>
                      )}
                      <Badge
                        variant='outline'
                        className={`flex w-fit items-center gap-1 text-[10px] ${
                          kb.visibility === 'private'
                            ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400'
                            : kb.visibility === 'public'
                              ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400'
                              : 'bg-purple-500/10 text-purple-600 dark:text-purple-400'
                        }`}
                      >
                        {kb.visibility === 'private' && (
                          <Lock className='h-3 w-3' />
                        )}
                        {kb.visibility === 'public' && (
                          <Globe className='h-3 w-3' />
                        )}
                        {kb.visibility === 'shared' && (
                          <Users className='h-3 w-3' />
                        )}
                        {kb.visibility || 'private'}
                      </Badge>
                    </CardHeader>
                    <CardContent>
                      {kb.description && (
                        <CardDescription className='mb-3 line-clamp-2'>
                          {kb.description}
                        </CardDescription>
                      )}
                      <div className='flex flex-wrap gap-2 text-xs'>
                        <Badge variant='secondary'>
                          {kb.document_count}{' '}
                          {kb.document_count === 1 ? 'document' : 'documents'}
                        </Badge>
                        <Badge variant='secondary'>
                          {kb.total_chunks} chunks
                        </Badge>
                        {kb.embedding_model && (
                          <Badge variant='outline' className='text-[10px]'>
                            {kb.embedding_model}
                          </Badge>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </ScrollArea>

          {/* Delete Confirmation Dialog */}
          <AlertDialog
            open={deleteConfirm !== null}
            onOpenChange={(open) => !open && setDeleteConfirm(null)}
          >
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete Knowledge Base</AlertDialogTitle>
                <AlertDialogDescription>
                  Are you sure you want to delete this knowledge base? This will
                  permanently delete all documents and chunks. This action
                  cannot be undone.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={() => deleteConfirm && handleDelete(deleteConfirm)}
                  className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
                >
                  Delete
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>
    </div>
  )
}
