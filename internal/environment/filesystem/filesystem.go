package filesystem

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ko1N/zeebe-video-service/internal/storage"
)

type FileSystem interface {
	AddInput(input *storage.FileUrl) error
	AddOutput(output *storage.FileUrl) error

	RootPath() string
	ToFullPath(file *storage.FileUrl) (string, error)

	// Flush will ensure all the files have been uploaded
	Flush() error
	// Close will close the filesystem entirely
	Close() error
}

func IsSubPath(parent, sub string) (bool, error) {
	up := ".." + string(os.PathSeparator)

	// path-comparisons using filepath.Abs don't work reliably according to docs (no unique representation).
	rel, err := filepath.Rel(parent, sub)
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(rel, up) && rel != ".." {
		return true, nil
	}
	return false, nil
}
