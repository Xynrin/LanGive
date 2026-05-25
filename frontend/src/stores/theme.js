import { defineStore } from 'pinia'

const KEY = 'langive.theme'

export const useThemeStore = defineStore('theme', {
  state: () => ({ dark: false }),
  actions: {
    init() {
      const saved = localStorage.getItem(KEY)
      if (saved) this.dark = saved === 'dark'
      else this.dark = window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? false
      this.apply()
    },
    toggle() {
      this.dark = !this.dark
      this.apply()
      localStorage.setItem(KEY, this.dark ? 'dark' : 'light')
    },
    apply() {
      document.documentElement.classList.toggle('dark', this.dark)
    },
  },
})
