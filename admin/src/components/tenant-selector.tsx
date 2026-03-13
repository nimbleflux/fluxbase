import { Building2, Check, ChevronsUpDown } from 'lucide-react'
import { useState, useEffect } from 'react'
import { useTenants } from '@nimbleflux/fluxbase-sdk-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

export function TenantSelector() {
  const { tenants, isLoading, currentTenantId, setCurrentTenant } = useTenants()
  const [open, setOpen] = useState(false)

  // Hide selector if user only has access to one tenant
  if (!isLoading && tenants.length <= 1) {
    return null
  }

  const currentTenant = tenants.find((t) => t.id === currentTenantId)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant='outline'
          role='combobox'
          aria-expanded={open}
          aria-label='Select tenant'
          size='sm'
          className={cn(
            'w-[200px] justify-between',
            !currentTenant && 'text-muted-foreground'
          )}
        >
          <Building2 className='mr-2 h-4 w-4' />
          {currentTenant ? currentTenant.name : 'Select tenant...'}
          <ChevronsUpDown className='ml-auto h-4 w-4 shrink-0 opacity-50' />
        </Button>
      </PopoverTrigger>
      <PopoverContent className='w-[200px] p-0'>
        <Command>
          <CommandInput placeholder='Search tenants...' />
          <CommandList>
            <CommandEmpty>No tenants found.</CommandEmpty>
            <CommandGroup>
              {tenants.map((tenant) => (
                <CommandItem
                  key={tenant.id}
                  value={tenant.name}
                  onSelect={() => {
                    setCurrentTenant(tenant.id)
                    setOpen(false)
                  }}
                >
                  <Check
                    className={cn(
                      'mr-2 h-4 w-4',
                      currentTenantId === tenant.id
                        ? 'opacity-100'
                        : 'opacity-0'
                    )}
                  />
                  <div className='flex flex-col'>
                    <span>{tenant.name}</span>
                    {tenant.my_role === 'tenant_admin' && (
                      <span className='text-xs text-muted-foreground'>
                        Admin
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
  )
}
