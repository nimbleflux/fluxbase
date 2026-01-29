import { Outlet } from '@tanstack/react-router'
import { getCookie } from '@/lib/cookies'
import { cn } from '@/lib/utils'
import { LayoutProvider } from '@/context/layout-provider'
import { SearchProvider } from '@/context/search-provider'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { ConfigDrawer } from '@/components/config-drawer'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { Header } from '@/components/layout/header'
import { Search } from '@/components/search'
import { SkipToMain } from '@/components/skip-to-main'
import { ThemeSwitch } from '@/components/theme-switch'
import { ImpersonationSelector } from '@/features/impersonation/components/impersonation-selector'

type AuthenticatedLayoutProps = {
  children?: React.ReactNode
}

export function AuthenticatedLayout({ children }: AuthenticatedLayoutProps) {
  const defaultOpen = getCookie('sidebar_state') !== 'false'
  return (
    <SearchProvider>
      <LayoutProvider>
        <SidebarProvider defaultOpen={defaultOpen}>
          <SkipToMain />
          <AppSidebar />
          <SidebarInset
            className={cn(
              // Set content container, so we can use container queries
              '@container/content',

              // If layout is fixed, set the height
              // to 100svh to prevent overflow
              'has-[[data-layout=fixed]]:h-svh',

              // If layout is fixed and sidebar is inset,
              // set the height to 100svh - spacing (total margins) to prevent overflow
              'peer-data-[variant=inset]:has-[[data-layout=fixed]]:h-[calc(100svh-(var(--spacing)*4))]'
            )}
          >
            <Header fixed>
              <Search />
              <div className='ms-auto flex items-center space-x-4'>
                <ImpersonationSelector />
                <ThemeSwitch />
                <ConfigDrawer />
              </div>
            </Header>
            {children ?? <Outlet />}
          </SidebarInset>
        </SidebarProvider>
      </LayoutProvider>
    </SearchProvider>
  )
}
