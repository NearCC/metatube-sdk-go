package handler

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/middleware"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/model"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/util"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}

	var user model.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if !user.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "用户已被停用"})
		return
	}

	if !util.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	session := sessions.Default(c)
	session.Set(middleware.SessionKey, user.ID)
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "已退出登录"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if !util.CheckPassword(req.OldPassword, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
		return
	}

	hash, err := util.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	if err := h.db.Model(&user).Update("password_hash", hash).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修改密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup, authMw *middleware.AuthMiddleware) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.GET("/me", authMw.RequireAuth(), h.Me)
		auth.PUT("/password", authMw.RequireAuth(), h.ChangePassword)
	}
}

type AdminUserHandler struct {
	db      *gorm.DB
	logPath string
}

func NewAdminUserHandler(db *gorm.DB, logPath string) *AdminUserHandler {
	return &AdminUserHandler{db: db, logPath: logPath}
}

// Logs returns the tail end of the current log file.
func (h *AdminUserHandler) Logs(c *gin.Context) {
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "200")); err == nil && limit > 0 {
		c.Set("logLimit", limit)
	}

	f, err := os.Open(h.logPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"content": "", "truncated": false}})
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取日志文件失败"})
		return
	}

	limit, _ := c.Get("logLimit")
	maxLines, _ := limit.(int)
	if maxLines <= 0 {
		maxLines = 200
	}
	// Read at most 512KB from the tail; a line longer than that is dropped.
	const maxBytes = 512 * 1024
	start := int64(0)
	truncated := false
	if fi.Size() > maxBytes {
		start = fi.Size() - maxBytes
		truncated = true
	}

	buf := make([]byte, fi.Size()-start)
	if _, err := f.ReadAt(buf, start); err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取日志文件失败"})
		return
	}

	lines := strings.Split(strings.TrimRight(string(buf), "\n"), "\n")
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
		truncated = true
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"content":   strings.Join(lines, "\n"),
		"truncated": truncated,
	}})
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	var users []model.User
	h.db.Order("created_at DESC").Find(&users)

	result := make([]gin.H, len(users))
	for i, u := range users {
		result[i] = gin.H{
			"id":         u.ID,
			"username":   u.Username,
			"role":       u.Role,
			"enabled":    u.Enabled,
			"created_at": u.CreatedAt,
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AdminUserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	hash, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	role := req.Role
	if role == "" {
		role = "user"
	}
	if role != "admin" && role != "user" {
		role = "user"
	}

	user := model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Role:         role,
		Enabled:      true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在或其他错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
			"enabled":  user.Enabled,
		},
	})
}

func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不能为空"})
		return
	}

	hash, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	if err := h.db.Model(&model.User{}).Where("id = ?", id).Update("password_hash", hash).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})
}

func (h *AdminUserHandler) SetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色不能为空"})
		return
	}

	if req.Role != "admin" && req.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色，只能是 admin 或 user"})
		return
	}

	if err := h.db.Model(&model.User{}).Where("id = ?", id).Update("role", req.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置角色失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色设置成功"})
}

func (h *AdminUserHandler) SetEnabled(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := h.db.Model(&model.User{}).Where("id = ?", id).Update("enabled", req.Enabled).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户状态设置成功"})
}

func (h *AdminUserHandler) RegisterRoutes(r *gin.RouterGroup, authMw *middleware.AuthMiddleware) {
	admin := r.Group("/admin")
	admin.Use(authMw.RequireAuth(), authMw.RequireAdmin())
	{
		admin.GET("/users", h.ListUsers)
		admin.POST("/users", h.CreateUser)
		admin.PUT("/users/:id/password", h.ResetPassword)
		admin.PUT("/users/:id/role", h.SetRole)
		admin.PUT("/users/:id/enabled", h.SetEnabled)
		admin.GET("/logs", h.Logs)
	}
}

func splitAuthHeader(header string) []string {
	for i := 0; i < len(header); i++ {
		if header[i] == ' ' {
			return []string{header[:i], header[i+1:]}
		}
	}
	return nil
}
