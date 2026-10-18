<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-content">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>
        <div>
          <h1 class="page-title">最近访问</h1>
          <p class="page-subtitle">最近浏览的影片记录</p>
        </div>
      </div>
      <button v-if="movies.length > 0" class="clear-btn" @click="clearHistory">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 6 5 6 21 6"/>
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
        </svg>
        清空记录
      </button>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="movies.length === 0" class="empty-container">
      <div class="empty-icon">
        <svg viewBox="0 0 80 80" fill="none">
          <circle cx="40" cy="40" r="35" stroke="#e2e8f0" stroke-width="2"/>
          <path d="M35 35l10 10M45 35l-10 10" stroke="#cbd5e1" stroke-width="3" stroke-linecap="round"/>
        </svg>
      </div>
      <h3 class="empty-title">暂无历史记录</h3>
      <p class="empty-text">浏览的影片将显示在这里</p>
    </div>

    <div v-else class="movie-grid">
      <div
        v-for="movie in movies"
        :key="`${movie.provider}-${movie.id}`"
        class="movie-card"
        @click="goToMovie(movie)"
      >
        <div class="movie-cover">
          <img v-if="movie.local_cover_url || movie.cover_url" :src="getImageUrl(movie, 'poster')" :alt="movie.title" @error="(e) => (e.target as HTMLImageElement).style.display = 'none'" />
          <div v-if="movie.score" class="movie-score">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
            </svg>
            {{ movie.score.toFixed(1) }}
          </div>
        </div>
        <div class="movie-info">
          <div class="movie-number">{{ movie.number }}</div>
          <div class="movie-title">{{ movie.title }}</div>
        </div>
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
import { api, getImageUrl } from '@/api'
import type { MovieDTO } from '@/types'

const router = useRouter()
const movies = ref<MovieDTO[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = 30

const totalPages = computed(() => Math.ceil(total.value / pageSize))

async function loadHistory() {
  loading.value = true
  try {
    const res = await api.getHistory(pageSize, (currentPage.value - 1) * pageSize)
    movies.value = res.data || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function changePage(page: number) {
  currentPage.value = page
  await loadHistory()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function goToMovie(movie: MovieDTO) {
  await api.ensureMovie(movie.uuid, movie.provider, movie.id)
  router.push(`/movie/${movie.uuid}`)
}

async function clearHistory() {
  if (confirm('确定要清空所有历史记录吗？')) {
    try {
      await api.clearHistory()
      movies.value = []
    } catch (e) {
      console.error(e)
    }
  }
}

onMounted(loadHistory)
</script>

<style scoped>
.page-container {
  padding: 24px 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
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
  background: linear-gradient(135deg, #6366f1 0%, #818cf8 100%);
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

.clear-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.clear-btn svg {
  width: 16px;
  height: 16px;
}

.clear-btn:hover {
  background: #fef2f2;
  border-color: #fecaca;
  color: #ef4444;
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

.movie-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 20px;
}

.movie-card {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-sm);
}

.movie-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.12);
}

.movie-cover {
  position: relative;
  aspect-ratio: 2/3;
  overflow: hidden;
  background: var(--bg-tertiary);
}

.movie-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.movie-card:hover .movie-cover img {
  transform: scale(1.05);
}

.movie-score {
  position: absolute;
  top: 12px;
  right: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(0, 0, 0, 0.75);
  color: #fbbf24;
  padding: 6px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.movie-score svg {
  width: 12px;
  height: 12px;
}

.movie-info {
  padding: 16px;
}

.movie-number {
  color: var(--primary);
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 6px;
  letter-spacing: 0.5px;
}

.movie-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
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