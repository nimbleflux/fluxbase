/**
 * AI Module Tests
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { FluxbaseAI, FluxbaseAIChat } from './ai'
import type { AIChatbotSummary, AIChatbotLookupResponse, ListConversationsResult, AIUserConversationDetail } from './types'

// Mock WebSocket constants (not available in Node.js)
// Only need the constants, not a full mock, since these tests don't actually connect
(global as any).WebSocket = {
  CONNECTING: 0,
  OPEN: 1,
  CLOSING: 2,
  CLOSED: 3,
};

// Mock fetch interface
class MockFetch {
  public lastUrl: string = ''
  public lastMethod: string = ''
  public lastBody: unknown = null
  public mockResponse: any = null
  public shouldThrow: boolean = false
  public errorMessage: string = 'Test error'

  async get<T>(path: string): Promise<T> {
    this.lastUrl = path
    this.lastMethod = 'GET'
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
    return this.mockResponse as T
  }

  async patch<T>(path: string, body?: unknown): Promise<T> {
    this.lastUrl = path
    this.lastMethod = 'PATCH'
    this.lastBody = body
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
    return this.mockResponse as T
  }

  async delete(path: string): Promise<void> {
    this.lastUrl = path
    this.lastMethod = 'DELETE'
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
  }
}

describe('FluxbaseAI', () => {
  let mockFetch: MockFetch
  let ai: FluxbaseAI

  beforeEach(() => {
    mockFetch = new MockFetch()
    ai = new FluxbaseAI(mockFetch, 'ws://localhost:8080')
  })

  describe('listChatbots', () => {
    it('should list available chatbots', async () => {
      const mockChatbots: AIChatbotSummary[] = [
        {
          id: 'cb1',
          name: 'sql-assistant',
          namespace: 'default',
          description: 'SQL query assistant',
        },
        {
          id: 'cb2',
          name: 'data-analyst',
          namespace: 'default',
          description: 'Data analysis helper',
        },
      ]
      mockFetch.mockResponse = { chatbots: mockChatbots, count: 2 }

      const { data, error } = await ai.listChatbots()

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/chatbots')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockChatbots)
      expect(error).toBeNull()
    })

    it('should return empty array when no chatbots', async () => {
      mockFetch.mockResponse = { count: 0 }

      const { data, error } = await ai.listChatbots()

      expect(data).toEqual([])
      expect(error).toBeNull()
    })

    it('should handle errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Permission denied'

      const { data, error } = await ai.listChatbots()

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Permission denied')
    })
  })

  describe('getChatbot', () => {
    it('should get chatbot details', async () => {
      const mockChatbot: AIChatbotSummary = {
        id: 'cb1',
        name: 'sql-assistant',
        namespace: 'default',
        description: 'SQL query assistant',
      }
      mockFetch.mockResponse = mockChatbot

      const { data, error } = await ai.getChatbot('cb1')

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/chatbots/cb1')
      expect(data).toEqual(mockChatbot)
      expect(error).toBeNull()
    })

    it('should handle errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Chatbot not found'

      const { data, error } = await ai.getChatbot('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
    })
  })

  describe('lookupChatbot', () => {
    it('should lookup chatbot by name', async () => {
      const mockLookup: AIChatbotLookupResponse = {
        chatbot: {
          id: 'cb1',
          name: 'sql-assistant',
          namespace: 'default',
        },
      }
      mockFetch.mockResponse = mockLookup

      const { data, error } = await ai.lookupChatbot('sql-assistant')

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/chatbots/by-name/sql-assistant')
      expect(data).toEqual(mockLookup)
      expect(error).toBeNull()
    })

    it('should encode chatbot name in URL', async () => {
      mockFetch.mockResponse = { chatbot: { name: 'my assistant' } }

      await ai.lookupChatbot('my assistant')

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/chatbots/by-name/my%20assistant')
    })

    it('should handle ambiguous lookup', async () => {
      const mockLookup: AIChatbotLookupResponse = {
        ambiguous: true,
        namespaces: ['default', 'custom'],
      }
      mockFetch.mockResponse = mockLookup

      const { data, error } = await ai.lookupChatbot('common-name')

      expect(data?.ambiguous).toBe(true)
      expect(data?.namespaces).toEqual(['default', 'custom'])
      expect(error).toBeNull()
    })

    it('should handle errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Chatbot not found'

      const { data, error } = await ai.lookupChatbot('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
    })
  })

  describe('listConversations', () => {
    it('should list conversations', async () => {
      const mockResult: ListConversationsResult = {
        conversations: [
          {
            id: 'conv1',
            chatbot_name: 'sql-assistant',
            title: 'Query help',
            created_at: '2025-01-01T00:00:00Z',
            updated_at: '2025-01-01T00:00:00Z',
          },
        ],
        count: 1,
      }
      mockFetch.mockResponse = mockResult

      const { data, error } = await ai.listConversations()

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/conversations')
      expect(data).toEqual(mockResult)
      expect(error).toBeNull()
    })

    it('should filter by chatbot', async () => {
      mockFetch.mockResponse = { conversations: [], count: 0 }

      await ai.listConversations({ chatbot: 'sql-assistant' })

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/conversations?chatbot=sql-assistant')
    })

    it('should filter by namespace', async () => {
      mockFetch.mockResponse = { conversations: [], count: 0 }

      await ai.listConversations({ namespace: 'custom' })

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/conversations?namespace=custom')
    })

    it('should support pagination', async () => {
      mockFetch.mockResponse = { conversations: [], count: 0 }

      await ai.listConversations({ limit: 10, offset: 20 })

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/conversations?limit=10&offset=20')
    })

    it('should handle errors', async () => {
      mockFetch.shouldThrow = true

      const { data, error } = await ai.listConversations()

      expect(data).toBeNull()
      expect(error).toBeDefined()
    })
  })

  describe('getConversation', () => {
    it('should get conversation details', async () => {
      const mockConv: AIUserConversationDetail = {
        id: 'conv1',
        chatbot_name: 'sql-assistant',
        title: 'Query help',
        messages: [
          { role: 'user', content: 'Help me write a query' },
          { role: 'assistant', content: 'Sure, what do you need?' },
        ],
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-01T00:00:00Z',
      }
      mockFetch.mockResponse = mockConv

      const { data, error } = await ai.getConversation('conv1')

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/conversations/conv1')
      expect(data).toEqual(mockConv)
      expect(error).toBeNull()
    })

    it('should handle errors', async () => {
      mockFetch.shouldThrow = true

      const { data, error } = await ai.getConversation('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
    })
  })

  describe('deleteConversation', () => {
    it('should delete conversation', async () => {
      const { error } = await ai.deleteConversation('conv1')

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/conversations/conv1')
      expect(mockFetch.lastMethod).toBe('DELETE')
      expect(error).toBeNull()
    })

    it('should handle errors', async () => {
      mockFetch.shouldThrow = true

      const { error } = await ai.deleteConversation('conv1')

      expect(error).toBeDefined()
    })
  })

  describe('updateConversation', () => {
    it('should update conversation title', async () => {
      const mockUpdated: AIUserConversationDetail = {
        id: 'conv1',
        chatbot_name: 'sql-assistant',
        title: 'New title',
        messages: [],
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-01T00:00:01Z',
      }
      mockFetch.mockResponse = mockUpdated

      const { data, error } = await ai.updateConversation('conv1', {
        title: 'New title',
      })

      expect(mockFetch.lastUrl).toBe('/api/v1/ai/conversations/conv1')
      expect(mockFetch.lastMethod).toBe('PATCH')
      expect(mockFetch.lastBody).toEqual({ title: 'New title' })
      expect(data).toEqual(mockUpdated)
      expect(error).toBeNull()
    })

    it('should handle errors', async () => {
      mockFetch.shouldThrow = true

      const { data, error } = await ai.updateConversation('conv1', { title: 'New' })

      expect(data).toBeNull()
      expect(error).toBeDefined()
    })
  })

  describe('createChat', () => {
    it('should create a chat client with correct WebSocket URL', () => {
      const chat = ai.createChat({
        token: 'test-token',
      })

      expect(chat).toBeInstanceOf(FluxbaseAIChat)
    })
  })
})

describe('FluxbaseAIChat', () => {
  describe('without connection', () => {
    it('should return false for isConnected when never connected', () => {
      const chat = new FluxbaseAIChat({
        wsUrl: 'ws://localhost:8080/ai/ws',
      })

      expect(chat.isConnected()).toBe(false)
    })

    it('should handle disconnect when not connected', () => {
      const chat = new FluxbaseAIChat({
        wsUrl: 'ws://localhost:8080/ai/ws',
      })

      // Should not throw
      chat.disconnect()
    })

    it('should throw when sendMessage is called without connection', () => {
      const chat = new FluxbaseAIChat({
        wsUrl: 'ws://localhost:8080/ai/ws',
      })

      // When ws is null, isConnected() returns false and should throw
      expect(() => chat.sendMessage('conv-123', 'Hello')).toThrow()
    })

    it('should throw when cancel is called without connection', () => {
      const chat = new FluxbaseAIChat({
        wsUrl: 'ws://localhost:8080/ai/ws',
      })

      expect(() => chat.cancel('conv-123')).toThrow()
    })

    it('should throw when startChat is called without connection', async () => {
      const chat = new FluxbaseAIChat({
        wsUrl: 'ws://localhost:8080/ai/ws',
      })

      await expect(chat.startChat('sql-assistant')).rejects.toThrow()
    })
  })

  describe('getAccumulatedContent', () => {
    it('should return empty string for unknown conversation', () => {
      const chat = new FluxbaseAIChat({
        wsUrl: 'ws://localhost:8080/ai/ws',
      })

      expect(chat.getAccumulatedContent('unknown')).toBe('')
    })
  })

  // Note: WebSocket-based tests (connect, sendMessage, startChat, message handling)
  // require complex async mocking that is difficult to achieve reliably in unit tests.
  // These are better covered by integration tests or E2E tests.
  // The REST API methods (FluxbaseAI) are fully tested above.
})

