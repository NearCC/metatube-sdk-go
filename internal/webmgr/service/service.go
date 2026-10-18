package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/metatube"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/model"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/multiplexer"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/repository"
)

func generateUUID(provider, id string) string {
	hash := sha256.Sum256([]byte(provider + ":" + id))
	return hex.EncodeToString(hash[:])[:32]
}

// sanitizeSegment makes a provider/id safe to use as a single path segment.
// Non-ASCII is preserved (actor ids can be Japanese names); only characters
// that are unsafe on the filesystem or in a URL are replaced.
func sanitizeSegment(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r < 0x20 || r == 0x7f: // control chars
			b.WriteByte('_')
		case strings.ContainsRune(`<>:"/\|?*`, r): // Windows-reserved
			b.WriteByte('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	// "." and ".." would traverse; also guard Windows device names
	if out == "" || out == "." || out == ".." {
		return "_"
	}
	return out
}

func movieImageDir(baseDir, provider, id string) string {
	return filepath.Join(baseDir, sanitizeSegment(provider), sanitizeSegment(id))
}

func actorImageDir(baseDir, provider, id string) string {
	return filepath.Join(baseDir, "actors", sanitizeSegment(provider), sanitizeSegment(id))
}

func (s *Service) downloadImage(provider, id, imageType, url, localPath string) error {
	if url == "" {
		return nil
	}

	if _, err := os.Stat(localPath); err == nil {
		return nil
	}

	var data []byte
	var err error
	switch imageType {
	case "thumb":
		data, err = s.metatube.GetImageThumb(provider, id)
	case "cover":
		data, err = s.metatube.GetImageCover(provider, id)
	case "big-thumb":
		data, err = s.metatube.GetImageBigThumb(provider, id)
	case "big-cover":
		data, err = s.metatube.GetImageBigCover(provider, id)
	case "preview":
		data, err = s.downloadFromURL(url)
	default:
		data, err = s.metatube.GetImageCover(provider, id)
	}
	if err != nil || len(data) == 0 {
		return err
	}

	return os.WriteFile(localPath, data, 0644)
}

func (s *Service) downloadFromURL(rawURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func getFileExt(url string) string {
	ext := filepath.Ext(url)
	if ext == "" || len(ext) > 5 {
		ext = ".jpg"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return strings.ToLower(ext)
}

func (s *Service) downloadAndUpdateMovieImages(uuid string, info *metatube.MovieInfo) {
	baseDir := movieImageDir(s.imageBaseDir, info.Provider, info.ID)
	os.MkdirAll(baseDir, 0755)

	provider := info.Provider
	movieID := info.ID

	thumbPath := filepath.Join(baseDir, "thumb"+getFileExt(info.ThumbURL))
	coverPath := filepath.Join(baseDir, "cover"+getFileExt(info.CoverURL))
	bigCoverPath := filepath.Join(baseDir, "big_cover"+getFileExt(info.BigCoverURL))
	bigThumbPath := filepath.Join(baseDir, "big_thumb"+getFileExt(info.BigThumbURL))

	s.downloadImage(provider, movieID, "thumb", info.ThumbURL, thumbPath)
	s.downloadImage(provider, movieID, "cover", info.CoverURL, coverPath)
	s.downloadImage(provider, movieID, "big-cover", info.BigCoverURL, bigCoverPath)
	s.downloadImage(provider, movieID, "big-thumb", info.BigThumbURL, bigThumbPath)

	var previewPaths []string
	for i, url := range info.PreviewImages {
		path := filepath.Join(baseDir, fmt.Sprintf("preview_%d%s", i+1, getFileExt(url)))
		if err := s.downloadImage(provider, movieID, "preview", url, path); err == nil {
			previewPaths = append(previewPaths, cachedPreviewURL(uuid, i))
		}
	}

	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return
	}

	movie.LocalThumbURL = cachedImageURL(uuid, "thumb", thumbPath)
	movie.LocalCoverURL = cachedImageURL(uuid, "cover", coverPath)
	movie.LocalBigCoverURL = cachedImageURL(uuid, "big-cover", bigCoverPath)
	movie.LocalBigThumbURL = cachedImageURL(uuid, "big-thumb", bigThumbPath)
	movie.LocalPreviewImages = s.toJSONArray(previewPaths)

	s.repo.SaveMovie(movie)
}

// cachedImageURL returns a provider-agnostic URL for a cached image. The disk
// layout is data/images/<provider>/<id>/ but the URL only carries the uuid, so
// requests are resolved back to provider+id through the database.
func cachedImageURL(uuid, imageType, diskPath string) string {
	if !fileExists(diskPath) {
		return ""
	}
	return fmt.Sprintf("/images/%s/%s%s", uuid, imageType, getFileExt(diskPath))
}

func cachedPreviewURL(uuid string, index int) string {
	return fmt.Sprintf("/images/%s/preview/%d", uuid, index+1)
}

// cachedImageFile resolves a uuid-relative image URL back to its file on disk.
// Cached URLs carry no provider, so the lookup goes through the database.
func (s *Service) cachedImageFile(uuid, imageType, ext string) string {
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return ""
	}
	base := movieImageDir(s.imageBaseDir, movie.Provider, movie.ID)
	name := imageType
	if imageType == "big-cover" {
		name = "big_cover"
	}
	path := filepath.Join(base, name+ext)
	if !fileExists(path) {
		return ""
	}
	return path
}

func (s *Service) cachedPreviewFile(uuid string, index int) string {
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return ""
	}
	base := movieImageDir(s.imageBaseDir, movie.Provider, movie.ID)
	matches, err := filepath.Glob(filepath.Join(base, fmt.Sprintf("preview_%d.*", index+1)))
	if err != nil || len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func (s *Service) validateLocalPaths(m *model.Movie) {
	type field struct {
		url   *string
		kind  string
	}
	for _, f := range []field{
		{&m.LocalThumbURL, "thumb"},
		{&m.LocalCoverURL, "cover"},
		{&m.LocalBigCoverURL, "big-cover"},
		{&m.LocalBigThumbURL, "big-thumb"},
	} {
		if *f.url == "" {
			continue
		}
		kind, ext := splitCachedURL(*f.url)
		if kind == "" || s.cachedImageFile(m.UUID, kind, ext) == "" {
			*f.url = ""
		}
	}

	if m.LocalPreviewImages == "" {
		return
	}
	local := s.parseJSONArray(m.LocalPreviewImages)
	var kept []string
	for i, p := range local {
		if p == "" {
			continue
		}
		if s.cachedPreviewFile(m.UUID, i) != "" {
			kept = append(kept, p)
		}
	}
	if len(kept) > 0 {
		m.LocalPreviewImages = s.toJSONArray(kept)
	} else {
		m.LocalPreviewImages = ""
	}
}

// splitCachedURL pulls the image type and extension out of /images/<uuid>/<kind>.<ext>
func splitCachedURL(u string) (kind, ext string) {
	trimmed := strings.TrimPrefix(u, "/images/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 {
		return "", ""
	}
	kind = parts[1]
	if dot := strings.LastIndex(kind, "."); dot >= 0 {
		return kind[:dot], kind[dot:]
	}
	return kind, ""
}

func (s *Service) validateLocalActorImages(a *model.Actor) {
	if a.LocalImages == "" {
		return
	}
	base := actorImageDir(s.imageBaseDir, a.Provider, a.ID)
	local := s.parseJSONArray(a.LocalImages)
	var kept []string
	for i, p := range local {
		if p == "" {
			continue
		}
		matches, err := filepath.Glob(filepath.Join(base, fmt.Sprintf("%d.*", i)))
		if err != nil || len(matches) == 0 {
			continue
		}
		kept = append(kept, cachedActorImageURL(a.UUID, i, matches[0]))
	}
	if len(kept) > 0 {
		a.LocalImages = s.toJSONArray(kept)
	} else {
		a.LocalImages = ""
	}
}

func cachedActorImageURL(uuid string, index int, diskPath string) string {
	return fmt.Sprintf("/images/uuid/actor/%s/%d%s", uuid, index, getFileExt(diskPath))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type Service struct {
	repo          *repository.Repository
	metatube      *metatube.Client
	movieProvider string
	mux           *multiplexer.Multiplexer
	imageBaseDir  string
}

func New(repo *repository.Repository, metatubeClient *metatube.Client, movieProvider string, mux *multiplexer.Multiplexer, dataDir string) *Service {
	imageBaseDir := filepath.Join(dataDir, "images")
	if mux == nil {
		mux = multiplexer.Default()
	}
	return &Service{
		repo:          repo,
		metatube:      metatubeClient,
		movieProvider: movieProvider,
		mux:           mux,
		imageBaseDir:  imageBaseDir,
	}
}

func (s *Service) Search(query, searchType, systemID string) (interface{}, error) {
	switch searchType {
	case "movie":
		movies, err := s.SearchMovies(query, systemID)
		return map[string]interface{}{"movies": movies, "actors": []interface{}{}}, err
	case "actor":
		actors, err := s.SearchActors(query)
		return map[string]interface{}{"movies": []interface{}{}, "actors": actors}, err
	default:
		movies, _ := s.SearchMovies(query, systemID)
		actors, _ := s.SearchActors(query)
		return map[string]interface{}{
			"movies": movies,
			"actors": actors,
		}, nil
	}
}

func (s *Service) SearchMovies(query, systemID string) ([]MovieDTO, error) {
	results, err := s.searchMoviesBySystem(query, systemID)
	if err != nil {
		return nil, err
	}

	var movies []MovieDTO
	for _, r := range results {
		movies = append(movies, s.toMovieDTO(r.Provider, r.ID, &r, nil))
	}

	return movies, nil
}

// EnabledMovieSystems returns the list of movie search systems the user can
// actually search against.
func (s *Service) EnabledMovieSystems() []multiplexer.System {
	return s.mux.EnabledSystems()
}

// AllMovieSystems returns every configured system, including those marked
// disabled. The frontend uses this to render the full list with disabled
// options greyed out.
func (s *Service) AllMovieSystems() []multiplexer.System {
	return s.mux.Systems()
}

// searchMoviesBySystem fans out across the providers in the given system and
// de-duplicates results across those providers by ID (case-insensitive).
func (s *Service) searchMoviesBySystem(query, systemID string) ([]metatube.MovieSearchResult, error) {
	providers := s.mux.ProvidersFor(systemID)
	if len(providers) == 0 || systemID == "" {
		// Empty or unknown system: fall back to the default provider config.
		results, err := s.metatube.SearchMovies(query, s.movieProvider, true)
		if err != nil {
			return nil, err
		}
		return results, nil
	}

	type providerResult struct {
		results []metatube.MovieSearchResult
		err     error
	}

	resultsCh := make(chan providerResult, len(providers))
	var wg sync.WaitGroup
	for _, p := range providers {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			results, err := s.metatube.SearchMovies(query, p, true)
			resultsCh <- providerResult{results: results, err: err}
		}(p)
	}
	wg.Wait()
	close(resultsCh)

	seen := make(map[string]bool)
	merged := make([]metatube.MovieSearchResult, 0)
	for pr := range resultsCh {
		if pr.err != nil || len(pr.results) == 0 {
			continue
		}
		for _, r := range pr.results {
			key := strings.ToLower(r.Provider) + ":" + strings.ToLower(r.ID)
			if seen[key] {
				continue
			}
			seen[key] = true
			merged = append(merged, r)
		}
	}
	return merged, nil
}

func (s *Service) SearchActors(query string) ([]ActorDTO, error) {
	dbActors, err := s.repo.SearchActorsByName(query)
	if err == nil && len(dbActors) > 0 {
		var actors []ActorDTO
		for _, a := range dbActors {
			s.validateLocalActorImages(&a)
			actors = append(actors, ActorDTO{
				UUID:        a.UUID,
				ID:          a.ID,
				Provider:    a.Provider,
				Name:        a.Name,
				Homepage:    a.Homepage,
				Aliases:     s.parseJSONArray(a.Aliases),
				Images:      s.parseJSONArray(a.Images),
				LocalImages: s.parseJSONArray(a.LocalImages),
			})
		}
		go func() {
			results, _ := s.metatube.SearchActors(query, "", true)
			for _, r := range results {
				info, _ := s.metatube.GetActorInfo(r.Provider, r.ID)
				if info != nil {
					s.cacheActor(info)
				}
			}
		}()
		return actors, nil
	}

	results, err := s.metatube.SearchActors(query, "", true)
	if err != nil {
		return nil, err
	}

	var actors []ActorDTO
	for _, r := range results {
		uuid := generateUUID(r.Provider, r.ID)
		actors = append(actors, s.toActorDTO(r.Provider, r.ID, &r))
		existing, _ := s.repo.GetActorByUUID(uuid)
		if existing == nil {
			info, _ := s.metatube.GetActorInfo(r.Provider, r.ID)
			if info != nil {
				s.cacheActor(info)
			}
		}
	}

	return actors, nil
}

func (s *Service) GetMovieInfo(provider, id string, userID uint) (*MovieDetailDTO, error) {
	uuid := generateUUID(provider, id)

	movie, err := s.repo.GetMovieByUUID(uuid)
	if err == nil && movie.LocalThumbURL != "" {
		favorite, folders := s.favoriteState(uuid, userID)
		return s.toMovieDetailDTOWithFavorite(movie, favorite, folders), nil
	}

	info, err := s.metatube.GetMovieInfo(provider, id)
	if err != nil {
		if movie, err := s.repo.GetMovie(provider, id); err == nil {
			return s.toMovieDetailDTO(movie, userID)
		}
		return nil, err
	}

	s.cacheMovie(info)

	movie, _ = s.repo.GetMovieByUUID(uuid)
	favorite, folders := s.favoriteState(uuid, userID)
	return s.toMovieDetailDTOWithFavorite(movie, favorite, folders), nil
}

func (s *Service) GetMovieInfoByUUID(uuid string, userID uint) (*MovieDetailDTO, error) {
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return nil, err
	}

	if movie.Title != "" && movie.Summary != "" {
		favorite, folders := s.favoriteState(uuid, userID)
		return s.toMovieDetailDTOWithFavorite(movie, favorite, folders), nil
	}

	info, err := s.metatube.GetMovieInfo(movie.Provider, movie.ID)
	if err != nil {
		favorite, folders := s.favoriteState(uuid, userID)
		return s.toMovieDetailDTOWithFavorite(movie, favorite, folders), nil
	}

	s.cacheMovie(info)

	movie, _ = s.repo.GetMovieByUUID(uuid)
	favorite, folders := s.favoriteState(uuid, userID)
	return s.toMovieDetailDTOWithFavorite(movie, favorite, folders), nil
}

