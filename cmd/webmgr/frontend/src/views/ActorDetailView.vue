<template>
  <div class="page-container">
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="actor" class="actor-detail">
      <div class="actor-left">
        <div class="actor-avatar">
          <img v-if="actor.images?.[0]" :src="getActorImageUrl(actor)" :alt="actor.name" @error="(e) => (e.target as HTMLImageElement).style.display = 'none'" />
          <button class="follow-btn" :class="{ followed: actor.is_followed }" @click="toggleFollow">
            <svg v-if="actor.is_followed" viewBox="0 0 24 24" fill="currentColor">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>
            </svg>
            {{ actor.is_followed ? '已关注' : '关注' }}
          </button>
        </div>

        <div class="actor-info">
          <h1 class="actor-name">{{ actor.name }}</h1>
          <div v-if="actor.aliases?.length" class="actor-aliases">
            <span class="alias-label">别名</span>
            {{ actor.aliases.join(', ') }}
          </div>

          <div class="actor-stats">
            <div v-if="actor.height" class="stat-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 2v20M2 12h20"/>
              </svg>
              {{ actor.height }}cm
            </div>
            <div v-if="actor.blood_type" class="stat-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 22c-4.97 0-9-2.24-9-5v-4c0-2.76 4.03-5 9-5s9 2.24 9 5v4c0 2.76-4.03 5-9 5z"/>
              </svg>
              {{ actor.blood_type }}型
            </div>
            <div v-if="actor.cup_size" class="stat-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 2l3 7h8l-6.5 5 2.5 8-7-5-7 5 2.5-8L1 9h8l3-7z"/>
              </svg>
              {{ actor.cup_size }}
            </div>
          </div>
        </div>

        <div v-if="actor.summary" class="actor-section">
          <h3 class="section-title">简介</h3>
          <p class="summary-text">{{ actor.summary }}</p>
        </div>

        <div v-if="actor.hobby || actor.skill" class="actor-section">
          <div v-if="actor.hobby" class="info-item">
            <span class="info-label">爱好</span>
            <span class="info-value">{{ actor.hobby }}</span>
          </div>
          <div v-if="actor.skill" class="info-item">
            <span class="info-label">特长</span>
            <span class="info-value">{{ actor.skill }}</span>
          </div>
        </div>
      </div>

      <div class="actor-right">
        <h3 class="section-title">参演影片</h3>
        <div class="section-divider"></div>
        <div v-if="loadingMovies" class="loading-container small">
          <div class="loading-spinner"></div>
        </div>
        <div v-else-if="actorMovies.length === 0" class="empty-state">
          暂无参演影片
        </div>
        <div v-else class="movie-grid">
          <MovieCardDirect
            v-for="movie in actorMovies"
            :key="`${movie.provider}-${movie.id}`"
            :movie="movie"
            @click="goToMovie"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, getActorImageUrl } from '@/api'
import MovieCardDirect from '@/components/MovieCardDirect.vue'
import type { ActorDetailDTO, MovieDTO } from '@/types'

const route = useRoute()
const router = useRouter()

const actor = ref<ActorDetailDTO | null>(null)
const actorMovies = ref<MovieDTO[]>([])
const loading = ref(false)
const loadingMovies = ref(false)

async function loadData() {
  const { uuid } = route.params
  loading.value = true
  loadingMovies.value = true
  try {
    const [actorRes, moviesRes] = await Promise.all([
      api.getActorByUUID(uuid as string),
      api.getActorMoviesByUUID(uuid as string),
    ])
    actor.value = actorRes.data
    actorMovies.value = moviesRes.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
    loadingMovies.value = false
  }
}

async function goToMovie(movie: MovieDTO) {
  await api.ensureMovie(movie.uuid, movie.provider, movie.id)
  router.push(`/movie/${movie.uuid}`)
}

async function toggleFollow() {
  if (!actor.value) return
  try {
    if (actor.value.is_followed) {
      await api.unfollowActorByUUID(actor.value.uuid)
    } else {
      await api.followActorByUUID(actor.value.uuid)
    }
    actor.value.is_followed = !actor.value.is_followed
  } catch (e) {
    console.error(e)
  }
}

onMounted(loadData)
</script>

<style scoped>
.page-container {
  padding: 0;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  color: var(--text-muted);
}

.loading-container.small {
  padding: 40px 0;
}

.loading-spinner {
  width: 36px;
  height: 36px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.actor-detail {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 32px;
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  padding: 24px;
  box-shadow: var(--shadow-md);
}

.actor-left {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.actor-avatar {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.actor-avatar img {
  width: 200px;
  height: 200px;
  border-radius: var(--radius-lg);
  object-fit: cover;
  box-shadow: var(--shadow-md);
}

.follow-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 24px;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-light) 100%);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  justify-content: center;
}

.follow-btn svg {
  width: 16px;
  height: 16px;
}

.follow-btn.followed {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.follow-btn:hover:not(.followed) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(13, 148, 136, 0.4);
}

.actor-info {
  display: flex;
  flex-direction: column;
}

.actor-name {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
  text-align: center;
}

.actor-aliases {
  color: var(--text-muted);
  font-size: 13px;
  margin-bottom: 16px;
  text-align: center;
}

.alias-label {
  display: inline-block;
  padding: 2px 8px;
  background: var(--bg-tertiary);
  border-radius: 4px;
  font-size: 12px;
  margin-right: 8px;
}

.actor-stats {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 500;
}

.stat-item svg {
  width: 16px;
  height: 16px;
  color: var(--text-muted);
}

.actor-section {
  margin-top: 4px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
}

.summary-text {
  color: var(--text-secondary);
  line-height: 1.6;
  font-size: 13px;
}

.info-item {
  display: flex;
  gap: 12px;
  margin-bottom: 8px;
}

.info-label {
  color: var(--text-muted);
  font-size: 13px;
  min-width: 40px;
}

.info-value {
  color: var(--text-primary);
  font-size: 13px;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}

.actor-right {
  min-width: 0;
}

.actor-right .section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.section-divider {
  height: 1px;
  background: var(--border);
  margin-bottom: 16px;
}

.movie-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 16px;
}
</style>
