<template>
  <div class="movie-card" @click="$emit('click', movie)">
    <div class="movie-cover">
      <img
        v-if="movie.local_thumb_url || movie.thumb_url || movie.local_cover_url || movie.cover_url"
        :src="getImageUrl(movie, 'poster')"
        :alt="movie.title"
        @error="(e) => (e.target as HTMLImageElement).style.display = 'none'"
      />
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
</template>

<script setup lang="ts">
import type { MovieDTO } from '@/types'
import { getImageUrl } from '@/api'

defineProps<{
  movie: MovieDTO
}>()

defineEmits<{
  click: [movie: MovieDTO]
}>()
</script>

<style scoped>
.movie-card {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-sm);
}

.movie-card:hover {
  transform: translateY(-6px);
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
  top: 10px;
  right: 10px;
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(0, 0, 0, 0.75);
  color: #fbbf24;
  padding: 5px 8px;
  border-radius: 16px;
  font-size: 11px;
  font-weight: 600;
}

.movie-score svg {
  width: 10px;
  height: 10px;
}

.movie-info {
  padding: 14px;
}

.movie-number {
  color: var(--primary);
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 5px;
  letter-spacing: 0.5px;
}

.movie-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}
</style>