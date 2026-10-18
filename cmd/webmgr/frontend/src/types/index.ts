export interface MovieDTO {
  uuid: string
  id: string
  provider: string
  number: string
  title: string
  homepage: string
  thumb_url: string
  local_thumb_url?: string
  cover_url: string
  local_cover_url?: string
  score: number
  actors: string[]
  release_date: string
  is_favorite: boolean
  favorite_folders?: number[]
}

export interface MovieDetailDTO extends MovieDTO {
  summary: string
  director: string
  big_thumb_url: string
  local_big_thumb_url?: string
  big_cover_url: string
  local_big_cover_url?: string
  preview_video_url: string
  preview_video_hls_url: string
  preview_images: string[]
  local_preview_images?: string[]
  maker: string
  label: string
  series: string
  genres: string[]
  runtime: number
}

export interface ActorDTO {
  uuid: string
  id: string
  provider: string
  name: string
  homepage: string
  aliases: string[]
  images: string[]
  local_images?: string[]
}

export interface ActorDetailDTO extends ActorDTO {
  summary: string
  hobby: string
  skill: string
  blood_type: string
  cup_size: string
  measurements: string
  nationality: string
  height: number
  birthday: string
  debut_date: string
  is_followed: boolean
}

export interface FavoriteFolderDTO {
  id: number
  name: string
  created_at: string
}

export interface SearchResponse {
  movies: MovieDTO[]
  actors: ActorDTO[]
}

export interface ProvidersResponse {
  actor_providers: Record<string, string>
  movie_providers: Record<string, string>
}

export interface StatsResponse {
  movies: number
  actors: number
  favorites: number
  history: number
  follows: number
  new_movies: number
}
