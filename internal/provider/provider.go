package provider

import (
	"errors"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/heuristic"
	"github.com/antzucaro/matchr"
	"go-micro.dev/v4/logger"
	"io/fs"
	"math"
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
	absolutePath := filepath.Join(p.directory, "music", basePath)
	trackName = heuristic.Normalize(trackName)
	found := ""
	minDistance := math.MaxInt
	similarPath := ""

	err := filepath.WalkDir(absolutePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".mp3" {
			return nil
		}

		normalized := heuristic.Normalize(d.Name())
		pos := strings.Index(normalized, trackName)
		distance := matchr.Levenshtein(normalized, trackName)
		if distance < minDistance {
			minDistance = distance
			similarPath = path
		}
		if pos >= 0 {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == "" {
		if minDistance != math.MaxInt && minDistance < len(trackName)/2 {
			logger.Debugf("Approximated result for '%s' = %s, distance = %d", trackName, similarPath, minDistance)
			return os.ReadFile(similarPath)
		}
		return nil, ErrNotFound
	}
	return os.ReadFile(found)
}
