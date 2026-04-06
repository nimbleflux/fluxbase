import { useState, useEffect, useCallback } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { RefreshCw, Save, Trash2, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { knowledgeBasesApi, type KnowledgeBase } from '@/lib/api'
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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { KnowledgeBaseHeader } from '@/components/knowledge-bases/knowledge-base-header'
import { useTenantStore } from '@/stores/tenant-store'

export const Route = createFileRoute(
  '/_authenticated/knowledge-bases/$id/settings'
)({
  component: KnowledgeBaseSettingsPage,
})

function KnowledgeBaseSettingsPage() {
  const { id } = Route.useParams()
  const navigate = useNavigate()
  const currentTenantId = useTenantStore((state) => state.currentTenant?.id)
  const [knowledgeBase, setKnowledgeBase] = useState<KnowledgeBase | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    enabled: true,
  })

  const fetchKnowledgeBase = useCallback(async () => {
    setLoading(true)
    try {
      const kb = await knowledgeBasesApi.get(id)
      setKnowledgeBase(kb)
      setFormData({
        name: kb.name,
        description: kb.description || '',
        enabled: kb.enabled,
      })
    } catch {
      toast.error('Failed to fetch knowledge base')
    } finally {
      setLoading(false)
    }
  }, [id])

  const handleSave = async () => {
    if (!formData.name.trim()) {
      toast.error('Name is required')
      return
    }

    setSaving(true)
    try {
      await knowledgeBasesApi.update(id, {
        name: formData.name,
        description: formData.description || undefined,
        enabled: formData.enabled,
      })
      toast.success('Settings saved')
      await fetchKnowledgeBase()
    } catch {
      toast.error('Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    try {
      await knowledgeBasesApi.delete(id)
      toast.success('Knowledge base deleted')
      navigate({ to: '/knowledge-bases' })
    } catch {
      toast.error('Failed to delete knowledge base')
    } finally {
      setDeleteConfirm(false)
    }
  }

  useEffect(() => {
    fetchKnowledgeBase()
  }, [fetchKnowledgeBase, currentTenantId])

  if (loading) {
    return (
      <div className='flex h-96 items-center justify-center'>
        <RefreshCw className='text-muted-foreground h-8 w-8 animate-spin' />
      </div>
    )
  }

  if (!knowledgeBase) {
    return (
      <div className='flex h-96 flex-col items-center justify-center gap-4'>
        <p className='text-muted-foreground'>Knowledge base not found</p>
      </div>
    )
  }

  return (
    <div className='flex flex-1 flex-col gap-6 p-6'>
      <KnowledgeBaseHeader knowledgeBase={knowledgeBase} activeTab='settings' />

      <div className='grid gap-6'>
        <Card>
          <CardHeader>
            <CardTitle>General Settings</CardTitle>
            <CardDescription>
              Update the name and description of this knowledge base
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className='grid gap-4'>
              <div className='grid gap-2'>
                <Label htmlFor='name'>Name</Label>
                <Input
                  id='name'
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                />
              </div>
              <div className='grid gap-2'>
                <Label htmlFor='description'>Description</Label>
                <Textarea
                  id='description'
                  value={formData.description}
                  onChange={(e) =>
                    setFormData({ ...formData, description: e.target.value })
                  }
                  placeholder='Describe what this knowledge base contains...'
                />
              </div>
              <div className='flex items-center justify-between'>
                <div className='space-y-0.5'>
                  <Label htmlFor='enabled'>Enabled</Label>
                  <p className='text-muted-foreground text-sm'>
                    Enable or disable this knowledge base for search
                  </p>
                </div>
                <Switch
                  id='enabled'
                  checked={formData.enabled}
                  onCheckedChange={(checked) =>
                    setFormData({ ...formData, enabled: checked })
                  }
                />
              </div>
              <div className='flex justify-end'>
                <Button onClick={handleSave} disabled={saving}>
                  {saving ? (
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  ) : (
                    <Save className='mr-2 h-4 w-4' />
                  )}
                  Save Changes
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Configuration (Read-only)</CardTitle>
            <CardDescription>
              These settings were configured when the knowledge base was created
              and cannot be changed
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className='grid gap-4 md:grid-cols-2'>
              <div className='space-y-1'>
                <Label className='text-muted-foreground text-sm'>
                  Embedding Model
                </Label>
                <div>
                  <Badge variant='outline'>
                    {knowledgeBase.embedding_model || 'Default (not set)'}
                  </Badge>
                </div>
              </div>
              <div className='space-y-1'>
                <Label className='text-muted-foreground text-sm'>
                  Embedding Dimensions
                </Label>
                <div>
                  <Badge variant='secondary'>
                    {knowledgeBase.embedding_dimensions}
                  </Badge>
                </div>
              </div>
              <div className='space-y-1'>
                <Label className='text-muted-foreground text-sm'>
                  Chunk Size
                </Label>
                <div>
                  <Badge variant='secondary'>
                    {knowledgeBase.chunk_size} characters
                  </Badge>
                </div>
              </div>
              <div className='space-y-1'>
                <Label className='text-muted-foreground text-sm'>
                  Chunk Overlap
                </Label>
                <div>
                  <Badge variant='secondary'>
                    {knowledgeBase.chunk_overlap} characters
                  </Badge>
                </div>
              </div>
              <div className='space-y-1'>
                <Label className='text-muted-foreground text-sm'>
                  Chunk Strategy
                </Label>
                <div>
                  <Badge variant='secondary'>
                    {knowledgeBase.chunk_strategy}
                  </Badge>
                </div>
              </div>
              <div className='space-y-1'>
                <Label className='text-muted-foreground text-sm'>Created</Label>
                <div>
                  <Badge variant='secondary'>
                    {new Date(knowledgeBase.created_at).toLocaleString()}
                  </Badge>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className='border-destructive'>
          <CardHeader>
            <CardTitle className='text-destructive'>Danger Zone</CardTitle>
            <CardDescription>
              Permanently delete this knowledge base and all its documents
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              variant='destructive'
              onClick={() => setDeleteConfirm(true)}
            >
              <Trash2 className='mr-2 h-4 w-4' />
              Delete Knowledge Base
            </Button>
          </CardContent>
        </Card>
      </div>

      <AlertDialog open={deleteConfirm} onOpenChange={setDeleteConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Knowledge Base</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{knowledgeBase.name}"? This will
              permanently delete all documents and chunks. This action cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
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
