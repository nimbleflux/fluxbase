/**
 * Index Exports Tests
 *
 * This test verifies that all expected exports are available from the main entry point.
 */

import { describe, it, expect } from 'vitest'
import * as FluxbaseSDK from './index'

describe('SDK Exports', () => {
  describe('Core Client', () => {
    it('should export FluxbaseClient', () => {
      expect(FluxbaseSDK.FluxbaseClient).toBeDefined()
    })

    it('should export createClient', () => {
      expect(FluxbaseSDK.createClient).toBeDefined()
      expect(typeof FluxbaseSDK.createClient).toBe('function')
    })
  })

  describe('Auth Module', () => {
    it('should export FluxbaseAuth', () => {
      expect(FluxbaseSDK.FluxbaseAuth).toBeDefined()
    })
  })

  describe('Database Query Builders', () => {
    it('should export QueryBuilder', () => {
      expect(FluxbaseSDK.QueryBuilder).toBeDefined()
    })

    it('should export SchemaQueryBuilder', () => {
      expect(FluxbaseSDK.SchemaQueryBuilder).toBeDefined()
    })
  })

  describe('Realtime Module', () => {
    it('should export FluxbaseRealtime', () => {
      expect(FluxbaseSDK.FluxbaseRealtime).toBeDefined()
    })

    it('should export RealtimeChannel', () => {
      expect(FluxbaseSDK.RealtimeChannel).toBeDefined()
    })

    it('should export ExecutionLogsChannel', () => {
      expect(FluxbaseSDK.ExecutionLogsChannel).toBeDefined()
    })
  })

  describe('Storage Module', () => {
    it('should export FluxbaseStorage', () => {
      expect(FluxbaseSDK.FluxbaseStorage).toBeDefined()
    })

    it('should export StorageBucket', () => {
      expect(FluxbaseSDK.StorageBucket).toBeDefined()
    })
  })

  describe('Functions Module', () => {
    it('should export FluxbaseFunctions', () => {
      expect(FluxbaseSDK.FluxbaseFunctions).toBeDefined()
    })

    it('should export FluxbaseAdminFunctions', () => {
      expect(FluxbaseSDK.FluxbaseAdminFunctions).toBeDefined()
    })
  })

  describe('Jobs Module', () => {
    it('should export FluxbaseJobs', () => {
      expect(FluxbaseSDK.FluxbaseJobs).toBeDefined()
    })

    it('should export FluxbaseAdminJobs', () => {
      expect(FluxbaseSDK.FluxbaseAdminJobs).toBeDefined()
    })
  })

  describe('AI Module', () => {
    it('should export FluxbaseAI', () => {
      expect(FluxbaseSDK.FluxbaseAI).toBeDefined()
    })

    it('should export FluxbaseAIChat', () => {
      expect(FluxbaseSDK.FluxbaseAIChat).toBeDefined()
    })

    it('should export FluxbaseAdminAI', () => {
      expect(FluxbaseSDK.FluxbaseAdminAI).toBeDefined()
    })
  })

  describe('Vector Search Module', () => {
    it('should export FluxbaseVector', () => {
      expect(FluxbaseSDK.FluxbaseVector).toBeDefined()
    })
  })

  describe('GraphQL Module', () => {
    it('should export FluxbaseGraphQL', () => {
      expect(FluxbaseSDK.FluxbaseGraphQL).toBeDefined()
    })
  })

  describe('Branching Module', () => {
    it('should export FluxbaseBranching', () => {
      expect(FluxbaseSDK.FluxbaseBranching).toBeDefined()
    })
  })

  describe('RPC Module', () => {
    it('should export FluxbaseRPC', () => {
      expect(FluxbaseSDK.FluxbaseRPC).toBeDefined()
    })

    it('should export FluxbaseAdminRPC', () => {
      expect(FluxbaseSDK.FluxbaseAdminRPC).toBeDefined()
    })
  })

  describe('Admin Module', () => {
    it('should export FluxbaseAdmin', () => {
      expect(FluxbaseSDK.FluxbaseAdmin).toBeDefined()
    })

    it('should export FluxbaseAdminMigrations', () => {
      expect(FluxbaseSDK.FluxbaseAdminMigrations).toBeDefined()
    })

    it('should export FluxbaseAdminStorage', () => {
      expect(FluxbaseSDK.FluxbaseAdminStorage).toBeDefined()
    })

    it('should export FluxbaseAdminRealtime', () => {
      expect(FluxbaseSDK.FluxbaseAdminRealtime).toBeDefined()
    })
  })

  describe('Management Module', () => {
    it('should export FluxbaseManagement', () => {
      expect(FluxbaseSDK.FluxbaseManagement).toBeDefined()
    })

    it('should export ClientKeysManager', () => {
      expect(FluxbaseSDK.ClientKeysManager).toBeDefined()
    })

    it('should export WebhooksManager', () => {
      expect(FluxbaseSDK.WebhooksManager).toBeDefined()
    })

    it('should export InvitationsManager', () => {
      expect(FluxbaseSDK.InvitationsManager).toBeDefined()
    })

    it('should export deprecated APIKeysManager', () => {
      expect(FluxbaseSDK.APIKeysManager).toBeDefined()
    })
  })

  describe('Settings Module', () => {
    it('should export FluxbaseSettings', () => {
      expect(FluxbaseSDK.FluxbaseSettings).toBeDefined()
    })

    it('should export SystemSettingsManager', () => {
      expect(FluxbaseSDK.SystemSettingsManager).toBeDefined()
    })

    it('should export AppSettingsManager', () => {
      expect(FluxbaseSDK.AppSettingsManager).toBeDefined()
    })

    it('should export EmailTemplateManager', () => {
      expect(FluxbaseSDK.EmailTemplateManager).toBeDefined()
    })

    it('should export EmailSettingsManager', () => {
      expect(FluxbaseSDK.EmailSettingsManager).toBeDefined()
    })

    it('should export SettingsClient', () => {
      expect(FluxbaseSDK.SettingsClient).toBeDefined()
    })
  })

  describe('DDL Module', () => {
    it('should export DDLManager', () => {
      expect(FluxbaseSDK.DDLManager).toBeDefined()
    })
  })

  describe('OAuth Module', () => {
    it('should export FluxbaseOAuth', () => {
      expect(FluxbaseSDK.FluxbaseOAuth).toBeDefined()
    })

    it('should export OAuthProviderManager', () => {
      expect(FluxbaseSDK.OAuthProviderManager).toBeDefined()
    })

    it('should export AuthSettingsManager', () => {
      expect(FluxbaseSDK.AuthSettingsManager).toBeDefined()
    })
  })

  describe('Impersonation Module', () => {
    it('should export ImpersonationManager', () => {
      expect(FluxbaseSDK.ImpersonationManager).toBeDefined()
    })
  })

  describe('HTTP Client', () => {
    it('should export FluxbaseFetch', () => {
      expect(FluxbaseSDK.FluxbaseFetch).toBeDefined()
    })
  })

  describe('Type Guards', () => {
    it('should export isFluxbaseError', () => {
      expect(FluxbaseSDK.isFluxbaseError).toBeDefined()
      expect(typeof FluxbaseSDK.isFluxbaseError).toBe('function')
    })

    it('should export isFluxbaseSuccess', () => {
      expect(FluxbaseSDK.isFluxbaseSuccess).toBeDefined()
      expect(typeof FluxbaseSDK.isFluxbaseSuccess).toBe('function')
    })

    it('should export isAuthError', () => {
      expect(FluxbaseSDK.isAuthError).toBeDefined()
      expect(typeof FluxbaseSDK.isAuthError).toBe('function')
    })

    it('should export isAuthSuccess', () => {
      expect(FluxbaseSDK.isAuthSuccess).toBeDefined()
      expect(typeof FluxbaseSDK.isAuthSuccess).toBe('function')
    })

    it('should export type checking utilities', () => {
      expect(FluxbaseSDK.isObject).toBeDefined()
      expect(FluxbaseSDK.isArray).toBeDefined()
      expect(FluxbaseSDK.isString).toBeDefined()
      expect(FluxbaseSDK.isNumber).toBeDefined()
      expect(FluxbaseSDK.isBoolean).toBeDefined()
      expect(FluxbaseSDK.assertType).toBeDefined()
    })
  })

  describe('Bundling Module', () => {
    it('should export bundleCode', () => {
      expect(FluxbaseSDK.bundleCode).toBeDefined()
      expect(typeof FluxbaseSDK.bundleCode).toBe('function')
    })

    it('should export loadImportMap', () => {
      expect(FluxbaseSDK.loadImportMap).toBeDefined()
      expect(typeof FluxbaseSDK.loadImportMap).toBe('function')
    })

    it('should export denoExternalPlugin', () => {
      expect(FluxbaseSDK.denoExternalPlugin).toBeDefined()
    })
  })
})
