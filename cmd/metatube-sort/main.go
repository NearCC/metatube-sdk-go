package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/metatube-community/metatube-sdk-go/cmd/metatube-sort/internal/apply"
	"github.com/metatube-community/metatube-sdk-go/cmd/metatube-sort/internal/enginex"
	"github.com/metatube-community/metatube-sdk-go/cmd/metatube-sort/internal/parse"
	"github.com/metatube-community/metatube-sdk-go/cmd/metatube-sort/internal/plan"
	"github.com/metatube-community/metatube-sdk-go/cmd/metatube-sort/internal/resolve"
	"github.com/metatube-community/metatube-sdk-go/cmd/metatube-sort/internal/scan"
)

type config struct {
	srcDir  string
	dstDir  string
	dbDSN   string
	apply   bool
	workers int
	verbose bool
	proxy   string
	diag    bool
	delay   time.Duration
}

func parseFlags() *config {
	c := &config{}
	flag.StringVar(&c.srcDir, "src", "", "source directory to scan (required)")
	flag.StringVar(&c.dstDir, "dst", "", "destination directory (required)")
	flag.StringVar(&c.dbDSN, "db", "./metatube-sort.db", "SQLite DSN")
	flag.BoolVar(&c.apply, "apply", false, "actually move files (default: dry-run)")
	flag.IntVar(&c.workers, "j", 1, "concurrency (keep low — Cloudflare rate-limits burst traffic from the same IP)")
	flag.BoolVar(&c.verbose, "v", false, "verbose logging")
	flag.StringVar(&c.proxy, "proxy", "", "HTTP/SOCKS5 proxy override (otherwise uses SDK default / env)")
	flag.BoolVar(&c.diag, "diag", false, "run network diagnostics (probe JavBus + report egress IP) and exit")
	flag.DurationVar(&c.delay, "delay", 1*time.Second, "delay between JavBus requests (per worker). Raise if you see most requests time out / fail with EOF.")
	flag.Parse()

	if !c.diag && (c.srcDir == "" || c.dstDir == "") {
		fmt.Fprintln(os.Stderr, "ERROR: -src and -dst are required (or pass -diag)")
		flag.Usage()
		os.Exit(2)
	}
	if c.workers < 1 {
		c.workers = 1
	}
	return c
}

func main() {
	cfg := parseFlags()

	if cfg.diag {
		runDiag()
		return
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg *config) error {
	absSrc, _ := filepath.Abs(cfg.srcDir)
	absDst, _ := filepath.Abs(cfg.dstDir)
	if absSrc == absDst {
		return fmt.Errorf("src and dst must be different")
	}

	if !cfg.apply {
		fmt.Printf("=== DRY RUN (use --apply to actually move) ===\n")
	}
	fmt.Printf("src: %s\n", absSrc)
	fmt.Printf("dst: %s\n", absDst)
	fmt.Printf("db:  %s\n\n", cfg.dbDSN)

	eng, db, err := enginex.New(cfg.dbDSN, cfg.proxy)
	if err != nil {
		return fmt.Errorf("init engine: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	entries, err := scan.Entries(cfg.srcDir)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}
	fmt.Printf("found %d entries in %s\n\n", len(entries), cfg.srcDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := resolve.New(eng, db, cfg.verbose, cfg.proxy)
	if err != nil {
		return fmt.Errorf("init resolver: %w", err)
	}
	plans := resolveAll(ctx, entries, r, cfg.workers, cfg.verbose, cfg.delay)

	if cfg.verbose {
		for _, p := range plans {
			fmt.Printf("parsed: %s -> %s\n", p.Src, p.Number)
		}
		fmt.Println()
	}

	printer := plan.NewPrinter(os.Stdout)
	for _, p := range plans {
		printer.Print(p)
	}
	fmt.Printf("\ntotal: %d (matched=%d, unmatched=%d)\n",
		len(plans),
		countMatched(plans),
		countUnmatched(plans),
	)

	if !cfg.apply {
		return nil
	}

	fmt.Println()
	return apply.Run(plans, cfg.dstDir)
}

func resolveAll(
	ctx context.Context,
	entries []scan.Entry,
	r *resolve.Resolver,
	workers int,
	verbose bool,
	delay time.Duration,
) []plan.Entry {
	out := make([]plan.Entry, len(entries))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for i, e := range entries {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, e scan.Entry) {
			defer wg.Done()
			defer func() { <-sem }()

			number := parse.NumberFromName(e.Name)
			if number == "" {
				out[i] = plan.Entry{Src: e.Path, RawName: e.Name, IsDir: e.IsDir}
				return
			}

			// Throttle per-worker. Cloudflare rate-limits burst
			// requests from the same IP — even with -j 1 you can trip
			// it if you fire requests back-to-back. The negative
			// cache means even a slow run completes in bounded time:
			// failed numbers get cached and we never re-fetch them
			// within the TTL.
			if delay > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
				}
			}

			resCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			actor, ok := r.Resolve(resCtx, number)
			cancel()

			out[i] = plan.Entry{
				Src:     e.Path,
				RawName: e.Name,
				IsDir:   e.IsDir,
				Number:  number,
				Actor:   actor,
			}
			if !ok {
				if verbose {
					fmt.Printf("  [skip] %s (no actors)\n", e.Path)
				}
			}
		}(i, e)
	}
	wg.Wait()
	return out
}

func countMatched(p []plan.Entry) int {
	n := 0
	for _, x := range p {
		if x.Number != "" && x.Actor != "" {
			n++
		}
	}
	return n
}

func countUnmatched(p []plan.Entry) int {
	n := 0
	for _, x := range p {
		if x.Number == "" || x.Actor == "" {
			n++
		}
	}
	return n
}
