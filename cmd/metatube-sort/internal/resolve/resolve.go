// Package resolve turns a番号 into the display actor name.
//
// Lookup order:
//  1. local SQLite cache (table metatube_sort_alias, populated by previous
//     runs) — fastest, no network
//  2. JavBus direct HTTP for the movie's first actor (returns the
//     Japanese name)
//  3. Gfriends Filetree.json (best-effort: keys are like "葵つかさ.jpg",
//     values are studio IDs like "7-S1" — there is NO Chinese name
//     mapping in the public dataset, so this step usually misses)
//  4. fall back to the actor's original Japanese name
//
// Why Japanese names and not Chinese: the SDK has no provider that
// returns Chinese actor names out of the box. FANZA's Chinese branch is
// gated by region, and no other in-tree provider exposes `ChineseName`.
// To get Chinese names you would need to plug in OpenAI / DeepL
// translation, or a hand-curated jp→cn map. For now the CLI emits
// Japanese names and users can rename folders by hand.
package resolve

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/metatube-community/metatube-sdk-go/cmd/metatube-sort/internal/enginex"
	"github.com/metatube-community/metatube-sdk-go/engine"
	"gorm.io/gorm"
)

const (
	javbusProvider = "JavBus"
	gfriendsFolder = "https://raw.githubusercontent.com/gfriends/gfriends/master/Filetree.json"
	filetreeHTTPTO = 30 * time.Second
)

type aliasRow struct {
	Number        string `gorm:"column:number;primaryKey"`
	JPN           string `gorm:"column:jp_name"`
	CN            string `gorm:"column:cn_name"`
	LastAttemptAt int64  `gorm:"column:last_attempt_at"` // unix seconds
	LastStatus    string `gorm:"column:last_status"`     // "ok" | "fail"
}

func (*aliasRow) TableName() string { return "metatube_sort_alias" }

// Resolver is safe for concurrent use.
type Resolver struct {
	eng *engine.Engine
	db  *gorm.DB

	// Direct HTTP scraper for JavBus — bypasses the Engine / colly path
	// because that one gets blocked by Cloudflare.
	javbus *javbusHTTP

	verbose bool

	// Filetree cache: JP-name -> CN directory. Built lazily on first miss.
	once    sync.Once
	loadErr error
	names   map[string]string // JP name -> CN folder name
}

// New constructs a Resolver and creates the local alias table if missing.
// proxyOverride is the -proxy flag value, or "" to fall back to env /
// system proxy auto-detection.
func New(eng *engine.Engine, db *enginex.GormDB, verbose bool, proxyOverride string) (*Resolver, error) {
	if err := db.Raw().AutoMigrate(&aliasRow{}); err != nil {
		return nil, fmt.Errorf("migrate alias table: %w", err)
	}
	return &Resolver{
		eng:     eng,
		db:      db.Raw(),
		javbus:  newJavbusHTTP(proxyOverride),
		verbose: verbose,
	}, nil
}

// negativeCacheTTL is how long a failed lookup is remembered before
// being retried. Cloudflare / JavBus rate-limiting is bursty, so a
// short TTL is enough to stop hammering on the same number for a
// single CLI run while still recovering within hours.
const negativeCacheTTL = 6 * time.Hour

