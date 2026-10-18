package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Writer appends to a log file, rotating it once it grows past maxBytes.
// Rotated files are kept as <path>.1 .. <path>.<maxFiles-1>; older ones are
// deleted. A non-positive maxBytes or maxFiles disables rotation.
type Writer struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	maxFiles int
	file     *os.File
}

func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return 0, os.ErrClosed
	}
	if w.maxBytes > 0 && w.maxFiles > 0 {
		if fi, err := w.file.Stat(); err == nil && fi.Size()+int64(len(p)) > w.maxBytes {
			if err := w.rotate(); err != nil {
				return 0, err
			}
		}
	}
	return w.file.Write(p)
}

func (w *Writer) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}
	for i := w.maxFiles - 2; i >= 1; i-- {
		from := fmt.Sprintf("%s.%d", w.path, i)
		if _, err := os.Stat(from); err == nil {
			os.Remove(fmt.Sprintf("%s.%d", w.path, i+1))
			os.Rename(from, fmt.Sprintf("%s.%d", w.path, i+1))
		}
	}
	os.Remove(w.path + ".1")
	if err := os.Rename(w.path, w.path+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = f
	return nil
}

// Path returns the active log file path.
func (w *Writer) Path() string {
	return w.path
}

// Close releases the underlying file handle.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

// Init sends the standard logger to both stdout and a rotating file at
// dataDir/logs/filename. The returned Writer must be closed on shutdown.
func Init(dataDir, filename string, maxSizeMB, maxFiles int) (*Writer, error) {
	logDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(logDir, filename)

	var w *Writer
	if maxSizeMB > 0 && maxFiles > 0 {
		w = &Writer{path: path, maxBytes: int64(maxSizeMB) * 1024 * 1024, maxFiles: maxFiles}
		if err := w.open(); err != nil {
			return nil, err
		}
	} else {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		w = &Writer{path: path, file: f}
	}

	log.SetOutput(io.MultiWriter(os.Stdout, w))
	log.SetFlags(log.LstdFlags)
	return w, nil
}

func (w *Writer) open() error {
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = f
	return nil
}