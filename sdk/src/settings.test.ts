import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  SystemSettingsManager,
  AppSettingsManager,
  EmailTemplateManager,
  EmailSettingsManager,
  FluxbaseSettings,
  SettingsClient
} from './settings'
import type { FluxbaseFetch } from './fetch'
import type {
  SystemSetting,
  AppSettings,
  CustomSetting,
  EmailTemplate,
  EmailProviderSettings,
  SecretSettingMetadata,
  UserSetting,
  UserSettingWithSource,
} from './types'

describe('SystemSettingsManager', () => {
  let manager: SystemSettingsManager
  let mockFetch: any

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    manager = new SystemSettingsManager(mockFetch as unknown as FluxbaseFetch)
  })

  describe('list', () => {
    it('should list all system settings', async () => {
      const mockSettings: SystemSetting[] = [
        {
          id: 'setting-1',
          key: 'app.auth.enable_signup',
          value: { value: true },
          description: 'Enable user signup',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'setting-2',
          key: 'app.realtime.enabled',
          value: { value: true },
          description: 'Enable realtime features',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ]

      vi.mocked(mockFetch.get).mockResolvedValue(mockSettings)

      const result = await manager.list()

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/system/settings')
      expect(result.settings).toEqual(mockSettings)
      expect(result.settings).toHaveLength(2)
    })

    it('should handle empty settings array', async () => {
      vi.mocked(mockFetch.get).mockResolvedValue([])

      const result = await manager.list()

      expect(result.settings).toEqual([])
    })
  })

  describe('get', () => {
    it('should get a specific setting by key', async () => {
      const mockSetting: SystemSetting = {
        id: 'setting-1',
        key: 'app.auth.enable_signup',
        value: { value: true },
        description: 'Enable user signup',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      vi.mocked(mockFetch.get).mockResolvedValue(mockSetting)

      const result = await manager.get('app.auth.enable_signup')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/system/settings/app.auth.enable_signup')
      expect(result).toEqual(mockSetting)
      expect(result.key).toBe('app.auth.enable_signup')
    })

    it('should handle not found error', async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error('Setting not found'))

      await expect(manager.get('nonexistent.key')).rejects.toThrow('Setting not found')
    })
  })

  describe('update', () => {
    it('should update a setting', async () => {
      const mockSetting: SystemSetting = {
        id: 'setting-1',
        key: 'app.auth.enable_signup',
        value: { value: false },
        description: 'Enable user signup',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
      }

      vi.mocked(mockFetch.put).mockResolvedValue(mockSetting)

      const result = await manager.update('app.auth.enable_signup', {
        value: { value: false },
        description: 'Enable user signup',
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/system/settings/app.auth.enable_signup', {
        value: { value: false },
        description: 'Enable user signup',
      })
      expect(result.value.value).toBe(false)
    })

    it('should create a new setting if it does not exist', async () => {
      const mockSetting: SystemSetting = {
        id: 'setting-new',
        key: 'app.new.setting',
        value: { value: 'test' },
        description: 'New setting',
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
      }

      vi.mocked(mockFetch.put).mockResolvedValue(mockSetting)

      const result = await manager.update('app.new.setting', {
        value: { value: 'test' },
        description: 'New setting',
      })

      expect(result.key).toBe('app.new.setting')
    })
  })

  describe('delete', () => {
    it('should delete a setting', async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue(undefined)

      await manager.delete('app.auth.enable_signup')

      expect(mockFetch.delete).toHaveBeenCalledWith('/api/v1/admin/system/settings/app.auth.enable_signup')
    })

    it('should handle delete errors', async () => {
      vi.mocked(mockFetch.delete).mockRejectedValue(new Error('Setting not found'))

      await expect(manager.delete('nonexistent.key')).rejects.toThrow('Setting not found')
    })
  })
})