// Resolve returns the display name for the first actor of the movie
// identified by number, or ("", false) if the number can't be resolved.
//
// Behavior:
//   - On positive cache hit (last_status="ok"): return cached actor.
//   - On negative cache hit within negativeCacheTTL: return "", false
//     immediately without contacting JavBus. This prevents the
//     "always-match-a-few-then-stop" pathology where transient
//     failures weren't cached so the next run would re-attempt and
//     randomly succeed on a different subset of the input.
//   - Otherwise, contact JavBus. On success, write positive cache and
//     return. On failure, write negative cache with current timestamp.
func (r *Resolver) Resolve(ctx context.Context, number string) (string, bool) {
	row, err := r.lookupLocal(number)
	if err == nil {
		switch row.LastStatus {
		case "ok":
			if r.verbose {
				fmt.Printf("  [res ] %s: cache hit (%s)\n", number, row.actorName())
			}
			return row.actorName(), true
		case "fail":
			age := time.Since(time.Unix(row.LastAttemptAt, 0))
			if age < negativeCacheTTL {
				if r.verbose {
					fmt.Printf("  [res ] %s: negative cache (%s ago, retry at %s)\n",
						number, age.Round(time.Minute),
						time.Unix(row.LastAttemptAt, 0).Add(negativeCacheTTL).Format("15:04:05"))
				}
				return "", false
			}
			if r.verbose {
				fmt.Printf("  [res ] %s: negative cache expired (%s ago), retrying\n",
					number, age.Round(time.Minute))
			}
		}
	}

	info, err := r.javbus.fetchMovieActors(ctx, number)
	if err != nil || len(info) == 0 {
		_ = r.markFail(number)
		return "", false
	}
	first := info[0]

	cn, _ := r.lookupChinese(first)
	if cn == "" {
		cn = first // fallback to original name
	}
	_ = r.markOk(number, first, cn)
	return cn, true
}

// lookupLocal returns the aliasRow for number, or an error if no row.
func (r *Resolver) lookupLocal(number string) (aliasRow, error) {
	var row aliasRow
	err := r.db.Where("number = ?", number).First(&row).Error
	return row, err
}

// actorName returns the best display name cached on the row. Prefers
// the Chinese alias if present, otherwise falls back to the Japanese.
func (row *aliasRow) actorName() string {
	if row.CN != "" {
		return row.CN
	}
	return row.JPN
}

func (r *Resolver) markOk(number, jp, cn string) error {
	return r.db.Exec(`
		INSERT INTO metatube_sort_alias (number, jp_name, cn_name, last_attempt_at, last_status)
		VALUES (?, ?, ?, ?, 'ok')
		ON CONFLICT(number) DO UPDATE SET
			jp_name = excluded.jp_name,
			cn_name = excluded.cn_name,
			last_attempt_at = excluded.last_attempt_at,
			last_status = 'ok'
	`, number, jp, cn, time.Now().Unix()).Error
}

func (r *Resolver) markFail(number string) error {
	return r.db.Exec(`
		INSERT INTO metatube_sort_alias (number, jp_name, cn_name, last_attempt_at, last_status)
		VALUES (?, '', '', ?, 'fail')
		ON CONFLICT(number) DO UPDATE SET
			last_attempt_at = excluded.last_attempt_at,
			last_status = 'fail'
	`, number, time.Now().Unix()).Error
}

// lookupChinese finds the Chinese directory name for a Japanese actor
// name, by querying the Gfriends Filetree.json. The JSON shape is:
//
//	{"Content": {"<CN folder>": {"<JP name>": "<filename>.jpg", ...}, ...}, ...}
//
// So we build (JP name -> CN folder) at startup.
func (r *Resolver) lookupChinese(jp string) (string, bool) {
	r.once.Do(r.loadFiletree)
	if r.names != nil {
		if cn, ok := r.names[jp]; ok {
			return cn, true
		}
	}
	return "", false
}

func (r *Resolver) loadFiletree() {
	r.names = make(map[string]string)
	client := &http.Client{Timeout: filetreeHTTPTO}
	resp, err := client.Get(gfriendsFolder)
	if err != nil {
		r.loadErr = err
		if r.verbose {
			fmt.Printf("filetree fetch failed: %v\n", err)
		}
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		r.loadErr = fmt.Errorf("filetree http %d", resp.StatusCode)
		return
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		r.loadErr = err
		return
	}
	var ft struct {
		Content map[string]map[string]string `json:"Content"`
	}
	if err := json.Unmarshal(body, &ft); err != nil {
		r.loadErr = err
		return
	}
	for cn, actors := range ft.Content {
		for jp := range actors {
			// Filenames in gfriends are sometimes "<jp>.<ext>" — Trim
			// strips the extension if present (it doesn't, but harmless).
			// Multiple folders may share the same JP key; first wins.
			if _, exists := r.names[jp]; !exists {
				r.names[jp] = cn
			}
		}
	}
	if r.verbose {
		fmt.Printf("filetree loaded: %d actors\n", len(r.names))
	}
}
