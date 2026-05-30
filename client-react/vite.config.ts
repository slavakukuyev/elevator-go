import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'node:path';

// SPA build. Keeps the static GitHub Pages + nginx deploy intact (see north-star design §4.1).
// Base path mirrors the existing Svelte client subdirectory so the deploy story is unchanged.
export default defineConfig({
  plugins: [react()],
  base: process.env.BASE_PATH ?? '/',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5174, // distinct from the Svelte dev server (5173) so both can run side by side
    proxy: {
      // REST is reached directly via VITE_API_URL in dev; proxy kept for parity with the Svelte client.
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
