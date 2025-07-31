import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/kit/vite';

/** @type {import('@sveltejs/kit').Config} */
const config = {
    preprocess: vitePreprocess(),

    kit: {
        adapter: adapter({
            pages: 'dist',
            assets: 'dist',
            fallback: 'index.html',
            precompress: false,
            strict: true
        }),
        prerender: {
            handleHttpError: 'warn'
        },
        paths: {
            base: process.argv.includes('dev') ? '' : process.env.BASE_PATH || '/elevator'
        }
    }
};

export default config; 