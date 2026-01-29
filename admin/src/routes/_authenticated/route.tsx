import { createFileRoute, redirect } from '@tanstack/react-router'
import { isAuthenticated } from '@/lib/auth'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ location }) => {
    // Check if user is authenticated
    if (!isAuthenticated()) {
      // Redirect to login, preserving the intended destination
      throw redirect({
        to: '/login',
        search: {
          redirect: location.href,
        },
      })
    }
  },
  component: AuthenticatedLayout,
})
