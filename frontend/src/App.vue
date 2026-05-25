<template>
  <div class="flex h-full">
    <Sidebar />
    <main class="flex-1 overflow-y-auto">
      <div class="max-w-6xl mx-auto p-8">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>
    <IncomingDialog />
    <UpdateDialog :show="updateShow" :info="updateInfo" @close="updateShow = false" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Sidebar from './components/Sidebar.vue'
import IncomingDialog from './components/IncomingDialog.vue'
import UpdateDialog from './components/UpdateDialog.vue'
import { CheckUpdate, GetAutoUpdate } from '../wailsjs/go/main/LanGiveApp'

const updateShow = ref(false)
const updateInfo = ref({})

onMounted(async () => {
  try {
    if (!(await GetAutoUpdate())) return
    const info = await CheckUpdate()
    if (info && info.has_update) { updateInfo.value = info; updateShow.value = true }
  } catch (e) {}
})
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity .2s ease, transform .2s ease; }
.fade-enter-from { opacity: 0; transform: translateY(6px); }
.fade-leave-to { opacity: 0; transform: translateY(-6px); }
</style>
