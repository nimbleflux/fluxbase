import { useState, useEffect, useRef, useCallback, memo } from 'react'
import { FluxbaseAIChat } from '@fluxbase/sdk'
import {
  Bot,
  Send,
  Loader2,
  AlertCircle,
  ChevronDown,
  ChevronUp,
  User,
  Copy,
  Check,
} from 'lucide-react'
import { toast } from 'sonner'
import type { AIChatbotSummary } from '@/lib/api'
import { getAccessToken } from '@/lib/auth'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@/components/ui/sheet'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { UserSearch } from '@/features/impersonation/components/user-search'

// Helper to get active token (impersonation takes precedence, same pattern as api.ts)
function getActiveToken(): string | null {
  const impersonationToken = localStorage.getItem(
    'fluxbase_impersonation_token'
  )
  return impersonationToken || getAccessToken()
}

interface ChatbotTestDialogProps {
  chatbot: AIChatbotSummary
  open: boolean
  onOpenChange: (open: boolean) => void
}

interface QueryResultMetadata {
  query: string
  summary: string
  rowCount: number
  data: Record<string, unknown>[]
}

interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  metadata?: {
    isStreaming?: boolean
    queryResults?: QueryResultMetadata[]
    type?: 'info' | 'error'
  }
}

function getWebSocketUrl(): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  return `${protocol}//${host}/ai/ws`
}

