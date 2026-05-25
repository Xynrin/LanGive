<template>
  <aside class="w-60 shrink-0 h-full p-4">
    <div class="glass rounded-2xl h-full flex flex-col p-4">
      <div class="flex items-center gap-3 px-2 pb-4 border-b border-slate-200/60 dark:border-white/10">
        <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-brand-400 to-brand-600 grid place-items-center shadow-lg shadow-brand-500/30">
          <Send class="w-5 h-5 text-white" />
        </div>
        <div>
          <div class="font-semibold tracking-tight">LanGive</div>
          <div class="text-xs text-slate-500 dark:text-slate-400">v{{ version }}</div>
        </div>
      </div>

      <nav class="flex-1 mt-4 space-y-1">
        <router-link
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          v-slot="{ isActive }"
          custom
        >
          <a
            :href="item.to"
            @click.prevent="$router.push(item.to)"
            :class="[
              'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200',
              isActive
                ? 'bg-brand-500 text-white shadow-lg shadow-brand-500/30'
                : 'text-slate-600 dark:text-slate-300 hover:bg-white/60 dark:hover:bg-white/5'
            ]"
          >
            <component :is="item.icon" class="w-4 h-4" />
            <span>{{ item.label }}</span>
          </a>
        </router-link>
      </nav>

      <button
        @click="theme.toggle()"
        class="btn-ghost justify-center"
        :title="theme.dark ? '切换到浅色' : '切换到深色'"
      >
        <component :is="theme.dark ? Sun : Moon" class="w-4 h-4" />
        <span class="text-xs">{{ theme.dark ? '浅色' : '深色' }}</span>
      </button>
    </div>
  </aside>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Home, MonitorSmartphone, ArrowDownUp, Settings, Send, Moon, Sun } from 'lucide-vue-next'
import { useThemeStore } from '../stores/theme'
import { GetVersion } from '../../wailsjs/go/main/LanGiveApp'

const theme = useThemeStore()
const version = ref('1.0.0')

const nav = [
  { to: '/',          label: '首页',  icon: Home },
  { to: '/devices',   label: '设备',  icon: MonitorSmartphone },
  { to: '/transfers', label: '传输',  icon: ArrowDownUp },
  { to: '/settings',  label: '设置',  icon: Settings },
]

onMounted(async () => {
  try { version.value = await GetVersion() } catch (e) {}
})
</script>
