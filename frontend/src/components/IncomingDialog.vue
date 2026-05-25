<template>
  <Transition name="modal">
    <div v-if="current" class="fixed inset-0 z-[60] grid place-items-center bg-black/50 backdrop-blur-sm">
      <div class="card w-[440px] max-w-[92vw] p-6 animate-fadeIn">
        <div class="flex items-start gap-3 mb-4">
          <div class="w-10 h-10 rounded-xl bg-brand-500/15 grid place-items-center text-brand-500 shrink-0">
            <Inbox class="w-5 h-5" />
          </div>
          <div class="flex-1">
            <h3 class="text-lg font-semibold">收到传输请求</h3>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{{ current.from_name }} · {{ current.from_addr }}</p>
          </div>
        </div>
        <div class="rounded-xl bg-slate-100 dark:bg-white/5 p-3 text-sm">
          <div class="flex items-center gap-2">
            <FileText class="w-4 h-4 text-slate-400" />
            <span class="font-medium truncate">{{ current.file_name }}</span>
          </div>
          <div class="text-xs text-slate-500 dark:text-slate-400 mt-1">{{ formatSize(current.total_size) }}</div>
        </div>
        <div v-if="queue.length > 1" class="text-xs text-slate-400 mt-2">还有 {{ queue.length - 1 }} 个等待处理…</div>
        <div class="flex gap-2 mt-5">
          <button @click="reject" class="btn-ghost flex-1 justify-center">拒绝</button>
          <button @click="approve" class="btn-primary flex-1 justify-center">接收</button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { Inbox, FileText } from 'lucide-vue-next'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { ApproveIncoming, RejectIncoming, PendingIncomingRequests } from '../../wailsjs/go/main/LanGiveApp'

const queue = ref([])
const current = computed(() => queue.value[0] || null)

function formatSize(b) {
  b = Number(b) || 0
  if (b < 1024) return b + ' B'
  if (b < 1024**2) return (b/1024).toFixed(1) + ' KB'
  if (b < 1024**3) return (b/1024**2).toFixed(1) + ' MB'
  return (b/1024**3).toFixed(2) + ' GB'
}

async function approve() {
  const r = current.value
  if (!r) return
  try { await ApproveIncoming(r.id) } catch (e) {}
  queue.value.shift()
}
async function reject() {
  const r = current.value
  if (!r) return
  try { await RejectIncoming(r.id) } catch (e) {}
  queue.value.shift()
}

function onIncoming(req) {
  if (!queue.value.find(q => q.id === req.id)) queue.value.push(req)
}

let poll = null
onMounted(async () => {
  EventsOn('transfer:incoming', onIncoming)
  try {
    const list = await PendingIncomingRequests()
    if (list) for (const r of list) onIncoming(r)
  } catch (e) {}
  poll = setInterval(async () => {
    try {
      const list = await PendingIncomingRequests() || []
      const ids = new Set(list.map(r => r.id))
      queue.value = queue.value.filter(q => ids.has(q.id))
      for (const r of list) onIncoming(r)
    } catch (e) {}
  }, 2000)
})
onBeforeUnmount(() => {
  EventsOff('transfer:incoming')
  if (poll) clearInterval(poll)
})
</script>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: opacity .2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
