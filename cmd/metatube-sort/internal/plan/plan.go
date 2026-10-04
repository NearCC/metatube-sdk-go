// Package plan defines the per-entry plan produced by the resolver and the
// printer that displays it.
package plan

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// Entry is one resolved (or unresolved) source item, ready to be moved.
type Entry struct {
	Src     string // absolute source path
	RawName string // base filename / foldername as scanned
	IsDir   bool
	Number  string // parsed番号; "" if not parseable
	Actor   string // resolved actor display name (Chinese preferred); "" if unresolved
}

// Resolved is true when Number and Actor are both present.
func (e Entry) Resolved() bool {
	return e.Number != "" && e.Actor != ""
}

// Target returns the proposed destination path under dst. It does NOT
// perform any disk IO or collision handling — that's apply's job.
func (e Entry) Target(dst string) string {
	if !e.Resolved() {
		return ""
	}
	return filepath.Join(dst, sanitizeFolderName(e.Actor), e.RawName)
}

func sanitizeFolderName(name string) string {
	// Cross-platform safety: strip characters that are illegal or painful
	// in Windows / NTFS paths. We keep spaces and most unicode (Chinese
	// names pass through untouched).
	illegal := `<>:"/\|?*`
	var b strings.Builder
	for _, r := range name {
		if strings.ContainsRune(illegal, r) || r < 32 {
			continue
		}
		b.WriteRune(r)
	}
	s := strings.TrimSpace(b.String())
	if s == "" {
		s = "_"
	}
	return s
}

// Printer renders plans as a two-column table.
type Printer struct {
	w io.Writer
}

func NewPrinter(w io.Writer) *Printer { return &Printer{w: w} }

func (p *Printer) Print(e Entry) {
	if !e.Resolved() {
		fmt.Fprintf(p.w, "  [skip]  %s\n", e.Src)
		return
	}
	fmt.Fprintf(p.w, "  [move]  %s\n         -> %s\n", e.Src, e.Target(""))
}
