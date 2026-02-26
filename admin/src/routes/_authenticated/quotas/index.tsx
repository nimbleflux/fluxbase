import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import api from '@/lib/api'
import { Alert, AlertDescription } from '@/components/ui/alert'
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
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface UserQuota {
  user_id: string
  max_documents: number
  max_chunks: number
  max_storage_bytes: number
  used_documents: number
  used_chunks: number
  used_storage_bytes: number
}

interface UserWithQuota {
  id: string
  email: string
  full_name: string | null
  quota: UserQuota | null
}

function formatBytes(bytes: number): string {
  if (bytes >= 1024 * 1024 * 1024) {
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
  }
  if (bytes >= 1024 * 1024) {
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }
  if (bytes >= 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`
  }
  return `${bytes} bytes`
}

function calculatePercentage(used: number, limit: number): number {
  if (limit === 0) return 0
  return Math.min(100, (used / limit) * 100)
}

export const Route = createFileRoute('/_authenticated/quotas/')({
  component: UserQuotasPage,
})

function UserQuotasPage() {
  const queryClient = useQueryClient()
  const [editingUser, setEditingUser] = useState<UserWithQuota | null>(null)
  const [editedQuota, setEditedQuota] = useState({
    maxDocuments: 0,
    maxChunks: 0,
    maxStorageMB: 0,
  })

  // Fetch users with their quotas
  const { data: users, isLoading } = useQuery({
    queryKey: ['users-with-quotas'],
    queryFn: async () => {
      const response = await api.get<UserWithQuota[]>(
        '/api/v1/admin/users-with-quotas'
      )
      return response.data
    },
  })

  const updateQuotaMutation = useMutation({
    mutationFn: async ({
      userId,
      quota,
    }: {
      userId: string
      quota: typeof editedQuota
    }) => {
      const response = await api.put(`/api/v1/admin/users/${userId}/quota`, {
        max_documents: quota.maxDocuments,
        max_chunks: quota.maxChunks,
        max_storage_bytes: quota.maxStorageMB * 1024 * 1024, // Convert MB to bytes
      })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users-with-quotas'] })
      setEditingUser(null)
    },
  })

  const handleEditQuota = (user: UserWithQuota) => {
    setEditingUser(user)
    if (user.quota) {
      setEditedQuota({
        maxDocuments: user.quota.max_documents,
        maxChunks: user.quota.max_chunks,
        maxStorageMB: user.quota.max_storage_bytes / (1024 * 1024), // Convert bytes to MB
      })
    } else {
      // Use system defaults
      setEditedQuota({
        maxDocuments: 10000,
        maxChunks: 500000,
        maxStorageMB: 10240, // 10GB
      })
    }
  }

  const handleSaveQuota = () => {
    if (!editingUser) return
    updateQuotaMutation.mutate({
      userId: editingUser.id,
      quota: editedQuota,
    })
  }

  return (
    <div className='space-y-6'>
      <div>
        <h1 className='text-3xl font-bold tracking-tight'>User Quotas</h1>
        <p className='text-muted-foreground'>
          Manage resource quotas for users. Limits apply across all knowledge
          bases.
        </p>
      </div>

      <Alert>
        <AlertDescription>
          <strong>System Defaults:</strong> 10,000 documents, 500,000 chunks, 10
          GB storage per user. Customize limits per user below.
        </AlertDescription>
      </Alert>

      {isLoading ? (
        <div className='flex justify-center py-8'>
          <div className='text-muted-foreground'>Loading users...</div>
        </div>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>User Quotas</CardTitle>
            <CardDescription>
              View and manage resource limits for each user
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Documents</TableHead>
                  <TableHead>Chunks</TableHead>
                  <TableHead>Storage</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users?.map((user: UserWithQuota) => {
                  const quota = user.quota || {
                    max_documents: 10000,
                    max_chunks: 500000,
                    max_storage_bytes: 10 * 1024 * 1024 * 1024,
                    used_documents: 0,
                    used_chunks: 0,
                    used_storage_bytes: 0,
                  }

                  const docsPercent = calculatePercentage(
                    quota.used_documents,
                    quota.max_documents
                  )
                  const chunksPercent = calculatePercentage(
                    quota.used_chunks,
                    quota.max_chunks
                  )
                  const storagePercent = calculatePercentage(
                    quota.used_storage_bytes,
                    quota.max_storage_bytes
                  )

                  return (
                    <TableRow key={user.id}>
                      <TableCell>
                        <div>
                          <div className='font-medium'>
                            {user.full_name || user.email}
                          </div>
                          <div className='text-muted-foreground text-sm'>
                            {user.email}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className='space-y-1'>
                          <div className='flex items-center gap-2 text-sm'>
                            <span>{quota.used_documents.toLocaleString()}</span>
                            <span className='text-muted-foreground'>
                              / {quota.max_documents.toLocaleString()}
                            </span>
                          </div>
                          <Progress value={docsPercent} className='h-2' />
                          <div className='text-muted-foreground text-xs'>
                            {docsPercent.toFixed(1)}%
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className='space-y-1'>
                          <div className='flex items-center gap-2 text-sm'>
                            <span>{quota.used_chunks.toLocaleString()}</span>
                            <span className='text-muted-foreground'>
                              / {quota.max_chunks.toLocaleString()}
                            </span>
                          </div>
                          <Progress value={chunksPercent} className='h-2' />
                          <div className='text-muted-foreground text-xs'>
                            {chunksPercent.toFixed(1)}%
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className='space-y-1'>
                          <div className='flex items-center gap-2 text-sm'>
                            <span>{formatBytes(quota.used_storage_bytes)}</span>
                            <span className='text-muted-foreground'>
                              / {formatBytes(quota.max_storage_bytes)}
                            </span>
                          </div>
                          <Progress value={storagePercent} className='h-2' />
                          <div className='text-muted-foreground text-xs'>
                            {storagePercent.toFixed(1)}%
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            storagePercent >= 90
                              ? 'destructive'
                              : storagePercent >= 75
                                ? 'secondary'
                                : 'default'
                          }
                        >
                          {storagePercent >= 90
                            ? 'Near Limit'
                            : storagePercent >= 75
                              ? 'Warning'
                              : 'OK'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => handleEditQuota(user)}
                        >
                          Edit Quota
                        </Button>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* Edit Quota Dialog */}
      <Dialog
        open={!!editingUser}
        onOpenChange={(open) => !open && setEditingUser(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit User Quota</DialogTitle>
            <DialogDescription>
              Set custom resource limits for {editingUser?.email}
            </DialogDescription>
          </DialogHeader>

          <div className='space-y-4 py-4'>
            <div className='space-y-2'>
              <Label htmlFor='quota-max-documents'>Max Documents</Label>
              <Input
                id='quota-max-documents'
                type='number'
                min={1}
                max={1000000}
                value={editedQuota.maxDocuments}
                onChange={(e) =>
                  setEditedQuota({
                    ...editedQuota,
                    maxDocuments: parseInt(e.target.value) || 0,
                  })
                }
              />
              <p className='text-muted-foreground text-xs'>
                System default: 10,000 documents
              </p>
            </div>

            <div className='space-y-2'>
              <Label htmlFor='quota-max-chunks'>Max Chunks</Label>
              <Input
                id='quota-max-chunks'
                type='number'
                min={1}
                max={10000000}
                value={editedQuota.maxChunks}
                onChange={(e) =>
                  setEditedQuota({
                    ...editedQuota,
                    maxChunks: parseInt(e.target.value) || 0,
                  })
                }
              />
              <p className='text-muted-foreground text-xs'>
                System default: 500,000 chunks
              </p>
            </div>

            <div className='space-y-2'>
              <Label htmlFor='quota-max-storage'>Max Storage (MB)</Label>
              <Input
                id='quota-max-storage'
                type='number'
                min={1}
                max={1024000} // 1TB max
                value={editedQuota.maxStorageMB}
                onChange={(e) =>
                  setEditedQuota({
                    ...editedQuota,
                    maxStorageMB: parseInt(e.target.value) || 0,
                  })
                }
              />
              <p className='text-muted-foreground text-xs'>
                System default: 10,240 MB (
                {formatBytes(editedQuota.maxStorageMB * 1024 * 1024)})
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button variant='outline' onClick={() => setEditingUser(null)}>
              Cancel
            </Button>
            <Button
              onClick={handleSaveQuota}
              disabled={updateQuotaMutation.isPending}
            >
              {updateQuotaMutation.isPending ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
