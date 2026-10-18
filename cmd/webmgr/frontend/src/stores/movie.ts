import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { MovieDTO } from '@/types'
import { api } from '@/api'

export interface MovieSystem {
  id: string
  name: string
  providers: string[]
  disabled: boolean
  disabled_hint?: string
}

export const useMovieStore = defineStore('movie', () => {
  const recentMovies = ref<MovieDTO[]>([])
  const favorites = ref<Map<string, MovieDTO[]>>(new Map())
  const history = ref<MovieDTO[]>([])
  const searchQuery = ref('')
  const searchResults = ref<MovieDTO[]>([])
  const loading = ref(false)
  const systems = ref<MovieSystem[]>([])
  const selectedSystem = ref('')

  async function loadSystems() {
    const res = await api.getSystems()
    systems.value = res.data || []
    if (systems.value.length === 0) return
    const current = systems.value.find(s => s.id === selectedSystem.value)
    if (!current || current.disabled) {
      const firstEnabled = systems.value.find(s => !s.disabled)
      selectedSystem.value = firstEnabled?.id || systems.value[0].id
    }
  }

  async function searchMovies(query: string) {
    searchQuery.value = query
    loading.value = true
    try {
      const res = await api.search(query, 'movie', selectedSystem.value)
      searchResults.value = res.data.movies || []
    } finally {
      loading.value = false
    }
  }

  function isSystemDisabled(id: string): boolean {
    const sys = systems.value.find(s => s.id === id)
    return !!sys?.disabled
  }

  function selectedSystemHint(): string {
    const sys = systems.value.find(s => s.id === selectedSystem.value)
    return sys?.disabled_hint || ''
  }

  function clearSearch() {
    searchQuery.value = ''
    searchResults.value = []
  }

  async function loadHistory() {
    const res = await api.getHistory(100, 0)
    history.value = res.data || []
  }

  async function loadFavorites(folderId: number) {
    const res = await api.getFavorites(folderId, 100, 0)
    favorites.value.set(folderId.toString(), res.data || [])
  }

  async function addToHistory(movie: MovieDTO) {
    await api.addHistoryByUUID(movie.uuid)
    await loadHistory()
  }

  async function toggleFavorite(movie: MovieDTO, folderId: number, isFavorite: boolean) {
    if (isFavorite) {
      await api.removeFavoriteByUUID(movie.uuid, folderId)
    } else {
      await api.addFavoriteByUUID(movie.uuid, folderId)
    }
    await loadFavorites(folderId)
  }

  return {
    recentMovies,
    favorites,
    history,
    searchQuery,
    searchResults,
    loading,
    systems,
    selectedSystem,
    loadSystems,
    searchMovies,
    clearSearch,
    loadHistory,
    loadFavorites,
    addToHistory,
    toggleFavorite,
    isSystemDisabled,
    selectedSystemHint,
  }
})
