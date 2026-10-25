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
          dock: 'var(--mim-bg-dock)',
          list: 'var(--mim-bg-list)',
          stage: 'var(--mim-bg-stage)',
          card: 'var(--mim-bg-card)',
          border: 'var(--mim-border)',
          text: 'var(--mim-text)',
          muted: 'var(--mim-muted)',
          accent: 'var(--mim-accent)',
          'accent-hover': 'var(--mim-accent-hover)',
          'accent-subtle': 'var(--mim-accent-subtle)',
          'accent-text': 'var(--mim-accent-text)',
          'bubble-self': 'var(--mim-bubble-self)',
          'bubble-self-text': 'var(--mim-bubble-self-text)',
          'bubble-peer': 'var(--mim-bubble-peer)',
          'bubble-peer-text': 'var(--mim-bubble-peer-text)',
          'group-badge': 'var(--mim-group-badge)',
          'group-badge-text': 'var(--mim-group-badge-text)',
          hover: 'var(--mim-hover)',
          dark: {
            dock: 'var(--mim-bg-dock)',
            list: 'var(--mim-bg-list)',
            stage: 'var(--mim-bg-stage)',
            card: 'var(--mim-bg-card)',
            border: 'var(--mim-border)',
            text: 'var(--mim-text)',
            muted: 'var(--mim-muted)',
            bubble: 'var(--mim-bubble-peer)',
            accent: 'var(--mim-accent)',
            hover: 'var(--mim-hover)'
          },
          light: {
            dock: 'var(--mim-bg-dock)',
            list: 'var(--mim-bg-list)',
            stage: 'var(--mim-bg-stage)',
            card: 'var(--mim-bg-card)',
            border: 'var(--mim-border)',
            text: 'var(--mim-text)',
            muted: 'var(--mim-muted)',
            bubble: 'var(--mim-bubble-peer)',
            accent: 'var(--mim-accent)',
            hover: 'var(--mim-hover)'
          }
        }
      },
      boxShadow: {
        'bite': '0 0 0 1px var(--mim-border), 0 0 0 3px var(--mim-accent-subtle), inset 0 1px 2px rgba(0, 0, 0, 0.2)',
        'bite-light': '0 0 0 1px var(--mim-border), 0 0 0 3px var(--mim-accent-subtle), inset 0 1px 2px rgba(0, 0, 0, 0.05)',
        'card': '0 4px 20px -2px rgba(0, 0, 0, 0.12)'
      },
      transitionTimingFunction: {
        'spring-damped': 'cubic-bezier(0.2, 0, 0, 1)',
        'spring-bouncy': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
      }
    },
  },
  plugins: [],
}
