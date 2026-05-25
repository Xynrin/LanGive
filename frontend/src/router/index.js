import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Devices from '../views/Devices.vue'
import Transfers from '../views/Transfers.vue'
import Settings from '../views/Settings.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/devices',
    name: 'Devices',
    component: Devices
  },
  {
    path: '/transfers',
    name: 'Transfers',
    component: Transfers
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
