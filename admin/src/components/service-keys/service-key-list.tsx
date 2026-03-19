import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  KeyRound,
  Plus,
  Trash2,
  Power,
  PowerOff,
  RefreshCw,
  Search,
  MoreHorizontal,
  Copy,
  ShieldAlert,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  platformServiceKeysApi,
  type PlatformServiceKey,
  type PlatformServiceKeyWithPlaintext,
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
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ConfigManagedBadge } from './config-managed-badge'
import { CreateKeyDialog } from './create-key-dialog'
import { RotateKeyDialog } from './rotate-key-dialog'

const KEY_TYPE_CONFIG: Record<string, { label: string; variant: 'default' | 'secondary' | 'outline' | 'destructive' }> = {
  anon: { label: 'Anonymous', variant: 'secondary' },
  publishable: { label: 'Publishable', variant: 'default' },
  tenant_service: { label: 'Tenant Service', variant: 'outline' },
  global_service: { label: 'Global Service', variant: 'default' },
}

interface ServiceKeyListProps {
  onCreateSuccess?: (key: PlatformServiceKeyWithPlaintext) => void
}

export function ServiceKeyList({ onCreateSuccess }: ServiceKeyListProps) {
  const queryClient = useQueryClient()
  const [searchQuery, setSearchQuery] = useState('')
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showRotateDialog, setShowRotateDialog] = useState(false)
  const [showKeyDialog, setShowKeyDialog] = useState(false)
  const [selectedKey, setSelectedKey] = useState<PlatformServiceKey | null>(null)
  const [createdKey, setCreatedKey] = useState<PlatformServiceKeyWithPlaintext | null>(null)

  const { data: serviceKeys, isLoading } = useQuery<PlatformServiceKey[]>({
    queryKey: ['platform-service-keys'],
    queryFn: platformServiceKeysApi.list,
  })

  const enableMutation = useMutation({
    mutationFn: platformServiceKeysApi.enable,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['platform-service-keys'] })
      toast.success('Service key enabled')
    },
    onError: () => toast.error('Failed to enable service key'),
  })

  const disableMutation = useMutation({
    mutationFn: platformServiceKeysApi.disable,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['platform-service-keys'] })
      toast.success('Service key disabled')
    },
    onError: () => toast.error('Failed to disable service key'),
  })

  const deleteMutation = useMutation({
    mutationFn: platformServiceKeysApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['platform-service-keys'] })
      toast.success('Service key deleted')
    },
    onError: () => toast.error('Failed to delete service key'),
  })

  const handleCreateSuccess = (key: PlatformServiceKeyWithPlaintext) => {
    setCreatedKey(key)
    setShowKeyDialog(true)
    onCreateSuccess?.(key)
  }

  const handleRotateSuccess = () => {
    setSelectedKey(null)
  }

  const openRotateDialog = (key: PlatformServiceKey) => {
    setSelectedKey(key)
    setShowRotateDialog(true)
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.success('Copied to clipboard')
  }

  const isExpired = (expiresAt?: string) => {
    if (!expiresAt) return false
    return new Date(expiresAt) < new Date()
  }

  const getKeyStatus = (key: PlatformServiceKey) => {
    if (key.revoked_at) return { label: 'Revoked', variant: 'destructive' as const }
    if (key.deprecated_at) {
      if (key.grace_period_ends_at && new Date(key.grace_period_ends_at) > new Date()) {
        return { label: 'Deprecated', variant: 'outline' as const }
      }
      return { label: 'Expired', variant: 'destructive' as const }
    }
    if (!key.is_active) return { label: 'Disabled', variant: 'secondary' as const }
    if (isExpired(key.expires_at)) return { label: 'Expired', variant: 'destructive' as const }
    return { label: 'Active', variant: 'default' as const }
  }

  const canModify = (key: PlatformServiceKey) => !key.revoked_at && !key.is_config_managed

  const filteredKeys = serviceKeys?.filter(
    (key) =>
      key.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      key.description?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      key.key_prefix.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by name, description, or key prefix..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8"
          />
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Key
        </Button>
      </div>

      {isLoading ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Key Type</TableHead>
              <TableHead>Tenant</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array(3)
              .fill(0)
              .map((_, i) => (
                <TableRow key={i}>
                  <TableCell><Skeleton className="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton className="h-5 w-20" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                  <TableCell><Skeleton className="h-5 w-16" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton className="h-8 w-8 ml-auto" /></TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
      ) : filteredKeys && filteredKeys.length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Key Type</TableHead>
              <TableHead>Tenant</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredKeys.map((key) => {
              const status = getKeyStatus(key)
              const keyTypeConfig = KEY_TYPE_CONFIG[key.key_type] || { label: key.key_type, variant: 'outline' as const }

              return (
                <TableRow key={key.id}>
                  <TableCell>
                    <div className="flex flex-col gap-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{key.name}</span>
                        <ConfigManagedBadge isConfigManaged={key.is_config_managed} />
                      </div>
                      {key.description && (
                        <span className="text-muted-foreground text-xs">{key.description}</span>
                      )}
                      <code className="text-xs text-muted-foreground">{key.key_prefix}...</code>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant={keyTypeConfig.variant}>{keyTypeConfig.label}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">
                    {key.tenant_id || '—'}
                  </TableCell>
                  <TableCell>
                    <Badge variant={status.variant}>{status.label}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">
                    {formatDistanceToNow(new Date(key.created_at), { addSuffix: true })}
                  </TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        {canModify(key) && key.is_active && !key.deprecated_at && (
                          <>
                            <DropdownMenuItem onClick={() => openRotateDialog(key)}>
                              <RefreshCw className="mr-2 h-4 w-4" />
                              Rotate Key
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                          </>
                        )}
                        {canModify(key) && !key.deprecated_at && (
                          key.is_active ? (
                            <DropdownMenuItem
                              onClick={() => disableMutation.mutate(key.id)}
                              disabled={disableMutation.isPending}
                            >
                              <PowerOff className="mr-2 h-4 w-4" />
                              Disable
                            </DropdownMenuItem>
                          ) : (
                            <DropdownMenuItem
                              onClick={() => enableMutation.mutate(key.id)}
                              disabled={enableMutation.isPending}
                            >
                              <Power className="mr-2 h-4 w-4" />
                              Enable
                            </DropdownMenuItem>
                          )
                        )}
                        {canModify(key) && <DropdownMenuSeparator />}
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <DropdownMenuItem
                              className="text-destructive focus:text-destructive"
                              onSelect={(e) => e.preventDefault()}
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete
                            </DropdownMenuItem>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Delete Service Key</AlertDialogTitle>
                              <AlertDialogDescription>
                                Are you sure you want to delete "{key.name}"? This action
                                cannot be undone and any applications using this key will
                                lose access immediately.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction
                                onClick={() => deleteMutation.mutate(key.id)}
                                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                              >
                                Delete
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <KeyRound className="mb-4 h-12 w-12 text-muted-foreground" />
          <p className="text-muted-foreground">
            {searchQuery ? 'No service keys match your search' : 'No service keys yet'}
          </p>
          {!searchQuery && (
            <Button onClick={() => setShowCreateDialog(true)} variant="outline" className="mt-4">
              Create Your First Service Key
            </Button>
          )}
        </div>
      )}

      <CreateKeyDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        onSuccess={handleCreateSuccess}
      />

      <RotateKeyDialog
        open={showRotateDialog}
        onOpenChange={setShowRotateDialog}
        serviceKey={selectedKey}
        onSuccess={handleRotateSuccess}
      />

      {createdKey && createdKey.key && (
        <AlertDialog open={showKeyDialog} onOpenChange={setShowKeyDialog}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Service Key Created</AlertDialogTitle>
              <AlertDialogDescription>
                Copy this key now. You won't be able to see it again!
              </AlertDialogDescription>
            </AlertDialogHeader>
            <div className="space-y-4 py-4">
              <div className="rounded-md bg-yellow-50 p-4 dark:bg-yellow-950">
                <div className="flex">
                  <ShieldAlert className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                      Important: Save this key
                    </h3>
                    <p className="mt-1 text-sm text-yellow-700 dark:text-yellow-300">
                      This is the only time you'll see the full key. Store it securely.
                    </p>
                  </div>
                </div>
              </div>
              <div className="flex gap-2">
                <Input
                  value={createdKey.key}
                  readOnly
                  className="font-mono text-xs"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => copyToClipboard(createdKey.key!)}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <AlertDialogFooter>
              <AlertDialogAction>I've Saved the Key</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </div>
  )
}
