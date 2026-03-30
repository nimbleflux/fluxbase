import { Outlet } from "@tanstack/react-router";
import { getCookie } from "@/lib/cookies";
import { cn } from "@/lib/utils";
import { LayoutProvider } from "@/context/layout-provider";
import { SearchProvider } from "@/context/search-provider";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { ConfigDrawer } from "@/components/config-drawer";
import { AppSidebar } from "@/components/layout/app-sidebar";
import { Header } from "@/components/layout/header";
import { Search } from "@/components/search";
import { SkipToMain } from "@/components/skip-to-main";
import { ThemeSwitch } from "@/components/theme-switch";
import { ImpersonationSelector } from "@/features/impersonation/components/impersonation-selector";
import { TenantSelector } from "@/components/tenant-selector";
import { BranchSelector } from "@/components/branch-selector";
import { useTenantQueryRefresh } from "@/hooks/use-tenant-query-refresh";
import { useBranchQueryRefresh } from "@/hooks/use-branch-query-refresh";

type AuthenticatedLayoutProps = {
  children?: React.ReactNode;
};

export function AuthenticatedLayout({ children }: AuthenticatedLayoutProps) {
  const defaultOpen = getCookie("sidebar_state") !== "false";

  // Refresh all queries when tenant context changes
  useTenantQueryRefresh();
  // Refresh all queries when branch context changes
  useBranchQueryRefresh();

  return (
    <SearchProvider>
      <LayoutProvider>
        <SidebarProvider defaultOpen={defaultOpen}>
          <SkipToMain />
          <AppSidebar />
          <SidebarInset
            className={cn(
              // Set content container, so we can use container queries
              "@container/content",

              // If layout is fixed, set the height
              // to 100svh to prevent overflow
              "has-[[data-layout=fixed]]:h-svh",

              // If layout is fixed and sidebar is inset,
              // set the height to 100svh - spacing (total margins) to prevent overflow
              "peer-data-[variant=inset]:has-[[data-layout=fixed]]:h-[calc(100svh-(var(--spacing)*4))]",
            )}
          >
            <Header fixed>
              <Search />
              <div className="ms-auto flex items-center space-x-4">
                <TenantSelector />
                <BranchSelector />
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
  );
}