// favoriteState reports whether the movie sits in any of the user's folders.
// userID 0 means unauthenticated, which has no favorites.
func (s *Service) favoriteState(uuid string, userID uint) (bool, []uint) {
	if userID == 0 {
		return false, nil
	}
	folderIDs, err := s.repo.GetFavoriteFoldersByMovie(uuid, userID)
	if err != nil || len(folderIDs) == 0 {
		return false, nil
	}
	return true, folderIDs
}

func (s *Service) EnsureMovie(provider, id string) error {
	_, err := s.repo.GetMovie(provider, id)
	if err == nil {
		return nil
	}
	info, err := s.metatube.GetMovieInfo(provider, id)
	if err != nil {
		return err
	}
	return s.cacheMovie(info)
}

func (s *Service) GetActorInfo(provider, id string) (*ActorDetailDTO, error) {
	info, err := s.metatube.GetActorInfo(provider, id)
	if err != nil {
		if actor, err := s.repo.GetActor(provider, id); err == nil {
			return s.toActorDetailDTO(actor), nil
		}
		return nil, err
	}

	s.cacheActor(info)

	isFollowed := s.repo.IsActorFollowed(id, provider, 0)

	return s.toActorDetailDTOFromInfo(info, isFollowed), nil
}

func (s *Service) GetActorMovies(provider, id string) ([]MovieDTO, error) {
	movies, err := s.repo.GetActorMovies(id, provider)
	if err != nil || len(movies) > 0 {
		return s.convertCachedMovies(movies, 0), err
	}

	results, err := s.metatube.SearchMovies(id, "JAV321", true)
	if err == nil && len(results) > 0 {
		var movieDTOs []MovieDTO
		for _, r := range results {
			movieDTOs = append(movieDTOs, s.toMovieDTO(r.Provider, r.ID, &r, nil))
		}
		return movieDTOs, nil
	}

	results, err = s.metatube.SearchMovies(id, "", true)
	if err != nil {
		return []MovieDTO{}, nil
	}

	var movieDTOs []MovieDTO
	for _, r := range results {
		movieDTOs = append(movieDTOs, s.toMovieDTO(r.Provider, r.ID, &r, nil))
	}

	return movieDTOs, nil
}

