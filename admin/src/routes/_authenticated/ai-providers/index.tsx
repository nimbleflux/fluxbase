import { createFileRoute } from '@tanstack/react-router'
import { Bot } from 'lucide-react'
import { AIProvidersTab } from '@/components/ai-providers/ai-providers-tab'

const AIProvidersPage = () => {
  return (
    <div className='flex flex-1 flex-col gap-6 p-6'>
      <div>
        <h1 className='flex items-center gap-2 text-3xl font-bold tracking-tight'>
          <Bot className='h-8 w-8' />
          AI Providers
        </h1>
        <p className='text-muted-foreground mt-2 text-sm'>
          Configure AI providers for chatbots and intelligent features
        </p>
      </div>

      <AIProvidersTab />
    </div>
  )
}

export const Route = createFileRoute('/_authenticated/ai-providers/')({
  component: AIProvidersPage,
})
