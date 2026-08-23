import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  base: '/',
  resolve: {
    alias: {
      '@workspace': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  define: {
    __RESUME_WORKSPACE__: JSON.stringify('offline-template-draft-consistency'),
    __EXPORT_NETWORK_DISABLED__: JSON.stringify(true),
  },
  build: {
    target: 'es2022',
    outDir: 'dist/resume-workspace',
    sourcemap: false,
    reportCompressedSize: true,
  },
  server: {
    host: '127.0.0.1',
    port: 5173,
    strictPort: true,
    proxy: {
      '/api/v1': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: false,
      },
    },
  },
  test: {
    environment: 'jsdom',
    environmentOptions: { jsdom: { url: 'http://resume-workspace.local/' } },
    globals: true,
  },
})