func (s *Service) GetFavoriteFolders(userID uint) ([]FavoriteFolderDTO, error) {
	folders, err := s.repo.GetFavoriteFolders(userID)
	if err != nil {
		return nil, err
	}

	var result []FavoriteFolderDTO
	for _, f := range folders {
		result = append(result, FavoriteFolderDTO{
			ID:        f.ID,
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
		})
	}

	return result, nil
}

func (s *Service) CreateFavoriteFolder(name string, userID uint) (*FavoriteFolderDTO, error) {
	folder, err := s.repo.CreateFavoriteFolder(name, userID)
	if err != nil {
		return nil, err
	}

	return &FavoriteFolderDTO{
		ID:        folder.ID,
		Name:      folder.Name,
		CreatedAt: folder.CreatedAt,
	}, nil
}

func (s *Service) UpdateFavoriteFolder(id uint, name string, userID uint) error {
	return s.repo.UpdateFavoriteFolder(id, name, userID)
}

func (s *Service) DeleteFavoriteFolder(id uint, userID uint) error {
	return s.repo.DeleteFavoriteFolder(id, userID)
}

func (s *Service) AddFavorite(provider, id string, folderID uint, userID uint) error {
	if _, err := s.GetMovieInfo(provider, id, 0); err != nil {
		return err
	}

	uuid := generateUUID(provider, id)
	return s.repo.AddFavorite(uuid, folderID, userID)
}

