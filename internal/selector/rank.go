package selector

import "github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"

type rankFunc func(list []*models.SearchTorrentsResult) []float32

func makeRankFunc(funcs ...rankFunc) rankFunc {
	return func(list []*models.SearchTorrentsResult) []float32 {
		ranks := make([]float32, len(list))
		for _, f := range funcs {
			fRanks := f(list)
			for i := range ranks {
				ranks[i] += fRanks[i]
			}
		}
		return ranks
	}
}
