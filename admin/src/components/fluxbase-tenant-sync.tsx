import { useEffect } from "react";
import { useFluxbaseClient } from "@nimbleflux/fluxbase-sdk-react";
import { useTenantStore } from "@/stores/tenant-store";

/**
 * Component that synchronizes the Fluxbase SDK client with the tenant store.
 * When a tenant is selected in the UI, this updates the SDK client to include
 * the X-FB-Tenant header in all requests.
 *
 * This is necessary because the SDK client maintains its own state separate
 * from the axios instance used by the rest of the admin UI.
 */
export function FluxbaseTenantSync({
  children,
}: {
  children: React.ReactNode;
}) {
  const client = useFluxbaseClient();
  const currentTenant = useTenantStore((state) => state.currentTenant);

  useEffect(() => {
    if (currentTenant?.id) {
      client.setTenant(currentTenant.id);
    } else {
      client.setTenant(undefined);
    }
  }, [client, currentTenant?.id]);

  return <>{children}</>;
}
