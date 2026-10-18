<template>
  <div class="page-container">
    <div class="hero-section">
      <h1 class="hero-title">探索你的影片库</h1>
      <p class="hero-subtitle">输入番号或演员名称，快速找到想要的影片</p>
      <div class="search-bar">
        <SystemSelect
          v-model="selectedSystem"
          :systems="systems"
          :disabled="systems.length === 0"
        />
        <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <path d="m21 21-4.35-4.35"/>
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索番号、演员..."
          class="search-input"
          @keyup.enter="handleSearch"
        />
        <button class="search-btn" @click="handleSearch">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M5 12h14M12 5l7 7-7 7"/>
          </svg>
        </button>
      </div>
      <div v-if="store.isSystemDisabled(selectedSystem)" class="system-hint">
        {{ store.selectedSystemHint() }}
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>搜索中...</p>
    </div>

    <div v-else-if="searchResults.length === 0" class="empty-container">
      <div class="empty-illustration">
        <svg viewBox="0 0 120 120" fill="none">
          <circle cx="60" cy="60" r="50" stroke="#e2e8f0" stroke-width="2"/>
          <circle cx="60" cy="55" r="20" stroke="#cbd5e1" stroke-width="2"/>
          <path d="M75 75c0-8.284-6.716-15-15-15" stroke="#cbd5e1" stroke-width="2" stroke-linecap="round"/>
        </svg>
      </div>
      <h3 class="empty-title">开始搜索</h3>
      <p class="empty-text">输入关键字搜索影片或演员</p>
    </div>

    <div v-else class="results-section">
      <div class="results-header">
        <h2 class="results-title">找到 {{ searchResults.length }} 部影片</h2>
      </div>
      <div class="movie-grid">
        <MovieCardDirect
          v-for="movie in searchResults"
          :key="`${movie.provider}-${movie.id}`"
          :movie="movie"
          @click="goToDetail"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useMovieStore } from '@/stores/movie'
import { api } from '@/api'
import MovieCardDirect from '@/components/MovieCardDirect.vue'
import SystemSelect from '@/components/SystemSelect.vue'
import type { MovieDTO } from '@/types'

const router = useRouter()
const store = useMovieStore()
const { searchQuery, searchResults, loading, systems, selectedSystem } = storeToRefs(store)

if (systems.value.length === 0) {
  store.loadSystems()
}

async function handleSearch() {
  const query = searchQuery.value.trim()
  if (!query || query.length < 4) return
  if (store.isSystemDisabled(selectedSystem.value)) {
    alert(store.selectedSystemHint() || '该搜索源未启用')
    return
  }
  await store.searchMovies(query)
}

async function goToDetail(movie: MovieDTO) {
  await api.ensureMovie(movie.uuid, movie.provider, movie.id)
  router.push(`/movie/${movie.uuid}`)
}
</script>

<style scoped>
.page-container {
  min-height: calc(100vh - 200px);
}

.hero-section {
  text-align: center;
  padding: 48px 0 40px;
}

.hero-title {
  font-size: 36px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 12px;
  letter-spacing: -0.5px;
}

.hero-subtitle {
  font-size: 16px;
  color: var(--text-muted);
  margin-bottom: 32px;
}

.search-bar {
  display: flex;
  align-items: center;
  max-width: 760px;
  margin: 0 auto;
  background: var(--bg-primary);
  border-radius: var(--radius-xl);
  padding: 8px;
  box-shadow: var(--shadow-lg), 0 0 0 1px var(--border);
  transition: all 0.2s ease;
}

.search-bar:focus-within {
  box-shadow: var(--shadow-lg), 0 0 0 2px var(--primary);
}

.search-icon {
  width: 20px;
  height: 20px;
  color: var(--text-muted);
  margin: 0 12px;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  padding: 12px 16px;
  border: none;
  background: transparent;
  font-size: 16px;
  color: var(--text-primary);
  min-width: 0;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-input:focus {
  outline: none;
}

.system-hint {
  max-width: 760px;
  margin: 12px auto 0;
  font-size: 13px;
  color: var(--text-muted);
  text-align: center;
}

.search-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-light) 100%);
  border: none;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.search-btn svg {
  width: 18px;
  height: 18px;
  color: white;
}

.search-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(13, 148, 136, 0.4);
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

.empty-illustration svg {
  width: 120px;
  height: 120px;
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

.results-section {
  margin-top: 24px;
}

.results-header {
  margin-bottom: 24px;
}

.results-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
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

.movie-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.7) 0%, transparent 50%);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding-bottom: 16px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.movie-card:hover .movie-overlay {
  opacity: 1;
}

.view-btn {
  background: var(--bg-primary);
  color: var(--text-primary);
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
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
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}

.movie-actors {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
}

.movie-actors svg {
  width: 12px;
  height: 12px;
  flex-shrink: 0;
}
</style>