<template>
  <div class="page-container">
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="movie" class="movie-detail">
      <div class="movie-header-section">
        <h1 class="movie-title">
          <span class="movie-number-inline">{{ movie.number }}</span>
          {{ movie.title }}
        </h1>
      </div>

      <div class="movie-main">
        <div class="movie-cover" @click="openLightbox(getImageUrl(movie, 'cover'), false)">
          <img :src="getImageUrl(movie, 'cover')" :alt="movie.title" />
          <div v-if="movie.score" class="cover-score">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
            </svg>
            {{ movie.score.toFixed(1) }}
          </div>
        </div>

        <div class="movie-info">
          <Teleport to="body">
            <div v-if="showFolderDialog" class="folder-modal-overlay" @click.self="cancelAddFavorite">
              <div class="folder-modal">
                <div class="folder-modal-header">
                  <h3>添加到收藏夹</h3>
                  <button class="close-btn" @click="cancelAddFavorite">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="18" y1="6" x2="6" y2="18"/>
                      <line x1="6" y1="6" x2="18" y2="18"/>
                    </svg>
                  </button>
                </div>
                <div class="folder-modal-body">
                  <div v-if="isCreatingFolder" class="new-folder-input">
                    <input
                      v-model="newFolderName"
                      type="text"
                      placeholder="收藏夹名称"
                      @keyup.enter="createAndSelectFolder"
                    />
                    <button class="create-btn" @click="createAndSelectFolder">创建</button>
                  </div>
                  <div v-else class="folder-list">
                    <div
                      v-for="folder in folders"
                      :key="folder.id"
                      class="folder-item"
                      :class="{ selected: selectedFolderIds.includes(folder.id) }"
                      @click="toggleFolder(folder.id)"
                    >
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
                      </svg>
                      <span>{{ folder.name }}</span>
                      <svg v-if="selectedFolderIds.includes(folder.id)" class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"/>
                      </svg>
                    </div>
                    <button class="add-folder-btn" @click="isCreatingFolder = true">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <line x1="12" y1="5" x2="12" y2="19"/>
                        <line x1="5" y1="12" x2="19" y2="12"/>
                      </svg>
                      新建收藏夹
                    </button>
                  </div>
                </div>
                <div class="folder-modal-footer">
                  <button class="cancel-btn" @click="cancelAddFavorite">取消</button>
                  <button class="confirm-btn" @click="confirmAddFavorite">确定</button>
                </div>
              </div>
            </div>
          </Teleport>

          <div v-if="movie.release_date" class="info-row">
            <span class="info-label">上映日期</span>
            <span class="info-value">{{ formatDate(movie.release_date) }}</span>
          </div>

          <div v-if="movie.runtime" class="info-row">
            <span class="info-label">时长</span>
            <span class="info-value">{{ movie.runtime }} 分钟</span>
          </div>

          <div v-if="movie.series" class="info-row">
            <span class="info-label">系列</span>
            <span class="info-value">{{ movie.series }}</span>
          </div>

          <div v-if="movie.maker" class="info-row">
            <span class="info-label">厂商</span>
            <span class="info-value">{{ movie.maker }}</span>
          </div>

          <div v-if="movie.label" class="info-row">
            <span class="info-label">类别</span>
            <span class="info-value">{{ movie.label }}</span>
          </div>

          <div v-if="movie.genres?.length" class="info-section">
            <span class="info-label">类型</span>
            <div class="tag-list">
              <span v-for="genre in movie.genres" :key="genre" class="tag">
                {{ genre }}
              </span>
            </div>
          </div>

          <div v-if="movie.actors?.length" class="info-section">
            <span class="info-label">演员</span>
            <div class="actors-list">
              <span
                v-for="actor in movie.actors"
                :key="actor"
                class="actor-tag"
                @click="goToActor(actor)"
              >
                {{ actor }}
              </span>
            </div>
          </div>
        </div>
      </div>

        <div class="action-row">
          <button class="action-btn primary" :class="{ favorited: isFavorite }" @click="toggleFavorite">
            <svg v-if="isFavorite" viewBox="0 0 24 24" fill="currentColor">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>
            </svg>
            {{ isFavorite ? '取消收藏' : '收藏' }}
          </button>
        </div>

      <div v-if="movie.summary" class="summary-section">
        <h3 class="section-title">简介</h3>
        <p class="summary-text">{{ movie.summary }}</p>
      </div>

      <div v-if="previewImages.length" class="preview-section">
        <h3 class="section-title">预览图</h3>
        <div class="preview-grid">
          <img
            v-for="(img, idx) in previewImages"
            :key="idx"
            :src="getPreviewUrl(movie, idx)"
            :alt="`预览图 ${idx + 1}`"
            @click="showPreview(idx)"
          />
        </div>
      </div>
    </div>

    <Teleport to="body">
      <div v-if="lightboxVisible" class="lightbox" @click.self="closeLightbox">
        <button class="lightbox-close" @click="closeLightbox">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
        <button v-if="lightboxIndex >= 0 && previewImages.length > 1" class="lightbox-nav prev" @click.stop="prevImage">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </button>
        <div class="lightbox-content" @click.self="closeLightbox">
          <img :src="lightboxImage" alt="Preview" class="lightbox-image" @click.stop />
        </div>
        <button v-if="lightboxIndex >= 0 && previewImages.length > 1" class="lightbox-nav next" @click.stop="nextImage">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="9 18 15 12 9 6"/>
          </svg>
        </button>
        <div v-if="lightboxIndex >= 0 && previewImages.length > 1" class="lightbox-counter">
          {{ lightboxIndex + 1 }} / {{ previewImages.length }}
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, getImageUrl, getPreviewUrl } from '@/api'
import type { MovieDetailDTO, FavoriteFolderDTO } from '@/types'

