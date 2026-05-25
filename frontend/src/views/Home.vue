<template>
  <div class="home">
    <header class="page-header">
      <h1>欢迎使用 LanGive</h1>
      <p class="subtitle">简单快捷的局域网文件传输工具</p>
    </header>

    <div class="quick-actions">
      <div class="action-card" @click="goToDevices">
        <div class="action-icon">📤</div>
        <h3>发送文件</h3>
        <p>选择设备并发送文件或文件夹</p>
      </div>
      
      <div class="action-card" @click="openDownloadFolder">
        <div class="action-icon">📁</div>
        <h3>下载文件夹</h3>
        <p>打开接收文件的目录</p>
      </div>
    </div>

    <div class="stats-section">
      <div class="stat-card">
        <div class="stat-value">{{ onlineDevices }}</div>
        <div class="stat-label">在线设备</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ completedTransfers }}</div>
        <div class="stat-label">已完成传输</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ activeTransfers }}</div>
        <div class="stat-label">进行中</div>
      </div>
    </div>

    <div class="recent-transfers" v-if="recentTransfers.length > 0">
      <h2>最近传输</h2>
      <div class="transfer-list">
        <div
          v-for="transfer in recentTransfers"
          :key="transfer.id"
          class="transfer-item"
        >
          <div class="transfer-icon">
            {{ transfer.type === 'send' ? '📤' : '📥' }}
          </div>
          <div class="transfer-info">
            <div class="transfer-name">{{ transfer.file_name }}</div>
            <div class="transfer-meta">
              {{ formatSize(transfer.total_size) }} · {{ transfer.peer_addr }}
            </div>
          </div>
          <div class="transfer-status" :class="transfer.status">
            {{ formatStatus(transfer.status) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const onlineDevices = ref(0)
const completedTransfers = ref(0)
const activeTransfers = ref(0)
const recentTransfers = ref([])
let refreshInterval

const goToDevices = () => {
  router.push('/devices')
}

const openDownloadFolder = async () => {
  try {
    const { GetDownloadPath } = await import('../../wailsjs/go/main/App')
    const path = await GetDownloadPath()
    // 使用系统默认方式打开文件夹
  } catch (e) {
    console.error('Failed to open download folder:', e)
  }
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatStatus = (status) => {
  const statusMap = {
    'pending': '等待中',
    'transferring': '传输中',
    'completed': '已完成',
    'failed': '失败',
    'cancelled': '已取消'
  }
  return statusMap[status] || status
}

const loadData = async () => {
  try {
    const { GetDevices, GetTransfers } = await import('../../wailsjs/go/main/App')
    const devices = await GetDevices()
    const transfers = await GetTransfers()
    
    onlineDevices.value = devices.length
    completedTransfers.value = transfers.filter(t => t.status === 'completed').length
    activeTransfers.value = transfers.filter(t => t.status === 'transferring').length
    recentTransfers.value = transfers.slice(0, 5)
  } catch (e) {
    console.error('Failed to load data:', e)
  }
}

onMounted(() => {
  loadData()
  refreshInterval = setInterval(loadData, 5000)
})

onUnmounted(() => {
  clearInterval(refreshInterval)
})
</script>

<style scoped>
.home {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 1.875rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.subtitle {
  color: var(--text-secondary);
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.action-card {
  background: var(--card-bg);
  border-radius: 0.75rem;
  padding: 1.5rem;
  box-shadow: var(--shadow);
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;
}

.action-card:hover {
  border-color: var(--primary-color);
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.action-icon {
  font-size: 2.5rem;
  margin-bottom: 1rem;
}

.action-card h3 {
  font-size: 1.125rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.action-card p {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.stats-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: var(--card-bg);
  border-radius: 0.75rem;
  padding: 1.25rem;
  text-align: center;
  box-shadow: var(--shadow);
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--primary-color);
}

.stat-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-top: 0.25rem;
}

.recent-transfers h2 {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 1rem;
}

.transfer-list {
  background: var(--card-bg);
  border-radius: 0.75rem;
  box-shadow: var(--shadow);
  overflow: hidden;
}

.transfer-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.transfer-item:last-child {
  border-bottom: none;
}

.transfer-icon {
  font-size: 1.5rem;
}

.transfer-info {
  flex: 1;
}

.transfer-name {
  font-weight: 500;
  margin-bottom: 0.25rem;
}

.transfer-meta {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.transfer-status {
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.transfer-status.completed {
  background: #dcfce7;
  color: #166534;
}

.transfer-status.transferring {
  background: #dbeafe;
  color: #1e40af;
}

.transfer-status.failed {
  background: #fee2e2;
  color: #991b1b;
}
</style>
