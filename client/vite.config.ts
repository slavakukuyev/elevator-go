import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'node:path';

// SPA build. Keeps the static GitHub Pages + nginx deploy intact (see north-star design §4.1).
export default defineConfig({
  plugins: [react()],
  base: process.env.BASE_PATH ?? '/',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173, // single client dev server
    proxy: {
      // REST is reached directly via VITE_API_URL in dev; proxy kept for local convenience.
      '/api': {
        target: 'http://localhost:6660',
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api/, ''),
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
  },
});