const route = useRoute()
const router = useRouter()

const movie = ref<MovieDetailDTO | null>(null)
const loading = ref(false)
const isFavorite = ref(false)
const lightboxImage = ref('')
const lightboxIndex = ref(0)
const lightboxVisible = ref(false)
const folders = ref<FavoriteFolderDTO[]>([])
const showFolderDialog = ref(false)
const selectedFolderIds = ref<number[]>([])
const isCreatingFolder = ref(false)
const newFolderName = ref('')

const previewImages = computed(() => {
  if (!movie.value) return []
  const local = movie.value.local_preview_images || []
  const remote = movie.value.preview_images || []
  if (local.length === 0) return remote
  return local
})

async function loadData() {
  const { uuid } = route.params
  loading.value = true
  try {
    const res = await api.getMovieByUUID(uuid as string)
    movie.value = res.data
    isFavorite.value = movie.value?.is_favorite || false
    await api.addHistoryByUUID(uuid as string)
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })
}

async function goToActor(actorName: string) {
  try {
    const res = await api.search(actorName, 'actor')
    const actors = res.data?.actors || []
    if (actors.length > 0) {
      const actor = actors[0]
      router.push(`/actor/${actor.uuid}`)
    }
  } catch (e) {
    console.error(e)
  }
}

async function toggleFavorite() {
  if (!movie.value) return
  const res = await api.getFavoriteFolders()
  folders.value = res.data || []
  selectedFolderIds.value = [...(movie.value.favorite_folders || [])]
  showFolderDialog.value = true
}

function toggleFolder(folderId: number) {
  const idx = selectedFolderIds.value.indexOf(folderId)
  if (idx >= 0) {
    selectedFolderIds.value.splice(idx, 1)
  } else {
    selectedFolderIds.value.push(folderId)
  }
}

async function createAndSelectFolder() {
  if (!newFolderName.value.trim()) return
  try {
    const res = await api.createFavoriteFolder(newFolderName.value.trim())
    const newFolder = res.data
    folders.value.push(newFolder)
    selectedFolderIds.value.push(newFolder.id)
    newFolderName.value = ''
    isCreatingFolder.value = false
  } catch (e) {
    console.error(e)
  }
}

