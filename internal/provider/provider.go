package provider

import (
	"errors"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/heuristic"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var ErrNotFound = errors.New("nothing found")

type ContentProvider interface {
	FindTrack(basePath, trackName string) ([]byte, error)
}

type contentProvider struct {
	directory string
}

func NewContentProvider(directory string) ContentProvider {
	return &contentProvider{directory: directory}
}

func (p contentProvider) FindTrack(basePath, trackName string) ([]byte, error) {
	absolutePath := filepath.Join(p.directory, basePath)
	trackName = heuristic.Normalize(trackName)
	found := ""

	err := filepath.WalkDir(absolutePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		normalized := heuristic.Normalize(d.Name())
		if strings.Index(normalized, trackName) >= 0 {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == "" {
		return nil, ErrNotFound
	}
	return os.ReadFile(found)
}
