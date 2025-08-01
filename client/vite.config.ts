import { defineConfig } from 'vite';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
    plugins: [sveltekit()],
    base: '/elevator/client/', // GitHub Pages subdirectory
    build: {
        outDir: 'dist',
        minify: 'esbuild',
        sourcemap: true,
        rollupOptions: {
            output: {
                manualChunks: {
                    vendor: ['svelte']
                }
            }
        }
    },
    server: {
        proxy: {
            '/api': {
                target: 'http://localhost:6660',
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/api/, '')
            }
        }
    }
}); 