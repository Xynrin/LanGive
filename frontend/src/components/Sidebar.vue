<template>
  <aside class="sidebar">
    <div class="logo">
      <div class="logo-icon">LG</div>
      <span class="logo-text">LanGive</span>
    </div>
    
    <nav class="nav">
      <router-link
        v-for="item in menuItems"
        :key="item.path"
        :to="item.path"
        class="nav-item"
        :class="{ active: $route.path === item.path }"
      >
        <span class="nav-icon">{{ item.icon }}</span>
        <span class="nav-text">{{ item.name }}</span>
      </router-link>
    </nav>
    
    <div class="device-info">
      <div class="device-name">{{ deviceName }}</div>
      <div class="device-status">
        <span class="status-dot online"></span>
        在线
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const deviceName = ref('')

const menuItems = [
  { path: '/', name: '首页', icon: '🏠' },
  { path: '/devices', name: '设备', icon: '💻' },
  { path: '/transfers', name: '传输', icon: '📤' },
  { path: '/settings', name: '设置', icon: '⚙️' }
]

onMounted(async () => {
  try {
    const { GetDeviceName } = await import('../../wailsjs/go/main/App')
    deviceName.value = await GetDeviceName()
  } catch (e) {
    deviceName.value = 'My Device'
  }
})
</script>

<style scoped>
.sidebar {
  width: 240px;
  background: var(--card-bg);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  padding: 1.5rem;
}

.logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 2rem;
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 1rem;
}

.logo-text {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
}

.nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 0.5rem;
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.2s;
}

.nav-item:hover {
  background: var(--bg-color);
  color: var(--text-primary);
}

.nav-item.active {
  background: var(--primary-color);
  color: white;
}

.nav-icon {
  font-size: 1.25rem;
}

.nav-text {
  font-size: 0.875rem;
  font-weight: 500;
}

.device-info {
  padding-top: 1rem;
  border-top: 1px solid var(--border-color);
}

.device-name {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.device-status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-top: 0.25rem;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.online {
  background: var(--secondary-color);
}
</style>