async function confirmAddFavorite() {
  if (!movie.value) return
  try {
    await api.setFavoritesByUUID(movie.value.uuid, selectedFolderIds.value)
    movie.value.favorite_folders = [...selectedFolderIds.value]
    isFavorite.value = selectedFolderIds.value.length > 0
    showFolderDialog.value = false
  } catch (e) {
    console.error(e)
  }
}

async function cancelAddFavorite() {
  showFolderDialog.value = false
  isCreatingFolder.value = false
  newFolderName.value = ''
}

function showPreview(idx: number) {
  const url = getPreviewUrl(movie.value, idx)
  lightboxIndex.value = idx
  openLightbox(url, true)
}

function openLightbox(url: string, isPreview = false) {
  lightboxImage.value = url
  lightboxVisible.value = true
  if (!isPreview) {
    lightboxIndex.value = -1
  }
}

function closeLightbox() {
  lightboxVisible.value = false
  lightboxImage.value = ''
}

function prevImage() {
  if (!previewImages.value.length) return
  lightboxIndex.value = (lightboxIndex.value - 1 + previewImages.value.length) % previewImages.value.length
  lightboxImage.value = getPreviewUrl(movie.value, lightboxIndex.value)
}

function nextImage() {
  if (!previewImages.value.length) return
  lightboxIndex.value = (lightboxIndex.value + 1) % previewImages.value.length
  lightboxImage.value = getPreviewUrl(movie.value, lightboxIndex.value)
}

function handleKeydown(e: KeyboardEvent) {
  if (!lightboxVisible.value) return
  if (e.key === 'ArrowLeft') prevImage()
  if (e.key === 'ArrowRight') nextImage()
  if (e.key === 'Escape') closeLightbox()
}

