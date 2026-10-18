package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/model"
	"gorm.io/gorm"
)

const (
	ContextUserKey = "currentUser"
	SessionKey     = "user_id"
)

type AuthMiddleware struct {
	db *gorm.DB
}

func NewAuthMiddleware(db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(SessionKey)

		if userID == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}

		var user model.User
		if err := m.db.First(&user, uid).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			return
		}

		if !user.Enabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "用户已被停用"})
			return
		}

		c.Set(ContextUserKey, &user)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(ContextUserKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}

		u := user.(*model.User)
		if u.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			return
		}

		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) *model.User {
	user, exists := c.Get(ContextUserKey)
	if !exists {
		return nil
	}
	return user.(*model.User)
}

func GetSessionUserID(c *gin.Context) uint {
	session := sessions.Default(c)
	userID := session.Get(SessionKey)
	if userID == nil {
		return 0
	}
	uid, ok := userID.(uint)
	if !ok {
		return 0
	}
	return uid
}