func (s *Service) RemoveFavorite(provider, id string, folderID uint, userID uint) error {
	uuid := generateUUID(provider, id)
	return s.repo.RemoveFavorite(uuid, folderID, userID)
}

func (s *Service) RemoveFavoriteByUUID(uuid string, folderID uint, userID uint) error {
	return s.repo.RemoveFavorite(uuid, folderID, userID)
}

func (s *Service) AddFavoriteByUUID(uuid string, folderID uint, userID uint) error {
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return err
	}

	if _, err := s.GetMovieInfo(movie.Provider, movie.ID, 0); err != nil {
		return err
	}

	if err := s.repo.AddFavorite(uuid, folderID, userID); err != nil {
		return err
	}

	go func() {
		info, err := s.metatube.GetMovieInfo(movie.Provider, movie.ID)
		if err != nil {
			return
		}
		s.downloadAndUpdateMovieImages(uuid, info)
	}()

	return nil
}

func (s *Service) SetFavoritesByUUID(uuid string, folderIDs []uint, userID uint) error {
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return err
	}

	if err := s.repo.SetFavorites(uuid, folderIDs, userID); err != nil {
		return err
	}

	if len(folderIDs) > 0 {
		go func() {
			info, err := s.metatube.GetMovieInfo(movie.Provider, movie.ID)
			if err != nil {
				return
			}
			s.downloadAndUpdateMovieImages(uuid, info)
		}()
	}

	return nil
}

func (s *Service) GetFavorites(folderID uint, limit, offset int, userID uint) ([]MovieDTO, int64, error) {
	movies, err := s.repo.GetMoviesByFolder(folderID, limit, offset, userID)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.repo.GetFavoriteCount(folderID, userID)
	if err != nil {
		return nil, 0, err
	}
	return s.convertCachedMovies(movies, userID), count, nil
}

func (s *Service) AddHistory(provider, id string, userID uint) error {
	if _, err := s.GetMovieInfo(provider, id, 0); err != nil {
		return err
	}

	uuid := generateUUID(provider, id)
	if err := s.repo.AddHistory(uuid, userID); err != nil {
		return err
	}

	return s.repo.TrimHistory(100, userID)
}

func (s *Service) AddHistoryByUUID(uuid string, userID uint) error {
	if err := s.repo.AddHistory(uuid, userID); err != nil {
		return err
	}

	return s.repo.TrimHistory(100, userID)
}

func (s *Service) GetHistory(limit, offset int, userID uint) ([]MovieDTO, int64, error) {
	movies, err := s.repo.GetHistory(limit, offset, userID)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.repo.GetHistoryCount(userID)
	if err != nil {
		return nil, 0, err
	}
	return s.convertCachedMovies(movies, userID), count, nil
}

func (s *Service) ClearHistory(userID uint) error {
	return s.repo.ClearHistory(userID)
}

func (s *Service) GetFollows(limit, offset int, userID uint) ([]ActorDTO, int64, error) {
	follows, total, err := s.repo.GetFollowsPaginated(limit, offset, userID)
	if err != nil {
		return nil, 0, err
	}

	var actors []ActorDTO
	for _, f := range follows {
		if actor, err := s.repo.GetActor(f.ActorProvider, f.ActorID); err == nil {
			s.validateLocalActorImages(actor)
			actors = append(actors, ActorDTO{
				UUID:        actor.UUID,
				ID:          actor.ID,
				Name:        actor.Name,
				Provider:    actor.Provider,
				Images:      s.parseJSONArray(actor.Images),
				LocalImages: s.parseJSONArray(actor.LocalImages),
			})
		}
	}
	return actors, total, nil
}

