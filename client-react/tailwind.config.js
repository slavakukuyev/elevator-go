/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Semantic elevator status palette — single source of truth for the viz.
        idle: { DEFAULT: '#22c55e', soft: '#bbf7d0' },
        moving: { DEFAULT: '#3b82f6', soft: '#bfdbfe' },
        deleting: { DEFAULT: '#eab308', soft: '#fef08a' },
        fault: { DEFAULT: '#ef4444', soft: '#fecaca' },
      },
      boxShadow: {
        car: '0 4px 14px -2px rgba(0,0,0,0.25)',
      },
      transitionTimingFunction: {
        car: 'cubic-bezier(0.4, 0, 0.2, 1)',
      },
    },
  },
  plugins: [],
};
