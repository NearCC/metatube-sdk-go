<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-content">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>
          </svg>
        </div>
        <div>
          <h1 class="page-title">我的收藏</h1>
          <p class="page-subtitle">收藏的影片将显示在这里</p>
        </div>
      </div>
    </div>

    <div class="favorites-layout">
      <aside class="folder-sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">收藏夹</span>
          <button class="new-folder-btn" @click="showNewFolderInput = !showNewFolderInput">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"/>
              <line x1="5" y1="12" x2="19" y2="12"/>
            </svg>
          </button>
        </div>

        <div class="new-folder-form" v-if="showNewFolderInput">
          <input
            v-model="newFolderName"
            placeholder="收藏夹名称"
            @keyup.enter="createFolder"
            @keyup.esc="cancelNewFolder"
            ref="newFolderInput"
          />
          <div class="form-actions">
            <button class="confirm-btn" @click="createFolder">创建</button>
            <button class="cancel-btn" @click="cancelNewFolder">取消</button>
          </div>
        </div>

        <div class="folder-list">
          <div
            v-for="folder in folders"
            :key="folder.id"
            class="folder-item"
            :class="{ active: selectedFolderId === folder.id }"
            @click="selectFolder(folder.id)"
          >
            <div class="folder-info">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
              </svg>
              <span class="folder-name">{{ folder.name }}</span>
            </div>
            <button class="delete-btn" @click.stop="deleteFolder(folder.id)" title="删除收藏夹">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="3 6 5 6 21 6"/>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
              </svg>
            </button>
          </div>
        </div>
      </aside>

      <main class="movie-content">
        <div v-if="loading" class="loading-container">
          <div class="loading-spinner"></div>
          <p>加载中...</p>
        </div>

        <div v-else-if="movies.length === 0" class="empty-container">
          <div class="empty-icon">
            <svg viewBox="0 0 80 80" fill="none">
              <circle cx="40" cy="40" r="35" stroke="#e2e8f0" stroke-width="2"/>
              <path d="M28 38l8 8 16-16" stroke="#cbd5e1" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <h3 class="empty-title">该收藏夹暂无影片</h3>
          <p class="empty-text">在影片详情页收藏到此文件夹</p>
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
              <div class="unfavorite-overlay" @click.stop="unfavoriteMovie(movie)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
                取消收藏
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
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, getImageUrl } from '@/api'
import type { MovieDTO, FavoriteFolderDTO } from '@/types'

const router = useRouter()
const movies = ref<MovieDTO[]>([])
const folders = ref<FavoriteFolderDTO[]>([])
const selectedFolderId = ref<number>(0)
const loading = ref(false)
const showNewFolderInput = ref(false)
const newFolderName = ref('')
const newFolderInput = ref<HTMLInputElement | null>(null)
const total = ref(0)
const currentPage = ref(1)
const pageSize = 30

const totalPages = computed(() => Math.ceil(total.value / pageSize))

async function loadFolders() {
  try {
    const res = await api.getFavoriteFolders()
    folders.value = res.data || []
    if (folders.value.length > 0 && selectedFolderId.value === 0) {
      selectedFolderId.value = folders.value[0].id
    }
  } catch (e) {
    console.error(e)
  }
}