func (s *Service) FollowActor(provider, id string, userID uint) error {
	if s.repo.IsActorFollowed(id, provider, userID) {
		return nil
	}
	return s.repo.FollowActor(id, provider, userID)
}

func (s *Service) UnfollowActor(provider, id string, userID uint) error {
	return s.repo.UnfollowActor(id, provider, userID)
}

func (s *Service) FollowActorByUUID(uuid string, userID uint) error {
	if s.repo.IsActorFollowedByUUID(uuid, userID) {
		return nil
	}
	return s.repo.FollowActorByUUID(uuid, userID)
}

func (s *Service) UnfollowActorByUUID(uuid string, userID uint) error {
	return s.repo.UnfollowActorByUUID(uuid, userID)
}

func (s *Service) GetActorInfoByUUID(uuid string, userID uint) (*ActorDetailDTO, error) {
	actor, err := s.repo.GetActorByUUID(uuid)
	if err != nil {
		return nil, err
	}

	isFollowed := s.repo.IsActorFollowedByUUID(uuid, userID)
	return s.toActorDetailDTOWithFollow(actor, isFollowed), nil
}

func (s *Service) GetActorMoviesByUUID(uuid string) ([]MovieDTO, error) {
	actor, err := s.repo.GetActorByUUID(uuid)
	if err != nil {
		return nil, err
	}

	results, err := s.metatube.SearchMovies(actor.ID, s.movieProvider, true)
	if err != nil {
		return []MovieDTO{}, nil
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].ReleaseDate > results[j].ReleaseDate
	})
	var movieDTOs []MovieDTO
	for _, r := range results {
		movieDTOs = append(movieDTOs, s.toMovieDTO(r.Provider, r.ID, &r, nil))
	}
	return movieDTOs, nil
}

func (s *Service) GetFollowCount(userID uint) int64 {
	return s.repo.GetFollowCount(userID)
}

// readCachedImage returns the on-disk file for a uuid image, or "" when the
// movie is unknown or the file has not been cached yet.
func (s *Service) readCachedImage(uuid, imageType string) []byte {
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return nil
	}
	base := movieImageDir(s.imageBaseDir, movie.Provider, movie.ID)
	name := imageType
	if imageType == "big-cover" {
		name = "big_cover"
	}
	matches, err := filepath.Glob(filepath.Join(base, name+".*"))
	if err != nil || len(matches) == 0 {
		return nil
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		return nil
	}
	return data
}

func (s *Service) GetImageThumb(uuid string) ([]byte, error) {
	if data := s.readCachedImage(uuid, "thumb"); len(data) > 0 {
		return data, nil
	}
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return nil, err
	}
	return s.metatube.GetImageThumb(movie.Provider, movie.ID)
}

func (s *Service) GetImageCover(uuid string) ([]byte, error) {
	if data := s.readCachedImage(uuid, "cover"); len(data) > 0 {
		return data, nil
	}
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return nil, err
	}
	return s.metatube.GetImageCover(movie.Provider, movie.ID)
}

func (s *Service) GetImagePoster(uuid string) ([]byte, error) {
	if data := s.readCachedImage(uuid, "poster"); len(data) > 0 {
		return data, nil
	}
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return nil, err
	}
	return s.metatube.GetImagePoster(movie.Provider, movie.ID)
}

func (s *Service) GetImageBigThumb(uuid string) ([]byte, error) {
	if data := s.readCachedImage(uuid, "big-thumb"); len(data) > 0 {
		return data, nil
	}
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return nil, err
	}
	return s.metatube.GetImageBigThumb(movie.Provider, movie.ID)
}

func (s *Service) GetImageBigCover(uuid string) ([]byte, error) {
	if data := s.readCachedImage(uuid, "big-cover"); len(data) > 0 {
		return data, nil
	}
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return nil, err
	}
	return s.metatube.GetImageBigCover(movie.Provider, movie.ID)
}

func (s *Service) GetImagePreview(uuid string, index int) ([]byte, error) {
	movie, err := s.repo.GetMovieByUUID(uuid)
	if err != nil {
		return nil, err
	}

	base := movieImageDir(s.imageBaseDir, movie.Provider, movie.ID)
	if matches, err := filepath.Glob(filepath.Join(base, fmt.Sprintf("preview_%d.*", index+1))); err == nil && len(matches) > 0 {
		if data, err := os.ReadFile(matches[0]); err == nil {
			return data, nil
		}
	}

	info, err := s.metatube.GetMovieInfo(movie.Provider, movie.ID)
	if err != nil {
		return nil, err
	}

	if index >= len(info.PreviewImages) || index < 0 {
		return nil, fmt.Errorf("preview index out of range")
	}

	return s.downloadFromURL(info.PreviewImages[index])
}

func (s *Service) GetImageThumbByProvider(provider, id string) ([]byte, error) {
	return s.metatube.GetImageThumb(provider, id)
}

func (s *Service) GetImageCoverByProvider(provider, id string) ([]byte, error) {
	return s.metatube.GetImageCover(provider, id)
}

func (s *Service) GetImagePosterByProvider(provider, id string) ([]byte, error) {
	return s.metatube.GetImagePoster(provider, id)
}

func (s *Service) GetImageBigThumbByProvider(provider, id string) ([]byte, error) {
	return s.metatube.GetImageBigThumb(provider, id)
}

func (s *Service) GetImageBigCoverByProvider(provider, id string) ([]byte, error) {
	return s.metatube.GetImageBigCover(provider, id)
}