onMounted(() => {
  loadData()
  window.addEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.page-container {
  padding: 24px 0;
}

.page-header {
  margin-bottom: 24px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.back-btn svg {
  width: 18px;
  height: 18px;
}

.back-btn:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.back-btn-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: var(--bg-tertiary);
  border: none;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  margin-bottom: 16px;
  transition: all 0.2s ease;
}

.back-btn-inline svg {
  width: 16px;
  height: 16px;
}

.back-btn-inline:hover {
  background: var(--border);
  color: var(--text-primary);
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

.movie-detail {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  padding: 32px;
  box-shadow: var(--shadow-md);
}

.movie-header-section {
  margin-bottom: 32px;
  text-align: center;
}

.movie-title {
  font-size: 26px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 12px;
  line-height: 1.3;
}

.movie-number-inline {
  color: var(--primary);
  font-weight: 600;
  margin-right: 8px;
}

.movie-main {
  display: grid;
  grid-template-columns: 4fr 1fr;
  gap: 40px;
  margin-bottom: 32px;
}

.movie-cover {
  position: relative;
  border-radius: var(--radius-lg);
  overflow: hidden;
  background: var(--bg-tertiary);
}

.movie-cover img {
  width: 100%;
  height: auto;
  display: block;
}

.cover-score {
  position: absolute;
  top: 16px;
  left: 16px;
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(0, 0, 0, 0.8);
  color: #fbbf24;
  padding: 10px 16px;
  border-radius: 20px;
  font-size: 15px;
  font-weight: 600;
}

.cover-score svg {
  width: 18px;
  height: 18px;
}

.movie-info {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border);
}

.info-label {
  color: var(--text-muted);
  font-size: 14px;
  font-weight: 500;
  min-width: 70px;
}

.info-value {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
}

.info-section {
  padding-bottom: 12px;
}

.info-section .info-label {
  display: block;
  margin-bottom: 10px;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag {
  padding: 6px 14px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}

.actors-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.actor-tag {
  padding: 8px 16px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border-radius: 20px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.actor-tag:hover {
  background: var(--primary);
  color: white;
}

.action-row {
  margin-top: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: fit-content;
  padding: 10px 20px;
  border: none;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn svg {
  width: 20px;
  height: 20px;
}

.action-btn.primary {
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-light) 100%);
  color: white;
}

.action-btn.primary.favorited {
  background: linear-gradient(135deg, #f43f5e 0%, #fb7185 100%);
}

.action-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(13, 148, 136, 0.4);
}

.action-btn.primary.favorited:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(244, 63, 94, 0.4);
}

.action-btn.primary svg {
  fill: white;
}

.section-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  padding-top: 24px;
  margin-bottom: 20px;
  border-top: 1px solid var(--border);
}

.summary-section {
  padding-top: 8px;
}

.preview-section {
  padding-top: 8px;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
}

.preview-grid img {
  width: 100%;
  aspect-ratio: 16/10;
  object-fit: cover;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
}

.preview-grid img:hover {
  transform: scale(1.03);
  box-shadow: var(--shadow-lg);
}

.lightbox {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.lightbox-content {
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: default;
}

.lightbox-image {
  width: auto;
  height: auto;
  max-width: 80vw;
  max-height: 80vh;
  object-fit: contain;
  cursor: default;
}

.lightbox-content {
  width: 80vw;
  height: 80vh;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: default;
}

.lightbox-close {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 44px;
  height: 44px;
  background: rgba(255, 255, 255, 0.1);
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.lightbox-close svg {
  width: 24px;
  height: 24px;
  color: white;
}

.lightbox-close:hover {
  background: rgba(255, 255, 255, 0.2);
}

.lightbox-nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 56px;
  height: 56px;
  background: rgba(255, 255, 255, 0.15);
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.lightbox-nav svg {
  width: 28px;
  height: 28px;
  color: white;
}

.lightbox-nav:hover {
  background: rgba(255, 255, 255, 0.25);
}

.lightbox-nav.prev {
  left: 24px;
}

.lightbox-nav.next {
  right: 24px;
}

.lightbox-counter {
  position: absolute;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 8px 20px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 500;
}

.lightbox-image {
  max-width: 80vw;
  max-height: 80vh;
  object-fit: contain;
  cursor: default;
}

.lightbox-content {
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.folder-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.folder-modal {
  background: var(--bg-primary);
  border-radius: 16px;
  width: 420px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  overflow: hidden;
}

.folder-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border);
}

.folder-modal-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.close-btn svg {
  width: 18px;
  height: 18px;
}

.folder-modal-body {
  padding: 16px 24px;
  overflow-y: auto;
  flex: 1;
}

.folder-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.folder-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.folder-item:hover {
  background: var(--bg-tertiary);
}

.folder-item.selected {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.2);
}

.folder-item input {
  display: none;
}

.folder-item svg:nth-child(1) {
  width: 16px;
  height: 16px;
  color: var(--text-muted);
}

.folder-item span {
  flex: 1;
  font-size: 14px;
  color: var(--text-primary);
}

.folder-item .check-icon {
  width: 14px;
  height: 14px;
  color: var(--primary);
}

.add-folder-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px dashed var(--border);
  background: transparent;
  color: var(--text-muted);
  font-size: 14px;
  width: 100%;
  margin-top: 8px;
}

.add-folder-btn:hover {
  background: var(--bg-tertiary);
  color: var(--primary);
  border-color: var(--primary);
}

.add-folder-btn svg {
  width: 18px;
  height: 18px;
}

.new-folder-input {
  display: flex;
  gap: 8px;
}

.new-folder-input input {
  flex: 1;
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 14px;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.new-folder-input input:focus {
  outline: none;
  border-color: var(--primary);
}

.create-btn {
  padding: 10px 16px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
}

.folder-modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px;
  border-top: 1px solid var(--border);
  background: var(--bg-secondary);
}

.folder-modal-footer .cancel-btn {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.folder-modal-footer .confirm-btn {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  background: var(--primary);
  color: white;
}

.folder-modal-footer .confirm-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.folder-modal-footer .confirm-btn:not(:disabled):hover {
  background: #e11d48;
}

.folder-modal-footer .cancel-btn:hover {
  background: var(--bg-hover);
}
</style>