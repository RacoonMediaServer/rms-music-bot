package connectivity

import (
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

type Interlayer struct {
	Services   servicemgr.ServiceFactory
	Discovery  DiscoveryServiceFactory
	Downloader Downloader
}

type Downloader interface {
	Download(content []byte) ([]string, error)
	GetFile(path string) ([]byte, error)
}

type DiscoveryServiceFactory interface {
	New(token string) (*client.Client, runtime.ClientAuthInfoWriter)
}

type discoveryFactory struct {
	remote config.Remote
}

func New(remote config.Remote, factory servicemgr.ClientFactory) Interlayer {
	return Interlayer{
		Services:  servicemgr.NewServiceFactory(factory),
		Discovery: &discoveryFactory{remote: remote},
	}
}

func (d discoveryFactory) New(token string) (*client.Client, runtime.ClientAuthInfoWriter) {
	tr := httptransport.New(d.remote.Host, d.remote.Path, []string{d.remote.Scheme})
	auth := httptransport.APIKeyAuth("X-Token", "header", token)
	discoveryClient := client.New(tr, strfmt.Default)
	return discoveryClient, auth
}
