import z from "zod";
import { createFileRoute, getRouteApi } from "@tanstack/react-router";
import { Key, Settings, Users, Building2, Shield } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  OAuthProvidersTab,
  SAMLProvidersTab,
  AuthSettingsTab,
  ActiveSessionsTab,
} from "@/components/authentication";

const authenticationSearchSchema = z.object({
  tab: z.string().optional().catch("providers"),
});

const route = getRouteApi("/_authenticated/authentication/");

const AuthenticationPage = () => {
  const search = route.useSearch();
  const navigate = route.useNavigate();

  return (
    <div className="flex h-full flex-col">
      <div className="bg-background flex items-center justify-between border-b px-6 py-4">
        <div className="flex items-center gap-3">
          <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg">
            <Shield className="text-primary h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-semibold">Authentication</h1>
            <p className="text-muted-foreground text-sm">
              Manage OAuth providers, auth settings, and user sessions
            </p>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-6">
        <Tabs
          value={search.tab || "providers"}
          onValueChange={(tab) => navigate({ search: { tab } })}
          className="w-full"
        >
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="providers">
              <Key className="mr-2 h-4 w-4" />
              OAuth Providers
            </TabsTrigger>
            <TabsTrigger value="saml">
              <Building2 className="mr-2 h-4 w-4" />
              SAML SSO
            </TabsTrigger>
            <TabsTrigger value="settings">
              <Settings className="mr-2 h-4 w-4" />
              Auth Settings
            </TabsTrigger>
            <TabsTrigger value="sessions">
              <Users className="mr-2 h-4 w-4" />
              Active Sessions
            </TabsTrigger>
          </TabsList>

          <TabsContent value="providers" className="space-y-4">
            <OAuthProvidersTab />
          </TabsContent>

          <TabsContent value="saml" className="space-y-4">
            <SAMLProvidersTab />
          </TabsContent>

          <TabsContent value="settings" className="space-y-4">
            <AuthSettingsTab />
          </TabsContent>

          <TabsContent value="sessions" className="space-y-4">
            <ActiveSessionsTab />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
};

export const Route = createFileRoute("/_authenticated/authentication/")({
  validateSearch: authenticationSearchSchema,
  component: AuthenticationPage,
});
