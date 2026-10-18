import type {
  MovieDTO,
  MovieDetailDTO,
  ActorDTO,
  ActorDetailDTO,
  FavoriteFolderDTO,
  SearchResponse,
  ProvidersResponse,
  StatsResponse,
} from '@/types'

const API_BASE = '/api'

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...options,
    credentials: 'include',
  })
  if (res.status === 401 && !window.location.pathname.includes('/login')) {
    window.location.href = '/login'
    throw new Error('未登录')
  }
  if (!res.ok) {
    const err = await res.json().catch(() => null)
    if (err && err.error) {
      const msg = typeof err.error === 'string' ? err.error : err.error?.message
      throw new Error(msg || `HTTP ${res.status}`)
    }
    throw new Error(`HTTP ${res.status}`)
  }
  return res.json()
}

export function getImageUrl(movie: { local_thumb_url?: string; thumb_url?: string; local_cover_url?: string; cover_url?: string; uuid?: string }, type: 'thumb' | 'cover' | 'poster' | 'big-thumb' | 'big-cover' = 'thumb'): string {
  if (movie.uuid) {
    if (type === 'thumb') return `${API_BASE}/images/uuid/thumb/${encodeURIComponent(movie.uuid)}`
    if (type === 'cover') return `${API_BASE}/images/uuid/cover/${encodeURIComponent(movie.uuid)}`
    if (type === 'poster') return `${API_BASE}/images/uuid/poster/${encodeURIComponent(movie.uuid)}`
    if (type === 'big-thumb') return `${API_BASE}/images/uuid/big-thumb/${encodeURIComponent(movie.uuid)}`
    if (type === 'big-cover') return `${API_BASE}/images/uuid/big-cover/${encodeURIComponent(movie.uuid)}`
  }

  if (type === 'thumb' || type === 'big-thumb') {
    if (movie.thumb_url) return movie.thumb_url
  } else {
    if (movie.cover_url) return movie.cover_url
  }
  return ''
}

export function getImageUrlByProvider(movie: { provider?: string; id?: string }, type: 'thumb' | 'cover' | 'poster' | 'big-thumb' | 'big-cover' = 'thumb'): string {
  if (!movie.provider || !movie.id) return ''
  if (type === 'thumb') return `${API_BASE}/images/thumb/${encodeURIComponent(movie.provider)}/${encodeURIComponent(movie.id)}`
  if (type === 'cover') return `${API_BASE}/images/cover/${encodeURIComponent(movie.provider)}/${encodeURIComponent(movie.id)}`
  if (type === 'poster') return `${API_BASE}/images/poster/${encodeURIComponent(movie.provider)}/${encodeURIComponent(movie.id)}`
  if (type === 'big-thumb') return `${API_BASE}/images/big-thumb/${encodeURIComponent(movie.provider)}/${encodeURIComponent(movie.id)}`
  if (type === 'big-cover') return `${API_BASE}/images/big-cover/${encodeURIComponent(movie.provider)}/${encodeURIComponent(movie.id)}`
  return ''
}

export function getActorImageUrl(actor: { provider?: string; id?: string; uuid?: string; local_images?: string[]; images?: string[] }, index: number = 0): string {
  if (actor.uuid) {
    return `${API_BASE}/images/uuid/actor/${encodeURIComponent(actor.uuid)}${index > 0 ? `/${index}` : ''}`
  }
  if (actor.provider && actor.id) {
    return `${API_BASE}/images/actor/${encodeURIComponent(actor.provider)}/${encodeURIComponent(actor.id)}${index > 0 ? `/${index}` : ''}`
  }
  if (actor.images?.[index]) {
    return actor.images[index]
  }
  return ''
}

