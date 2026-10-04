// Package scan reads a single directory and returns its top-level entries.
//
// Non-recursive by design: caller wants to match either a file or a folder
// at the same level, not everything under it.
package scan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var videoExts = map[string]struct{}{
	".mp4":  {},
	".mkv":  {},
	".avi":  {},
	".wmv":  {},
	".ts":   {},
	".mov":  {},
	".m4v":  {},
	".webm": {},
	".flv":  {},
	".mpg":  {},
	".mpeg": {},
	".rmvb": {},
	".strm": {},
}

// Entry describes a single top-level item.
type Entry struct {
	Name  string
	Path  string
	IsDir bool
}

// Entries returns the top-level contents of src, in directory order.
// Files are filtered to known video extensions; directories are kept as-is
// (caller decides what to do with a directory entry — typically by parsing
// the folder name).
func Entries(src string) ([]Entry, error) {
	dir, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", src, err)
	}
	out := make([]Entry, 0, len(dir))
	for _, d := range dir {
		name := d.Name()
		path := filepath.Join(src, name)
		isDir := d.IsDir()
		if !isDir && !isVideo(name) {
			continue
		}
		out = append(out, Entry{Name: name, Path: path, IsDir: isDir})
	}
	return out, nil
}

func isVideo(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	_, ok := videoExts[ext]
	return ok
}
