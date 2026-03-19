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
} from 'lucide-react'
import { type SidebarData } from '../types'

export type VisibilityLevel = 'all' | 'instance-only' | 'tenant-only'

export interface SidebarItem {
  title: string
  url: string
  icon: React.ComponentType<{ className?: string }>
  visibility?: VisibilityLevel
  badge?: string
}

export interface SidebarGroup {
  title: string
  collapsible?: boolean
  items: SidebarItem[]
  visibility?: VisibilityLevel
}

export interface SidebarDataWithVisibility extends Omit<SidebarData, 'navGroups'> {
  navGroups: SidebarGroup[]
}

export const sidebarData: SidebarDataWithVisibility = {
  user: {
    name: 'Admin',
    email: 'admin@fluxbase.eu',
    avatar: '',
  },
  teams: [
    {
      name: 'Fluxbase',
      logo: Command,
      plan: 'Backend as a Service',
    },
  ],
  navGroups: [
    {
      title: 'Overview',
      items: [
        {
          title: 'Dashboard',
          url: '/',
          icon: LayoutDashboard,
          visibility: 'all',
        },
      ],
    },
    {
      title: 'Database',
      collapsible: true,
      items: [
        {
          title: 'Tables',
          url: '/tables',
          icon: Database,
          visibility: 'all',
        },
        {
          title: 'Schema Viewer',
          url: '/schema',
          icon: GitFork,
          visibility: 'all',
        },
        {
          title: 'SQL Editor',
          url: '/sql-editor',
          icon: Code,
          visibility: 'all',
        },
      ],
    },
    {
      title: 'Users & Authentication',
      collapsible: true,
      items: [
        {
          title: 'Users',
          url: '/users',
          icon: Users,
          visibility: 'all',
        },
        {
          title: 'Tenants',
          url: '/tenants',
          icon: Building2,
          visibility: 'instance-only',
        },
        {
          title: 'Authentication',
          url: '/authentication',
          icon: Shield,
          visibility: 'all',
        },
      ],
    },
    {
      title: 'AI',
      collapsible: true,
      items: [
        {
          title: 'Knowledge Bases',
          url: '/knowledge-bases',
          icon: BookOpen,
          visibility: 'all',
        },
        {
          title: 'AI Chatbots',
          url: '/chatbots',
          icon: Bot,
          visibility: 'all',
        },
        {
          title: 'Quotas',
          url: '/quotas',
          icon: Shield,
          visibility: 'instance-only',
        },
        {
          title: 'MCP Tools',
          url: '/mcp-tools',
          icon: Wrench,
          visibility: 'all',
        },
      ],
    },
    {
      title: 'API & Services',
      collapsible: true,
      items: [
        {
          title: 'API Explorer',
          url: '/api/rest',
          icon: Code2,
          visibility: 'all',
        },
        {
          title: 'Realtime',
          url: '/realtime',
          icon: Radio,
          visibility: 'all',
        },
        {
          title: 'Storage',
          url: '/storage',
          icon: FolderOpen,
          visibility: 'all',
        },
        {
          title: 'Functions',
          url: '/functions',
          icon: FileCode,
          visibility: 'all',
        },
        {
          title: 'Jobs',
          url: '/jobs',
          icon: ListTodo,
          visibility: 'all',
        },
        {
          title: 'RPC',
          url: '/rpc',
          icon: Terminal,
          visibility: 'all',
        },
        {
          title: 'Configuration',
          url: '/features',
          icon: Zap,
          visibility: 'instance-only',
        },
        {
          title: 'Extensions',
          url: '/extensions',
          icon: Puzzle,
          visibility: 'instance-only',
        },
        {
          title: 'Email',
          url: '/email-settings',
          icon: Mail,
          visibility: 'all',
        },
        {
          title: 'Storage Config',
          url: '/storage-config',
          icon: HardDrive,
          visibility: 'all',
        },
        {
          title: 'AI Providers',
          url: '/ai-providers',
          icon: Bot,
          visibility: 'all',
        },
        {
          title: 'Database Config',
          url: '/database-config',
          icon: Database,
          visibility: 'instance-only',
        },
      ],
    },
    {
      title: 'Security',
      collapsible: true,
      items: [
        {
          title: 'RLS Policies',
          url: '/policies',
          icon: ShieldAlert,
          visibility: 'all',
        },
        {
          title: 'Security Settings',
          url: '/security-settings',
          icon: ShieldCheck,
          visibility: 'instance-only',
        },
        {
          title: 'Secrets',
          url: '/secrets',
          icon: Lock,
          visibility: 'all',
        },
        {
          title: 'Client Keys',
          url: '/client-keys',
          icon: Key,
          visibility: 'all',
        },
        {
          title: 'Service Keys',
          url: '/service-keys',
          icon: KeyRound,
          visibility: 'all',
        },
        {
          title: 'Webhooks',
          url: '/webhooks',
          icon: Webhook,
          visibility: 'all',
        },
      ],
    },
    {
      title: 'Monitoring',
      collapsible: true,
      items: [
        {
          title: 'Log Stream',
          url: '/logs',
          icon: ScrollText,
          visibility: 'all',
        },
        {
          title: 'Monitoring',
          url: '/monitoring',
          icon: Activity,
          visibility: 'instance-only',
        },
      ],
    },
    {
      title: 'Account settings',
      collapsible: true,
      items: [
        {
          title: 'Account',
          url: '/settings',
          icon: Settings,
          visibility: 'all',
        },
        {
          title: 'Appearance',
          url: '/settings/appearance',
          icon: Palette,
          visibility: 'all',
        },
      ],
    },
  ],
}

export function filterSidebarForContext(
  groups: SidebarGroup[],
  options: { isInstanceAdmin: boolean; actingAsTenantAdmin: boolean }
): SidebarGroup[] {
  const { isInstanceAdmin, actingAsTenantAdmin } = options

  // If instance admin AND NOT acting as tenant admin, show everything
  if (isInstanceAdmin && !actingAsTenantAdmin) {
    return groups
  }

  // Otherwise (tenant admin mode), filter out instance-only items
  return groups
    .map((group) => {
      // Filter items within the group
      const filteredItems = group.items.filter((item) => {
        const visibility = item.visibility || 'all'
        // Hide instance-only items when acting as tenant admin
        if (visibility === 'instance-only' && (!isInstanceAdmin || actingAsTenantAdmin)) {
          return false
        }
        return true
      })

      return {
        ...group,
        items: filteredItems,
      }
    })
    .filter((group) => {
      // Hide group if all items are filtered out
      return group.items.length > 0
    })
}
