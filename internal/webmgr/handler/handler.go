package handler

import (
	"net/http"
	"strconv"

	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/middleware"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup, authMw *middleware.AuthMiddleware) {
	auth := api.Group("")
	auth.Use(authMw.RequireAuth())
	{
		auth.GET("/search", h.search)
		auth.GET("/systems", h.systems)
		auth.GET("/providers", h.providers)
		auth.GET("/stats", h.stats)

		auth.GET("/movies", h.listMovies)
		auth.GET("/movies/:uuid", h.getMovieByUUID)
		auth.POST("/movies/ensure", h.ensureMovie)
		auth.POST("/movies/:uuid/favorite", h.addFavoriteByUUID)
		auth.PUT("/movies/:uuid/favorite", h.setFavoritesByUUID)
		auth.DELETE("/movies/:uuid/favorite", h.removeFavoriteByUUID)
		auth.POST("/movies/:uuid/history", h.addHistoryByUUID)

		auth.GET("/actors", h.listActors)
		auth.GET("/actors/:provider/:id", h.getActor)
		auth.GET("/actors/:provider/:id/movies", h.getActorMovies)
		auth.POST("/actors/:provider/:id/follow", h.followActor)
		auth.DELETE("/actors/:provider/:id/follow", h.unfollowActor)

		auth.GET("/actors/actor/:uuid", h.getActorByUUID)
		auth.GET("/actors/actor/:uuid/movies", h.getActorMoviesByUUID)
		auth.POST("/actors/actor/:uuid/follow", h.followActorByUUID)
		auth.DELETE("/actors/actor/:uuid/follow", h.unfollowActorByUUID)

		auth.GET("/favorites", h.getFavorites)
		auth.GET("/favorites/folders", h.listFavoriteFolders)
		auth.POST("/favorites/folders", h.createFavoriteFolder)
		auth.PUT("/favorites/folders/:id", h.updateFavoriteFolder)
		auth.DELETE("/favorites/folders/:id", h.deleteFavoriteFolder)

		auth.GET("/history", h.getHistory)
		auth.DELETE("/history", h.clearHistory)

		auth.GET("/follows", h.listFollows)
		auth.GET("/follows/count", h.getFollowCount)

		auth.GET("/images/thumb/:provider/:id", h.getImageThumbByProvider)
		auth.GET("/images/cover/:provider/:id", h.getImageCoverByProvider)
		auth.GET("/images/poster/:provider/:id", h.getImagePosterByProvider)
		auth.GET("/images/big-thumb/:provider/:id", h.getImageBigThumbByProvider)
		auth.GET("/images/big-cover/:provider/:id", h.getImageBigCoverByProvider)
		auth.GET("/images/preview/:provider/:id/:index", h.getImagePreviewByProvider)

		auth.GET("/images/actor/:provider/:id", h.getActorImageByProvider)
		auth.GET("/images/actor/:provider/:id/:index", h.getActorImageByProvider)

		auth.GET("/images/uuid/thumb/:uuid", h.getImageThumb)
		auth.GET("/images/uuid/cover/:uuid", h.getImageCover)
		auth.GET("/images/uuid/poster/:uuid", h.getImagePoster)
		auth.GET("/images/uuid/big-thumb/:uuid", h.getImageBigThumb)
		auth.GET("/images/uuid/big-cover/:uuid", h.getImageBigCover)
		auth.GET("/images/uuid/preview/:uuid/:index", h.getImagePreview)
		auth.GET("/images/uuid/actor/:uuid", h.getActorImageByUUID)
		auth.GET("/images/uuid/actor/:uuid/:index", h.getActorImageByUUID)
	}
}

