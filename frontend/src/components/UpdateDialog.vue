<template>
  <Transition name="modal">
    <div v-if="show" class="fixed inset-0 z-[60] grid place-items-center bg-black/50 backdrop-blur-sm">
      <div class="card w-[480px] max-w-[92vw] p-6 animate-fadeIn">
        <div class="flex items-start gap-3 mb-4">
          <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-brand-400 to-brand-600 grid place-items-center text-white shadow-lg shadow-brand-500/30">
            <Sparkles class="w-5 h-5" />
          </div>
          <div class="flex-1">
            <h3 class="text-lg font-semibold">发现新版本</h3>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              v{{ info.current_version }} → <span class="text-brand-500 font-medium">v{{ info.latest_version }}</span>
              <span v-if="info.published_at" class="ml-2">· {{ info.published_at.slice(0,10) }}</span>
            </p>
          </div>
          <button v-if="!installing" @click="$emit('close')" class="text-slate-400 hover:text-slate-600">
            <X class="w-4 h-4" />
          </button>
        </div>

        <div class="rounded-xl bg-slate-100 dark:bg-white/5 p-3 max-h-60 overflow-y-auto text-sm whitespace-pre-line">
          <div v-if="info.release_notes" v-html="renderedNotes"></div>
          <div v-else class="text-slate-400 text-xs">无更新日志</div>
        </div>

        <div v-if="error" class="mt-3 text-sm text-rose-500">{{ error }}</div>

        <div class="flex gap-2 mt-5">
          <button @click="$emit('close')" class="btn-ghost flex-1 justify-center" :disabled="installing">稍后</button>
          <button @click="install" class="btn-primary flex-1 justify-center" :disabled="installing">
            <component :is="installing ? Loader2 : Download" class="w-4 h-4" :class="installing && 'animate-spin'" />
            {{ installing ? '安装中…' : '立即更新' }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Sparkles, X, Download, Loader2 } from 'lucide-vue-next'
import { DownloadAndInstall, Restart } from '../../wailsjs/go/main/LanGiveApp'

const props = defineProps({ show: Boolean, info: { type: Object, default: () => ({}) } })
const emit = defineEmits(['close'])

const installing = ref(false)
const error = ref('')

const renderedNotes = computed(() => {
  if (!props.info.release_notes) return ''
  return String(props.info.release_notes).replace(/<script[\s\S]*?<\/script>/gi, '')
})

async function install() {
  if (!props.info.download_url) { error.value = '当前平台暂未提供下载包'; return }
  installing.value = true
  error.value = ''
  try {
    await DownloadAndInstall(props.info.download_url)
    await Restart()
  } catch (e) {
    error.value = '更新失败：' + e
    installing.value = false
  }
}
</script>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: opacity .2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
