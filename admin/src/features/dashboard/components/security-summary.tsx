import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import {
  AlertCircle,
  AlertTriangle,
  Info,
  Shield,
  CheckCircle2,
} from 'lucide-react'
import { policyApi } from '@/lib/api'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export function SecuritySummary() {
  const { data: warningsData, isLoading } = useQuery({
    queryKey: ['security-warnings'],
    queryFn: () => policyApi.getSecurityWarnings(),
    staleTime: 60000, // Cache for 1 minute
  })

  if (isLoading) {
    return (
      <div className='grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4'>
        {[1, 2, 3, 4].map((i) => (
          <Card key={i}>
            <CardHeader className='pb-2'>
              <Skeleton className='h-4 w-24' />
              <Skeleton className='mt-2 h-8 w-12' />
            </CardHeader>
          </Card>
        ))}
      </div>
    )
  }

  const summary = warningsData?.summary || {
    critical: 0,
    high: 0,
    medium: 0,
    low: 0,
    total: 0,
  }
  const hasIssues = summary.total > 0

  return (
    <div className='space-y-4'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <Shield className='h-5 w-5' />
          <h2 className='text-lg font-semibold'>Security Overview</h2>
        </div>
        <Button variant='outline' size='sm' asChild>
          <Link to='/policies'>View All Policies</Link>
        </Button>
      </div>

      {!hasIssues ? (
        <Card className='border-green-500/50 bg-green-500/5'>
          <CardHeader className='pb-4'>
            <div className='flex items-center gap-3'>
              <CheckCircle2 className='h-8 w-8 text-green-500' />
              <div>
                <CardTitle className='text-lg'>All Clear</CardTitle>
                <CardDescription>
                  No security warnings detected. Your RLS policies are properly
                  configured.
                </CardDescription>
              </div>
            </div>
          </CardHeader>
        </Card>
      ) : (
        <div className='grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4'>
          <Card
            className={cn(
              summary.critical > 0 && 'border-red-500/50 bg-red-500/5'
            )}
          >
            <CardHeader className='pb-2'>
              <CardDescription>Critical Issues</CardDescription>
              <CardTitle className='flex items-center gap-2 text-2xl'>
                <AlertCircle
                  className={cn(
                    'h-5 w-5',
                    summary.critical > 0
                      ? 'text-red-500'
                      : 'text-muted-foreground'
                  )}
                />
                {summary.critical}
              </CardTitle>
            </CardHeader>
          </Card>

          <Card
            className={cn(
              summary.high > 0 && 'border-orange-500/50 bg-orange-500/5'
            )}
          >
            <CardHeader className='pb-2'>
              <CardDescription>High Priority</CardDescription>
              <CardTitle className='flex items-center gap-2 text-2xl'>
                <AlertTriangle
                  className={cn(
                    'h-5 w-5',
                    summary.high > 0
                      ? 'text-orange-500'
                      : 'text-muted-foreground'
                  )}
                />
                {summary.high}
              </CardTitle>
            </CardHeader>
          </Card>

          <Card
            className={cn(
              summary.medium > 0 && 'border-yellow-500/50 bg-yellow-500/5'
            )}
          >
            <CardHeader className='pb-2'>
              <CardDescription>Medium Priority</CardDescription>
              <CardTitle className='flex items-center gap-2 text-2xl'>
                <Info
                  className={cn(
                    'h-5 w-5',
                    summary.medium > 0
                      ? 'text-yellow-500'
                      : 'text-muted-foreground'
                  )}
                />
                {summary.medium}
              </CardTitle>
            </CardHeader>
          </Card>

          <Card>
            <CardHeader className='pb-2'>
              <CardDescription>Low Priority</CardDescription>
              <CardTitle className='flex items-center gap-2 text-2xl'>
                <Info
                  className={cn(
                    'h-5 w-5',
                    summary.low > 0 ? 'text-blue-500' : 'text-muted-foreground'
                  )}
                />
                {summary.low}
              </CardTitle>
            </CardHeader>
          </Card>
        </div>
      )}
    </div>
  )
}
