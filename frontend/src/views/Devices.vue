<template>
  <div class="space-y-6 animate-fadeIn">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">设备</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">局域网内的 LanGive 设备 · 每 5 秒刷新</p>
      </div>
      <button @click="refresh" class="btn-ghost"><RefreshCw class="w-4 h-4" /> 刷新</button>
    </header>

    <div v-if="loading && devices.length === 0" class="card text-center text-slate-400 py-12">扫描中…</div>
    <div v-else-if="devices.length === 0" class="card text-center text-slate-400 py-12">
      <Wifi class="w-8 h-8 mx-auto mb-2" />
      <div>暂无可发现的设备</div>
      <div class="text-xs mt-1">确认对端在同一局域网且已开启 LanGive</div>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <article
        v-for="d in devices"
        :key="d.uuid"
        class="card card-hover group"
      >
        <div class="flex items-start gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-brand-400 to-brand-600 grid place-items-center text-white shadow-md shadow-brand-500/30">
            <component :is="platformIcon(d.platform)" class="w-5 h-5" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="font-medium truncate">{{ d.name || '(未命名)' }}</div>
            <div class="text-xs text-slate-500 dark:text-slate-400 truncate">{{ d.address }}:{{ d.port }}</div>
            <div class="flex flex-wrap gap-1 mt-2">
              <span class="badge bg-slate-500/10 text-slate-500">{{ d.platform || '?' }}</span>
              <span v-if="d.privacy" class="badge bg-amber-500/15 text-amber-600">隐私</span>
              <span v-else class="badge bg-emerald-500/15 text-emerald-600">公共</span>
            </div>
          </div>
        </div>
        <button
          class="btn-primary w-full justify-center mt-4"
          @click="openSend(d)"
        >
          <Send class="w-4 h-4" /> 发送
        </button>
      </article>
    </div>

    <Transition name="modal">
      <div v-if="target" class="fixed inset-0 z-50 grid place-items-center bg-black/40 backdrop-blur-sm" @click.self="target = null">
        <div class="card w-[420px] max-w-[92vw] p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-semibold">发送到 {{ target.name }}</h3>
            <button @click="target = null" class="text-slate-400 hover:text-slate-600"><X class="w-4 h-4" /></button>
          </div>
          <div class="space-y-2">
            <button @click="sendFiles" class="btn-ghost w-full justify-start" :disabled="busy">
              <FileText class="w-4 h-4" /> 选择文件…
            </button>
            <button @click="sendFolder" class="btn-ghost w-full justify-start" :disabled="busy">
              <Folder class="w-4 h-4" /> 选择文件夹…
            </button>
          </div>
          <div v-if="error" class="mt-4 text-sm text-rose-500">{{ error }}</div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import {
  RefreshCw, Wifi, Send, X, FileText, Folder,
  Monitor, Apple, Smartphone, Laptop
} from 'lucide-vue-next'
import {
  GetPublicDevices, SendPath, SelectFiles, SelectFolder
} from '../../wailsjs/go/main/LanGiveApp'

const devices = ref([])
const loading = ref(true)
const target = ref(null)
const busy = ref(false)
const error = ref('')

let timer = null

function platformIcon(p) {
  switch ((p || '').toLowerCase()) {
    case 'darwin': return Apple
    case 'windows': return Monitor
    case 'android':
    case 'ios': return Smartphone
    default: return Laptop
  }
}

async function refresh() {
  try {
    devices.value = await GetPublicDevices() || []
  } catch (e) {}
  loading.value = false
}

function openSend(d) { target.value = d; error.value = '' }

async function sendFiles() {
  if (!target.value) return
  try {
    busy.value = true
    const files = await SelectFiles()
    if (!files || files.length === 0) { busy.value = false; return }
    for (const f of files) {
      await SendPath(target.value.uuid, f)
    }
    target.value = null
  } catch (e) { error.value = String(e) }
  finally { busy.value = false }
}

async function sendFolder() {
  if (!target.value) return
  try {
    busy.value = true
    const dir = await SelectFolder()
    if (!dir) { busy.value = false; return }
    await SendPath(target.value.uuid, dir)
    target.value = null
  } catch (e) { error.value = String(e) }
  finally { busy.value = false }
}

onMounted(async () => { await refresh(); timer = setInterval(refresh, 5000) })
onBeforeUnmount(() => { if (timer) clearInterval(timer) })
</script>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: opacity .2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
