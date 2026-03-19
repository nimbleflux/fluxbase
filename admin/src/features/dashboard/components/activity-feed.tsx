import { CheckCircle2, XCircle, AlertTriangle, Info, Clock, Wifi, WifiOff } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { useActivityLogs, type ActivityLog } from '@/hooks/use-activity-logs'

type ActivityType = 'success' | 'error' | 'warning' | 'info'

// Map log level and category to activity type
const getActivityType = (level: string, category?: string): ActivityType => {
  if (level === 'error' || level === 'fatal') return 'error'
  if (level === 'warn') return 'warning'
  if (category === 'security' || category === 'auth') return 'success'
  return 'info'
}

// Get source label from log entry
const getSourceLabel = (component?: string, customCategory?: string): string | undefined => {
  return customCategory || component
}

// Extract security event details from log
const getSecurityEventDetails = (log: ActivityLog): string[] => {
  const details: string[] = []

  // Extract relevant fields for security events
  if (log.fields) {
    // Prefer email over user_id for display
    if (log.fields.email) {
      details.push(`User: ${String(log.fields.email)}`)
    } else if (log.user_id) {
      details.push(`User: ${log.user_id}`)
    }

    if (log.fields.action) {
      details.push(`Action: ${String(log.fields.action)}`)
    }
    if (log.fields.resource_type) {
      details.push(`Resource: ${String(log.fields.resource_type)}`)
    }
    if (log.fields.reason) {
      details.push(`Reason: ${String(log.fields.reason)}`)
    }
    if (log.fields.target_user) {
      details.push(`Target: ${String(log.fields.target_user)}`)
    }
  } else if (log.user_id) {
    details.push(`User: ${log.user_id}`)
  }

  if (log.ip_address) {
    details.push(`IP: ${log.ip_address}`)
  }

  return details
}

const formatRelativeTime = (date: string): string => {
  const now = new Date()
  const then = new Date(date)
  const diffMs = now.getTime() - then.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  return then.toLocaleDateString()
}

export function ActivityFeed() {
  const { logs, loading, error, isSubscribed } = useActivityLogs({
    enabled: true,
    maxLogs: 50,
    timeRangeHours: 24,
  })

  // Sort logs by timestamp descending (most recent first)
  const sortedLogs = [...logs].sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  )

  // Convert log to activity item
  const logToActivity = (log: ActivityLog) => ({
    id: log.id,
    type: getActivityType(log.level, log.category),
    message: log.message,
    timestamp: log.timestamp,
    source: getSourceLabel(log.component, log.custom_category),
  })

  const getActivityIcon = (type: ActivityType) => {
    switch (type) {
      case 'success':
        return <CheckCircle2 className='h-4 w-4 text-green-500' />
      case 'error':
        return <XCircle className='h-4 w-4 text-red-500' />
      case 'warning':
        return <AlertTriangle className='h-4 w-4 text-yellow-500' />
      case 'info':
        return <Info className='h-4 w-4 text-blue-500' />
    }
  }

  const getActivityBadgeVariant = (type: ActivityType) => {
    switch (type) {
      case 'success':
        return 'secondary'
      case 'error':
        return 'destructive'
      case 'warning':
        return 'secondary'
      case 'info':
        return 'outline'
    }
  }

  // Connection status indicator
  const getConnectionStatus = () => {
    if (error) {
      return (
        <div className='flex items-center gap-2 text-destructive text-xs'>
          <WifiOff className='h-3 w-3' />
          Connection error
        </div>
      )
    }
    if (isSubscribed) {
      return (
        <div className='flex items-center gap-2 text-green-600 text-xs'>
          <Wifi className='h-3 w-3' />
          Live
        </div>
      )
    }
    return (
      <div className='flex items-center gap-2 text-muted-foreground text-xs'>
        <WifiOff className='h-3 w-3' />
        Connecting...
      </div>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center justify-between text-base'>
          <div className='flex items-center gap-2'>
            <Clock className='h-4 w-4' />
            Recent Activity
          </div>
          {!loading && getConnectionStatus()}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className='space-y-3'>
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className='flex gap-3'>
                <Skeleton className='h-4 w-4 rounded-full' />
                <div className='flex-1 space-y-1'>
                  <Skeleton className='h-4 w-48' />
                  <Skeleton className='h-3 w-24' />
                </div>
              </div>
            ))}
          </div>
        ) : error ? (
          <div className='py-8 text-center'>
            <p className='text-destructive text-sm'>{error.message}</p>
            <p className='text-muted-foreground mt-2 text-xs'>
              Failed to load activity logs
            </p>
          </div>
        ) : sortedLogs.length > 0 ? (
          <div className='space-y-3'>
            {sortedLogs.slice(0, 10).map((log) => {
              const activity = logToActivity(log)
              const securityDetails = log.category === 'security' ? getSecurityEventDetails(log) : []
              return (
                <div
                  key={activity.id}
                  className='group flex gap-3 rounded-lg p-2 transition-colors hover:bg-muted/50'
                >
                  <div className='mt-0.5 shrink-0'>{getActivityIcon(activity.type)}</div>
                  <div className='min-w-0 flex-1'>
                    <p className='text-sm leading-tight'>{activity.message}</p>
                    {securityDetails.length > 0 && (
                      <div className='mt-1 flex flex-wrap gap-x-2 gap-y-0.5 text-xs text-muted-foreground'>
                        {securityDetails.map((detail, idx) => (
                          <span key={idx} className='flex items-center gap-1'>
                            <span className='font-medium'>{detail.split(':')[0]}:</span>
                            <span>{detail.split(':').slice(1).join(':')}</span>
                            {idx < securityDetails.length - 1 && <span className='text-muted-foreground/30'>|</span>}
                          </span>
                        ))}
                      </div>
                    )}
                    <div className='mt-1 flex items-center gap-2'>
                      <span className='text-muted-foreground text-xs'>
                        {formatRelativeTime(activity.timestamp)}
                      </span>
                      {activity.source && (
                        <>
                          <span className='text-muted-foreground/50'>•</span>
                          <Badge variant={getActivityBadgeVariant(activity.type)} className='h-4 px-1 py-0 text-[10px]'>
                            {activity.source}
                          </Badge>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <div className='py-8 text-center text-sm text-muted-foreground'>
            Waiting for logs...
          </div>
        )}
      </CardContent>
    </Card>
  )
}
