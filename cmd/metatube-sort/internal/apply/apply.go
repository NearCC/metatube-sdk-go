// Package apply executes the move plan. Resolved entries are moved into
// <dst>/<actor>/; unresolved entries are left untouched (per spec).
//
// On collision, the destination is renamed with a __N suffix — but the
// extension is preserved (a.mp4 -> a__1.mp4, never a.mp4__1).
package apply

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/metatube-community/metatube-sdk-go/cmd/metatube-sort/internal/plan"
)

// Run walks plans and renames each resolved entry to its target. Returns
// the first hard error; per-file failures are reported on stdout.
func Run(plans []plan.Entry, dstRoot string) error {
	if err := os.MkdirAll(dstRoot, 0o755); err != nil {
		return fmt.Errorf("mkdir dst: %w", err)
	}

	var moved, skipped int
	for _, e := range plans {
		if !e.Resolved() {
			skipped++
			continue
		}

		target := e.Target(dstRoot)
		actorDir := filepath.Dir(target)
		if err := os.MkdirAll(actorDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "  [err ] mkdir %s: %v\n", actorDir, err)
			continue
		}

		final, err := uniquePath(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [err ] %s: %v\n", e.Src, err)
			continue
		}

		if err := os.Rename(e.Src, final); err != nil {
			fmt.Fprintf(os.Stderr, "  [err ] rename %s -> %s: %v\n", e.Src, final, err)
			continue
		}
		fmt.Printf("  [ok  ] %s\n         -> %s\n", e.Src, final)
		moved++
	}
	fmt.Printf("\nmoved=%d skipped=%d\n", moved, skipped)
	return nil
}

// uniquePath returns target if it doesn't exist, otherwise target with a
// __1 / __2 / ... suffix inserted before the extension. The extension is
// never touched.
func uniquePath(target string) (string, error) {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return target, nil
	}
	dir, base := filepath.Split(target)
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]

	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s__%d%s", stem, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("too many collisions for %s", target)
}