describe('AppSettingsManager', () => {
  let manager: AppSettingsManager
  let mockFetch: any

  const mockAppSettings: AppSettings = {
    authentication: {
      enable_signup: true,
      enable_magic_link: true,
      password_min_length: 8,
      require_email_verification: false,
    },
    features: {
      enable_realtime: true,
      enable_storage: true,
      enable_functions: true,
    },
    email: {
      enabled: false,
      provider: 'smtp',
    },
    security: {
      enable_global_rate_limit: false,
    },
  }

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    manager = new AppSettingsManager(mockFetch as unknown as FluxbaseFetch)
  })

  describe('get', () => {
    it('should get all app settings', async () => {
      vi.mocked(mockFetch.get).mockResolvedValue(mockAppSettings)

      const result = await manager.get()

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/app/settings')
      expect(result).toEqual(mockAppSettings)
      expect(result.authentication.enable_signup).toBe(true)
      expect(result.features.enable_realtime).toBe(true)
    })
  })

  describe('update', () => {
    it('should update authentication settings', async () => {
      const updatedSettings = {
        ...mockAppSettings,
        authentication: {
          ...mockAppSettings.authentication,
          enable_signup: false,
          password_min_length: 12,
        },
      }

      vi.mocked(mockFetch.put).mockResolvedValue(updatedSettings)

      const result = await manager.update({
        authentication: {
          enable_signup: false,
          password_min_length: 12,
        },
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        authentication: {
          enable_signup: false,
          password_min_length: 12,
        },
      })
      expect(result.authentication.enable_signup).toBe(false)
      expect(result.authentication.password_min_length).toBe(12)
    })

    it('should update multiple categories at once', async () => {
      const updatedSettings = {
        ...mockAppSettings,
        authentication: { ...mockAppSettings.authentication, enable_signup: false },
        features: { ...mockAppSettings.features, enable_realtime: false },
        security: { ...mockAppSettings.security, enable_global_rate_limit: true },
      }

      vi.mocked(mockFetch.put).mockResolvedValue(updatedSettings)

      const result = await manager.update({
        authentication: { enable_signup: false },
        features: { enable_realtime: false },
        security: { enable_global_rate_limit: true },
      })

      expect(result.authentication.enable_signup).toBe(false)
      expect(result.features.enable_realtime).toBe(false)
      expect(result.security.enable_global_rate_limit).toBe(true)
    })
  })

  describe('reset', () => {
    it('should reset all settings to defaults', async () => {
      vi.mocked(mockFetch.post).mockResolvedValue(mockAppSettings)

      const result = await manager.reset()

      expect(mockFetch.post).toHaveBeenCalledWith('/api/v1/admin/app/settings/reset', {})
      expect(result).toEqual(mockAppSettings)
    })
  })

  describe('convenience methods', () => {
    it('should enable signup', async () => {
      const updatedSettings = {
        ...mockAppSettings,
        authentication: { ...mockAppSettings.authentication, enable_signup: true },
      }

      vi.mocked(mockFetch.put).mockResolvedValue(updatedSettings)

      const result = await manager.enableSignup()

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        authentication: { enable_signup: true },
      })
      expect(result.authentication.enable_signup).toBe(true)
    })

    it('should disable signup', async () => {
      const updatedSettings = {
        ...mockAppSettings,
        authentication: { ...mockAppSettings.authentication, enable_signup: false },
      }

      vi.mocked(mockFetch.put).mockResolvedValue(updatedSettings)

      const result = await manager.disableSignup()

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        authentication: { enable_signup: false },
      })
      expect(result.authentication.enable_signup).toBe(false)
    })

    it('should set password min length', async () => {
      const updatedSettings = {
        ...mockAppSettings,
        authentication: { ...mockAppSettings.authentication, password_min_length: 16 },
      }

      vi.mocked(mockFetch.put).mockResolvedValue(updatedSettings)

      const result = await manager.setPasswordMinLength(16)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        authentication: { password_min_length: 16 },
      })
      expect(result.authentication.password_min_length).toBe(16)
    })

    it('should reject invalid password length', async () => {
      await expect(manager.setPasswordMinLength(7)).rejects.toThrow('Password minimum length must be between 8 and 128')
      await expect(manager.setPasswordMinLength(129)).rejects.toThrow('Password minimum length must be between 8 and 128')
    })

    it('should enable feature', async () => {
      const updatedSettings = {
        ...mockAppSettings,
        features: { ...mockAppSettings.features, enable_realtime: true },
      }

      vi.mocked(mockFetch.put).mockResolvedValue(updatedSettings)

      const result = await manager.setFeature('realtime', true)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        features: { enable_realtime: true },
      })
      expect(result.features.enable_realtime).toBe(true)
    })

    it('should disable feature', async () => {
      const updatedSettings = {
        ...mockAppSettings,
        features: { ...mockAppSettings.features, enable_storage: false },
      }

      vi.mocked(mockFetch.put).mockResolvedValue(updatedSettings)

      const result = await manager.setFeature('storage', false)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        features: { enable_storage: false },
      })
      expect(result.features.enable_storage).toBe(false)
    })

    it('should enable rate limiting', async () => {
      const updatedSettings = {
        ...mockAppSettings,
        security: { ...mockAppSettings.security, enable_global_rate_limit: true },
      }

      vi.mocked(mockFetch.put).mockResolvedValue(updatedSettings)

      const result = await manager.setRateLimiting(true)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        security: { enable_global_rate_limit: true },
      })
      expect(result.security.enable_global_rate_limit).toBe(true)
    })
  })
})

