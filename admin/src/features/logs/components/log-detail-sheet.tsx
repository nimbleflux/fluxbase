import { Copy, ExternalLink } from 'lucide-react'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { LOG_CATEGORY_CONFIG } from '../constants'
import type { LogEntry, LogCategory } from '../types'
import { LogLevelBadge } from './log-level-badge'

interface LogDetailSheetProps {
  log: LogEntry | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function LogDetailSheet({
  log,
  open,
  onOpenChange,
}: LogDetailSheetProps) {
  if (!log) return null

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} copied to clipboard`)
  }

  const copyAllDetails = () => {
    const details = JSON.stringify(log, null, 2)
    copyToClipboard(details, 'Log details')
  }

  const categoryConfig = LOG_CATEGORY_CONFIG[log.category as LogCategory]
  const CategoryIcon = categoryConfig?.icon

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className='flex w-[500px] flex-col sm:max-w-[500px]'>
        <SheetHeader>
          <SheetTitle className='flex items-center gap-2'>
            Log Details
            <Button variant='ghost' size='sm' onClick={copyAllDetails}>
              <Copy className='h-3 w-3' />
            </Button>
          </SheetTitle>
        </SheetHeader>

        <ScrollArea className='mt-4 min-h-0 flex-1'>
          <div className='space-y-4 px-4'>
            {/* Core Info */}
            <div className='grid grid-cols-2 gap-4'>
              <div>
                <label className='text-muted-foreground text-xs'>Level</label>
                <div className='mt-1'>
                  <LogLevelBadge level={log.level} />
                </div>
              </div>
              <div>
                <label className='text-muted-foreground text-xs'>
                  Category
                </label>
                <div className='mt-1 flex items-center gap-1.5'>
                  {CategoryIcon && (
                    <CategoryIcon className='text-muted-foreground h-3.5 w-3.5' />
                  )}
                  <span className='text-sm font-medium capitalize'>
                    {log.category}
                  </span>
                  {log.custom_category && (
                    <Badge variant='outline' className='ml-1 text-xs'>
                      {log.custom_category}
                    </Badge>
                  )}
                </div>
              </div>
              <div>
                <label className='text-muted-foreground text-xs'>
                  Timestamp
                </label>
                <p className='font-mono text-sm'>
                  {new Date(log.timestamp).toLocaleString()}
                </p>
              </div>
              <div>
                <label className='text-muted-foreground text-xs'>
                  Component
                </label>
                <p className='font-mono text-sm'>{log.component || '-'}</p>
              </div>
            </div>

            {/* Message */}
            <div>
              <label className='text-muted-foreground text-xs'>Message</label>
              <div className='bg-muted mt-1 rounded-md p-3'>
                <pre className='font-mono text-sm break-words whitespace-pre-wrap'>
                  {log.message}
                </pre>
              </div>
            </div>

            {/* Correlation IDs */}
            {(log.request_id || log.trace_id) && (
              <div className='space-y-3'>
                <label className='text-muted-foreground text-xs'>
                  Correlation IDs
                </label>
                <div className='grid grid-cols-1 gap-2'>
                  {log.request_id && (
                    <div className='bg-muted/50 flex items-center justify-between rounded p-2'>
                      <div>
                        <span className='text-muted-foreground text-xs'>
                          Request ID
                        </span>
                        <p className='max-w-[300px] truncate font-mono text-xs'>
                          {log.request_id}
                        </p>
                      </div>
                      <Button
                        variant='ghost'
                        size='sm'
                        onClick={() =>
                          copyToClipboard(log.request_id!, 'Request ID')
                        }
                      >
                        <Copy className='h-3 w-3' />
                      </Button>
                    </div>
                  )}
                  {log.trace_id && (
                    <div className='bg-muted/50 flex items-center justify-between rounded p-2'>
                      <div>
                        <span className='text-muted-foreground text-xs'>
                          Trace ID
                        </span>
                        <p className='max-w-[300px] truncate font-mono text-xs'>
                          {log.trace_id}
                        </p>
                      </div>
                      <Button
                        variant='ghost'
                        size='sm'
                        onClick={() =>
                          copyToClipboard(log.trace_id!, 'Trace ID')
                        }
                      >
                        <Copy className='h-3 w-3' />
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* User Info */}
            {(log.user_id || log.ip_address) && (
              <div className='grid grid-cols-2 gap-4'>
                {log.user_id && (
                  <div>
                    <label className='text-muted-foreground text-xs'>
                      User ID
                    </label>
                    <p className='truncate font-mono text-sm'>{log.user_id}</p>
                  </div>
                )}
                {log.ip_address && (
                  <div>
                    <label className='text-muted-foreground text-xs'>
                      IP Address
                    </label>
                    <p className='font-mono text-sm'>{log.ip_address}</p>
                  </div>
                )}
              </div>
            )}

            {/* Execution Info */}
            {log.execution_id && (
              <div className='bg-muted/50 rounded-md border p-3'>
                <label className='text-muted-foreground text-xs'>
                  Execution
                </label>
                <div className='mt-1 flex items-center gap-2'>
                  <Badge variant='outline'>{log.execution_type}</Badge>
                  <span className='flex-1 truncate font-mono text-sm'>
                    {log.execution_id}
                  </span>
                  {log.line_number !== undefined && (
                    <Badge variant='secondary'>Line {log.line_number}</Badge>
                  )}
                  <Button variant='ghost' size='sm' asChild>
                    <a href={`/${log.execution_type}s`}>
                      <ExternalLink className='h-3 w-3' />
                    </a>
                  </Button>
                </div>
              </div>
            )}

            {/* Additional Fields */}
            {log.fields && Object.keys(log.fields).length > 0 && (
              <div>
                <div className='mb-1 flex items-center justify-between'>
                  <label className='text-muted-foreground text-xs'>
                    Additional Fields
                  </label>
                  <Button
                    variant='ghost'
                    size='sm'
                    onClick={() =>
                      copyToClipboard(
                        JSON.stringify(log.fields, null, 2),
                        'Fields'
                      )
                    }
                  >
                    <Copy className='h-3 w-3' />
                  </Button>
                </div>
                <div className='bg-muted rounded-md p-3'>
                  <pre className='overflow-auto font-mono text-xs'>
                    {JSON.stringify(log.fields, null, 2)}
                  </pre>
                </div>
              </div>
            )}

            {/* Log ID */}
            <div className='border-t pt-4'>
              <div className='flex items-center justify-between'>
                <div>
                  <label className='text-muted-foreground text-xs'>
                    Log ID
                  </label>
                  <p className='text-muted-foreground font-mono text-xs'>
                    {log.id}
                  </p>
                </div>
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={() => copyToClipboard(log.id, 'Log ID')}
                >
                  <Copy className='h-3 w-3' />
                </Button>
              </div>
            </div>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}
