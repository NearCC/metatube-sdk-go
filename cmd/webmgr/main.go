package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "go.uber.org/automaxprocs"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/metatube-community/metatube-sdk-go/database"
	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/config"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/handler"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/logger"
	webmgrmetatube "github.com/metatube-community/metatube-sdk-go/internal/webmgr/metatube"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/middleware"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/model"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/multiplexer"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/repository"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/service"
	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/util"
	"github.com/metatube-community/metatube-sdk-go/route"
	sdkauth "github.com/metatube-community/metatube-sdk-go/route/auth"
)

//go:embed all:dist
var distFS embed.FS

var (
	Version   = "0.1.0"
	BuildTime = "unknown"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	dataDir := cfg.App.DataDir
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create data directory: %v\n", err)
		os.Exit(1)
	}

	logWriter, err := logger.Init(dataDir, cfg.Log.Filename, cfg.Log.MaxSizeMB, cfg.Log.MaxFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize file logging: %v\n", err)
		os.Exit(1)
	}
	defer logWriter.Close()

	dbPath := cfg.GetDBPath()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create database directory: %v\n", err)
		os.Exit(1)
	}

	sdkDB, err := database.Open(&database.Config{
		DSN:                  "file:" + dbPath + "?cache=shared&_journal_mode=WAL",
		DisableAutomaticPing: true,
		LogLevel:             gormlogger.Silent,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open SDK database: %v\n", err)
		os.Exit(1)
	}

	sdkEngine := engine.New(sdkDB, engine.WithRequestTimeout(engine.DefaultRequestTimeout))
	if err := sdkEngine.DBAutoMigrate(true); err != nil {
		fmt.Fprintf(os.Stderr, "SDK auto-migrate failed: %v\n", err)
		os.Exit(1)
	}

	repo, err := repository.New("file:" + dbPath + "?cache=shared&_journal_mode=WAL")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize webmgr database: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()
	silenceGorm(repo.DB())

	if err := initDefaultUser(repo); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize default user: %v\n", err)
		os.Exit(1)
	}

	mtClient := webmgrmetatube.New(sdkEngine)
	mux := multiplexer.New(cfg.Multiplexer.MovieSystems)
	svc := service.New(repo, mtClient, cfg.Metatube.MovieProvider, mux, dataDir)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: io.MultiWriter(os.Stdout, logWriter)}))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	store := cookie.NewStore([]byte("metatube-session-secret-change-in-production"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("metatube_session", store))

	authMw := middleware.NewAuthMiddleware(repo.DB())
	authHandler := handler.NewAuthHandler(repo.DB())
	adminUserHandler := handler.NewAdminUserHandler(repo.DB(), logWriter.Path())
	h := handler.New(svc)

	api := r.Group("/api")
	authHandler.RegisterRoutes(api, authMw)
	adminUserHandler.RegisterRoutes(api, authMw)
	h.RegisterRoutes(api, authMw)

	staticFS, _ := fs.Sub(distFS, "dist")

	r.Use(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.Next()
			return
		}
		if strings.HasPrefix(path, "/images/") {
			session := sessions.Default(c)
			if session.Get(middleware.SessionKey) == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
				return
			}
			http.StripPrefix("/images/", http.FileServer(http.Dir(filepath.Join(dataDir, "images")))).ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		stripped := strings.TrimPrefix(path, "/")
		if _, err := staticFS.Open(stripped); err == nil {
			http.FileServer(http.FS(staticFS)).ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		c.Request.URL.Path = "/"
		http.FileServer(http.FS(staticFS)).ServeHTTP(c.Writer, c.Request)
		c.Abort()
	})

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("webmgr %s (built %s) UI listening on %s", Version, BuildTime, addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	var apiToken sdkauth.Validator
	if cfg.Metatube.Token != "" {
		apiToken = sdkauth.Token(cfg.Metatube.Token)
	}
	apiRouter := route.New(sdkEngine, apiToken)
	apiAddr := fmt.Sprintf("%s:%d", cfg.Metatube.APIHost, cfg.Metatube.APIPort)
	apiSrv := &http.Server{
		Addr:    apiAddr,
		Handler: apiRouter,
	}
	go func() {
		log.Printf("metatube REST API listening on %s", apiAddr)
		if err := apiSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}
	if err := apiSrv.Shutdown(ctx); err != nil {
		log.Printf("API server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}

func initDefaultUser(repo *repository.Repository) error {
	var count int64
	repo.DB().Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}
	hash, err := util.HashPassword("admin")
	if err != nil {
		return err
	}
	return repo.DB().Create(&model.User{
		Username:     "admin",
		PasswordHash: hash,
		Role:         "admin",
		Enabled:      true,
	}).Error
}

func silenceGorm(db *gorm.DB) {
	db.Config.Logger = gormlogger.Default.LogMode(gormlogger.Silent)
}
