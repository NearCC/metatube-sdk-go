// Package enginex constructs an Engine and its underlying *gorm.DB for the
// CLI. It also triggers blank imports of the providers we use — keeping the
// SDK source untouched and letting the binary pick up just what it needs.
package enginex

import (
	"fmt"
	"time"

	"github.com/metatube-community/metatube-sdk-go/database"
	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/internal/envconfig"
	"github.com/metatube-community/metatube-sdk-go/model"
	"gorm.io/gorm/logger"

	_ "github.com/metatube-community/metatube-sdk-go/provider/gfriends"
	_ "github.com/metatube-community/metatube-sdk-go/provider/javbus"
)

// New opens the SQLite DB, auto-migrates the metadata tables, and returns
// a ready Engine plus the *gorm.DB (caller owns the close).
//
// proxy is an optional HTTP/SOCKS5 URL applied to all providers via the
// engine's standard `proxy` config key — this is the documented SDK way
// to route through a residential proxy so Cloudflare stops blocking the
// data-center / ISP IP you're on.
func New(dsn, proxy string) (*engine.Engine, *GormDB, error) {
	db, err := database.Open(&database.Config{
		DSN:      dsn,
		LogLevel: logger.Silent,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.AutoMigrate(
		&model.MovieInfo{},
		&model.ActorInfo{},
		&model.MovieReviewInfo{},
	); err != nil {
		return nil, nil, fmt.Errorf("migrate: %w", err)
	}

	opts := []engine.Option{engine.WithRequestTimeout(45 * time.Second)}
	if proxy != "" {
		// The SDK wires proxy through envconfig.Config under the "proxy"
		// key, which applyProviderConfig forwards to the provider's
		// ProxySetter.SetProxy method. Same path the metatube-server uses.
		proxyCfg := envconfig.NewConfig()
		proxyCfg.Set("proxy", proxy)
		opts = append(opts,
			engine.WithMovieProviderConfig("JavBus", proxyCfg),
			engine.WithActorProviderConfig("Gfriends", proxyCfg),
		)
	}
	eng := engine.New(db, opts...)
	return eng, wrapDB(db), nil
}
