import { useState } from "react";
import { Building2, Check, ChevronsUpDown } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { useTenantStore, type Tenant } from "@/stores/tenant-store";
import { tenantsApi } from "@/lib/api";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

import { checkIsInstanceAdmin } from "@/lib/route-guards";

interface TenantRequiredGuardProps {
  children: React.ReactNode;
}

/**
 * Guard that shows an inline tenant selector when an instance admin
 * visits a tenant-required page without a tenant selected.
 * When tenant is selected or user is a non-instance admin, renders children.
 */
export function TenantRequiredGuard({ children }: TenantRequiredGuardProps) {
  const { currentTenant, isInstanceAdmin } = useTenantStore();

  // If tenant is selected, or user is a tenant admin (non-instance), show content
  if (currentTenant || !isInstanceAdmin) {
    return <>{children}</>;
  }

  // Instance admin without tenant selected on a tenant-required page
  return (
    <div className="flex items-center justify-center p-8">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-orange-100 dark:bg-orange-950">
            <Building2 className="h-6 w-6 text-orange-600 dark:text-orange-400" />
          </div>
          <CardTitle>Select a tenant</CardTitle>
          <CardDescription>
            This page displays tenant-scoped data. Choose a tenant to continue.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex justify-center">
          <InlineTenantSelector />
        </CardContent>
      </Card>
    </div>
  );
}

function InlineTenantSelector() {
  const { tenants, setCurrentTenant, setIsInstanceAdmin } = useTenantStore();
  const [open, setOpen] = useState(false);
  const storeIsInstanceAdmin = checkIsInstanceAdmin();

  const { data: fetchedTenants } = useQuery({
    queryKey: ["tenants", storeIsInstanceAdmin, setIsInstanceAdmin],
    queryFn: async () => {
      setIsInstanceAdmin(storeIsInstanceAdmin);
      if (storeIsInstanceAdmin) {
        return await tenantsApi.list();
      }
      return await tenantsApi.listMine();
    },
  });

  const tenantList = fetchedTenants
    ? fetchedTenants.map(
        (t) =>
          ({
            id: t.id,
            slug: t.slug,
            name: t.name,
            is_default: t.is_default,
            metadata: t.metadata || {},
            created_at: t.created_at,
            updated_at: t.updated_at,
          }) as Tenant,
      )
    : tenants;

  const handleSelect = (tenant: Tenant) => {
    setCurrentTenant(tenant);
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-[240px] justify-between"
        >
          Select tenant...
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[260px] p-0" align="center">
        <Command>
          <CommandInput placeholder="Search tenants..." />
          <CommandList>
            <CommandEmpty>No tenants found.</CommandEmpty>
            <CommandGroup>
              {tenantList.map((tenant) => (
                <CommandItem
                  key={tenant.id}
                  value={`${tenant.name} ${tenant.id}`}
                  onSelect={() => handleSelect(tenant)}
                >
                  <Check className={cn("mr-2 h-4 w-4 opacity-0")} />
                  <div className="flex flex-col">
                    <span>{tenant.name}</span>
                    {tenant.is_default && (
                      <span className="text-xs text-muted-foreground">
                        Default
                      </span>
                    )}
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