export function getPreviewUrl(movie: { local_preview_images?: string[]; preview_images?: string[]; uuid?: string } | null, index: number): string {
  if (!movie) return ''
  const uuid = movie.uuid
  if (uuid) {
    return `${API_BASE}/images/uuid/preview/${encodeURIComponent(uuid)}/${index}`
  }
  const localArr = movie.local_preview_images
  const local = localArr ? Array.from(localArr) : []
  if (local[index]) return local[index]
  return ''
}

export const api = {
  isLoggedIn() {
    return !!localStorage.getItem('metatube_user')
  },

  async login(username: string, password: string) {
    const data = await fetchJSON<{ user: any }>(
      `${API_BASE}/auth/login`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      }
    )
    return { user: data.user }
  },

  async logout() {
    await fetchJSON<{ message: string }>(`${API_BASE}/auth/logout`, { method: 'POST' })
  },

  async getMe() {
    return fetchJSON<{ user: any }>(`${API_BASE}/auth/me`)
  },

  async changePassword(oldPassword: string, newPassword: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/auth/password`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
      }
    )
  },

  search(query: string, type: 'movie' | 'actor' | 'all' = 'all', system: string = '') {
    const sysParam = system ? `&system=${encodeURIComponent(system)}` : ''
    return fetchJSON<{ data: SearchResponse }>(
      `${API_BASE}/search?q=${encodeURIComponent(query)}&type=${type}${sysParam}`
    )
  },

  getSystems() {
    return fetchJSON<{ data: { id: string; name: string; providers: string[]; disabled: boolean; disabled_hint?: string }[] }>(
      `${API_BASE}/systems`
    )
  },

  getMovie(provider: string, id: string) {
    return fetchJSON<{ data: MovieDetailDTO }>(
      `${API_BASE}/movies/${provider}/${id}`
    )
  },

  getMovieByUUID(uuid: string) {
    return fetchJSON<{ data: MovieDetailDTO }>(
      `${API_BASE}/movies/${uuid}`
    )
  },

  ensureMovie(uuid: string, provider: string, id: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/movies/ensure`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ uuid, provider, id }),
      }
    )
  },

  getActor(provider: string, id: string) {
    return fetchJSON<{ data: ActorDetailDTO }>(
      `${API_BASE}/actors/${provider}/${id}`
    )
  },

  getActorMovies(provider: string, id: string) {
    return fetchJSON<{ data: MovieDTO[] }>(
      `${API_BASE}/actors/${provider}/${id}/movies`
    )
  },

  getActorByUUID(uuid: string) {
    return fetchJSON<{ data: ActorDetailDTO }>(
      `${API_BASE}/actors/actor/${uuid}`
    )
  },

  getActorMoviesByUUID(uuid: string) {
    return fetchJSON<{ data: MovieDTO[] }>(
      `${API_BASE}/actors/actor/${uuid}/movies`
    )
  },

  followActorByUUID(uuid: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/actors/actor/${uuid}/follow`,
      { method: 'POST' }
    )
  },

  unfollowActorByUUID(uuid: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/actors/actor/${uuid}/follow`,
      { method: 'DELETE' }
    )
  },

  getFavoriteFolders() {
    return fetchJSON<{ data: FavoriteFolderDTO[] }>(
      `${API_BASE}/favorites/folders`
    )
  },

  createFavoriteFolder(name: string) {
    return fetchJSON<{ data: FavoriteFolderDTO }>(
      `${API_BASE}/favorites/folders`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
      }
    )
  },

  updateFavoriteFolder(id: number, name: string) {
    return fetchJSON<{ data: FavoriteFolderDTO }>(
      `${API_BASE}/favorites/folders/${id}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
      }
    )
  },

  deleteFavoriteFolder(id: number) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/favorites/folders/${id}`,
      { method: 'DELETE' }
    )
  },

  getFavorites(folderId: number, limit = 30, offset = 0) {
    return fetchJSON<{ data: MovieDTO[]; total: number }>(
      `${API_BASE}/favorites?folder_id=${folderId}&limit=${limit}&offset=${offset}`
    )
  },

  addFavorite(provider: string, id: string, folderId = 1) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/movies/${provider}/${id}/favorite`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: `folder_id=${folderId}`,
      }
    )
  },

  addFavoriteByUUID(uuid: string, folderId: number) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/movies/${uuid}/favorite`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ folder_id: folderId }),
      }
    )
  },

  removeFavorite(provider: string, id: string, folderId: number) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/movies/${provider}/${id}/favorite?folder_id=${folderId}`,
      { method: 'DELETE' }
    )
  },

  removeFavoriteByUUID(uuid: string, folderId: number) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/movies/${uuid}/favorite?folder_id=${folderId}`,
      { method: 'DELETE' }
    )
  },

  setFavoritesByUUID(uuid: string, folderIds: number[]) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/movies/${uuid}/favorite`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ folder_ids: folderIds }),
      }
    )
  },

  getHistory(limit = 30, offset = 0) {
    return fetchJSON<{ data: MovieDTO[]; total: number }>(
      `${API_BASE}/history?limit=${limit}&offset=${offset}`
    )
  },

  addHistory(provider: string, id: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/movies/${provider}/${id}/history`,
      { method: 'POST' }
    )
  },

  addHistoryByUUID(uuid: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/movies/${uuid}/history`,
      { method: 'POST' }
    )
  },

  clearHistory() {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/history`,
      { method: 'DELETE' }
    )
  },

  getFollows(limit = 30, offset = 0) {
    return fetchJSON<{ data: ActorDTO[]; total: number }>(
      `${API_BASE}/follows?limit=${limit}&offset=${offset}`
    )
  },

  followActor(provider: string, id: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/actors/${provider}/${id}/follow`,
      { method: 'POST' }
    )
  },

  unfollowActor(provider: string, id: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/actors/${provider}/${id}/follow`,
      { method: 'DELETE' }
    )
  },

  getFollowCount() {
    return fetchJSON<{ data: number }>(
      `${API_BASE}/follows/count`
    )
  },

  getProviders() {
    return fetchJSON<{ data: ProvidersResponse }>(
      `${API_BASE}/providers`
    )
  },

  getStats() {
    return fetchJSON<{ data: StatsResponse }>(
      `${API_BASE}/stats`
    )
  },

  getImageThumbUrl(provider: string, id: string) {
    return `${API_BASE}/images/thumb/${encodeURIComponent(provider)}/${encodeURIComponent(id)}`
  },

  getImageCoverUrl(provider: string, id: string) {
    return `${API_BASE}/images/cover/${encodeURIComponent(provider)}/${encodeURIComponent(id)}`
  },

  getImageBigThumbUrl(provider: string, id: string) {
    return `${API_BASE}/images/big-thumb/${encodeURIComponent(provider)}/${encodeURIComponent(id)}`
  },

  getImageBigCoverUrl(provider: string, id: string) {
    return `${API_BASE}/images/big-cover/${encodeURIComponent(provider)}/${encodeURIComponent(id)}`
  },

  listUsers() {
    return fetchJSON<{ data: any[] }>(`${API_BASE}/admin/users`)
  },

  getLogs(limit = 200) {
    return fetchJSON<{ data: { content: string; truncated: boolean } }>(
      `${API_BASE}/admin/logs?limit=${limit}`
    )
  },

  createUser(username: string, password: string, role: string) {
    return fetchJSON<{ data: any }>(
      `${API_BASE}/admin/users`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password, role }),
      }
    )
  },

  resetUserPassword(id: number, password: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/admin/users/${id}/password`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password }),
      }
    )
  },

  setUserRole(id: number, role: string) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/admin/users/${id}/role`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role }),
      }
    )
  },

  setUserEnabled(id: number, enabled: boolean) {
    return fetchJSON<{ message: string }>(
      `${API_BASE}/admin/users/${id}/enabled`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled }),
      }
    )
  },
}
