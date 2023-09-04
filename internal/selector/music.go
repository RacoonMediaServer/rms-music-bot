package selector

import (
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"github.com/antzucaro/matchr"
	"go-micro.dev/v4/logger"
	"strings"
)

type MusicSelector struct {
	Query               string
	Discography         bool
	MinSeedersThreshold int64
	MaxSizeMB           int64
	Format              string
}

func (s MusicSelector) Select(list []*models.SearchTorrentsResult) *models.SearchTorrentsResult {
	if len(list) == 0 {
		return nil
	}

	funcs := []rankFunc{s.limitBySize, s.rankBySeeders, s.rankByFormat}
	funcs = append(funcs, s.getRankByTextFunc()...)

	rank := makeRankFunc(funcs...)
	ranks := rank(list)
	_, _, best := findMax(ranks, func(elem float32) float32 {
		return elem
	})
	for i := range ranks {
		logger.Tracef("%d rank: %.4f", i, ranks[i])
	}
	sel := list[best]
	logger.Infof("Selected { Title: %s, Format: %s, Size: %d, Seeders: %d }", getString(sel.Title), sel.Format, getValue(sel.Size), getValue(sel.Seeders))
	return sel
}

func (s MusicSelector) limitBySize(list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))

	for i, t := range list {
		size := getValue(t.Size)

		if size >= s.MaxSizeMB {
			ranks[i] = -1
			logger.Tracef("%d limit by size: %.4f", i, ranks[i])
		}
	}
	return ranks
}

func (s MusicSelector) rankBySeeders(list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))
	_, max, _ := findMax(list, func(t *models.SearchTorrentsResult) int64 {
		return getValue(t.Seeders)
	})

	for i, t := range list {
		seeders := getValue(t.Seeders)
		if seeders < s.MinSeedersThreshold {
			ranks[i] = float32(seeders) / float32(max)
		} else {
			ranks[i] = 1
		}
		logger.Tracef("%d rank by seeders: %.4f", i, ranks[i])
	}
	return ranks
}

func (s MusicSelector) rankByFormat(list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))
	for i, t := range list {
		if t.Format == s.Format {
			ranks[i] = 1
		}
	}
	return ranks
}

func (s MusicSelector) getRankByTextFunc() []rankFunc {
	var result []rankFunc
	if s.Discography {
		result = append(result, func(list []*models.SearchTorrentsResult) []float32 {
			return s.rankByText(s.Query+" дискография", list)
		})
		result = append(result, func(list []*models.SearchTorrentsResult) []float32 {
			return s.rankByText(s.Query+" discography", list)
		})
	} else {
		result = append(result, func(list []*models.SearchTorrentsResult) []float32 {
			return s.rankByText(s.Query, list)
		})
	}
	return result
}

func (s MusicSelector) rankByText(query string, list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))
	distance := make([]int, len(list))

	target := strings.ToLower(query)
	for i, t := range list {
		title := strings.ToLower(*t.Title)
		distance[i] = matchr.Levenshtein(title, target)
	}

	_, max, _ := findMax(distance, func(elem int) int {
		return elem
	})
	for j, d := range distance {
		ranks[j] = 1 - (float32(d) / float32(max))
		logger.Tracef("%d rank by voice: %.4f", j, ranks[j])
	}
	return ranks
}
