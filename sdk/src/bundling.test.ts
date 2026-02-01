/**
 * Bundling Module Tests
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  loadEsbuild,
  getEsbuild,
  denoExternalPlugin,
  loadImportMap,
  bundleCode
} from './bundling'

// Mock modules
vi.mock('esbuild', () => ({
  build: vi.fn(),
}))

vi.mock('fs', () => ({
  readFileSync: vi.fn(),
}))

describe('Bundling Module', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('loadEsbuild', () => {
    it('should return true when esbuild is available', async () => {
      const result = await loadEsbuild()
      expect(result).toBe(true)
    })

    it('should return true on subsequent calls (cached)', async () => {
      const result1 = await loadEsbuild()
      const result2 = await loadEsbuild()
      expect(result1).toBe(true)
      expect(result2).toBe(true)
    })
  })

  describe('getEsbuild', () => {
    it('should return the esbuild instance after loading', async () => {
      await loadEsbuild()
      const esbuild = getEsbuild()
      expect(esbuild).toBeDefined()
    })
  })

  describe('denoExternalPlugin', () => {
    it('should have correct name', () => {
      expect(denoExternalPlugin.name).toBe('deno-external')
    })

    it('should register resolve handlers for npm: imports', () => {
      const mockOnResolve = vi.fn()
      const mockBuild = { onResolve: mockOnResolve }

      denoExternalPlugin.setup(mockBuild as any)

      expect(mockOnResolve).toHaveBeenCalledTimes(3)

      // Check npm: filter
      expect(mockOnResolve).toHaveBeenCalledWith(
        { filter: /^npm:/ },
        expect.any(Function)
      )
    })

    it('should register resolve handlers for https:// imports', () => {
      const mockOnResolve = vi.fn()
      const mockBuild = { onResolve: mockOnResolve }

      denoExternalPlugin.setup(mockBuild as any)

      // Check https:// filter
      expect(mockOnResolve).toHaveBeenCalledWith(
        { filter: /^https?:\/\// },
        expect.any(Function)
      )
    })

    it('should register resolve handlers for jsr: imports', () => {
      const mockOnResolve = vi.fn()
      const mockBuild = { onResolve: mockOnResolve }

      denoExternalPlugin.setup(mockBuild as any)

      // Check jsr: filter
      expect(mockOnResolve).toHaveBeenCalledWith(
        { filter: /^jsr:/ },
        expect.any(Function)
      )
    })

    it('should mark npm: imports as external', () => {
      let npmCallback: ((args: { path: string }) => { path: string; external: boolean }) | null = null
      const mockOnResolve = vi.fn((opts, cb) => {
        if (opts.filter.toString().includes('npm:')) {
          npmCallback = cb
        }
      })
      const mockBuild = { onResolve: mockOnResolve }

      denoExternalPlugin.setup(mockBuild as any)

      expect(npmCallback).not.toBeNull()
      const result = npmCallback!({ path: 'npm:some-package' })
      expect(result).toEqual({ path: 'npm:some-package', external: true })
    })

    it('should mark https:// imports as external', () => {
      let httpsCallback: ((args: { path: string }) => { path: string; external: boolean }) | null = null
      const mockOnResolve = vi.fn((opts, cb) => {
        if (opts.filter.toString().includes('https')) {
          httpsCallback = cb
        }
      })
      const mockBuild = { onResolve: mockOnResolve }

      denoExternalPlugin.setup(mockBuild as any)

      expect(httpsCallback).not.toBeNull()
      const result = httpsCallback!({ path: 'https://example.com/module.ts' })
      expect(result).toEqual({ path: 'https://example.com/module.ts', external: true })
    })

    it('should mark jsr: imports as external', () => {
      let jsrCallback: ((args: { path: string }) => { path: string; external: boolean }) | null = null
      const mockOnResolve = vi.fn((opts, cb) => {
        if (opts.filter.toString().includes('jsr:')) {
          jsrCallback = cb
        }
      })
      const mockBuild = { onResolve: mockOnResolve }

      denoExternalPlugin.setup(mockBuild as any)

      expect(jsrCallback).not.toBeNull()
      const result = jsrCallback!({ path: 'jsr:@scope/package' })
      expect(result).toEqual({ path: 'jsr:@scope/package', external: true })
    })
  })

  describe('loadImportMap', () => {
    it('should load import map from deno.json', async () => {
      const mockFs = await import('fs')
      vi.mocked(mockFs.readFileSync).mockReturnValue(JSON.stringify({
        imports: {
          '@lib/utils': './lib/utils.ts',
          'lodash': 'npm:lodash@4.17.21',
        }
      }))

      const result = await loadImportMap('/path/to/deno.json')

      expect(result).toEqual({
        '@lib/utils': './lib/utils.ts',
        'lodash': 'npm:lodash@4.17.21',
      })
    })

    it('should return null when imports not present', async () => {
      const mockFs = await import('fs')
      vi.mocked(mockFs.readFileSync).mockReturnValue(JSON.stringify({
        compilerOptions: {}
      }))

      const result = await loadImportMap('/path/to/deno.json')

      expect(result).toBeNull()
    })

    it('should return null and warn on parse error', async () => {
      const mockFs = await import('fs')
      vi.mocked(mockFs.readFileSync).mockReturnValue('invalid json')
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

      const result = await loadImportMap('/path/to/deno.json')

      expect(result).toBeNull()
      expect(consoleSpy).toHaveBeenCalled()
      consoleSpy.mockRestore()
    })

    it('should return null and warn when file not found', async () => {
      const mockFs = await import('fs')
      vi.mocked(mockFs.readFileSync).mockImplementation(() => {
        throw new Error('ENOENT: no such file')
      })
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

      const result = await loadImportMap('/path/to/nonexistent.json')

      expect(result).toBeNull()
      expect(consoleSpy).toHaveBeenCalled()
      consoleSpy.mockRestore()
    })
  })

  describe('bundleCode', () => {
    beforeEach(async () => {
      // Ensure esbuild is loaded
      await loadEsbuild()
    })

    it('should bundle simple code', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'bundled code' }],
        errors: [],
        warnings: [],
        metafile: undefined,
        mangleCache: undefined,
      } as any)

      const result = await bundleCode({
        code: 'export default function() { return "hello" }',
      })

      expect(result.code).toBe('bundled code')
      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          bundle: true,
          write: false,
          format: 'esm',
        })
      )
    })

    it('should apply minification when specified', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'minified code' }],
        errors: [],
        warnings: [],
      } as any)

      await bundleCode({
        code: 'export default function() { return "hello" }',
        minify: true,
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          minify: true,
        })
      )
    })

    it('should handle external modules', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'code with externals' }],
        errors: [],
        warnings: [],
      } as any)

      await bundleCode({
        code: 'import lodash from "lodash"; export default lodash',
        external: ['lodash'],
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          external: expect.arrayContaining(['lodash']),
        })
      )
    })

    it('should handle import map with npm: imports as external', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'code' }],
        errors: [],
        warnings: [],
      } as any)

      await bundleCode({
        code: 'import { something } from "@lib/pkg"',
        importMap: {
          '@lib/pkg': 'npm:some-package@1.0.0',
        },
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          external: expect.arrayContaining(['@lib/pkg']),
        })
      )
    })

    it('should handle import map with https:// imports as external', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'code' }],
        errors: [],
        warnings: [],
      } as any)

      await bundleCode({
        code: 'import { something } from "remote-lib"',
        importMap: {
          'remote-lib': 'https://deno.land/x/lib/mod.ts',
        },
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          external: expect.arrayContaining(['remote-lib']),
        })
      )
    })

    it('should handle import map with local file paths as aliases', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'code' }],
        errors: [],
        warnings: [],
      } as any)

      await bundleCode({
        code: 'import { something } from "@/utils"',
        importMap: {
          '@/utils': './src/utils.ts',
        },
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          alias: { '@/utils': './src/utils.ts' },
        })
      )
    })

    it('should set baseDir as resolveDir', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'code' }],
        errors: [],
        warnings: [],
      } as any)

      await bundleCode({
        code: 'export default 1',
        baseDir: '/custom/base/dir',
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          absWorkingDir: '/custom/base/dir',
          stdin: expect.objectContaining({
            resolveDir: '/custom/base/dir',
          }),
        })
      )
    })

    it('should pass nodePaths option', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'code' }],
        errors: [],
        warnings: [],
      } as any)

      await bundleCode({
        code: 'export default 1',
        nodePaths: ['/custom/node_modules'],
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          nodePaths: ['/custom/node_modules'],
        })
      )
    })

    it('should pass define option', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'code' }],
        errors: [],
        warnings: [],
      } as any)

      await bundleCode({
        code: 'export default process.env.NODE_ENV',
        define: { 'process.env.NODE_ENV': '"production"' },
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          define: { 'process.env.NODE_ENV': '"production"' },
        })
      )
    })

    it('should enable inline sourcemap when specified', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'code with sourcemap' }],
        errors: [],
        warnings: [],
      } as any)

      const result = await bundleCode({
        code: 'export default 1',
        sourcemap: true,
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          sourcemap: 'inline',
        })
      )
      expect(result.sourceMap).toBeDefined()
    })

    it('should throw when no output generated', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [],
        errors: [],
        warnings: [],
      } as any)

      await expect(bundleCode({
        code: 'export default 1',
      })).rejects.toThrow('Bundling failed: no output generated')
    })

    it('should handle bare specifiers in import map as external', async () => {
      const mockEsbuild = await import('esbuild')
      vi.mocked(mockEsbuild.build).mockResolvedValue({
        outputFiles: [{ text: 'code' }],
        errors: [],
        warnings: [],
      } as any)

      await bundleCode({
        code: 'import express from "express"',
        importMap: {
          'express': 'express',
        },
      })

      expect(mockEsbuild.build).toHaveBeenCalledWith(
        expect.objectContaining({
          external: expect.arrayContaining(['express']),
        })
      )
    })
  })
})
