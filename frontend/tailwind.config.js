/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{vue,js,ts}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
      colors: {
        brand: {
          50:  '#eef4ff',
          100: '#dce7ff',
          200: '#bdd2ff',
          300: '#90b2ff',
          400: '#5d87ff',
          500: '#3a64ff',
          600: '#2848f5',
          700: '#1f37dc',
          800: '#1d30b1',
          900: '#1f2f8b',
        },
      },
      boxShadow: {
        glass: '0 8px 32px 0 rgba(31, 38, 135, 0.18)',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0', transform: 'translateY(4px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        progressGlow: {
          '0%, 100%': { opacity: '0.85' },
          '50%': { opacity: '1' },
        },
      },
      animation: {
        fadeIn: 'fadeIn 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
        progressGlow: 'progressGlow 2s ease-in-out infinite',
      },
    },
  },
  plugins: [],
}
