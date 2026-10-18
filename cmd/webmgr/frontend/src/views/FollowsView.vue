<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-content">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
            <circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
          </svg>
        </div>
        <div>
          <h1 class="page-title">我的关注</h1>
          <p class="page-subtitle">已关注 {{ total }} 位演员</p>
        </div>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="actors.length === 0" class="empty-container">
      <div class="empty-icon">
        <svg viewBox="0 0 80 80" fill="none">
          <circle cx="40" cy="40" r="35" stroke="#e2e8f0" stroke-width="2"/>
          <path d="M30 40a10 10 0 1 1 20 0" stroke="#cbd5e1" stroke-width="3" stroke-linecap="round"/>
          <circle cx="40" cy="40" r="4" fill="#cbd5e1"/>
        </svg>
      </div>
      <h3 class="empty-title">暂无关注</h3>
      <p class="empty-text">在演员详情页关注演员</p>
    </div>

    <div v-else class="actor-grid">
      <div
        v-for="actor in actors"
        :key="`${actor.provider}-${actor.id}`"
        class="actor-card"
      >
        <div class="actor-avatar" @click="goToActor(actor)">
          <img v-if="actor.images?.[0]" :src="getActorImageUrl(actor)" :alt="actor.name" @error="(e) => (e.target as HTMLImageElement).style.display = 'none'" />
          <div class="unfollow-overlay" @click.stop="unfollowActor(actor)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
            取消关注
          </div>
        </div>
        <div class="actor-name">{{ actor.name }}</div>
      </div>
    </div>

    <div v-if="totalPages > 1" class="pagination">
      <button :disabled="currentPage <= 1" @click="changePage(currentPage - 1)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
      </button>
      <span>{{ currentPage }} / {{ totalPages }}</span>
      <button :disabled="currentPage >= totalPages" @click="changePage(currentPage + 1)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, getActorImageUrl } from '@/api'
import type { ActorDTO } from '@/types'

const router = useRouter()
const actors = ref<ActorDTO[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = 30

const totalPages = computed(() => Math.ceil(total.value / pageSize))

async function loadFollows() {
  loading.value = true
  try {
    const res = await api.getFollows(pageSize, (currentPage.value - 1) * pageSize)
    actors.value = res.data || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function goToActor(actor: ActorDTO) {
  router.push(`/actor/${actor.uuid}`)
}

async function unfollowActor(actor: ActorDTO) {
  try {
    await api.unfollowActorByUUID(actor.uuid)
    const idx = actors.value.findIndex(a => a.uuid === actor.uuid)
    if (idx !== -1) actors.value.splice(idx, 1)
    total.value--
  } catch (e) {
    console.error(e)
  }
}

async function changePage(page: number) {
  currentPage.value = page
  await loadFollows()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(loadFollows)
</script>

<style scoped>
.page-container {
  padding: 24px 0;
}

.page-header {
  margin-bottom: 32px;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background: linear-gradient(135deg, #8b5cf6 0%, #a78bfa 100%);
  border-radius: var(--radius-lg);
}

.header-icon svg {
  width: 28px;
  height: 28px;
  color: white;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  color: var(--text-muted);
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
}

.empty-icon svg {
  width: 80px;
  height: 80px;
  margin-bottom: 24px;
}

.empty-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.empty-text {
  color: var(--text-muted);
  font-size: 14px;
}

.actor-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 24px;
}

.actor-card {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-sm);
}

.actor-card:hover {
  transform: translateY(-6px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.12);
}

.actor-avatar {
  position: relative;
  aspect-ratio: 1;
  overflow: hidden;
  background: var(--bg-tertiary);
  cursor: pointer;
}

.unfollow-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px;
  background: #8b5cf6;
  color: white;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transform: translateY(100%);
  transition: transform 0.2s;
}

.actor-card:hover .unfollow-overlay {
  transform: translateY(0);
}

.unfollow-overlay svg {
  width: 14px;
  height: 14px;
}

.actor-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.actor-card:hover .actor-avatar img {
  transform: scale(1.05);
}

.unfollow-btn {
  position: absolute;
  top: 10px;
  left: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  cursor: pointer;
  opacity: 0;
  transition: all 0.2s;
}

.actor-card:hover .unfollow-btn {
  opacity: 1;
}

.unfollow-btn:hover {
  background: rgba(139, 92, 246, 0.9);
  transform: scale(1.1);
}

.unfollow-btn svg {
  width: 14px;
  height: 14px;
}

.actor-name {
  padding: 14px;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 32px;
  padding: 20px 0;
}

.pagination button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s;
}

.pagination button:hover:not(:disabled) {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
}

.pagination button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pagination button svg {
  width: 18px;
  height: 18px;
}

.pagination span {
  font-size: 14px;
  color: var(--text-secondary);
  min-width: 60px;
  text-align: center;
}
</style>