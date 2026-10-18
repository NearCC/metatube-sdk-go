package model

import (
	"time"
)

type TimeTracker struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Movie struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	Provider           string    `json:"provider" gorm:"primaryKey"`
	UUID               string    `json:"uuid" gorm:"uniqueIndex;not null"`
	Number             string    `json:"number" gorm:"not null"`
	Title              string    `json:"title"`
	Summary            string    `json:"summary"`
	Homepage           string    `json:"homepage"`
	Director           string    `json:"director"`
	Actors             string    `json:"actors"` // JSON array
	ThumbURL           string    `json:"thumb_url"`
	LocalThumbURL      string    `json:"local_thumb_url"`
	CoverURL           string    `json:"cover_url"`
	LocalCoverURL      string    `json:"local_cover_url"`
	BigCoverURL        string    `json:"big_cover_url"`
	LocalBigCoverURL   string    `json:"local_big_cover_url"`
	BigThumbURL        string    `json:"big_thumb_url"`
	LocalBigThumbURL   string    `json:"local_big_thumb_url"`
	PreviewVideoURL    string    `json:"preview_video_url"`
	PreviewImages      string    `json:"preview_images"`       // JSON array
	LocalPreviewImages string    `json:"local_preview_images"` // JSON array
	Maker              string    `json:"maker"`
	Label              string    `json:"label"`
	Series             string    `json:"series"`
	Genres             string    `json:"genres"` // JSON array
	Score              float64   `json:"score"`
	Runtime            int       `json:"runtime"`
	ReleaseDate        string    `json:"release_date"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (Movie) TableName() string {
	return "movies"
}

type Actor struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Provider     string    `json:"provider" gorm:"primaryKey"`
	UUID         string    `json:"uuid" gorm:"index"`
	Name         string    `json:"name" gorm:"not null"`
	Homepage     string    `json:"homepage"`
	Summary      string    `json:"summary"`
	Hobby        string    `json:"hobby"`
	Skill        string    `json:"skill"`
	BloodType    string    `json:"blood_type"`
	CupSize      string    `json:"cup_size"`
	Measurements string    `json:"measurements"`
	Nationality  string    `json:"nationality"`
	Height       int       `json:"height"`
	Aliases      string    `json:"aliases"`      // JSON array
	Images       string    `json:"images"`       // JSON array
	LocalImages  string    `json:"local_images"` // JSON array
	Birthday     string    `json:"birthday"`
	DebutDate    string    `json:"debut_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Actor) TableName() string {
	return "actors"
}

type FavoriteFolder struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

func (FavoriteFolder) TableName() string {
	return "favorite_folders"
}

type Favorite struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	MovieUUID string    `json:"movie_uuid" gorm:"not null;index"`
	FolderID  uint      `json:"folder_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`

	Folder FavoriteFolder `json:"folder,omitempty" gorm:"foreignKey:FolderID;constraint:OnDelete:CASCADE"`
}

func (Favorite) TableName() string {
	return "favorites"
}

type History struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	MovieUUID string    `json:"movie_uuid" gorm:"uniqueIndex"`
	VisitedAt time.Time `json:"visited_at" gorm:"autoCreateTime"`
}

func (History) TableName() string {
	return "history"
}

type ActorFollow struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        uint      `json:"user_id" gorm:"not null;index"`
	ActorUUID     string    `json:"actor_uuid" gorm:"not null;index"`
	ActorID       string    `json:"actor_id" gorm:"not null"`
	ActorProvider string    `json:"actor_provider" gorm:"not null"`
	FollowedAt    time.Time `json:"followed_at"`
	LastCheckAt   time.Time `json:"last_check_at"`
}

func (ActorFollow) TableName() string {
	return "actor_follows"
}

type ActorMovie struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ActorID       string    `json:"actor_id" gorm:"not null"`
	ActorProvider string    `json:"actor_provider" gorm:"not null"`
	MovieUUID     string    `json:"movie_uuid" gorm:"not null;index"`
	AddedAt       time.Time `json:"added_at"`
}

func (ActorMovie) TableName() string {
	return "actor_movies"
}

type ActorNewMovie struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	FollowID   uint      `json:"follow_id" gorm:"not null"`
	MovieUUID  string    `json:"movie_uuid" gorm:"not null;index"`
	DetectedAt time.Time `json:"detected_at"`
}

func (ActorNewMovie) TableName() string {
	return "actor_new_movies"
}

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string    `json:"username" gorm:"unique;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:user"` // admin or user
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
}

func (User) TableName() string {
	return "users"
}