describe('AppSettingsManager - Custom Settings', () => {
  let manager: AppSettingsManager
  let mockFetch: any

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    manager = new AppSettingsManager(mockFetch as unknown as FluxbaseFetch)
  })

  describe('getSetting', () => {
    it('should get setting value without metadata', async () => {
      const mockSetting: CustomSetting = {
        id: 'setting-1',
        key: 'features.beta_enabled',
        value: { enabled: true },
        value_type: 'json',
        description: 'Beta feature toggle',
        editable_by: ['instance_admin'],
        metadata: {},
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      vi.mocked(mockFetch.get).mockResolvedValue(mockSetting)

      const result = await manager.getSetting('features.beta_enabled')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/settings/custom/features.beta_enabled')
      expect(result).toEqual({ enabled: true })
      expect(result).not.toHaveProperty('id')
    })

    it('should handle errors', async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error('Setting not found'))

      await expect(manager.getSetting('nonexistent')).rejects.toThrow('Setting not found')
    })
  })

  describe('getSettings', () => {
    it('should get multiple setting values', async () => {
      const mockSettings: CustomSetting[] = [
        {
          id: 'setting-1',
          key: 'features.beta_enabled',
          value: { enabled: true },
          value_type: 'json',
          description: 'Beta toggle',
          editable_by: ['instance_admin'],
          metadata: {},
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'setting-2',
          key: 'features.dark_mode',
          value: { enabled: false },
          value_type: 'json',
          description: 'Dark mode toggle',
          editable_by: ['instance_admin'],
          metadata: {},
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ]

      vi.mocked(mockFetch.post).mockResolvedValue(mockSettings)

      const result = await manager.getSettings(['features.beta_enabled', 'features.dark_mode'])

      expect(mockFetch.post).toHaveBeenCalledWith('/api/v1/settings/batch', {
        keys: ['features.beta_enabled', 'features.dark_mode'],
      })
      expect(result).toEqual({
        'features.beta_enabled': { enabled: true },
        'features.dark_mode': { enabled: false },
      })
    })

    it('should handle empty keys array', async () => {
      vi.mocked(mockFetch.post).mockResolvedValue([])

      const result = await manager.getSettings([])

      expect(result).toEqual({})
    })
  })

  describe('setSetting', () => {
    it('should create new setting if not found', async () => {
      const mockSetting: CustomSetting = {
        id: 'setting-1',
        key: 'billing.tiers',
        value: { free: 1000, pro: 10000 },
        value_type: 'json',
        description: 'API quotas',
        editable_by: ['instance_admin'],
        metadata: {},
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      vi.mocked(mockFetch.put).mockRejectedValue({ status: 404, message: 'not found' })
      vi.mocked(mockFetch.post).mockResolvedValue(mockSetting)

      const result = await manager.setSetting('billing.tiers', { free: 1000, pro: 10000 }, {
        description: 'API quotas'
      })

      expect(mockFetch.post).toHaveBeenCalledWith('/api/v1/admin/settings/custom', {
        key: 'billing.tiers',
        value: { free: 1000, pro: 10000 },
        value_type: 'json',
        description: 'API quotas',
        is_public: false,
        is_secret: false,
      })
      expect(result).toEqual(mockSetting)
    })
  })
})

describe('SettingsClient', () => {
  let client: SettingsClient
  let mockFetch: any

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    client = new SettingsClient(mockFetch as unknown as FluxbaseFetch)
  })

  describe('get', () => {
    it('should get public setting value', async () => {
      vi.mocked(mockFetch.get).mockResolvedValue({ value: { enabled: true } })

      const result = await client.get('features.beta_enabled')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/settings/features.beta_enabled')
      expect(result).toEqual({ enabled: true })
    })

    it('should handle settings with special characters in key', async () => {
      vi.mocked(mockFetch.get).mockResolvedValue({ value: '1.0.0' })

      const result = await client.get('public.app_version')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/settings/public.app_version')
      expect(result).toBe('1.0.0')
    })

    it('should throw error for unauthorized access', async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error('Forbidden'))

      await expect(client.get('internal.secret')).rejects.toThrow('Forbidden')
    })

    it('should throw error for non-existent setting', async () => {
      vi.mocked(mockFetch.get).mockRejectedValue(new Error('Not found'))

      await expect(client.get('nonexistent.key')).rejects.toThrow('Not found')
    })
  })

  describe('getMany', () => {
    it('should get multiple public settings', async () => {
      vi.mocked(mockFetch.post).mockResolvedValue([
        { key: 'features.beta_enabled', value: { enabled: true } },
        { key: 'features.dark_mode', value: { enabled: false } },
        { key: 'public.app_version', value: '1.0.0' },
      ])

      const result = await client.getMany([
        'features.beta_enabled',
        'features.dark_mode',
        'public.app_version',
      ])

      expect(mockFetch.post).toHaveBeenCalledWith('/api/v1/settings/batch', {
        keys: ['features.beta_enabled', 'features.dark_mode', 'public.app_version'],
      })
      expect(result).toEqual({
        'features.beta_enabled': { enabled: true },
        'features.dark_mode': { enabled: false },
        'public.app_version': '1.0.0',
      })
    })

    it('should filter unauthorized settings (RLS)', async () => {
      // Backend filters out settings user can't access
      vi.mocked(mockFetch.post).mockResolvedValue([
        { key: 'features.beta_enabled', value: { enabled: true } },
        { key: 'features.dark_mode', value: { enabled: false } },
        // 'internal.secret' is omitted by RLS
      ])

      const result = await client.getMany([
        'features.beta_enabled',
        'features.dark_mode',
        'internal.secret',
      ])

      expect(result).toEqual({
        'features.beta_enabled': { enabled: true },
        'features.dark_mode': { enabled: false },
      })
      expect(result).not.toHaveProperty('internal.secret')
    })

    it('should handle empty keys array', async () => {
      vi.mocked(mockFetch.post).mockResolvedValue([])

      const result = await client.getMany([])

      expect(result).toEqual({})
    })
  })
})

