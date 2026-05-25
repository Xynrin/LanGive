<template>
  <div class="devices">
    <header class="page-header">
      <h1>设备列表</h1>
      <p class="subtitle">发现 {{ devices.length }} 台在线设备</p>
    </header>

    <div v-if="devices.length === 0" class="empty-state card">
      <div class="empty-state-icon">🔍</div>
      <h3>未发现设备</h3>
      <p>请确保其他设备已启动 LanGive 并连接到同一局域网</p>
    </div>

    <div v-else class="device-grid">
      <div
        v-for="device in devices"
        :key="device.id"
        class="device-card"
        @click="selectDevice(device)"
      >
        <div class="device-platform">
          {{ getPlatformIcon(device.platform) }}
        </div>
        <div class="device-info">
          <h3 class="device-name">{{ device.name }}</h3>
          <p class="device-address">{{ device.address }}</p>
        </div>
        <button class="btn btn-primary" @click.stop="sendToDevice(device)">
          发送
        </button>
      </div>
    </div>

    <!-- 发送文件对话框 -->
    <div v-if="showSendDialog" class="dialog-overlay" @click="closeDialog">
      <div class="dialog" @click.stop>
        <h3>发送文件到 {{ selectedDevice?.name }}</h3>
        <div class="dialog-content">
          <div class="file-input-area">
            <input
              type="file"
              ref="fileInput"
              multiple
              @change="handleFileSelect"
              style="display: none"
            />
            <button class="btn btn-secondary" @click="$refs.fileInput.click()">
              选择文件
            </button>
            <button class="btn btn-secondary" @click="selectFolder">
              选择文件夹
            </button>
          </div>
          <div v-if="selectedFiles.length > 0" class="selected-files">
            <p>已选择 {{ selectedFiles.length }} 个文件:</p>
            <ul>
              <li v-for="file in selectedFiles" :key="file">{{ file }}</li>
            </ul>
          </div>
        </div>
        <div class="dialog-actions">
          <button class="btn btn-secondary" @click="closeDialog">取消</button>
          <button
            class="btn btn-primary"
            @click="confirmSend"
            :disabled="selectedFiles.length === 0"
          >
            发送
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const devices = ref([])
const showSendDialog = ref(false)
const selectedDevice = ref(null)
const selectedFiles = ref([])
let refreshInterval

const getPlatformIcon = (platform) => {
  const icons = {
    'windows': '💻',
    'darwin': '🍎',
    'linux': '🐧',
    'android': '📱',
    'ios': '📱'
  }
  return icons[platform] || '💻'
}

const selectDevice = (device) => {
  selectedDevice.value = device
}

const sendToDevice = (device) => {
  selectedDevice.value = device
  showSendDialog.value = true
  selectedFiles.value = []
}

const handleFileSelect = (event) => {
  const files = Array.from(event.target.files)
  selectedFiles.value = files.map(f => f.path || f.name)
}

const selectFolder = async () => {
  // 调用后端选择文件夹
  try {
    const { SelectFolder } = await import('../../wailsjs/go/main/App')
    const folder = await SelectFolder()
    if (folder) {
      selectedFiles.value = [folder]
    }
  } catch (e) {
    console.error('Failed to select folder:', e)
  }
}

const confirmSend = async () => {
  if (!selectedDevice.value || selectedFiles.value.length === 0) return
  
  try {
    const { SendFiles, SendFolder } = await import('../../wailsjs/go/main/App')
    
    // 判断是文件还是文件夹
    const isFolder = selectedFiles.value.length === 1 && !selectedFiles.value[0].includes('.')
    
    if (isFolder) {
      await SendFolder(selectedDevice.value.id, selectedFiles.value[0])
    } else {
      await SendFiles(selectedDevice.value.id, selectedFiles.value)
    }
    
    closeDialog()
    alert('文件发送成功！')
  } catch (e) {
    alert('发送失败: ' + e.message)
  }
}

const closeDialog = () => {
  showSendDialog.value = false
  selectedDevice.value = null
  selectedFiles.value = []
}

const loadDevices = async () => {
  try {
    const { GetDevices } = await import('../../wailsjs/go/main/App')
    devices.value = await GetDevices()
  } catch (e) {
    console.error('Failed to load devices:', e)
  }
}

onMounted(() => {
  loadDevices()
  refreshInterval = setInterval(loadDevices, 5000)
})

onUnmounted(() => {
  clearInterval(refreshInterval)
})
</script>

<style scoped>
.devices {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 1.875rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.subtitle {
  color: var(--text-secondary);
}

.device-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.device-card {
  background: var(--card-bg);
  border-radius: 0.75rem;
  padding: 1.5rem;
  box-shadow: var(--shadow);
  display: flex;
  align-items: center;
  gap: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.device-card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

.device-platform {
  font-size: 2rem;
}

.device-info {
  flex: 1;
}

.device-name {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.device-address {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.dialog {
  background: var(--card-bg);
  border-radius: 0.75rem;
  padding: 1.5rem;
  width: 90%;
  max-width: 500px;
  box-shadow: var(--shadow-lg);
}

.dialog h3 {
  margin-bottom: 1rem;
}

.dialog-content {
  margin-bottom: 1.5rem;
}

.file-input-area {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.selected-files {
  background: var(--bg-color);
  padding: 1rem;
  border-radius: 0.5rem;
}

.selected-files ul {
  margin-top: 0.5rem;
  padding-left: 1.5rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
