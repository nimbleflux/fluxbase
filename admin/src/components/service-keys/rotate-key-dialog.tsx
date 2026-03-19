import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, Loader2, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { platformServiceKeysApi, type PlatformServiceKey } from '@/lib/api'

interface RotateKeyDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  serviceKey: PlatformServiceKey | null
  onSuccess?: (newKey: PlatformServiceKey & { key?: string; grace_period_ends_at: string }) => void
}

const GRACE_PERIOD_OPTIONS = [
  { value: '1', label: '1 hour' },
  { value: '6', label: '6 hours' },
  { value: '24', label: '24 hours (default)' },
  { value: '72', label: '3 days' },
  { value: '168', label: '1 week' },
]

export function RotateKeyDialog({
  open,
  onOpenChange,
  serviceKey,
  onSuccess,
}: RotateKeyDialogProps) {
  const queryClient = useQueryClient()
  const [gracePeriodHours, setGracePeriodHours] = useState('24')
  const [newKeyName, setNewKeyName] = useState('')

  const rotateMutation = useMutation({
    mutationFn: () => {
      if (!serviceKey) throw new Error('No service key selected')
      return platformServiceKeysApi.rotate(serviceKey.id, {
        grace_period_hours: parseInt(gracePeriodHours, 10),
        new_key_name: newKeyName.trim() || undefined,
      })
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['platform-service-keys'] })
      toast.success('Service key rotated successfully')
      setGracePeriodHours('24')
      setNewKeyName('')
      onSuccess?.(data)
      onOpenChange(false)
    },
    onError: (error: Error) => {
      toast.error(`Failed to rotate service key: ${error.message}`)
    },
  })

  const handleRotate = () => {
    rotateMutation.mutate()
  }

  if (!serviceKey) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <RefreshCw className="h-5 w-5" />
            Rotate Service Key
          </DialogTitle>
          <DialogDescription>
            Create a new key to replace "{serviceKey.name}". The old key will remain
            active during the grace period.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="rounded-md bg-yellow-50 p-4 dark:bg-yellow-950">
            <div className="flex">
              <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
              <div className="ml-3">
                <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                  Important: Key Rotation
                </h3>
                <div className="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
                  <p>
                    A new key will be generated. After the grace period ends, the old
                    key will stop working. Make sure to update your applications before
                    then.
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="currentKey">Current Key</Label>
            <Input
              id="currentKey"
              value={serviceKey.key_prefix + '...'}
              readOnly
              className="bg-muted"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="gracePeriod">Grace Period</Label>
            <Select value={gracePeriodHours} onValueChange={setGracePeriodHours}>
              <SelectTrigger>
                <SelectValue placeholder="Select grace period" />
              </SelectTrigger>
              <SelectContent>
                {GRACE_PERIOD_OPTIONS.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-muted-foreground text-xs">
              The old key will work during this period, giving you time to update your
              applications.
            </p>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="newKeyName">New Key Name (optional)</Label>
            <Input
              id="newKeyName"
              placeholder={`${serviceKey.name} (rotated)`}
              value={newKeyName}
              onChange={(e) => setNewKeyName(e.target.value)}
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button onClick={handleRotate} disabled={rotateMutation.isPending}>
            {rotateMutation.isPending && (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            )}
            Rotate Key
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