describe('AppSettingsManager - Email Configuration', () => {
  let manager: AppSettingsManager
  let mockFetch: any

  const mockAppSettings: AppSettings = {
    authentication: { enable_signup: true },
    features: { enable_realtime: true },
    email: { enabled: true, provider: 'smtp' },
    security: { enable_global_rate_limit: false },
  }

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    manager = new AppSettingsManager(mockFetch as unknown as FluxbaseFetch)
  })

  describe('configureSMTP', () => {
    it('should configure SMTP provider', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.configureSMTP({
        host: 'smtp.gmail.com',
        port: 587,
        username: 'test@gmail.com',
        password: 'app-password',
        use_tls: true,
        from_address: 'noreply@app.com',
        from_name: 'My App',
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        email: {
          enabled: true,
          provider: 'smtp',
          from_address: 'noreply@app.com',
          from_name: 'My App',
          reply_to_address: undefined,
          smtp: {
            host: 'smtp.gmail.com',
            port: 587,
            username: 'test@gmail.com',
            password: 'app-password',
            use_tls: true,
          },
        },
      })
    })
  })

  describe('configureSendGrid', () => {
    it('should configure SendGrid provider', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.configureSendGrid('SG.xxx', {
        from_address: 'noreply@app.com',
        from_name: 'My App',
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        email: {
          enabled: true,
          provider: 'sendgrid',
          from_address: 'noreply@app.com',
          from_name: 'My App',
          reply_to_address: undefined,
          sendgrid: {
            api_key: 'SG.xxx',
          },
        },
      })
    })

    it('should work without options', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.configureSendGrid('SG.xxx')

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        email: {
          enabled: true,
          provider: 'sendgrid',
          from_address: undefined,
          from_name: undefined,
          reply_to_address: undefined,
          sendgrid: {
            api_key: 'SG.xxx',
          },
        },
      })
    })
  })

  describe('configureMailgun', () => {
    it('should configure Mailgun provider', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.configureMailgun('key-xxx', 'mg.app.com', {
        eu_region: true,
        from_address: 'noreply@app.com',
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        email: {
          enabled: true,
          provider: 'mailgun',
          from_address: 'noreply@app.com',
          from_name: undefined,
          reply_to_address: undefined,
          mailgun: {
            api_key: 'key-xxx',
            domain: 'mg.app.com',
            eu_region: true,
          },
        },
      })
    })

    it('should default eu_region to false', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.configureMailgun('key-xxx', 'mg.app.com')

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', expect.objectContaining({
        email: expect.objectContaining({
          mailgun: expect.objectContaining({
            eu_region: false,
          }),
        }),
      }))
    })
  })

  describe('configureSES', () => {
    it('should configure AWS SES provider', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.configureSES('AKIAIOSFODNN7EXAMPLE', 'secret-key', 'us-east-1', {
        from_address: 'noreply@app.com',
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        email: {
          enabled: true,
          provider: 'ses',
          from_address: 'noreply@app.com',
          from_name: undefined,
          reply_to_address: undefined,
          ses: {
            access_key_id: 'AKIAIOSFODNN7EXAMPLE',
            secret_access_key: 'secret-key',
            region: 'us-east-1',
          },
        },
      })
    })
  })

  describe('setEmailEnabled', () => {
    it('should enable email', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.setEmailEnabled(true)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        email: { enabled: true },
      })
    })

    it('should disable email', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.setEmailEnabled(false)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        email: { enabled: false },
      })
    })
  })
})

