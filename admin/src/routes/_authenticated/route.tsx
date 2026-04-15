import { createFileRoute, isRedirect, redirect } from "@tanstack/react-router";
import { isAuthenticated } from "@/lib/auth";
import { adminAuthAPI } from "@/lib/api/auth";
import { AuthenticatedLayout } from "@/components/layout/authenticated-layout";

export const Route = createFileRoute("/_authenticated")({
  beforeLoad: async ({ location }) => {
    if (!isAuthenticated()) {
      // Check if initial setup is needed (no admin user exists yet).
      // Only redirect to setup on a successful response indicating
      // needs_setup: true. On API errors (rate limit, server error),
      // fall through to the login redirect — the setup page won't work
      // anyway if the API is down.
      try {
        const status = await adminAuthAPI.getSetupStatus();
        if (status.needs_setup) {
          throw redirect({ to: "/setup" });
        }
      } catch (err) {
        // Re-throw TanStack Router redirects
        if (isRedirect(err)) throw err;
        // Swallow API errors — fall through to login redirect
      }

      // Not authenticated and setup is complete — redirect to login
      throw redirect({
        to: "/login",
        search: {
          redirect: location.href,
        },
      });
    }
  },
  component: AuthenticatedLayout,
});
