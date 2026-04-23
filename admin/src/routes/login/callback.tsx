import { useEffect, useRef } from "react";
import { createFileRoute } from "@tanstack/react-router";
import { toast } from "sonner";
import { useAuthStore } from "@/stores/auth-store";
import type { DashboardUser } from "@/lib/auth";
import { decodeJWT } from "@/lib/jwt";

export const Route = createFileRoute("/login/callback")({
  component: SSOCallbackPage,
});

function SSOCallbackPage() {
  const { auth } = useAuthStore();
  const processedRef = useRef(false);

  // Parse tokens from URL hash (fragments) for OAuth/SAML callbacks,
  // errors from query params (error redirects use ?error=...)
  const hashParams = new URLSearchParams(window.location.hash.substring(1));
  const searchParams = new URLSearchParams(window.location.search);

  const access_token = hashParams.get("access_token");
  const refresh_token = hashParams.get("refresh_token");
  const redirect_to = hashParams.get("redirect_to");
  const error = searchParams.get("error");

  useEffect(() => {
    // Prevent double processing
    if (processedRef.current) return;
    processedRef.current = true;

    const handleCallback = async () => {
      // Handle error
      if (error) {
        toast.error("SSO Login Failed", {
          description: error,
        });
        window.location.href = "/admin/login";
        return;
      }

      // Handle missing tokens
      if (!access_token || !refresh_token) {
        toast.error("SSO Login Failed", {
          description: "No authentication tokens received",
        });
        window.location.href = "/admin/login";
        return;
      }

      try {
        auth.setTokens(access_token, refresh_token);

        const tokenPayload = decodeJWT(access_token);

        const user: DashboardUser = {
          id: tokenPayload.user_id,
          email: tokenPayload.email,
          email_verified: true,
          full_name: tokenPayload.user_metadata?.name || null,
          avatar_url: tokenPayload.user_metadata?.avatar || null,
          totp_enabled: false,
          is_active: true,
          is_locked: false,
          last_login_at: new Date().toISOString(),
          created_at: tokenPayload.iat
            ? new Date(tokenPayload.iat * 1000).toISOString()
            : new Date().toISOString(),
          updated_at: new Date().toISOString(),
          role: tokenPayload.role,
        };

        auth.setUser({
          accountNo: tokenPayload.user_id,
          email: tokenPayload.email,
          role: [tokenPayload.role || "tenant_admin"],
          exp: tokenPayload.exp
            ? tokenPayload.exp * 1000
            : Date.now() + 24 * 60 * 60 * 1000,
        });

        localStorage.setItem("fluxbase_admin_user", JSON.stringify(user));

        toast.success("Welcome!", {
          description: "You have successfully logged in via SSO.",
        });

        // Redirect to the intended destination or dashboard
        const destination =
          redirect_to && redirect_to !== "/" ? redirect_to : "/admin";
        window.location.href = destination;
      } catch (_error) {
        toast.error("SSO Login Failed", {
          description: "Failed to complete authentication",
        });
        window.location.href = "/admin/login";
      }
    };

    handleCallback();
  }, [access_token, refresh_token, redirect_to, error, auth]);

  return (
    <div className="from-background to-muted flex min-h-screen flex-col items-center justify-center bg-gradient-to-br p-4">
      <div className="space-y-4 text-center">
        <div className="border-primary mx-auto h-12 w-12 animate-spin rounded-full border-b-2" />
        <p className="text-muted-foreground">Completing SSO login...</p>
      </div>
    </div>
  );
}