describe('AppSettingsManager - Additional Methods', () => {
  let manager: AppSettingsManager
  let mockFetch: any

  const mockAppSettings: AppSettings = {
    authentication: {
      enable_signup: true,
      password_min_length: 8,
      password_require_uppercase: true,
      password_require_lowercase: true,
      password_require_number: true,
      password_require_special: true,
      session_timeout_minutes: 30,
      max_sessions_per_user: 3,
      require_email_verification: true,
    },
    features: { enable_realtime: true, enable_storage: true, enable_functions: true },
    email: { enabled: true, provider: 'smtp' },
    security: { enable_global_rate_limit: false },
  }

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    manager = new AppSettingsManager(mockFetch as unknown as FluxbaseFetch)
  })

  describe('setPasswordComplexity', () => {
    it('should set password complexity requirements', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.setPasswordComplexity({
        min_length: 12,
        require_uppercase: true,
        require_lowercase: true,
        require_number: true,
        require_special: true,
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        authentication: {
          password_min_length: 12,
          password_require_uppercase: true,
          password_require_lowercase: true,
          password_require_number: true,
          password_require_special: true,
        },
      })
    })
  })

  describe('setSessionSettings', () => {
    it('should set session timeout and max sessions', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.setSessionSettings(30, 3)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        authentication: {
          session_timeout_minutes: 30,
          max_sessions_per_user: 3,
        },
      })
    })
  })

  describe('setEmailVerificationRequired', () => {
    it('should enable email verification', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.setEmailVerificationRequired(true)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        authentication: { require_email_verification: true },
      })
    })

    it('should disable email verification', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.setEmailVerificationRequired(false)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        authentication: { require_email_verification: false },
      })
    })
  })

  describe('setFeature - functions', () => {
    it('should enable functions feature', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockAppSettings)

      await manager.setFeature('functions', true)

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/app/settings', {
        features: { enable_functions: true },
      })
    })
  })

  describe('listSettings', () => {
    it('should list all custom settings', async () => {
      const mockSettings: CustomSetting[] = [
        {
          id: '1',
          key: 'billing.tiers',
          value: { free: 1000 },
          value_type: 'json',
          description: '',
          editable_by: [],
          metadata: {},
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ]
      vi.mocked(mockFetch.get).mockResolvedValue(mockSettings)

      const result = await manager.listSettings()

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/settings/custom')
      expect(result).toEqual(mockSettings)
    })
  })

  describe('deleteSetting', () => {
    it('should delete a custom setting', async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue(undefined)

      await manager.deleteSetting('billing.tiers')

      expect(mockFetch.delete).toHaveBeenCalledWith('/api/v1/admin/settings/custom/billing.tiers')
    })
  })
})