func (s *Service) GetImagePreviewByProvider(provider, id string, index int) ([]byte, error) {
	info, err := s.metatube.GetMovieInfo(provider, id)
	if err != nil {
		return nil, err
	}

	if index >= len(info.PreviewImages) || index < 0 {
		return nil, fmt.Errorf("preview index out of range")
	}

	return s.downloadFromURL(info.PreviewImages[index])
}

func (s *Service) GetActorImageByUUID(uuid string, index int) ([]byte, error) {
	actor, err := s.repo.GetActorByUUID(uuid)
	if err != nil {
		return nil, err
	}

	base := actorImageDir(s.imageBaseDir, actor.Provider, actor.ID)
	if matches, err := filepath.Glob(filepath.Join(base, fmt.Sprintf("%d.*", index))); err == nil && len(matches) > 0 {
		if data, err := os.ReadFile(matches[0]); err == nil {
			return data, nil
		}
	}

	return s.GetActorImageByProvider(actor.Provider, actor.ID, index)
}

func (s *Service) GetActorImageByProvider(provider, id string, index int) ([]byte, error) {
	info, err := s.metatube.GetActorInfo(provider, id)
	if err != nil {
		return nil, err
	}

	if index >= len(info.Images) || index < 0 {
		return nil, fmt.Errorf("actor image index out of range")
	}

	return s.downloadFromURL(info.Images[index])
}

func (s *Service) GetStats() map[string]int64 {
	return s.repo.GetStats()
}

func (s *Service) GetProviders() (*metatube.ProvidersResponse, error) {
	return s.metatube.GetProviders()
}

func (s *Service) cacheMovie(info *metatube.MovieInfo) error {
	uuid := generateUUID(info.Provider, info.ID)

	existing, _ := s.repo.GetMovie(info.Provider, info.ID)

	localThumbURL := ""
	localCoverURL := ""
	localBigThumbURL := ""
	localBigCoverURL := ""
	localPreviewImages := ""

	if existing != nil {
		localThumbURL = existing.LocalThumbURL
		localCoverURL = existing.LocalCoverURL
		localBigThumbURL = existing.LocalBigThumbURL
		localBigCoverURL = existing.LocalBigCoverURL
		localPreviewImages = existing.LocalPreviewImages
	}

	movie := &model.Movie{
		ID:                 info.ID,
		Provider:           info.Provider,
		UUID:               uuid,
		Number:             info.Number,
		Title:              info.Title,
		Summary:            info.Summary,
		Homepage:           info.Homepage,
		Director:           info.Director,
		Actors:             s.toJSONArray(info.Actors),
		ThumbURL:           info.ThumbURL,
		LocalThumbURL:      localThumbURL,
		CoverURL:           info.CoverURL,
		LocalCoverURL:      localCoverURL,
		BigCoverURL:        info.BigCoverURL,
		LocalBigCoverURL:   localBigCoverURL,
		BigThumbURL:        info.BigThumbURL,
		LocalBigThumbURL:   localBigThumbURL,
		PreviewVideoURL:    info.PreviewVideoURL,
		PreviewImages:      s.toJSONArray(info.PreviewImages),
		LocalPreviewImages: localPreviewImages,
		Maker:              info.Maker,
		Label:              info.Label,
		Series:             info.Series,
		Genres:             s.toJSONArray(info.Genres),
		Score:              info.Score,
		Runtime:            info.Runtime,
		ReleaseDate:        info.ReleaseDate,
	}
	return s.repo.SaveMovie(movie)
}

func (s *Service) cacheActor(info *metatube.ActorInfo) error {
	uuid := generateUUID(info.Provider, info.ID)
	actor := &model.Actor{
		ID:           info.ID,
		Provider:     info.Provider,
		UUID:         uuid,
		Name:         info.Name,
		Homepage:     info.Homepage,
		Summary:      info.Summary,
		Hobby:        info.Hobby,
		Skill:        info.Skill,
		BloodType:    info.BloodType,
		CupSize:      info.CupSize,
		Measurements: info.Measurements,
		Nationality:  info.Nationality,
		Height:       info.Height,
		Aliases:      s.toJSONArray(info.Aliases),
		Images:       s.toJSONArray(info.Images),
		Birthday:     info.Birthday,
		DebutDate:    info.DebutDate,
	}
	if err := s.repo.SaveActor(actor); err != nil {
		return err
	}
	s.cacheActorImages(uuid, info.Provider, info.ID, info.Images)
	return nil
}

func (s *Service) cacheActorImages(uuid, provider, id string, images []string) {
	if len(images) == 0 {
		return
	}

	actorDir := actorImageDir(s.imageBaseDir, provider, id)
	os.MkdirAll(actorDir, 0755)

	var localPaths []string
	for i, imgURL := range images {
		ext := getFileExt(imgURL)
		localPath := filepath.Join(actorDir, fmt.Sprintf("%d%s", i, ext))
		data, err := s.downloadFromURL(imgURL)
		if err == nil && len(data) > 0 {
			os.WriteFile(localPath, data, 0644)
			localPaths = append(localPaths, cachedActorImageURL(uuid, i, localPath))
		} else {
			localPaths = append(localPaths, "")
		}
	}

	actor, err := s.repo.GetActorByUUID(uuid)
	if err != nil {
		return
	}
	actor.LocalImages = s.toJSONArray(localPaths)
	s.repo.SaveActor(actor)
}

