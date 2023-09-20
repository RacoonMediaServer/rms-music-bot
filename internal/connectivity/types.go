package connectivity

import (
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
	"github.com/RacoonMediaServer/rms-music-bot/internal/provider"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/go-openapi/runtime"
)

type Interlayer struct {
	Services        servicemgr.ServiceFactory
	Discovery       DiscoveryServiceFactory
	TorrentManager  TorrentManager
	Registry        registry.Registry
	ContentManager  ContentManager
	ContentProvider provider.ContentProvider
}

type TorrentManager interface {
	Add(content []byte) (string, error)
	GetFile(path string) ([]byte, error)
}

type ContentManager interface {
	AddContent(artistName string, content model.Content) error
	GetContent(artistName string) ([]model.Content, error)
}

type DiscoveryServiceFactory interface {
	New(token string) (*client.Client, runtime.ClientAuthInfoWriter)
}