describe('AppSettingsManager - Secret Settings', () => {
  let manager: AppSettingsManager
  let mockFetch: any

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    manager = new AppSettingsManager(mockFetch as unknown as FluxbaseFetch)
  })

  describe('setSecretSetting', () => {
    it('should update existing secret', async () => {
      const mockMetadata: SecretSettingMetadata = {
        key: 'stripe_api_key',
        description: 'Stripe key',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      vi.mocked(mockFetch.put).mockResolvedValue(mockMetadata)

      const result = await manager.setSecretSetting('stripe_api_key', 'sk-xxx', {
        description: 'Stripe key',
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/settings/custom/secret/stripe_api_key', {
        value: 'sk-xxx',
        description: 'Stripe key',
      })
      expect(result).toEqual(mockMetadata)
    })

    it('should create new secret if not found', async () => {
      const mockMetadata: SecretSettingMetadata = {
        key: 'new_secret',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      vi.mocked(mockFetch.put).mockRejectedValue({ status: 404 })
      vi.mocked(mockFetch.post).mockResolvedValue(mockMetadata)

      const result = await manager.setSecretSetting('new_secret', 'secret-value')

      expect(mockFetch.post).toHaveBeenCalledWith('/api/v1/admin/settings/custom/secret', {
        key: 'new_secret',
        value: 'secret-value',
        description: undefined,
      })
      expect(result).toEqual(mockMetadata)
    })

    it('should rethrow non-404 errors', async () => {
      vi.mocked(mockFetch.put).mockRejectedValue({ status: 500, message: 'Server error' })

      await expect(manager.setSecretSetting('key', 'value')).rejects.toEqual({ status: 500, message: 'Server error' })
    })
  })

  describe('getSecretSetting', () => {
    it('should get secret metadata', async () => {
      const mockMetadata: SecretSettingMetadata = {
        key: 'stripe_api_key',
        description: 'Stripe key',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      vi.mocked(mockFetch.get).mockResolvedValue(mockMetadata)

      const result = await manager.getSecretSetting('stripe_api_key')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/settings/custom/secret/stripe_api_key')
      expect(result).toEqual(mockMetadata)
    })
  })

  describe('listSecretSettings', () => {
    it('should list all secret settings', async () => {
      const mockSecrets: SecretSettingMetadata[] = [
        { key: 'stripe_api_key', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
        { key: 'openai_key', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
      ]
      vi.mocked(mockFetch.get).mockResolvedValue(mockSecrets)

      const result = await manager.listSecretSettings()

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/settings/custom/secrets')
      expect(result).toEqual(mockSecrets)
    })
  })

  describe('deleteSecretSetting', () => {
    it('should delete secret', async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue(undefined)

      await manager.deleteSecretSetting('stripe_api_key')

      expect(mockFetch.delete).toHaveBeenCalledWith('/api/v1/admin/settings/custom/secret/stripe_api_key')
    })
  })

  describe('getUserSecretValue', () => {
    it('should get decrypted user secret value', async () => {
      vi.mocked(mockFetch.get).mockResolvedValue({ value: 'decrypted-secret' })

      const result = await manager.getUserSecretValue('user-123', 'api_key')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/settings/user/user-123/secret/api_key/decrypt')
      expect(result).toBe('decrypted-secret')
    })
  })
})

describe('AppSettingsManager - setSetting edge cases', () => {
  let manager: AppSettingsManager
  let mockFetch: any

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    manager = new AppSettingsManager(mockFetch as unknown as FluxbaseFetch)
  })

  it('should update existing setting', async () => {
    const mockSetting: CustomSetting = {
      id: '1',
      key: 'existing.key',
      value: { updated: true },
      value_type: 'json',
      description: '',
      editable_by: [],
      metadata: {},
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    vi.mocked(mockFetch.put).mockResolvedValue(mockSetting)

    const result = await manager.setSetting('existing.key', { updated: true })

    expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/settings/custom/existing.key', {
      value: { updated: true },
      value_type: 'json',
      description: undefined,
      is_public: undefined,
      is_secret: undefined,
    })
    expect(result).toEqual(mockSetting)
  })

  it('should wrap primitive values in object', async () => {
    const mockSetting: CustomSetting = {
      id: '1',
      key: 'primitive.key',
      value: { value: 42 },
      value_type: 'json',
      description: '',
      editable_by: [],
      metadata: {},
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    vi.mocked(mockFetch.put).mockResolvedValue(mockSetting)

    await manager.setSetting('primitive.key', 42)

    expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/settings/custom/primitive.key', expect.objectContaining({
      value: { value: 42 },
    }))
  })

  it('should wrap array values in object', async () => {
    const mockSetting: CustomSetting = {
      id: '1',
      key: 'array.key',
      value: { value: [1, 2, 3] },
      value_type: 'json',
      description: '',
      editable_by: [],
      metadata: {},
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    vi.mocked(mockFetch.put).mockResolvedValue(mockSetting)

    await manager.setSetting('array.key', [1, 2, 3])

    expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/settings/custom/array.key', expect.objectContaining({
      value: { value: [1, 2, 3] },
    }))
  })

  it('should wrap null in object', async () => {
    const mockSetting: CustomSetting = {
      id: '1',
      key: 'null.key',
      value: { value: null },
      value_type: 'json',
      description: '',
      editable_by: [],
      metadata: {},
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    vi.mocked(mockFetch.put).mockResolvedValue(mockSetting)

    await manager.setSetting('null.key', null)

    expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/settings/custom/null.key', expect.objectContaining({
      value: { value: null },
    }))
  })

  it('should handle non-404 errors on setSetting', async () => {
    vi.mocked(mockFetch.put).mockRejectedValue({ status: 500, message: 'Server error' })

    await expect(manager.setSetting('key', { val: 1 })).rejects.toEqual({ status: 500, message: 'Server error' })
  })
})

describe('EmailTemplateManager', () => {
  let manager: EmailTemplateManager
  let mockFetch: any

  const mockTemplate: EmailTemplate = {
    type: 'magic_link',
    subject: 'Sign in to {{.AppName}}',
    html_body: '<a href="{{.MagicLink}}">Sign In</a>',
    text_body: 'Sign in: {{.MagicLink}}',
    is_custom: false,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  }

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    manager = new EmailTemplateManager(mockFetch as unknown as FluxbaseFetch)
  })

  describe('list', () => {
    it('should list all email templates', async () => {
      vi.mocked(mockFetch.get).mockResolvedValue([mockTemplate])

      const result = await manager.list()

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/email/templates')
      expect(result.templates).toHaveLength(1)
    })

    it('should handle non-array response', async () => {
      vi.mocked(mockFetch.get).mockResolvedValue(undefined)

      const result = await manager.list()

      expect(result.templates).toEqual([])
    })
  })

  describe('get', () => {
    it('should get template by type', async () => {
      vi.mocked(mockFetch.get).mockResolvedValue(mockTemplate)

      const result = await manager.get('magic_link')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/email/templates/magic_link')
      expect(result).toEqual(mockTemplate)
    })
  })

  describe('update', () => {
    it('should update template', async () => {
      const updated = { ...mockTemplate, subject: 'New Subject' }
      vi.mocked(mockFetch.put).mockResolvedValue(updated)

      const result = await manager.update('magic_link', {
        subject: 'New Subject',
        html_body: '<p>Custom</p>',
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/email/templates/magic_link', {
        subject: 'New Subject',
        html_body: '<p>Custom</p>',
      })
      expect(result.subject).toBe('New Subject')
    })
  })

  describe('reset', () => {
    it('should reset template to default', async () => {
      vi.mocked(mockFetch.post).mockResolvedValue(mockTemplate)

      const result = await manager.reset('magic_link')

      expect(mockFetch.post).toHaveBeenCalledWith('/api/v1/admin/email/templates/magic_link/reset', {})
      expect(result).toEqual(mockTemplate)
    })
  })

  describe('test', () => {
    it('should send test email', async () => {
      vi.mocked(mockFetch.post).mockResolvedValue(undefined)

      await manager.test('magic_link', 'test@example.com')

      expect(mockFetch.post).toHaveBeenCalledWith('/api/v1/admin/email/templates/magic_link/test', {
        recipient_email: 'test@example.com',
      })
    })
  })
})

