import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { api } from '@/api'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
  },
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/movie/:uuid',
    name: 'movie-detail',
    component: () => import('@/views/MovieDetailView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/actor/:uuid',
    name: 'actor-detail',
    component: () => import('@/views/ActorDetailView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/favorites',
    name: 'favorites',
    component: () => import('@/views/FavoritesView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/history',
    name: 'history',
    component: () => import('@/views/HistoryView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/follows',
    name: 'follows',
    component: () => import('@/views/FollowsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/change-password',
    name: 'change-password',
    component: () => import('@/views/ChangePasswordView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/admin',
    name: 'admin',
    component: () => import('@/views/AdminView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !api.isLoggedIn()) {
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }
  if (to.meta.requiresAdmin) {
    var userData = null
    try {
      var stored = localStorage.getItem('metatube_user')
      if (stored) userData = JSON.parse(stored)
    } catch (e) {}
    if (!userData || userData.role !== 'admin') {
      next({ name: 'home' })
      return
    }
  }
  if (to.name === 'login' && api.isLoggedIn()) {
    next({ name: 'home' })
    return
  }
  next()
})

export default router
