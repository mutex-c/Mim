/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        mim: {
          dark: {
            dock: '#0b0b0a',
            list: '#151515',
            stage: '#1a1a19',
            card: '#20201e',
            border: '#262624',
            text: '#e8e8e6',
            muted: '#8e8e8a',
            bubble: '#232321',
            accent: '#3b82f6',
            hover: '#1e1e1d'
          },
          light: {
            dock: '#f5f5f3',
            list: '#ffffff',
            stage: '#fcfcfb',
            card: '#ffffff',
            border: '#eaeae8',
            text: '#171717',
            muted: '#737373',
            bubble: '#f0f0ee',
            accent: '#2563eb',
            hover: '#f5f5f3'
          }
        }
      },
      boxShadow: {
        'bite': '0 0 0 1px rgba(255, 255, 255, 0.08), 0 0 0 3px rgba(59, 130, 246, 0.25), inset 0 1px 2px rgba(0, 0, 0, 0.2)',
        'bite-light': '0 0 0 1px rgba(0, 0, 0, 0.06), 0 0 0 3px rgba(37, 99, 235, 0.2), inset 0 1px 2px rgba(0, 0, 0, 0.05)',
        'card': '0 4px 20px -2px rgba(0, 0, 0, 0.25)'
      },
      transitionTimingFunction: {
        'spring-damped': 'cubic-bezier(0.2, 0, 0, 1)',
        'spring-bouncy': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
      }
    },
  },
  plugins: [],
}
