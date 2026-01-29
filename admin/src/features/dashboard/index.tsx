import { getRouteApi, Link } from '@tanstack/react-router'
import { LayoutDashboard } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { Main } from '@/components/layout/main'
import { FluxbaseStats } from './components/fluxbase-stats'
import { SecuritySummary } from './components/security-summary'

const route = getRouteApi('/_authenticated/')

export function Dashboard() {
  const search = route.useSearch()
  const navigate = route.useNavigate()
  return (
    <Main>
        <div className='mb-2 flex items-center justify-between space-y-2'>
          <div>
            <h1 className='flex items-center gap-2 text-3xl font-bold tracking-tight'>
              <LayoutDashboard className='h-8 w-8' />
              Dashboard
            </h1>
            <p className='text-muted-foreground mt-2 text-sm'>
              Monitor your Backend as a Service
            </p>
          </div>
        </div>
        <Tabs
          orientation='vertical'
          value={search.tab || 'overview'}
          onValueChange={(tab) => navigate({ search: { tab } })}
          className='space-y-4'
        >
          <TabsContent value='overview' className='space-y-4'>
            {/* Fluxbase System Stats */}
            <FluxbaseStats />

            {/* Security Summary */}
            <SecuritySummary />

            {/* Quick Actions */}
            <Card>
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
                <CardDescription>Common administrative tasks</CardDescription>
              </CardHeader>
              <CardContent>
                <div className='grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4'>
                  <div className='text-sm'>
                    <p className='text-muted-foreground mb-1'>Database</p>
                    <Link
                      to='/tables'
                      className='text-primary hover:underline'
                    >
                      Browse database tables →
                    </Link>
                  </div>
                  <div className='text-sm'>
                    <p className='text-muted-foreground mb-1'>Users</p>
                    <Link
                      to='/users'
                      className='text-primary hover:underline'
                    >
                      Manage user accounts →
                    </Link>
                  </div>
                  <div className='text-sm'>
                    <p className='text-muted-foreground mb-1'>Functions</p>
                    <Link
                      to='/functions'
                      className='text-primary hover:underline'
                    >
                      Manage Edge Functions →
                    </Link>
                  </div>
                  <div className='text-sm'>
                    <p className='text-muted-foreground mb-1'>Settings</p>
                    <Link
                      to='/settings'
                      className='text-primary hover:underline'
                    >
                      Configure system settings →
                    </Link>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
    </Main>
  )
}
