import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import type { ServerResponse } from 'http'
import { existsSync } from 'fs'

// Helper to resolve module path (supports both local and workspace node_modules)
const resolveModule = (moduleName: string): string => {
  const localPath = path.resolve(__dirname, 'node_modules', moduleName)
  const workspacePath = path.resolve(__dirname, '../node_modules', moduleName)
  return existsSync(localPath) ? localPath : workspacePath
}

// Helper to handle proxy errors gracefully during backend restarts
const handleProxyError = (err: Error, res: ServerResponse) => {
  // eslint-disable-next-line no-console
  console.error('Proxy error:', err.message)
  if (res && !res.headersSent && res.writeHead) {
    res.writeHead(503, {
      'Content-Type': 'text/html',
      'Cache-Control': 'no-store, no-cache, must-revalidate',
    })
    res.end(`
      <html>
        <head><title>Backend Unavailable</title></head>
        <body style="font-family: system-ui; padding: 2rem; text-align: center;">
          <h1>Backend Unavailable</h1>
          <p>The backend is restarting. Retrying in 2 seconds...</p>
          <script>setTimeout(() => location.reload(), 2000)</script>
        </body>
      </html>
    `)
  }
}

// https://vite.dev/config/
export default defineConfig({
  // Always use /admin as base path for consistency between dev and prod
  base: '/admin/',
  plugins: [
    tanstackRouter({
      target: 'react',
      autoCodeSplitting: true,
    }),
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@fluxbase/sdk': path.resolve(__dirname, '../sdk/dist/index.js'),
      '@fluxbase/sdk-react': path.resolve(__dirname, '../sdk-react/dist/index.mjs'),
      // Force all React imports to use a consistent version (supports workspace setup)
      react: resolveModule('react'),
      'react-dom': resolveModule('react-dom'),
      'react/jsx-runtime': resolveModule('react/jsx-runtime'),
      'react/jsx-dev-runtime': resolveModule('react/jsx-dev-runtime'),
      // Force React Query to use a consistent version for context
      '@tanstack/react-query': resolveModule('@tanstack/react-query'),
    },
  },
  optimizeDeps: {
    exclude: ['esbuild'],
  },
  build: {
    chunkSizeWarningLimit: 800,
    rollupOptions: {
      external: ['esbuild'],
    },
  },
  server: {
    host: '0.0.0.0', // Listen on all interfaces (required for devcontainer port forwarding)
    port: 5050,
    strictPort: true, // Fail if port is already in use
    proxy: {
      // Proxy v1 storage API with special handling for file uploads
      '/api/v1/storage': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        timeout: 600000, // 10 minute timeout for large file uploads
        proxyTimeout: 600000,
        configure: (proxy) => {
          proxy.on('error', (err, _req, res) => {
            handleProxyError(err, res as ServerResponse)
          })
          proxy.on('proxyReq', (proxyReq) => {
            // Set longer timeout on the proxy request
            proxyReq.setTimeout(600000)
          })
        },
      },
      // Proxy v1 API requests to the backend
      '/api/v1': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', (err, _req, res) => {
            handleProxyError(err, res as ServerResponse)
          })
        },
      },
      // Keep non-versioned endpoints
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', (err, _req, res) => {
            handleProxyError(err, res as ServerResponse)
          })
        },
      },
      '/openapi.json': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', (err, _req, res) => {
            handleProxyError(err, res as ServerResponse)
          })
        },
      },
      '/realtime': {
        target: 'ws://localhost:8080',
        ws: true,
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', (err, _req, res) => {
            handleProxyError(err, res as ServerResponse)
          })
        },
      },
      // Proxy AI WebSocket for chatbot testing
      '/ai/ws': {
        target: 'ws://localhost:8080',
        ws: true,
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', (err, _req, res) => {
            handleProxyError(err, res as ServerResponse)
          })
        },
      },
      // Proxy dashboard auth endpoints
      '/dashboard': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', (err, _req, res) => {
            handleProxyError(err, res as ServerResponse)
          })
        },
      },
    },
  },
})
