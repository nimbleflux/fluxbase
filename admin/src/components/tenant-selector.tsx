import { Building2, Check, ChevronsUpDown, Shield } from "lucide-react";
import { useState, useEffect } from "react";
import { useTenantStore, type Tenant } from "@/stores/tenant-store";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { tenantsApi } from "@/lib/api";
import { getStoredUser } from "@/lib/auth";

function checkIsInstanceAdmin(): boolean {
  const user = getStoredUser();
  if (!user) return false;
  if ("role" in user && user.role) {
    return Array.isArray(user.role)
      ? user.role.includes("instance_admin")
      : user.role === "instance_admin";
  }
  return false;
}

export function TenantSelector() {
  const {
    tenants,
    currentTenant,
    setCurrentTenant,
    setTenants,
    isInstanceAdmin: storeIsInstanceAdmin,
    setIsInstanceAdmin,
    actingAsTenantAdmin,
    setActingAsTenantAdmin,
  } = useTenantStore();
  const [open, setOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const admin = checkIsInstanceAdmin();
    setIsInstanceAdmin(admin);

    async function fetchTenants() {
      try {
        setIsLoading(true);
        if (admin) {
          const data = await tenantsApi.list();
          setTenants(
            data.map(
              (t) =>
                ({
                  id: t.id,
                  slug: t.slug,
                  name: t.name,
                  is_default: t.is_default,
                  metadata: t.metadata || {},
                  my_role: undefined,
                  created_at: t.created_at,
                  updated_at: t.updated_at,
                }) as Tenant,
            ),
            admin, // Pass isInstanceAdmin flag
          );
        } else {
          const data = await tenantsApi.listMine();
          setTenants(
            data.map(
              (t) =>
                ({
                  id: t.id,
                  slug: t.slug,
                  name: t.name,
                  is_default: t.is_default,
                  metadata: t.metadata || {},
                  my_role: t.my_role,
                  created_at: t.created_at,
                  updated_at: t.updated_at,
                }) as Tenant,
            ),
            admin, // Pass isInstanceAdmin flag
          );
        }
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error("Failed to fetch tenants:", error);
      } finally {
        setIsLoading(false);
      }
    }

    fetchTenants();
  }, [setTenants, setIsInstanceAdmin]);

  useEffect(() => {
    if (storeIsInstanceAdmin && currentTenant && !currentTenant.my_role) {
      setActingAsTenantAdmin(true);
    } else {
      setActingAsTenantAdmin(false);
    }
  }, [storeIsInstanceAdmin, currentTenant, setActingAsTenantAdmin]);

  const handleSelectTenant = (tenant: Tenant) => {
    setCurrentTenant(tenant);
    setOpen(false);
  };

  const handleClearTenant = () => {
    setCurrentTenant(null);
    setActingAsTenantAdmin(false);
    setOpen(false);
  };

  if (!isLoading && tenants.length <= 1 && !storeIsInstanceAdmin) {
    return null;
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant={actingAsTenantAdmin ? "default" : "outline"}
          role="combobox"
          aria-expanded={open}
          aria-label="Select tenant"
          size="sm"
          className={cn(
            "w-[200px] justify-between",
            !currentTenant && "text-muted-foreground",
            actingAsTenantAdmin &&
              "bg-orange-500 hover:bg-orange-600 text-white border-orange-500",
          )}
        >
          {actingAsTenantAdmin ? (
            <Shield className="mr-2 h-4 w-4" />
          ) : (
            <Building2 className="mr-2 h-4 w-4" />
          )}
          {currentTenant ? currentTenant.name : "Select tenant..."}
          <ChevronsUpDown className="ml-auto h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[240px] p-0">
        <Command>
          <CommandInput placeholder="Search tenants..." />
          <CommandList>
            <CommandEmpty>No tenants found.</CommandEmpty>
            <CommandGroup>
              {tenants.map((tenant) => (
                <CommandItem
                  key={tenant.id}
                  value={tenant.name}
                  onSelect={() => handleSelectTenant(tenant)}
                >
                  <Check
                    className={cn(
                      "mr-2 h-4 w-4",
                      currentTenant?.id === tenant.id
                        ? "opacity-100"
                        : "opacity-0",
                    )}
                  />
                  <div className="flex flex-col">
                    <span>{tenant.name}</span>
                    <div className="flex items-center gap-2">
                      {tenant.is_default && (
                        <span className="text-xs text-muted-foreground">
                          Default
                        </span>
                      )}
                      {tenant.my_role === "tenant_admin" && (
                        <span className="text-xs text-blue-500">Admin</span>
                      )}
                      {tenant.my_role === "tenant_member" && (
                        <span className="text-xs text-muted-foreground">
                          Member
                        </span>
                      )}
                    </div>
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
            {storeIsInstanceAdmin && currentTenant && (
              <>
                <CommandSeparator />
                <CommandGroup>
                  <CommandItem
                    onSelect={handleClearTenant}
                    className="text-muted-foreground"
                  >
                    <Building2 className="mr-2 h-4 w-4" />
                    Clear tenant context
                  </CommandItem>
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
