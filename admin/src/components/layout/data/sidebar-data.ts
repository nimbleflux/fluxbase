import {
  LayoutDashboard,
  Database,
  GitFork,
  Code,
  Code2,
  ScrollText,
  Activity,
  Users,
  Shield,
  Zap,
  FileCode,
  FolderOpen,
  Radio,
  ListTodo,
  Terminal,
  Bot,
  BookOpen,
  Wrench,
  Key,
  KeyRound,
  ShieldAlert,
  ShieldCheck,
  Webhook,
  Lock,
  Settings,
  Palette,
  Puzzle,
  Mail,
  HardDrive,
  Command,
  Building2,
  Server,
} from "lucide-react";
import { type SidebarData } from "../types";

export type VisibilityLevel = "all" | "instance-only" | "tenant-only";

export interface SidebarItem {
  title: string;
  url: string;
  icon: React.ComponentType<{ className?: string }>;
  visibility?: VisibilityLevel;
  badge?: string;
}

export interface SidebarGroup {
  title: string;
  collapsible?: boolean;
  items: SidebarItem[];
  visibility?: VisibilityLevel;
}

export interface SidebarDataWithVisibility extends Omit<
  SidebarData,
  "navGroups"
> {
  navGroups: SidebarGroup[];
}

export const sidebarData: SidebarDataWithVisibility = {
  user: {
    name: "Admin",
    email: "admin@fluxbase.eu",
    avatar: "",
  },
  teams: [
    {
      name: "Fluxbase",
      logo: Command,
      plan: "Backend as a Service",
    },
  ],
  navGroups: [
    {
      title: "Overview",
      items: [
        {
          title: "Dashboard",
          url: "/",
          icon: LayoutDashboard,
          visibility: "all",
        },
      ],
    },
    {
      title: "Database",
      collapsible: true,
      items: [
        {
          title: "Tables",
          url: "/tables",
          icon: Database,
          visibility: "all",
        },
        {
          title: "Schema Viewer",
          url: "/schema",
          icon: GitFork,
          visibility: "all",
        },
        {
          title: "SQL Editor",
          url: "/sql-editor",
          icon: Code,
          visibility: "tenant-only",
        },
      ],
    },
    {
      title: "Users & Authentication",
      collapsible: true,
      items: [
        {
          title: "Users",
          url: "/users",
          icon: Users,
          visibility: "all",
        },
        {
          title: "Authentication",
          url: "/authentication",
          icon: Shield,
          visibility: "all",
        },
      ],
    },
    {
      title: "AI",
      collapsible: true,
      items: [
        {
          title: "Knowledge Bases",
          url: "/knowledge-bases",
          icon: BookOpen,
          visibility: "tenant-only",
        },
        {
          title: "AI Chatbots",
          url: "/chatbots",
          icon: Bot,
          visibility: "tenant-only",
        },
        {
          title: "MCP Tools",
          url: "/mcp-tools",
          icon: Wrench,
          visibility: "tenant-only",
        },
      ],
    },
    {
      title: "API & Services",
      collapsible: true,
      items: [
        {
          title: "API Explorer",
          url: "/api/rest",
          icon: Code2,
          visibility: "all",
        },
        {
          title: "Realtime",
          url: "/realtime",
          icon: Radio,
          visibility: "tenant-only",
        },
        {
          title: "Storage",
          url: "/storage",
          icon: FolderOpen,
          visibility: "tenant-only",
        },
        {
          title: "Functions",
          url: "/functions",
          icon: FileCode,
          visibility: "tenant-only",
        },
        {
          title: "Jobs",
          url: "/jobs",
          icon: ListTodo,
          visibility: "tenant-only",
        },
        {
          title: "RPC",
          url: "/rpc",
          icon: Terminal,
          visibility: "tenant-only",
        },
        {
          title: "Email",
          url: "/email-settings",
          icon: Mail,
          visibility: "all",
        },
        {
          title: "Storage Config",
          url: "/storage-config",
          icon: HardDrive,
          visibility: "all",
        },
        {
          title: "AI Providers",
          url: "/ai-providers",
          icon: Bot,
          visibility: "tenant-only",
        },
      ],
    },
    {
      title: "Security",
      collapsible: true,
      items: [
        {
          title: "RLS Policies",
          url: "/policies",
          icon: ShieldAlert,
          visibility: "all",
        },
        {
          title: "Security Settings",
          url: "/security-settings",
          icon: ShieldCheck,
          visibility: "tenant-only",
        },
        {
          title: "Secrets",
          url: "/secrets",
          icon: Lock,
          visibility: "tenant-only",
        },
        {
          title: "Client Keys",
          url: "/client-keys",
          icon: Key,
          visibility: "tenant-only",
        },
        {
          title: "Service Keys",
          url: "/service-keys",
          icon: KeyRound,
          visibility: "tenant-only",
        },
        {
          title: "Webhooks",
          url: "/webhooks",
          icon: Webhook,
          visibility: "tenant-only",
        },
      ],
    },
    {
      title: "Monitoring",
      collapsible: true,
      items: [
        {
          title: "Log Stream",
          url: "/logs",
          icon: ScrollText,
          visibility: "all",
        },
        {
          title: "Monitoring",
          url: "/monitoring",
          icon: Activity,
          visibility: "all",
        },
      ],
    },
    {
      title: "Platform",
      collapsible: true,
      items: [
        {
          title: "Tenants",
          url: "/tenants",
          icon: Building2,
          visibility: "instance-only",
        },
        {
          title: "Configuration",
          url: "/features",
          icon: Zap,
          visibility: "instance-only",
        },
        {
          title: "Extensions",
          url: "/extensions",
          icon: Puzzle,
          visibility: "instance-only",
        },
        {
          title: "Database Config",
          url: "/database-config",
          icon: Database,
          visibility: "instance-only",
        },
        {
          title: "Instance Settings",
          url: "/instance-settings",
          icon: Server,
          visibility: "instance-only",
        },
      ],
    },
    {
      title: "Account settings",
      collapsible: true,
      items: [
        {
          title: "Account",
          url: "/settings",
          icon: Settings,
          visibility: "all",
        },
        {
          title: "Appearance",
          url: "/settings/appearance",
          icon: Palette,
          visibility: "all",
        },
      ],
    },
  ],
};

export function filterSidebarForContext(
  groups: SidebarGroup[],
  options: { isInstanceAdmin: boolean; actingAsTenantAdmin: boolean },
): SidebarGroup[] {
  const { isInstanceAdmin, actingAsTenantAdmin } = options;

  return groups
    .map((group) => {
      const filteredItems = group.items.filter((item) => {
        const visibility = item.visibility || "all";

        // Instance admin: always show instance-only + all
        // When acting as tenant, also show tenant-only items
        if (isInstanceAdmin) {
          if (actingAsTenantAdmin) {
            return true; // Show everything to instance admins in tenant context
          }
          return visibility !== "tenant-only"; // Hide tenant-only when not in tenant context
        }

        // Tenant admin: show tenant-only + all, hide instance-only
        if (visibility === "instance-only") {
          return false;
        }

        return true;
      });

      return { ...group, items: filteredItems };
    })
    .filter((group) => group.items.length > 0);
}
