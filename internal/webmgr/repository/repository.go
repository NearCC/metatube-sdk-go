package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func hashUUID(provider, id string) string {
	hash := sha256.Sum256([]byte(provider + ":" + id))
	return hex.EncodeToString(hash[:])[:32]
}

type Repository struct {
	db *gorm.DB
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(
		&model.Movie{},
		&model.Actor{},
		&model.FavoriteFolder{},
		&model.Favorite{},
		&model.History{},
		&model.ActorFollow{},
		&model.ActorMovie{},
		&model.ActorNewMovie{},
		&model.User{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Movie operations
func (r *Repository) GetMovie(provider, id string) (*model.Movie, error) {
	var movie model.Movie
	result := r.db.Where("id = ? AND provider = ?", id, provider).First(&movie)
	if result.Error != nil {
		return nil, result.Error
	}
	return &movie, nil
}

func (r *Repository) GetMovieByUUID(uuid string) (*model.Movie, error) {
	var movie model.Movie
	result := r.db.Where("uuid = ?", uuid).First(&movie)
	if result.Error != nil {
		return nil, result.Error
	}
	return &movie, nil
}

func (r *Repository) SaveMovie(movie *model.Movie) error {
	return r.db.Save(movie).Error
}

func (r *Repository) GetMoviesByFolder(folderID uint, limit, offset int, userID uint) ([]model.Movie, error) {
	var movies []model.Movie
	err := r.db.Raw(`
		SELECT m.* FROM movies m
		INNER JOIN favorites f ON m.uuid = f.movie_uuid
		WHERE f.folder_id = ? AND f.user_id = ?
		ORDER BY f.created_at DESC
		LIMIT ? OFFSET ?
	`, folderID, userID, limit, offset).Scan(&movies).Error
	return movies, err
}

func (r *Repository) GetFavoriteCount(folderID uint, userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Favorite{}).Where("folder_id = ? AND user_id = ?", folderID, userID).Count(&count).Error
	return count, err
}

// Actor operations
func (r *Repository) GetActor(provider, id string) (*model.Actor, error) {
	var actor model.Actor
	result := r.db.Where("id = ? AND provider = ?", id, provider).First(&actor)
	if result.Error != nil {
		return nil, result.Error
	}
	return &actor, nil
}

func (r *Repository) GetActorByUUID(uuid string) (*model.Actor, error) {
	var actor model.Actor
	result := r.db.Where("uuid = ?", uuid).First(&actor)
	if result.Error != nil {
		return nil, result.Error
	}
	return &actor, nil
}

func (r *Repository) SearchActorsByName(name string) ([]model.Actor, error) {
	var actors []model.Actor
	err := r.db.Where("name LIKE ?", "%"+name+"%").Limit(20).Find(&actors).Error
	return actors, err
}

func (r *Repository) SaveActor(actor *model.Actor) error {
	return r.db.Save(actor).Error
}

func (r *Repository) GetActorsByFollow(limit, offset int) ([]model.Actor, error) {
	var actors []model.Actor
	err := r.db.Raw(`
		SELECT a.* FROM actors a
		INNER JOIN actor_follows af ON a.id = af.actor_id AND a.provider = af.actor_provider
		ORDER BY af.followed_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset).Scan(&actors).Error
	return actors, err
}

// FavoriteFolder operations
func (r *Repository) GetFavoriteFolders(userID uint) ([]model.FavoriteFolder, error) {
	var folders []model.FavoriteFolder
	err := r.db.Where("user_id = ?", userID).Order("name").Find(&folders).Error
	return folders, err
}

func (r *Repository) CreateFavoriteFolder(name string, userID uint) (*model.FavoriteFolder, error) {
	folder := &model.FavoriteFolder{Name: name, UserID: userID}
	err := r.db.Create(folder).Error
	return folder, err
}

func (r *Repository) UpdateFavoriteFolder(id uint, name string, userID uint) error {
	return r.db.Model(&model.FavoriteFolder{}).Where("id = ? AND user_id = ?", id, userID).Update("name", name).Error
}

func (r *Repository) DeleteFavoriteFolder(id uint, userID uint) error {
	if err := r.db.Where("folder_id = ? AND user_id = ?", id, userID).Delete(&model.Favorite{}).Error; err != nil {
		return err
	}
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.FavoriteFolder{}).Error
}

// Favorite operations
func (r *Repository) AddFavorite(movieUUID string, folderID uint, userID uint) error {
	favorite := &model.Favorite{
		UserID:    userID,
		MovieUUID: movieUUID,
		FolderID:  folderID,
	}
	return r.db.Create(favorite).Error
}

func (r *Repository) RemoveFavorite(movieUUID string, folderID uint, userID uint) error {
	result := r.db.Where("movie_uuid = ? AND folder_id = ? AND user_id = ?",
		movieUUID, folderID, userID).Delete(&model.Favorite{})
	return result.Error
}

func (r *Repository) SetFavorites(movieUUID string, folderIDs []uint, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("movie_uuid = ? AND user_id = ?", movieUUID, userID).Delete(&model.Favorite{}).Error; err != nil {
			return err
		}
		for _, folderID := range folderIDs {
			favorite := &model.Favorite{
				UserID:    userID,
				MovieUUID: movieUUID,
				FolderID:  folderID,
			}
			if err := tx.Create(favorite).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetFavoriteFoldersByMovie(movieUUID string, userID uint) ([]uint, error) {
	var folderIDs []uint
	err := r.db.Model(&model.Favorite{}).
		Where("movie_uuid = ? AND user_id = ?", movieUUID, userID).
		Pluck("folder_id", &folderIDs).Error
	return folderIDs, err
}

func (r *Repository) IsMovieFavorited(movieUUID string, userID uint) bool {
	var count int64
	r.db.Model(&model.Favorite{}).Where("movie_uuid = ? AND user_id = ?",
		movieUUID, userID).Count(&count)
	return count > 0
}

// History operations
func (r *Repository) AddHistory(movieUUID string, userID uint) error {
	now := time.Now()
	err := r.db.Exec(`
		INSERT OR REPLACE INTO history (movie_uuid, user_id, visited_at)
		VALUES (?, ?, ?)
	`, movieUUID, userID, now).Error
	return err
}

func (r *Repository) GetHistory(limit, offset int, userID uint) ([]model.Movie, error) {
	var movies []model.Movie
	err := r.db.Raw(`SELECT m.* FROM movies m
		WHERE EXISTS (SELECT 1 FROM history h WHERE h.movie_uuid = m.uuid AND h.user_id = ?)
		ORDER BY (
			SELECT MAX(h2.visited_at) FROM history h2 WHERE h2.movie_uuid = m.uuid AND h2.user_id = ?
		) DESC
		LIMIT ? OFFSET ?
	`, userID, userID, limit, offset).Scan(&movies).Error
	return movies, err
}

func (r *Repository) GetHistoryCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.History{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *Repository) ClearHistory(userID uint) error {
	return r.db.Exec("DELETE FROM history WHERE user_id = ?", userID).Error
}

func (r *Repository) TrimHistory(maxRecords int, userID uint) error {
	return r.db.Exec(`
		DELETE FROM history WHERE user_id = ? AND id NOT IN (
			SELECT id FROM (
				SELECT id FROM history WHERE user_id = ?
				ORDER BY visited_at DESC
				LIMIT ?
			)
		)
	`, userID, userID, maxRecords).Error
}

// ActorFollow operations
func (r *Repository) FollowActorByUUID(uuid string, userID uint) error {
	actor, err := r.GetActorByUUID(uuid)
	if err != nil {
		return err
	}
	follow := &model.ActorFollow{
		UserID:        userID,
		ActorUUID:     uuid,
		ActorID:       actor.ID,
		ActorProvider: actor.Provider,
	}
	return r.db.Create(follow).Error
}

func (r *Repository) UnfollowActorByUUID(uuid string, userID uint) error {
	result := r.db.Where("actor_uuid = ? AND user_id = ?", uuid, userID).Delete(&model.ActorFollow{})
	return result.Error
}

func (r *Repository) IsActorFollowedByUUID(uuid string, userID uint) bool {
	var count int64
	r.db.Model(&model.ActorFollow{}).Where("actor_uuid = ? AND user_id = ?", uuid, userID).Count(&count)
	return count > 0
}

func (r *Repository) FollowActor(actorID, actorProvider string, userID uint) error {
	uuid := hashUUID(actorProvider, actorID)
	follow := &model.ActorFollow{
		UserID:        userID,
		ActorUUID:     uuid,
		ActorID:       actorID,
		ActorProvider: actorProvider,
	}
	return r.db.Create(follow).Error
}

func (r *Repository) UnfollowActor(actorID, actorProvider string, userID uint) error {
	uuid := hashUUID(actorProvider, actorID)
	result := r.db.Where("actor_uuid = ? AND user_id = ?", uuid, userID).Delete(&model.ActorFollow{})
	return result.Error
}

func (r *Repository) IsActorFollowed(actorID, actorProvider string, userID uint) bool {
	uuid := hashUUID(actorProvider, actorID)
	var count int64
	r.db.Model(&model.ActorFollow{}).Where("actor_uuid = ? AND user_id = ?", uuid, userID).Count(&count)
	return count > 0
}

func (r *Repository) GetFollowCount(userID uint) int64 {
	var count int64
	r.db.Model(&model.ActorFollow{}).Where("user_id = ?", userID).Count(&count)
	return count
}

func (r *Repository) GetFollows(userID uint) ([]model.ActorFollow, error) {
	var follows []model.ActorFollow
	err := r.db.Where("user_id = ?", userID).Order("followed_at DESC").Find(&follows).Error
	return follows, err
}

func (r *Repository) GetFollowsPaginated(limit, offset int, userID uint) ([]model.ActorFollow, int64, error) {
	var follows []model.ActorFollow
	var total int64
	err := r.db.Model(&model.ActorFollow{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Where("user_id = ?", userID).Order("followed_at DESC").Limit(limit).Offset(offset).Find(&follows).Error
	return follows, total, err
}

// ActorMovie operations
func (r *Repository) AddActorMovie(actorID, actorProvider, movieUUID string) error {
	actorMovie := &model.ActorMovie{
		ActorID:       actorID,
		ActorProvider: actorProvider,
		MovieUUID:     movieUUID,
	}
	return r.db.Create(actorMovie).Error
}

func (r *Repository) IsActorMovieExists(actorID, actorProvider, movieUUID string) bool {
	var count int64
	r.db.Model(&model.ActorMovie{}).Where(
		"actor_id = ? AND actor_provider = ? AND movie_uuid = ?",
		actorID, actorProvider, movieUUID).Count(&count)
	return count > 0
}

func (r *Repository) GetActorMovies(actorID, actorProvider string) ([]model.Movie, error) {
	var movies []model.Movie
	err := r.db.Raw(`
		SELECT m.* FROM movies m
		INNER JOIN actor_movies am ON m.uuid = am.movie_uuid
		WHERE am.actor_id = ? AND am.actor_provider = ?
		ORDER BY am.added_at DESC
	`, actorID, actorProvider).Scan(&movies).Error
	return movies, err
}

// ActorNewMovie operations
func (r *Repository) AddActorNewMovie(followID uint, movieUUID string) error {
	newMovie := &model.ActorNewMovie{
		FollowID:  followID,
		MovieUUID: movieUUID,
	}
	return r.db.Create(newMovie).Error
}

func (r *Repository) GetUnreadNewMovies() ([]model.ActorNewMovie, error) {
	var newMovies []model.ActorNewMovie
	err := r.db.Find(&newMovies).Error
	return newMovies, err
}

func (r *Repository) MarkNewMoviesAsRead(followID uint) error {
	return r.db.Where("follow_id = ?", followID).Delete(&model.ActorNewMovie{}).Error
}

func (r *Repository) GetUnreadNewMovieCount() int64 {
	var count int64
	r.db.Model(&model.ActorNewMovie{}).Count(&count)
	return count
}

// Stats
func (r *Repository) GetStats() map[string]int64 {
	var movieCount, actorCount, favoriteCount, historyCount, followCount int64

	r.db.Model(&model.Movie{}).Count(&movieCount)
	r.db.Model(&model.Actor{}).Count(&actorCount)
	r.db.Model(&model.Favorite{}).Count(&favoriteCount)
	r.db.Model(&model.History{}).Count(&historyCount)
	r.db.Model(&model.ActorFollow{}).Count(&followCount)

	return map[string]int64{
		"movies":     movieCount,
		"actors":     actorCount,
		"favorites":  favoriteCount,
		"history":    historyCount,
		"follows":    followCount,
		"new_movies": r.GetUnreadNewMovieCount(),
	}
}