func (s *Service) convertCachedMovies(movies []model.Movie, userID uint) []MovieDTO {
	var result []MovieDTO
	for _, m := range movies {
		s.validateLocalPaths(&m)
		result = append(result, MovieDTO{
			UUID:            m.UUID,
			ID:              m.ID,
			Provider:        m.Provider,
			Number:          m.Number,
			Title:           m.Title,
			Homepage:        m.Homepage,
			ThumbURL:        m.ThumbURL,
			LocalThumbURL:   m.LocalThumbURL,
			CoverURL:        m.CoverURL,
			LocalCoverURL:   m.LocalCoverURL,
			Score:           m.Score,
			Actors:          s.parseJSONArray(m.Actors),
			ReleaseDate:     m.ReleaseDate,
			IsFavorite:      false,
			FavoriteFolders: nil,
		})
	}
	return result
}

type MovieDTO struct {
	UUID            string   `json:"uuid"`
	ID              string   `json:"id"`
	Provider        string   `json:"provider"`
	Number          string   `json:"number"`
	Title           string   `json:"title"`
	Homepage        string   `json:"homepage"`
	ThumbURL        string   `json:"thumb_url"`
	LocalThumbURL   string   `json:"local_thumb_url"`
	CoverURL        string   `json:"cover_url"`
	LocalCoverURL   string   `json:"local_cover_url"`
	Score           float64  `json:"score"`
	Actors          []string `json:"actors"`
	ReleaseDate     string   `json:"release_date"`
	IsFavorite      bool     `json:"is_favorite"`
	FavoriteFolders []uint   `json:"favorite_folders,omitempty"`
}

type MovieDetailDTO struct {
	MovieDTO
	Summary            string   `json:"summary"`
	Director           string   `json:"director"`
	BigThumbURL        string   `json:"big_thumb_url"`
	LocalBigThumbURL   string   `json:"local_big_thumb_url"`
	BigCoverURL        string   `json:"big_cover_url"`
	LocalBigCoverURL   string   `json:"local_big_cover_url"`
	PreviewVideoURL    string   `json:"preview_video_url"`
	PreviewVideoHLSURL string   `json:"preview_video_hls_url"`
	PreviewImages      []string `json:"preview_images"`
	LocalPreviewImages []string `json:"local_preview_images"`
	Maker              string   `json:"maker"`
	Label              string   `json:"label"`
	Series             string   `json:"series"`
	Genres             []string `json:"genres"`
	Runtime            int      `json:"runtime"`
}

type ActorDTO struct {
	UUID        string   `json:"uuid"`
	ID          string   `json:"id"`
	Provider    string   `json:"provider"`
	Name        string   `json:"name"`
	Homepage    string   `json:"homepage"`
	Aliases     []string `json:"aliases"`
	Images      []string `json:"images"`
	LocalImages []string `json:"local_images"`
}

type ActorDetailDTO struct {
	ActorDTO
	Summary      string `json:"summary"`
	Hobby        string `json:"hobby"`
	Skill        string `json:"skill"`
	BloodType    string `json:"blood_type"`
	CupSize      string `json:"cup_size"`
	Measurements string `json:"measurements"`
	Nationality  string `json:"nationality"`
	Height       int    `json:"height"`
	Birthday     string `json:"birthday"`
	DebutDate    string `json:"debut_date"`
	IsFollowed   bool   `json:"is_followed"`
}

type FavoriteFolderDTO struct {
	ID        uint        `json:"id"`
	Name      string      `json:"name"`
	CreatedAt interface{} `json:"created_at"`
}

func (s *Service) toMovieDTO(provider, id string, r *metatube.MovieSearchResult, folderIDs []uint) MovieDTO {
	uuid := generateUUID(provider, id)
	return MovieDTO{
		UUID:            uuid,
		ID:              r.ID,
		Provider:        r.Provider,
		Number:          r.Number,
		Title:           r.Title,
		Homepage:        r.Homepage,
		ThumbURL:        r.ThumbURL,
		CoverURL:        r.CoverURL,
		Score:           r.Score,
		Actors:          r.Actors,
		ReleaseDate:     r.ReleaseDate,
		IsFavorite:      len(folderIDs) > 0,
		FavoriteFolders: folderIDs,
	}
}

func (s *Service) toMovieDetailDTO(m *model.Movie, userID uint) (*MovieDetailDTO, error) {
	s.validateLocalPaths(m)

	favorite, folderIDs := s.favoriteState(m.UUID, userID)

	return &MovieDetailDTO{
		MovieDTO: MovieDTO{
			UUID:            m.UUID,
			ID:              m.ID,
			Provider:        m.Provider,
			Number:          m.Number,
			Title:           m.Title,
			Homepage:        m.Homepage,
			ThumbURL:        m.ThumbURL,
			LocalThumbURL:   m.LocalThumbURL,
			CoverURL:        m.CoverURL,
			LocalCoverURL:   m.LocalCoverURL,
			Score:           m.Score,
			Actors:          s.parseJSONArray(m.Actors),
			ReleaseDate:     m.ReleaseDate,
			IsFavorite:      favorite,
			FavoriteFolders: folderIDs,
		},
		Summary:            m.Summary,
		Director:           m.Director,
		BigThumbURL:        m.BigThumbURL,
		LocalBigThumbURL:   m.LocalBigThumbURL,
		BigCoverURL:        m.BigCoverURL,
		LocalBigCoverURL:   m.LocalBigCoverURL,
		PreviewVideoURL:    m.PreviewVideoURL,
		PreviewVideoHLSURL: "",
		PreviewImages:      s.parseJSONArray(m.PreviewImages),
		LocalPreviewImages: s.parseJSONArray(m.LocalPreviewImages),
		Maker:              m.Maker,
		Label:              m.Label,
		Series:             m.Series,
		Genres:             s.parseJSONArray(m.Genres),
		Runtime:            m.Runtime,
	}, nil
}