func (h *Handler) getImageThumb(c *gin.Context) {
	uuid := c.Param("uuid")

	data, err := h.svc.GetImageThumb(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImageCover(c *gin.Context) {
	uuid := c.Param("uuid")

	data, err := h.svc.GetImageCover(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImagePoster(c *gin.Context) {
	uuid := c.Param("uuid")

	data, err := h.svc.GetImagePoster(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImageBigThumb(c *gin.Context) {
	uuid := c.Param("uuid")

	data, err := h.svc.GetImageBigThumb(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImageBigCover(c *gin.Context) {
	uuid := c.Param("uuid")

	data, err := h.svc.GetImageBigCover(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImagePreview(c *gin.Context) {
	uuid := c.Param("uuid")
	indexStr := c.Param("index")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
		return
	}

	data, err := h.svc.GetImagePreview(uuid, index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImageThumbByProvider(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	data, err := h.svc.GetImageThumbByProvider(provider, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImageCoverByProvider(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	data, err := h.svc.GetImageCoverByProvider(provider, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImagePosterByProvider(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	data, err := h.svc.GetImagePosterByProvider(provider, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImageBigThumbByProvider(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	data, err := h.svc.GetImageBigThumbByProvider(provider, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImageBigCoverByProvider(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	data, err := h.svc.GetImageBigCoverByProvider(provider, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getImagePreviewByProvider(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")
	indexStr := c.Param("index")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
		return
	}

	data, err := h.svc.GetImagePreviewByProvider(provider, id, index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getActorImageByProvider(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")
	indexStr := c.Param("index")

	index := 0
	if indexStr != "" {
		var err error
		index, err = strconv.Atoi(indexStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
			return
		}
	}

	data, err := h.svc.GetActorImageByProvider(provider, id, index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) getActorImageByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	indexStr := c.Param("index")

	index := 0
	if indexStr != "" {
		var err error
		index, err = strconv.Atoi(indexStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
			return
		}
	}

	data, err := h.svc.GetActorImageByUUID(uuid, index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) search(c *gin.Context) {
	q := c.Query("q")
	searchType := c.DefaultQuery("type", "all")
	systemID := c.Query("system")

	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	results, err := h.svc.Search(q, searchType, systemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

func (h *Handler) systems(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": h.svc.AllMovieSystems()})
}

func (h *Handler) providers(c *gin.Context) {
	providers, err := h.svc.GetProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": providers})
}

func (h *Handler) stats(c *gin.Context) {
	stats := h.svc.GetStats()
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func (h *Handler) listMovies(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

func (h *Handler) getMovie(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	var userID uint
	if user := middleware.GetCurrentUser(c); user != nil {
		userID = user.ID
	}

	movie, err := h.svc.GetMovieInfo(provider, id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movie})
}

func (h *Handler) getMovieByUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	var userID uint
	if user := middleware.GetCurrentUser(c); user != nil {
		userID = user.ID
	}

	movie, err := h.svc.GetMovieInfoByUUID(uuid, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movie})
}

func (h *Handler) ensureMovie(c *gin.Context) {
	var req struct {
		UUID     string `json:"uuid"`
		Provider string `json:"provider"`
		ID       string `json:"id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.svc.EnsureMovie(req.Provider, req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "movie ensured"})
}

func (h *Handler) addFavorite(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	folderIDStr := c.DefaultPostForm("folder_id", "1")
	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder_id"})
		return
	}

	if err := h.svc.AddFavorite(provider, id, uint(folderID), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "added to favorites"})
}

func (h *Handler) removeFavorite(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	folderIDStr := c.DefaultQuery("folder_id", "1")
	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder_id"})
		return
	}

	if err := h.svc.RemoveFavorite(provider, id, uint(folderID), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed from favorites"})
}

func (h *Handler) addFavoriteByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		FolderID uint `json:"folder_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.FolderID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "folder_id is required"})
		return
	}

	if err := h.svc.AddFavoriteByUUID(uuid, req.FolderID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "added to favorites"})
}

func (h *Handler) setFavoritesByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		FolderIDs []uint `json:"folder_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.svc.SetFavoritesByUUID(uuid, req.FolderIDs, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "favorites updated"})
}

func (h *Handler) removeFavoriteByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	folderIDStr := c.DefaultQuery("folder_id", "1")
	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder_id"})
		return
	}

	if err := h.svc.RemoveFavoriteByUUID(uuid, uint(folderID), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed from favorites"})
}

func (h *Handler) addHistory(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := h.svc.AddHistory(provider, id, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "added to history"})
}

func (h *Handler) addHistoryByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := h.svc.AddHistoryByUUID(uuid, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "added to history"})
}

func (h *Handler) listActors(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

func (h *Handler) getActor(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	actor, err := h.svc.GetActorInfo(provider, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": actor})
}

func (h *Handler) getActorMovies(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")

	movies, err := h.svc.GetActorMovies(provider, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movies})
}

func (h *Handler) followActor(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := h.svc.FollowActor(provider, id, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "actor followed"})
}

func (h *Handler) unfollowActor(c *gin.Context) {
	provider := c.Param("provider")
	id := c.Param("id")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := h.svc.UnfollowActor(provider, id, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "actor unfollowed"})
}

func (h *Handler) getActorByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	actor, err := h.svc.GetActorInfoByUUID(uuid, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": actor})
}

func (h *Handler) getActorMoviesByUUID(c *gin.Context) {
	uuid := c.Param("uuid")

	movies, err := h.svc.GetActorMoviesByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movies})
}

func (h *Handler) followActorByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := h.svc.FollowActorByUUID(uuid, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "actor followed"})
}

func (h *Handler) unfollowActorByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := h.svc.UnfollowActorByUUID(uuid, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "actor unfollowed"})
}

func (h *Handler) listFavoriteFolders(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	folders, err := h.svc.GetFavoriteFolders(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": folders})
}

func (h *Handler) createFavoriteFolder(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	folder, err := h.svc.CreateFavoriteFolder(req.Name, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": folder})
}

func (h *Handler) updateFavoriteFolder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if err := h.svc.UpdateFavoriteFolder(uint(id), req.Name, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "folder updated"})
}

func (h *Handler) deleteFavoriteFolder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := h.svc.DeleteFavoriteFolder(uint(id), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "folder deleted"})
}

func (h *Handler) getFavorites(c *gin.Context) {
	folderIDStr := c.DefaultQuery("folder_id", "1")
	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder_id"})
		return
	}
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	movies, total, err := h.svc.GetFavorites(uint(folderID), limit, offset, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movies, "total": total})
}

func (h *Handler) getHistory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	movies, total, err := h.svc.GetHistory(limit, offset, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movies, "total": total})
}

func (h *Handler) clearHistory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := h.svc.ClearHistory(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "history cleared"})
}

func (h *Handler) listFollows(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	actors, total, err := h.svc.GetFollows(limit, offset, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": actors, "total": total})
}

func (h *Handler) getFollowCount(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	count := h.svc.GetFollowCount(user.ID)
	c.JSON(http.StatusOK, gin.H{"data": count})
}
