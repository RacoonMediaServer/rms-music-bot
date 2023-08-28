package add

import (
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
)

func selectTorrent(list []*models.SearchTorrentsResult) *models.SearchTorrentsResult {
	return list[0]
}