function formatTimestamp(date: Date): string {
  return date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

const QueryResultDisplay = memo(function QueryResultDisplay({
  result,
}: {
  result: QueryResultMetadata
}) {
  const [expanded, setExpanded] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(result.query)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error('Failed to copy to clipboard')
    }
  }

  return (
    <div className='mt-3 w-full min-w-0 space-y-2 border-t pt-2'>
      <div>
        <div className='mb-1 flex items-center justify-between'>
          <span className='text-xs font-medium'>SQL Query</span>
          {result.data.length > 0 && (
            <Button
              variant='ghost'
              size='sm'
              className='h-6 px-2 text-xs'
              onClick={() => setExpanded(!expanded)}
            >
              {expanded ? (
                <>
                  <ChevronUp className='mr-1 h-3 w-3' />
                  Hide Data
                </>
              ) : (
                <>
                  <ChevronDown className='mr-1 h-3 w-3' />
                  Show Data
                </>
              )}
            </Button>
          )}
        </div>
        <div className='group relative'>
          <pre className='overflow-x-auto rounded bg-black/10 p-2 pr-10 text-xs dark:bg-white/10'>
            <code>{result.query}</code>
          </pre>
          <Button
            variant='ghost'
            size='icon'
            className='absolute top-1 right-1 h-6 w-6 opacity-0 transition-opacity group-hover:opacity-100'
            onClick={handleCopy}
          >
            <Copy className={copied ? 'hidden h-3 w-3' : 'h-3 w-3'} />
            <Check
              className={copied ? 'h-3 w-3 text-green-500' : 'hidden h-3 w-3'}
            />
          </Button>
        </div>
      </div>

      <div className='text-muted-foreground text-xs'>
        {result.summary} ({result.rowCount} rows)
      </div>

      {expanded && result.data.length > 0 && (
        <div className='max-h-60 w-full overflow-auto rounded border'>
          <Table>
            <TableHeader>
              <TableRow>
                {Object.keys(result.data[0]).map((key) => (
                  <TableHead key={key} className='px-2 py-1 text-xs'>
                    {key}
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {result.data.slice(0, 10).map((row, idx) => (
                <TableRow key={idx}>
                  {Object.values(row).map((value, vidx) => (
                    <TableCell key={vidx} className='px-2 py-1 text-xs'>
                      {value === null ? (
                        <span className='text-muted-foreground italic'>
                          null
                        </span>
                      ) : (
                        String(value)
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {result.data.length > 10 && (
            <div className='text-muted-foreground border-t py-2 text-center text-xs'>
              Showing 10 of {result.data.length} rows
            </div>
          )}
        </div>
      )}
    </div>
  )
})

const MessageBubble = memo(function MessageBubble({
  message,
}: {
  message: ChatMessage
}) {
  const isUser = message.role === 'user'
  const isSystem = message.role === 'system'
  const isError = message.metadata?.type === 'error'

  return (
    <div
      className={cn('flex w-full', isUser ? 'justify-end' : 'justify-start')}
    >
      <div
        className={cn(
          'max-w-[85%] min-w-0 overflow-hidden rounded-lg px-4 py-2',
          isUser && 'bg-primary text-primary-foreground',
          isSystem &&
            !isError &&
            'bg-muted text-muted-foreground w-full text-center text-sm italic',
          isSystem &&
            isError &&
            'bg-destructive/10 text-destructive w-full text-sm',
          !isUser && !isSystem && 'bg-muted'
        )}
      >
        <div className='break-words whitespace-pre-wrap'>
          {message.content}
          {message.metadata?.isStreaming && (
            <span className='ml-0.5 inline-block h-4 w-1.5 animate-pulse bg-current' />
          )}
        </div>

        {message.metadata?.queryResults &&
          message.metadata.queryResults.length > 0 && (
            <div className='space-y-3'>
              {message.metadata.queryResults.map((result, idx) => (
                <QueryResultDisplay key={idx} result={result} />
              ))}
            </div>
          )}

        {!isSystem && (
          <div className='mt-1 text-xs opacity-70'>
            {formatTimestamp(message.timestamp)}
          </div>
        )}
      </div>
    </div>
  )
})

function ConnectionLoadingState() {
  return (
    <div className='flex flex-1 items-center justify-center py-12'>
      <div className='space-y-2 text-center'>
        <Loader2 className='text-muted-foreground mx-auto h-8 w-8 animate-spin' />
        <p className='text-muted-foreground text-sm'>
          Connecting to chatbot...
        </p>
      </div>
    </div>
  )
}

function ConnectionErrorState({ error }: { error: string }) {
  return (
    <div className='flex flex-1 items-center justify-center py-12'>
      <div className='space-y-2 text-center'>
        <AlertCircle className='text-destructive mx-auto h-8 w-8' />
        <p className='text-muted-foreground text-sm'>
          Failed to connect to chatbot
        </p>
        <p className='text-destructive text-xs'>{error}</p>
      </div>
    </div>
  )
}

export function ChatbotTestDialog({
  chatbot,
  open,
  onOpenChange,
}: ChatbotTestDialogProps) {
  const [conversationId, setConversationId] = useState<string | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)
  const [connectionError, setConnectionError] = useState<string | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [inputValue, setInputValue] = useState('')
  const [isSending, setIsSending] = useState(false)
  const [isThinking, setIsThinking] = useState(false)
  const [currentProgress, setCurrentProgress] = useState<string | null>(null)

  // Impersonation state
  const [impersonateUserId, setImpersonateUserId] = useState<string | null>(
    null
  )
  const [impersonateUserEmail, setImpersonateUserEmail] = useState<
    string | null
  >(null)

  const chatClientRef = useRef<FluxbaseAIChat | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = useCallback(() => {
    // Use requestAnimationFrame to ensure DOM has updated before scrolling
    requestAnimationFrame(() => {
      messagesEndRef.current?.scrollIntoView({
        behavior: 'smooth',
        block: 'end',
      })
    })
  }, [])

  useEffect(() => {
    scrollToBottom()
  }, [messages, scrollToBottom])

  const addSystemMessage = useCallback(
    (content: string, type: 'info' | 'error' = 'info') => {
      const systemMsg: ChatMessage = {
        id: `system-${Date.now()}`,
        role: 'system',
        content,
        timestamp: new Date(),
        metadata: { type },
      }
      setMessages((prev) => [...prev, systemMsg])
    },
    []
  )

  const handleContentChunk = useCallback((delta: string, _convId: string) => {
    setMessages((prev) => {
      const lastMsg = prev[prev.length - 1]

      if (lastMsg?.role === 'assistant' && lastMsg.metadata?.isStreaming) {
        return prev.map((msg, idx) =>
          idx === prev.length - 1
            ? { ...msg, content: msg.content + delta }
            : msg
        )
      }

      return [
        ...prev,
        {
          id: `msg-${Date.now()}`,
          role: 'assistant' as const,
          content: delta,
          timestamp: new Date(),
          metadata: { isStreaming: true },
        },
      ]
    })
  }, [])

  const handleProgress = useCallback(
    (step: string, message: string, _convId: string) => {
      setCurrentProgress(`${step}: ${message}`)
      setIsThinking(true)
    },
    []
  )

  const handleQueryResult = useCallback(
    (
      query: string,
      summary: string,
      rowCount: number,
      data: Record<string, unknown>[],
      _convId: string
    ) => {
      const newResult: QueryResultMetadata = { query, summary, rowCount, data }

      setMessages((prev) => {
        const lastMsg = prev[prev.length - 1]

        // If the last message is an assistant message, append the query result to it
        if (lastMsg?.role === 'assistant') {
          return prev.map((msg, idx) =>
            idx === prev.length - 1
              ? {
                  ...msg,
                  metadata: {
                    ...msg.metadata,
                    queryResults: [
                      ...(msg.metadata?.queryResults || []),
                      newResult,
                    ],
                  },
                }
              : msg
          )
        }

        // If there's no assistant message yet, create one with the query result
        // This happens when the AI uses tool calls without streaming content first
        return [
          ...prev,
          {
            id: `msg-${Date.now()}`,
            role: 'assistant' as const,
            content: summary,
            timestamp: new Date(),
            metadata: {
              isStreaming: false,
              queryResults: [newResult],
            },
          },
        ]
      })
    },
    []
  )

  const handleDone = useCallback((_usage: unknown, _convId: string) => {
    setIsThinking(false)
    setCurrentProgress(null)
    setIsSending(false)

    setMessages((prev) =>
      prev.map((msg, idx) =>
        idx === prev.length - 1 && msg.role === 'assistant'
          ? { ...msg, metadata: { ...msg.metadata, isStreaming: false } }
          : msg
      )
    )
  }, [])

  const handleError = useCallback(
    (error: string, _code: string | undefined, _convId: string | undefined) => {
      setIsThinking(false)
      setCurrentProgress(null)
      setIsSending(false)

      addSystemMessage(`Error: ${error}`, 'error')
      toast.error(`Chatbot error: ${error}`)
    },
    [addSystemMessage]
  )

  useEffect(() => {
    if (!open) return

    let mounted = true

    const initializeConnection = async () => {
      setIsConnecting(true)
      setConnectionError(null)

      try {
        const wsUrl = getWebSocketUrl()
        const token = getActiveToken()

        const chatClient = new FluxbaseAIChat({
          wsUrl,
          token: token || undefined,
          onContent: handleContentChunk,
          onProgress: handleProgress,
          onQueryResult: handleQueryResult,
          onDone: handleDone,
          onError: handleError,
          reconnectAttempts: 0,
        })

        await chatClient.connect()

        if (!mounted) {
          chatClient.disconnect()
          return
        }

        chatClientRef.current = chatClient
        setIsConnected(true)

        const convId = await chatClient.startChat(
          chatbot.name,
          chatbot.namespace,
          undefined, // conversationId
          impersonateUserId || undefined
        )

        if (!mounted) {
          chatClient.disconnect()
          return
        }

        setConversationId(convId)
        addSystemMessage(
          impersonateUserEmail
            ? `Connected to ${chatbot.name} (testing as ${impersonateUserEmail})`
            : `Connected to ${chatbot.name}`
        )
      } catch (error) {
        if (mounted) {
          const errorMessage =
            error instanceof Error ? error.message : 'Unknown error'
          setConnectionError(errorMessage)
          toast.error('Failed to connect to chatbot')
        }
      } finally {
        if (mounted) {
          setIsConnecting(false)
        }
      }
    }

    initializeConnection()

    return () => {
      mounted = false
      if (chatClientRef.current) {
        chatClientRef.current.disconnect()
        chatClientRef.current = null
      }
      setIsConnected(false)
      setConversationId(null)
      setMessages([])
      setIsThinking(false)
      setCurrentProgress(null)
      setIsSending(false)
      setConnectionError(null)
    }
  }, [
    open,
    chatbot.name,
    chatbot.namespace,
    impersonateUserId,
    impersonateUserEmail,
    addSystemMessage,
    handleContentChunk,
    handleProgress,
    handleQueryResult,
    handleDone,
    handleError,
  ])

  const handleSendMessage = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()

      if (
        !inputValue.trim() ||
        !isConnected ||
        !chatClientRef.current ||
        !conversationId
      ) {
        return
      }

      const userMessage = inputValue.trim()
      setInputValue('')
      setIsSending(true)
      setIsThinking(true)

      const userMsg: ChatMessage = {
        id: `msg-${Date.now()}`,
        role: 'user',
        content: userMessage,
        timestamp: new Date(),
      }
      setMessages((prev) => [...prev, userMsg])

      try {
        chatClientRef.current.sendMessage(conversationId, userMessage)
      } catch {
        toast.error('Failed to send message')
        setIsSending(false)
        setIsThinking(false)
      }
    },
    [inputValue, isConnected, conversationId]
  )

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side='right'
        className='flex h-full w-full flex-col p-0 sm:max-w-2xl'
      >
        <SheetHeader className='shrink-0 border-b px-6 py-4'>
          <SheetTitle className='flex items-center gap-2'>
            <Bot className='h-5 w-5' />
            Test {chatbot.name}
          </SheetTitle>
          <SheetDescription>
            {chatbot.model ? (
              <>
                Model: <span className='font-medium'>{chatbot.model}</span>
              </>
            ) : (
              'Chat with this bot to test its responses'
            )}
          </SheetDescription>
        </SheetHeader>

        {/* User impersonation selector */}
        <div className='bg-muted/30 shrink-0 border-b px-6 py-3'>
          <div className='flex items-center gap-3'>
            <User className='text-muted-foreground h-4 w-4 shrink-0' />
            <div className='flex-1'>
              {impersonateUserId ? (
                <div className='flex items-center gap-2'>
                  <span className='text-sm'>
                    Testing as: <strong>{impersonateUserEmail}</strong>
                  </span>
                  <Button
                    variant='ghost'
                    size='sm'
                    className='h-6 px-2 text-xs'
                    onClick={() => {
                      setImpersonateUserId(null)
                      setImpersonateUserEmail(null)
                    }}
                  >
                    Clear
                  </Button>
                </div>
              ) : (
                <UserSearch
                  value={impersonateUserId || undefined}
                  onSelect={(userId, userEmail) => {
                    setImpersonateUserId(userId)
                    setImpersonateUserEmail(userEmail)
                  }}
                />
              )}
            </div>
          </div>
          {impersonateUserId && (
            <p className='text-muted-foreground mt-1.5 ml-7 text-xs'>
              Queries will run with this user's RLS context
            </p>
          )}
        </div>

        {isConnecting && <ConnectionLoadingState />}

        {connectionError && !isConnecting && (
          <ConnectionErrorState error={connectionError} />
        )}

        {isConnected && !connectionError && (
          <>
            <div className='min-h-0 flex-1 overflow-x-hidden overflow-y-auto px-6 py-4'>
              <div className='space-y-4'>
                {messages.map((msg) => (
                  <MessageBubble key={msg.id} message={msg} />
                ))}

                {isThinking && (
                  <div className='text-muted-foreground flex items-center gap-2 text-sm'>
                    <Loader2 className='h-4 w-4 animate-spin' />
                    {currentProgress || 'Thinking...'}
                  </div>
                )}

                <div ref={messagesEndRef} />
              </div>
            </div>

            <div className='shrink-0 border-t px-6 py-4'>
              <form onSubmit={handleSendMessage} className='flex gap-2'>
                <Input
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  placeholder='Ask a question...'
                  disabled={!isConnected || isSending}
                  className='flex-1'
                />
                <Button
                  type='submit'
                  disabled={!isConnected || isSending || !inputValue.trim()}
                  size='icon'
                >
                  {isSending ? (
                    <Loader2 className='h-4 w-4 animate-spin' />
                  ) : (
                    <Send className='h-4 w-4' />
                  )}
                </Button>
              </form>
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  )
}