describe('EmailSettingsManager', () => {
  let manager: EmailSettingsManager
  let mockFetch: any

  const mockSettings: EmailProviderSettings = {
    enabled: true,
    provider: 'smtp',
    from_address: 'noreply@app.com',
    from_name: 'My App',
    smtp_host: 'smtp.gmail.com',
    smtp_port: 587,
    smtp_username: 'user@gmail.com',
    smtp_password_set: true,
    smtp_tls: true,
    _overrides: {},
  }

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    manager = new EmailSettingsManager(mockFetch as unknown as FluxbaseFetch)
  })

  describe('get', () => {
    it('should get email settings', async () => {
      vi.mocked(mockFetch.get).mockResolvedValue(mockSettings)

      const result = await manager.get()

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/admin/email/settings')
      expect(result).toEqual(mockSettings)
    })
  })

  describe('update', () => {
    it('should update email settings', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue(mockSettings)

      const result = await manager.update({
        provider: 'sendgrid',
        sendgrid_api_key: 'SG.xxx',
      })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/email/settings', {
        provider: 'sendgrid',
        sendgrid_api_key: 'SG.xxx',
      })
      expect(result).toEqual(mockSettings)
    })
  })

  describe('test', () => {
    it('should send test email', async () => {
      const mockResponse = { success: true, message: 'Test email sent' }
      vi.mocked(mockFetch.post).mockResolvedValue(mockResponse)

      const result = await manager.test('admin@app.com')

      expect(mockFetch.post).toHaveBeenCalledWith('/api/v1/admin/email/settings/test', {
        recipient_email: 'admin@app.com',
      })
      expect(result).toEqual(mockResponse)
    })
  })

  describe('enable', () => {
    it('should enable email', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue({ ...mockSettings, enabled: true })

      const result = await manager.enable()

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/email/settings', { enabled: true })
      expect(result.enabled).toBe(true)
    })
  })

  describe('disable', () => {
    it('should disable email', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue({ ...mockSettings, enabled: false })

      const result = await manager.disable()

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/email/settings', { enabled: false })
      expect(result.enabled).toBe(false)
    })
  })

  describe('setProvider', () => {
    it('should set email provider', async () => {
      vi.mocked(mockFetch.put).mockResolvedValue({ ...mockSettings, provider: 'sendgrid' })

      const result = await manager.setProvider('sendgrid')

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/admin/email/settings', { provider: 'sendgrid' })
      expect(result.provider).toBe('sendgrid')
    })
  })
})

describe('SettingsClient - User Settings', () => {
  let client: SettingsClient
  let mockFetch: any

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    client = new SettingsClient(mockFetch as unknown as FluxbaseFetch)
  })

  describe('getSetting', () => {
    it('should get setting with fallback', async () => {
      const mockResult: UserSettingWithSource = {
        key: 'theme',
        value: { mode: 'dark' },
        source: 'user',
      }
      vi.mocked(mockFetch.get).mockResolvedValue(mockResult)

      const result = await client.getSetting('theme')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/settings/user/theme')
      expect(result).toEqual(mockResult)
    })
  })

  describe('getUserSetting', () => {
    it('should get user own setting', async () => {
      const mockSetting: UserSetting = {
        id: '1',
        key: 'theme',
        value: { mode: 'dark' },
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      vi.mocked(mockFetch.get).mockResolvedValue(mockSetting)

      const result = await client.getUserSetting('theme')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/settings/user/own/theme')
      expect(result).toEqual(mockSetting)
    })
  })

  describe('getSystemSetting', () => {
    it('should get system default setting', async () => {
      const mockSetting = { key: 'theme', value: { mode: 'light' } }
      vi.mocked(mockFetch.get).mockResolvedValue(mockSetting)

      const result = await client.getSystemSetting('theme')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/settings/user/system/theme')
      expect(result).toEqual(mockSetting)
    })
  })

  describe('setSetting', () => {
    it('should set user setting', async () => {
      const mockSetting: UserSetting = {
        id: '1',
        key: 'theme',
        value: { mode: 'dark' },
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      vi.mocked(mockFetch.put).mockResolvedValue(mockSetting)

      const result = await client.setSetting('theme', { mode: 'dark' }, { description: 'User theme' })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/settings/user/theme', {
        value: { mode: 'dark' },
        description: 'User theme',
      })
      expect(result).toEqual(mockSetting)
    })
  })

  describe('listSettings', () => {
    it('should list user settings', async () => {
      const mockSettings: UserSetting[] = [
        { id: '1', key: 'theme', value: { mode: 'dark' }, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
      ]
      vi.mocked(mockFetch.get).mockResolvedValue(mockSettings)

      const result = await client.listSettings()

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/settings/user/list')
      expect(result).toEqual(mockSettings)
    })
  })

  describe('deleteSetting', () => {
    it('should delete user setting', async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue(undefined)

      await client.deleteSetting('theme')

      expect(mockFetch.delete).toHaveBeenCalledWith('/api/v1/settings/user/theme')
    })
  })
})

