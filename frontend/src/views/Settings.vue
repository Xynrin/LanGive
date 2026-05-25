<template>
  <div class="settings">
    <header class="page-header">
      <h1>设置</h1>
    </header>

    <div class="settings-content">
      <!-- 设备信息 -->
      <div class="settings-section card">
        <h3>设备信息</h3>
        <div class="form-group">
          <label>设备名称</label>
          <div class="input-group">
            <input
              v-model="deviceName"
              type="text"
              class="input"
              placeholder="输入设备名称"
            />
            <button class="btn btn-primary" @click="saveDeviceName">
              保存
            </button>
          </div>
          <p class="help-text">此名称将显示在其他设备的设备列表中</p>
        </div>
        
        <div class="form-group">
          <label>设备ID</label>
          <div class="info-display">
            <code>{{ deviceUUID }}</code>
            <button class="btn btn-secondary btn-sm" @click="copyUUID">
              复制
            </button>
          </div>
          <p class="help-text">用于设备识别和连接验证</p>
        </div>
      </div>

      <!-- 隐私设置 -->
      <div class="settings-section card">
        <h3>隐私与会话</h3>
        
        <div class="form-group">
          <div class="toggle-group">
            <div class="toggle-info">
              <label>隐私模式</label>
              <p class="help-text">开启后不在公共会话中广播，仅能通过IP直接连接</p>
            </div>
            <button 
              class="toggle-switch"
              :class="{ active: privacyMode }"
              @click="togglePrivacyMode"
            >
              <span class="toggle-slider"></span>
            </button>
          </div>
        </div>

        <div class="form-group" v-if="!privacyMode">
          <label>当前会话</label>
          <div class="session-display">
            <span class="session-badge public">公共会话</span>
            <span class="help-text">所有未开启隐私模式的设备可见</span>
          </div>
        </div>

        <div class="form-group" v-else>
          <label>会话ID</label>
          <div class="info-display">
            <code>{{ sessionID }}</code>
            <button class="btn btn-secondary btn-sm" @click="copySessionID">
              复制
            </button>
          </div>
          <p class="help-text">其他设备可通过此ID直接连接到您</p>
        </div>
      </div>

      <!-- 存储设置 -->
      <div class="settings-section card">
        <h3>存储设置</h3>
        <div class="form-group">
          <label>下载路径</label>
          <div class="input-group">
            <input
              v-model="downloadPath"
              type="text"
              class="input"
              readonly
            />
            <button class="btn btn-secondary" @click="selectDownloadPath">
              更改
            </button>
          </div>
          <p class="help-text">接收的文件将保存到此目录</p>
        </div>
      </div>

      <!-- 扫描设置 -->
      <div class="settings-section card">
        <h3>扫描设置</h3>
        <div class="form-group">
          <label>扫描间隔</label>
          <div class="scan-options">
            <button 
              class="scan-option"
              :class="{ active: scanInterval === 5 }"
              @click="setScanInterval(5)"
            >
              快速 (5秒)
            </button>
            <button 
              class="scan-option"
              :class="{ active: scanInterval === 15 }"
              @click="setScanInterval(15)"
            >
              平衡 (15秒)
            </button>
            <button 
              class="scan-option"
              :class="{ active: scanInterval === 30 }"
              @click="setScanInterval(30)"
            >
              节能 (30秒)
            </button>
          </div>
          <p class="help-text">后台运行时将自动使用更长的扫描间隔</p>
        </div>
      </div>

      <!-- 更新设置 -->
      <div class="settings-section card">
        <h3>关于与更新</h3>
        <div class="about-info">
          <div class="about-item">
            <span class="about-label">当前版本</span>
            <span class="about-value">v{{ version }}</span>
          </div>
          <div class="about-item">
            <span class="about-label">自动更新</span>
            <button 
              class="toggle-switch small"
              :class="{ active: autoUpdate }"
              @click="toggleAutoUpdate"
            >
              <span class="toggle-slider"></span>
            </button>
          </div>
          <div class="about-item">
            <span class="about-label">检查更新</span>
            <button class="btn btn-secondary" @click="checkUpdate" :disabled="checking">
              {{ checking ? '检查中...' : '检查更新' }}
            </button>
          </div>
        </div>
        
        <div v-if="updateInfo" class="update-info">
          <div v-if="updateInfo.has_update" class="update-available">
            <div class="update-header">
              <span class="update-badge">发现新版本</span>
              <span class="update-version">v{{ updateInfo.version }}</span>
            </div>
            <pre class="changelog">{{ updateInfo.changelog }}</pre>
            <div class="update-actions">
              <button class="btn btn-primary" @click="downloadUpdate">
                下载更新
              </button>
              <button class="btn btn-secondary" @click="ignoreUpdate">
                忽略
              </button>
            </div>
          </div>
          <div v-else class="update-latest">
            <span class="latest-icon">✓</span>
            <span>当前已是最新版本</span>
          </div>
        </div>
      </div>

      <!-- 高级设置 -->
      <div class="settings-section card advanced">
        <h3>高级设置</h3>
        <div class="form-group">
          <label>服务端口</label>
          <div class="input-group">
            <input
              v-model.number="port"
              type="number"
              class="input"
              min="1024"
              max="65535"
            />
            <button class="btn btn-secondary" @click="savePort">
              保存
            </button>
          </div>
          <p class="help-text warning">修改端口后需要重启应用</p>
        </div>
        
        <div class="form-group">
          <div class="danger-zone">
            <label>危险操作</label>
            <button class="btn btn-danger" @click="resetConfig">
              重置所有设置
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const deviceName = ref('')
const deviceUUID = ref('')
const downloadPath = ref('')
const version = ref('')
const port = ref(5566)
const scanInterval = ref(5)
const privacyMode = ref(false)
const sessionID = ref('public')
const autoUpdate = ref(true)
const checking = ref(false)
const updateInfo = ref(null)