async function loadFavorites() {
  if (selectedFolderId.value === 0) return
  loading.value = true
  try {
    const res = await api.getFavorites(selectedFolderId.value, pageSize, (currentPage.value - 1) * pageSize)
    movies.value = res.data || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function selectFolder(id: number) {
  selectedFolderId.value = id
  currentPage.value = 1
  await loadFavorites()
}

async function changePage(page: number) {
  currentPage.value = page
  await loadFavorites()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function createFolder() {
  const name = newFolderName.value.trim()
  if (!name) return
  try {
    const res = await api.createFavoriteFolder(name)
    folders.value.push(res.data)
    selectedFolderId.value = res.data.id
    newFolderName.value = ''
    showNewFolderInput.value = false
    await loadFavorites()
  } catch (e) {
    console.error(e)
  }
}

function cancelNewFolder() {
  showNewFolderInput.value = false
  newFolderName.value = ''
}

async function deleteFolder(id: number) {
  if (!confirm('确定要删除该收藏夹吗？')) return
  try {
    await api.deleteFavoriteFolder(id)
    const idx = folders.value.findIndex(f => f.id === id)
    if (idx !== -1) folders.value.splice(idx, 1)
    if (selectedFolderId.value === id) {
      selectedFolderId.value = folders.value.length > 0 ? folders.value[0].id : 0
      movies.value = []
    }
  } catch (e) {
    console.error(e)
  }
}

async function goToMovie(movie: MovieDTO) {
  await api.ensureMovie(movie.uuid, movie.provider, movie.id)
  router.push(`/movie/${movie.uuid}`)
}

async function unfavoriteMovie(movie: MovieDTO) {
  try {
    await api.removeFavoriteByUUID(movie.uuid, selectedFolderId.value)
    const idx = movies.value.findIndex(m => m.uuid === movie.uuid)
    if (idx !== -1) movies.value.splice(idx, 1)
  } catch (e) {
    console.error(e)
  }
}

onMounted(async () => {
  await loadFolders()
  if (selectedFolderId.value > 0) {
    await loadFavorites()
  }
})
</script>

<style scoped>
.page-container {
  padding: 24px 0;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.page-header {
  margin-bottom: 24px;
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
  background: linear-gradient(135deg, #f43f5e 0%, #fb7185 100%);
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

.favorites-layout {
  display: flex;
  gap: 24px;
  flex: 1;
  min-height: 0;
}

.folder-sidebar {
  width: 260px;
  flex-shrink: 0;
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  padding: 16px;
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 200px);
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.sidebar-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.new-folder-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: none;
  background: var(--bg-tertiary);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
}

.new-folder-btn:hover {
  background: var(--primary);
  color: white;
}

.new-folder-btn svg {
  width: 16px;
  height: 16px;
}

.new-folder-form {
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border);
}

.new-folder-form input {
  width: 100%;
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 13px;
  margin-bottom: 8px;
}

.form-actions {
  display: flex;
  gap: 6px;
}

.form-actions button {
  flex: 1;
  padding: 6px 0;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  border: none;
  transition: opacity 0.2s;
}

.form-actions button:hover {
  opacity: 0.85;
}

.confirm-btn {
  background: var(--primary);
  color: white;
}

.cancel-btn {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.folder-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.folder-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.folder-item:hover {
  background: var(--bg-tertiary);
}

.folder-item.active {
  background: var(--primary);
  color: white;
}

.folder-item.active .folder-name {
  color: white;
}

.folder-info {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.folder-info svg {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.folder-name {
  font-size: 14px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.delete-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  opacity: 0;
  transition: all 0.2s;
}

.folder-item:hover .delete-btn {
  opacity: 1;
}

.folder-item.active .delete-btn {
  color: rgba(255, 255, 255, 0.7);
}

.folder-item.active .delete-btn:hover {
  color: white;
}

.delete-btn svg {
  width: 14px;
  height: 14px;
}

.movie-content {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
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

.unfavorite-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px;
  background: #f43f5e;
  color: white;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transform: translateY(100%);
  transition: transform 0.2s;
}

.movie-card:hover .unfavorite-overlay {
  transform: translateY(0);
}

.unfavorite-overlay svg {
  width: 14px;
  height: 14px;
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

.unfavorite-btn {
  position: absolute;
  top: 12px;
  left: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  cursor: pointer;
  opacity: 0;
  transition: all 0.2s;
}

.movie-card:hover .unfavorite-btn {
  opacity: 1;
}

.unfavorite-btn:hover {
  background: rgba(244, 63, 94, 0.9);
  transform: scale(1.1);
}

.unfavorite-btn svg {
  width: 16px;
  height: 16px;
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