describe('SettingsClient - User Secrets', () => {
  let client: SettingsClient
  let mockFetch: any

  beforeEach(() => {
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    }
    client = new SettingsClient(mockFetch as unknown as FluxbaseFetch)
  })

  describe('setSecret', () => {
    it('should update existing secret', async () => {
      const mockMetadata: SecretSettingMetadata = {
        key: 'openai_key',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      vi.mocked(mockFetch.put).mockResolvedValue(mockMetadata)

      const result = await client.setSecret('openai_key', 'sk-xxx', { description: 'OpenAI API key' })

      expect(mockFetch.put).toHaveBeenCalledWith('/api/v1/settings/secret/openai_key', {
        value: 'sk-xxx',
        description: 'OpenAI API key',
      })
      expect(result).toEqual(mockMetadata)
    })

    it('should create new secret if not found', async () => {
      const mockMetadata: SecretSettingMetadata = {
        key: 'new_key',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      vi.mocked(mockFetch.put).mockRejectedValue({ status: 404, message: 'not found' })
      vi.mocked(mockFetch.post).mockResolvedValue(mockMetadata)

      const result = await client.setSecret('new_key', 'secret')

      expect(mockFetch.post).toHaveBeenCalledWith('/api/v1/settings/secret/', {
        key: 'new_key',
        value: 'secret',
        description: undefined,
      })
      expect(result).toEqual(mockMetadata)
    })

    it('should rethrow non-404 errors', async () => {
      vi.mocked(mockFetch.put).mockRejectedValue({ status: 500, message: 'Server error' })

      await expect(client.setSecret('key', 'value')).rejects.toEqual({ status: 500, message: 'Server error' })
    })
  })

  describe('getSecret', () => {
    it('should get secret metadata', async () => {
      const mockMetadata: SecretSettingMetadata = {
        key: 'openai_key',
        description: 'OpenAI API key',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      vi.mocked(mockFetch.get).mockResolvedValue(mockMetadata)

      const result = await client.getSecret('openai_key')

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/settings/secret/openai_key')
      expect(result).toEqual(mockMetadata)
    })
  })

  describe('listSecrets', () => {
    it('should list all user secrets', async () => {
      const mockSecrets: SecretSettingMetadata[] = [
        { key: 'openai_key', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
        { key: 'stripe_key', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
      ]
      vi.mocked(mockFetch.get).mockResolvedValue(mockSecrets)

      const result = await client.listSecrets()

      expect(mockFetch.get).toHaveBeenCalledWith('/api/v1/settings/secret/')
      expect(result).toEqual(mockSecrets)
    })
  })

  describe('deleteSecret', () => {
    it('should delete user secret', async () => {
      vi.mocked(mockFetch.delete).mockResolvedValue(undefined)

      await client.deleteSecret('openai_key')

      expect(mockFetch.delete).toHaveBeenCalledWith('/api/v1/settings/secret/openai_key')
    })
  })
})

describe('FluxbaseSettings', () => {
  it('should initialize all managers', () => {
    const mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    } as unknown as FluxbaseFetch

    const settings = new FluxbaseSettings(mockFetch)

    expect(settings.system).toBeInstanceOf(SystemSettingsManager)
    expect(settings.app).toBeInstanceOf(AppSettingsManager)
    expect(settings.email).toBeInstanceOf(EmailSettingsManager)
  })

  describe('SystemSettingsManager - list with non-array response', () => {
    it('should handle non-array response', async () => {
      const mockFetch = {
        get: vi.fn().mockResolvedValue(null),
      } as unknown as FluxbaseFetch

      const manager = new SystemSettingsManager(mockFetch)
      const result = await manager.list()

      expect(result.settings).toEqual([])
    })
  })
})
