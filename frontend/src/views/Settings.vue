<template>
  <div class="space-y-6 animate-fadeIn">
    <header>
      <h1 class="text-3xl font-bold tracking-tight">设置</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">设备身份、网络、会话与更新</p>
    </header>

    <section class="card space-y-4">
      <h2 class="text-sm font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">设备</h2>

      <div>
        <label class="label">设备名称</label>
        <div class="flex gap-2">
          <input v-model="form.name" class="input flex-1" />
          <button @click="saveName" class="btn-primary">保存</button>
        </div>
      </div>

      <div>
        <label class="label">下载路径</label>
        <div class="flex gap-2">
          <input v-model="form.downloadPath" class="input flex-1" readonly />
          <button @click="pickFolder" class="btn-ghost"><Folder class="w-4 h-4" /> 浏览…</button>
          <button @click="openDownloads" class="btn-ghost" :disabled="!form.downloadPath"><FolderOpen class="w-4 h-4" /></button>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="label">服务端口</label>
          <input v-model.number="form.port" type="number" class="input" />
        </div>
        <div>
          <label class="label">扫描间隔（秒）</label>
          <input v-model.number="form.scanInterval" type="number" class="input" />
        </div>
      </div>
      <div class="flex justify-end">
        <button @click="saveNetwork" class="btn-primary">保存网络设置</button>
      </div>
    </section>

    <section class="card space-y-4">
      <h2 class="text-sm font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">会话</h2>

      <Toggle v-model="form.privacy" label="隐私模式" desc="开启后不广播到公共会话，仅可被知晓 IP 的对端连接" @change="savePrivacy" />
      <Toggle v-model="form.autoUpdate" label="自动检查更新" desc="启动时检查 GitHub Releases 是否有新版本" @change="saveAutoUpdate" />

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs text-slate-500 dark:text-slate-400 pt-2 border-t border-slate-200/60 dark:border-white/10">
        <div><span class="font-medium">UUID：</span><span class="font-mono">{{ uuid }}</span></div>
        <div><span class="font-medium">会话：</span><span class="font-mono">{{ sessionId }}</span></div>
      </div>
    </section>

    <section class="card space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-sm font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">关于</h2>
          <p class="mt-2 text-sm">LanGive <span class="font-mono">v{{ version }}</span></p>
        </div>
        <button @click="checkUpdate" class="btn-primary" :disabled="checking">
          <RefreshCw class="w-4 h-4" :class="checking && 'animate-spin'" /> 检查更新
        </button>
      </div>
      <div v-if="update && update.has_update" class="rounded-xl border border-brand-500/30 bg-brand-500/5 p-4 text-sm">
        <div class="font-medium">发现新版本 v{{ update.latest_version }}</div>
        <div v-if="update.release_notes" class="text-slate-500 dark:text-slate-400 mt-1 text-xs whitespace-pre-line">{{ update.release_notes }}</div>
        <button @click="install" class="btn-primary mt-3" :disabled="installing">
          <Download class="w-4 h-4" /> {{ installing ? '安装中…' : '立即更新' }}
        </button>
      </div>
      <div v-else-if="update" class="text-sm text-slate-500 dark:text-slate-400">已是最新版本</div>
    </section>

    <section class="card flex items-center justify-between">
      <div>
        <div class="font-medium">重置所有设置</div>
        <div class="text-xs text-slate-500 dark:text-slate-400">恢复到默认值，UUID 保留</div>
      </div>
      <button @click="reset" class="btn-danger"><RotateCcw class="w-4 h-4" /> 重置</button>
    </section>

    <Toast :msg="toast" />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, h } from 'vue'
import { Folder, FolderOpen, RefreshCw, Download, RotateCcw } from 'lucide-vue-next'
import {
  GetDeviceName, SetDeviceName, GetDeviceUUID, GetDownloadPath, SetDownloadPath,
  GetPort, SetPort, GetScanInterval, SetScanInterval,
  GetPrivacyMode, SetPrivacyMode, GetSessionID,
  GetAutoUpdate, SetAutoUpdate, ResetConfig,
  CheckUpdate, GetVersion, DownloadAndInstall, Restart,
  SelectFolder, OpenInExplorer
} from '../../wailsjs/go/main/LanGiveApp'

const form = reactive({
  name: '', downloadPath: '', port: 5566, scanInterval: 5,
  privacy: false, autoUpdate: true,
})
const uuid = ref('')
const sessionId = ref('')
const version = ref('1.0.0')
const update = ref(null)
const checking = ref(false)
const installing = ref(false)
const toast = ref('')

function flash(msg) { toast.value = msg; setTimeout(() => toast.value = '', 2000) }

async function load() {
  form.name = await GetDeviceName()
  form.downloadPath = await GetDownloadPath()
  form.port = await GetPort()
  form.scanInterval = await GetScanInterval()
  form.privacy = await GetPrivacyMode()
  form.autoUpdate = await GetAutoUpdate()
  uuid.value = await GetDeviceUUID()
  sessionId.value = await GetSessionID()
  version.value = await GetVersion()
}

async function saveName() { await SetDeviceName(form.name); flash('已保存') }
async function pickFolder() {
  const p = await SelectFolder()
  if (p) { form.downloadPath = p; await SetDownloadPath(p); flash('已保存') }
}
async function openDownloads() { try { await OpenInExplorer(form.downloadPath) } catch (e) {} }
async function saveNetwork() {
  await SetPort(form.port); await SetScanInterval(form.scanInterval); flash('已保存')
}
async function savePrivacy() { await SetPrivacyMode(form.privacy); sessionId.value = await GetSessionID() }
async function saveAutoUpdate() { await SetAutoUpdate(form.autoUpdate) }

async function checkUpdate() {
  checking.value = true
  try { update.value = await CheckUpdate() } catch (e) { flash('检查失败') }
  finally { checking.value = false }
}
async function install() {
  if (!update.value?.download_url) return
  installing.value = true
  try { await DownloadAndInstall(update.value.download_url); await Restart() }
  catch (e) { flash('更新失败：' + e); installing.value = false }
}
async function reset() {
  if (!confirm('确认重置所有设置？')) return
  await ResetConfig(); await load(); flash('已重置')
}

onMounted(load)

// 内联小组件
const Toggle = (props, { emit }) => h('label', { class: 'flex items-start gap-4 cursor-pointer select-none' }, [
  h('input', {
    type: 'checkbox',
    checked: props.modelValue,
    onChange: (e) => { emit('update:modelValue', e.target.checked); emit('change') },
    class: 'sr-only peer',
  }),
  h('span', { class: 'mt-1 w-10 h-6 rounded-full bg-slate-300 dark:bg-white/10 peer-checked:bg-brand-500 transition-colors relative shrink-0' }, [
    h('span', { class: 'absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform peer-checked:translate-x-4', style: { transform: props.modelValue ? 'translateX(1rem)' : 'none' } })
  ]),
  h('span', { class: 'flex-1' }, [
    h('div', { class: 'text-sm font-medium' }, props.label),
    h('div', { class: 'text-xs text-slate-500 dark:text-slate-400' }, props.desc),
  ]),
])
Toggle.props = ['modelValue', 'label', 'desc']
Toggle.emits = ['update:modelValue', 'change']

const Toast = (props) => props.msg ? h('div', {
  class: 'fixed bottom-6 right-6 px-4 py-2 rounded-xl glass shadow-lg text-sm animate-fadeIn'
}, props.msg) : null
Toast.props = ['msg']
</script>