const loadSettings = async () => {
  try {
    const { 
      GetDeviceName, GetDeviceUUID, GetDownloadPath, 
      GetVersion, GetPort, GetScanInterval,
      GetPrivacyMode, GetSessionID, GetAutoUpdate
    } = await import('../../wailsjs/go/main/App')
    
    deviceName.value = await GetDeviceName()
    deviceUUID.value = await GetDeviceUUID()
    downloadPath.value = await GetDownloadPath()
    version.value = await GetVersion()
    port.value = await GetPort()
    scanInterval.value = await GetScanInterval()
    privacyMode.value = await GetPrivacyMode()
    sessionID.value = await GetSessionID()
    autoUpdate.value = await GetAutoUpdate()
  } catch (e) {
    console.error('Failed to load settings:', e)
  }
}

const saveDeviceName = async () => {
  try {
    const { SetDeviceName } = await import('../../wailsjs/go/main/App')
    await SetDeviceName(deviceName.value)
    showToast('设备名称已保存')
  } catch (e) {
    showToast('保存失败: ' + e.message, 'error')
  }
}

const copyUUID = () => {
  navigator.clipboard.writeText(deviceUUID.value)
  showToast('已复制设备ID')
}

const copySessionID = () => {
  navigator.clipboard.writeText(sessionID.value)
  showToast('已复制会话ID')
}

const togglePrivacyMode = async () => {
  privacyMode.value = !privacyMode.value
  try {
    const { SetPrivacyMode, GetSessionID } = await import('../../wailsjs/go/main/App')
    await SetPrivacyMode(privacyMode.value)
    sessionID.value = await GetSessionID()
    showToast(privacyMode.value ? '隐私模式已开启' : '隐私模式已关闭')
  } catch (e) {
    showToast('设置失败: ' + e.message, 'error')
  }
}

const selectDownloadPath = async () => {
  try {
    const { SelectFolder, SetDownloadPath } = await import('../../wailsjs/go/main/App')
    const path = await SelectFolder()
    if (path) {
      await SetDownloadPath(path)
      downloadPath.value = path
      showToast('下载路径已更新')
    }
  } catch (e) {
    console.error('Failed to select folder:', e)
  }
}

const setScanInterval = async (interval) => {
  scanInterval.value = interval
  try {
    const { SetScanInterval } = await import('../../wailsjs/go/main/App')
    await SetScanInterval(interval)
    showToast('扫描间隔已更新')
  } catch (e) {
    showToast('设置失败: ' + e.message, 'error')
  }
}

const toggleAutoUpdate = async () => {
  autoUpdate.value = !autoUpdate.value
  try {
    const { SetAutoUpdate } = await import('../../wailsjs/go/main/App')
    await SetAutoUpdate(autoUpdate.value)
  } catch (e) {
    console.error('Failed to set auto update:', e)
  }
}

const savePort = async () => {
  if (port.value < 1024 || port.value > 65535) {
    showToast('端口号必须在 1024-65535 之间', 'error')
    return
  }
  try {
    const { SetPort } = await import('../../wailsjs/go/main/App')
    await SetPort(port.value)
    showToast('端口已保存，请重启应用', 'warning')
  } catch (e) {
    showToast('保存失败: ' + e.message, 'error')
  }
}

const checkUpdate = async () => {
  checking.value = true
  try {
    const { CheckUpdate } = await import('../../wailsjs/go/main/App')
    updateInfo.value = await CheckUpdate()
  } catch (e) {
    showToast('检查更新失败: ' + e.message, 'error')
  } finally {
    checking.value = false
  }
}

