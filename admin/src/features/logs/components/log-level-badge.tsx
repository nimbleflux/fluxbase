import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { LOG_LEVEL_CONFIG } from '../constants'
import type { LogLevel } from '../types'

interface LogLevelBadgeProps {
  level: LogLevel
  className?: string
  showIcon?: boolean
}

export function LogLevelBadge({
  level,
  className,
  showIcon = false,
}: LogLevelBadgeProps) {
  const config = LOG_LEVEL_CONFIG[level] || LOG_LEVEL_CONFIG.info
  const Icon = config.icon

  return (
    <Badge
      className={cn(
        'h-4 px-1.5 py-0 text-[10px] font-medium uppercase',
        config.color,
        'border-0 text-white',
        className
      )}
    >
      {showIcon && <Icon className='mr-1 h-2.5 w-2.5' />}
      {level}
    </Badge>
  )
}

interface LogLevelIndicatorProps {
  level: LogLevel
  className?: string
}

/**
 * Simple colored dot indicator for log level
 */
export function LogLevelIndicator({
  level,
  className,
}: LogLevelIndicatorProps) {
  const config = LOG_LEVEL_CONFIG[level] || LOG_LEVEL_CONFIG.info

  return (
    <span
      className={cn(
        'inline-block h-2 w-2 rounded-full',
        config.color,
        className
      )}
      title={config.label}
    />
  )
}
