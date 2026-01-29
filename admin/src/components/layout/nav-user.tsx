import { LogOut, User } from 'lucide-react'
import { logout } from '@/lib/auth'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { SidebarMenu, SidebarMenuItem } from '@/components/ui/sidebar'

// Component to display user name in sidebar footer
type NavUserProps = {
  user: {
    name: string
    email: string
    avatar: string
  }
}

export function NavUser({ user }: NavUserProps) {
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <div className='flex items-center gap-2 px-2 py-1.5'>
          <Avatar className='h-8 w-8'>
            <AvatarImage src={user.avatar} alt={user.name} />
            <AvatarFallback>
              <User className='h-4 w-4' />
            </AvatarFallback>
          </Avatar>
          <div className='grid flex-1 text-start text-sm leading-tight'>
            <span className='truncate font-semibold'>{user.name}</span>
            <span className='text-muted-foreground truncate text-xs'>
              {user.email}
            </span>
          </div>
          <Button
            variant='ghost'
            size='icon'
            className='h-8 w-8'
            onClick={logout}
            title='Sign out'
          >
            <LogOut className='h-4 w-4' />
          </Button>
        </div>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