const downloadUpdate = async () => {
  if (!updateInfo.value?.download_url) return
  
  try {
    const { DownloadAndInstall, Restart } = await import('../../wailsjs/go/main/App')
    await DownloadAndInstall(updateInfo.value.download_url)
    if (confirm('更新已下载，是否立即重启应用？')) {
      await Restart()
    }
  } catch (e) {
    showToast('下载更新失败: ' + e.message, 'error')
  }
}

const ignoreUpdate = () => {
  updateInfo.value = null
}

const resetConfig = async () => {
  if (!confirm('确定要重置所有设置吗？此操作不可撤销。')) {
    return
  }
  
  try {
    const { ResetConfig } = await import('../../wailsjs/go/main/App')
    await ResetConfig()
    showToast('设置已重置', 'warning')
    setTimeout(() => location.reload(), 1000)
  } catch (e) {
    showToast('重置失败: ' + e.message, 'error')
  }
}

const showToast = (message, type = 'success') => {
  // 简单的 toast 提示
  const toast = document.createElement('div')
  toast.className = `toast toast-${type}`
  toast.textContent = message
  document.body.appendChild(toast)
  setTimeout(() => toast.remove(), 3000)
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.settings {
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 1.875rem;
  font-weight: 700;
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.settings-section {
  padding: 1.5rem;
}

.settings-section h3 {
  font-size: 1.125rem;
  font-weight: 600;
  margin-bottom: 1.25rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-color);
}

.settings-section.advanced {
  border: 1px solid var(--danger-color);
}

.settings-section.advanced h3 {
  color: var(--danger-color);
  border-color: var(--danger-color);
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group > label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
}

.input-group {
  display: flex;
  gap: 0.5rem;
}

.input-group .input {
  flex: 1;
}

.help-text {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-top: 0.5rem;
}

.help-text.warning {
  color: var(--warning-color);
}

.info-display {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.info-display code {
  flex: 1;
  padding: 0.5rem;
  background: var(--bg-color);
  border-radius: 0.375rem;
  font-size: 0.75rem;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toggle-group {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toggle-info {
  flex: 1;
}

.toggle-info label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.25rem;
}

.toggle-switch {
  width: 48px;
  height: 24px;
  background: var(--border-color);
  border: none;
  border-radius: 12px;
  cursor: pointer;
  position: relative;
  transition: background 0.2s;
}

.toggle-switch.small {
  width: 40px;
  height: 20px;
}

.toggle-switch.active {
  background: var(--primary-color);
}

.toggle-slider {
  position: absolute;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  top: 2px;
  left: 2px;
  transition: transform 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}

.toggle-switch.small .toggle-slider {
  width: 16px;
  height: 16px;
}

.toggle-switch.active .toggle-slider {
  transform: translateX(24px);
}

.toggle-switch.small.active .toggle-slider {
  transform: translateX(20px);
}

.session-display {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.session-badge {
  display: inline-flex;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
  width: fit-content;
}

.session-badge.public {
  background: #dbeafe;
  color: #1e40af;
}

.scan-options {
  display: flex;
  gap: 0.5rem;
}

.scan-option {
  flex: 1;
  padding: 0.5rem 1rem;
  border: 1px solid var(--border-color);
  background: white;
  border-radius: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.scan-option:hover {
  border-color: var(--primary-color);
}

.scan-option.active {
  background: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}

.about-info {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.about-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.about-label {
  font-weight: 500;
}

.about-value {
  color: var(--text-secondary);
}

.update-info {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
}

.update-available {
  background: #dbeafe;
  border-radius: 0.5rem;
  padding: 1rem;
}

.update-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.update-badge {
  font-weight: 600;
  color: #1e40af;
}

.update-version {
  color: var(--primary-color);
  font-weight: 600;
}

.changelog {
  background: white;
  border-radius: 0.375rem;
  padding: 0.75rem;
  font-size: 0.75rem;
  max-height: 200px;
  overflow-y: auto;
  margin-bottom: 1rem;
  white-space: pre-wrap;
}

.update-actions {
  display: flex;
  gap: 0.5rem;
}

.update-latest {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem;
  background: #dcfce7;
  border-radius: 0.5rem;
  color: #166534;
}

.latest-icon {
  font-weight: bold;
}

.danger-zone {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: #fef2f2;
  border-radius: 0.5rem;
  border: 1px solid #fecaca;
}

.danger-zone label {
  font-weight: 600;
  color: var(--danger-color);
}

.btn-sm {
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
}

/* Toast styles */
:global(.toast) {
  position: fixed;
  bottom: 2rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 0.75rem 1.5rem;
  background: var(--text-primary);
  color: white;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  z-index: 1000;
  animation: toast-in 0.3s ease;
}

:global(.toast-success) {
  background: var(--secondary-color);
}

:global(.toast-error) {
  background: var(--danger-color);
}

:global(.toast-warning) {
  background: var(--warning-color);
}

@keyframes toast-in {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(1rem);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}
</style>
