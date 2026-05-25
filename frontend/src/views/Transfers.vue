<template>
  <div class="transfers">
    <header class="page-header">
      <h1>传输记录</h1>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="clearCompleted">
          清除已完成
        </button>
        <button class="btn btn-primary" @click="refresh">
          刷新
        </button>
      </div>
    </header>

    <div v-if="transfers.length === 0" class="empty-state card">
      <div class="empty-state-icon">📭</div>
      <h3>暂无传输记录</h3>
      <p>开始发送或接收文件以查看传输进度</p>
    </div>

    <div v-else class="transfer-list">
      <div
        v-for="transfer in transfers"
        :key="transfer.id"
        class="transfer-card"
        :class="transfer.status"
      >
        <div class="transfer-header">
          <div class="transfer-type-icon">
            {{ transfer.type === 'send' ? '📤' : '📥' }}
          </div>
          <div class="transfer-title">
            <h4>{{ transfer.file_name }}</h4>
            <span class="transfer-peer">{{ transfer.peer_addr }}</span>
          </div>
          <div class="transfer-status-badge" :class="transfer.status">
            {{ formatStatus(transfer.status) }}
          </div>
        </div>

        <div class="transfer-progress" v-if="transfer.status === 'transferring'">
          <div class="progress-info">
            <span>{{ formatSize(transfer.sent_size) }} / {{ formatSize(transfer.total_size) }}</span>
            <span>{{ transfer.progress.toFixed(1) }}%</span>
          </div>
          <div class="progress-bar">
            <div
              class="progress-bar-fill"
              :style="{ width: transfer.progress + '%' }"
            ></div>
          </div>
        </div>

        <div class="transfer-footer">
          <span class="transfer-size">{{ formatSize(transfer.total_size) }}</span>
          <button
            v-if="transfer.status === 'transferring'"
            class="btn btn-danger btn-sm"
            @click="cancelTransfer(transfer.id)"
          >
            取消
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const transfers = ref([])
let refreshInterval

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

const loadTransfers = async () => {
  try {
    const { GetTransfers } = await import('../../wailsjs/go/main/App')
    transfers.value = await GetTransfers()
  } catch (e) {
    console.error('Failed to load transfers:', e)
  }
}

const cancelTransfer = async (id) => {
  try {
    const { CancelTransfer } = await import('../../wailsjs/go/main/App')
    await CancelTransfer(id)
    loadTransfers()
  } catch (e) {
    console.error('Failed to cancel transfer:', e)
  }
}

const clearCompleted = () => {
  transfers.value = transfers.value.filter(t => t.status === 'transferring')
}

const refresh = () => {
  loadTransfers()
}

onMounted(() => {
  loadTransfers()
  refreshInterval = setInterval(loadTransfers, 2000)
})

onUnmounted(() => {
  clearInterval(refreshInterval)
})
</script>

<style scoped>
.transfers {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 1.875rem;
  font-weight: 700;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.transfer-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.transfer-card {
  background: var(--card-bg);
  border-radius: 0.75rem;
  padding: 1.25rem;
  box-shadow: var(--shadow);
  border-left: 4px solid var(--border-color);
}

.transfer-card.transferring {
  border-left-color: var(--primary-color);
}

.transfer-card.completed {
  border-left-color: var(--secondary-color);
}

.transfer-card.failed {
  border-left-color: var(--danger-color);
}

.transfer-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.transfer-type-icon {
  font-size: 1.5rem;
}

.transfer-title {
  flex: 1;
}

.transfer-title h4 {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.transfer-peer {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.transfer-status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.transfer-status-badge.completed {
  background: #dcfce7;
  color: #166534;
}

.transfer-status-badge.transferring {
  background: #dbeafe;
  color: #1e40af;
}

.transfer-status-badge.failed {
  background: #fee2e2;
  color: #991b1b;
}

.transfer-status-badge.pending {
  background: #fef3c7;
  color: #92400e;
}

.transfer-progress {
  margin-bottom: 1rem;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.transfer-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.transfer-size {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.btn-sm {
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
}
</style>
