<template>
  <div class="space-y-6 animate-fadeIn">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">传输</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">所有发送与接收记录</p>
      </div>
      <button @click="clear" class="btn-ghost" :disabled="!hasFinished">
        <Trash2 class="w-4 h-4" /> 清除已完成
      </button>
    </header>

    <div v-if="transfers.length === 0" class="card text-center text-slate-400 py-12">
      <Inbox class="w-8 h-8 mx-auto mb-2" />
      <div>还没有任何传输</div>
    </div>

    <ul v-else class="space-y-3">
      <li v-for="t in sorted" :key="t.id" class="card">
        <div class="flex items-center gap-3">
          <component :is="t.type === 'send' ? Upload : Download" class="w-4 h-4 text-brand-500 shrink-0" />
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-medium truncate">{{ t.file_name }}</span>
              <span :class="badgeClass(t.status)">{{ statusLabel(t.status) }}</span>
            </div>
            <div class="text-xs text-slate-500 dark:text-slate-400">
              {{ t.peer_addr }} · {{ formatSize(t.sent_size) }} / {{ formatSize(t.total_size) }}
            </div>
          </div>
          <button
            v-if="t.status === 'transferring'"
            @click="cancel(t.id)"
            class="btn-danger"
          ><X class="w-4 h-4" /> 取消</button>
        </div>
        <div class="mt-3 h-2 rounded-full bg-slate-200/60 dark:bg-white/5 overflow-hidden">
          <div
            class="h-full bg-gradient-to-r from-brand-400 to-brand-600 transition-[width] duration-300"
            :class="t.status === 'transferring' ? 'animate-progressGlow' : ''"
            :style="{ width: (t.progress || 0) + '%' }"
          />
        </div>
        <div v-if="t.error" class="mt-2 text-xs text-rose-500">{{ t.error }}</div>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { Upload, Download, X, Trash2, Inbox } from 'lucide-vue-next'
import {
  GetTransfers, CancelTransfer, ClearCompletedTransfers
} from '../../wailsjs/go/main/LanGiveApp'

const transfers = ref([])
let timer = null

const sorted = computed(() => [...transfers.value].reverse())
const hasFinished = computed(() =>
  transfers.value.some(t => ['completed', 'failed', 'cancelled'].includes(t.status))
)

async function refresh() {
  try { transfers.value = await GetTransfers() || [] } catch (e) {}
}
async function cancel(id) {
  try { await CancelTransfer(id); await refresh() } catch (e) {}
}
async function clear() {
  try { await ClearCompletedTransfers(); await refresh() } catch (e) {}
}

function statusLabel(s) {
  return ({ completed:'已完成', transferring:'传输中', failed:'失败', cancelled:'已取消', paused:'已暂停', pending:'等待' })[s] || s
}
function badgeClass(s) {
  return ({
    completed:    'badge bg-emerald-500/15 text-emerald-600 dark:text-emerald-400',
    transferring: 'badge bg-brand-500/15 text-brand-600 dark:text-brand-400',
    failed:       'badge bg-rose-500/15 text-rose-600 dark:text-rose-400',
    cancelled:    'badge bg-slate-500/15 text-slate-500',
    paused:       'badge bg-amber-500/15 text-amber-600',
  })[s] || 'badge bg-slate-500/15 text-slate-500'
}
function formatSize(b) {
  b = Number(b) || 0
  if (b < 1024) return b + ' B'
  if (b < 1024**2) return (b/1024).toFixed(1) + ' KB'
  if (b < 1024**3) return (b/1024**2).toFixed(1) + ' MB'
  return (b/1024**3).toFixed(2) + ' GB'
}

onMounted(async () => { await refresh(); timer = setInterval(refresh, 1500) })
onBeforeUnmount(() => { if (timer) clearInterval(timer) })
</script>