func (s *Service) toMovieDetailDTOWithFavorite(m *model.Movie, favorite bool, folderIDs []uint) *MovieDetailDTO {
	s.validateLocalPaths(m)
	return &MovieDetailDTO{
		MovieDTO: MovieDTO{
			UUID:            m.UUID,
			ID:              m.ID,
			Provider:        m.Provider,
			Number:          m.Number,
			Title:           m.Title,
			Homepage:        m.Homepage,
			ThumbURL:        m.ThumbURL,
			LocalThumbURL:   m.LocalThumbURL,
			CoverURL:        m.CoverURL,
			LocalCoverURL:   m.LocalCoverURL,
			Score:           m.Score,
			Actors:          s.parseJSONArray(m.Actors),
			ReleaseDate:     m.ReleaseDate,
			IsFavorite:      favorite,
			FavoriteFolders: folderIDs,
		},
		Summary:            m.Summary,
		Director:           m.Director,
		BigThumbURL:        m.BigThumbURL,
		LocalBigThumbURL:   m.LocalBigThumbURL,
		BigCoverURL:        m.BigCoverURL,
		LocalBigCoverURL:   m.LocalBigCoverURL,
		PreviewVideoURL:    m.PreviewVideoURL,
		PreviewVideoHLSURL: "",
		PreviewImages:      s.parseJSONArray(m.PreviewImages),
		LocalPreviewImages: s.parseJSONArray(m.LocalPreviewImages),
		Maker:              m.Maker,
		Label:              m.Label,
		Series:             m.Series,
		Genres:             s.parseJSONArray(m.Genres),
		Runtime:            m.Runtime,
	}
}

func (s *Service) toActorDTO(provider, id string, r *metatube.ActorSearchResult) ActorDTO {
	uuid := generateUUID(provider, id)
	return ActorDTO{
		UUID:     uuid,
		ID:       r.ID,
		Provider: r.Provider,
		Name:     r.Name,
		Homepage: r.Homepage,
		Aliases:  r.Aliases,
		Images:   r.Images,
	}
}

func (s *Service) toActorDetailDTO(a *model.Actor) *ActorDetailDTO {
	s.validateLocalActorImages(a)
	return &ActorDetailDTO{
		ActorDTO: ActorDTO{
			UUID:        a.UUID,
			ID:          a.ID,
			Provider:    a.Provider,
			Name:        a.Name,
			Homepage:    a.Homepage,
			Aliases:     s.parseJSONArray(a.Aliases),
			Images:      s.parseJSONArray(a.Images),
			LocalImages: s.parseJSONArray(a.LocalImages),
		},
		Summary:      a.Summary,
		Hobby:        a.Hobby,
		Skill:        a.Skill,
		BloodType:    a.BloodType,
		CupSize:      a.CupSize,
		Measurements: a.Measurements,
		Nationality:  a.Nationality,
		Height:       a.Height,
		Birthday:     a.Birthday,
		DebutDate:    a.DebutDate,
	}
}

func (s *Service) toActorDetailDTOWithFollow(a *model.Actor, isFollowed bool) *ActorDetailDTO {
	s.validateLocalActorImages(a)
	return &ActorDetailDTO{
		ActorDTO: ActorDTO{
			UUID:        a.UUID,
			ID:          a.ID,
			Provider:    a.Provider,
			Name:        a.Name,
			Homepage:    a.Homepage,
			Aliases:     s.parseJSONArray(a.Aliases),
			Images:      s.parseJSONArray(a.Images),
			LocalImages: s.parseJSONArray(a.LocalImages),
		},
		Summary:      a.Summary,
		Hobby:        a.Hobby,
		Skill:        a.Skill,
		BloodType:    a.BloodType,
		CupSize:      a.CupSize,
		Measurements: a.Measurements,
		Nationality:  a.Nationality,
		Height:       a.Height,
		Birthday:     a.Birthday,
		DebutDate:    a.DebutDate,
		IsFollowed:   isFollowed,
	}
}

func (s *Service) toActorDetailDTOFromInfo(info *metatube.ActorInfo, isFollowed bool) *ActorDetailDTO {
	uuid := generateUUID(info.Provider, info.ID)
	return &ActorDetailDTO{
		ActorDTO: ActorDTO{
			UUID:     uuid,
			ID:       info.ID,
			Provider: info.Provider,
			Name:     info.Name,
			Homepage: info.Homepage,
			Aliases:  info.Aliases,
			Images:   info.Images,
		},
		Summary:      info.Summary,
		Hobby:        info.Hobby,
		Skill:        info.Skill,
		BloodType:    info.BloodType,
		CupSize:      info.CupSize,
		Measurements: info.Measurements,
		Nationality:  info.Nationality,
		Height:       info.Height,
		Birthday:     info.Birthday,
		DebutDate:    info.DebutDate,
		IsFollowed:   isFollowed,
	}
}

func (s *Service) toJSONArray(arr []string) string {
	if arr == nil {
		return "[]"
	}
	data, _ := json.Marshal(arr)
	return string(data)
}

func (s *Service) parseJSONArray(str string) []string {
	if str == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(str), &arr); err != nil {
		return nil
	}
	return arr
}
