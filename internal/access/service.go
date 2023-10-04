package access

import (
	"context"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	rms_users "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-users"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"go-micro.dev/v4/client"
	"time"
)

const usersRequestTimeout = 15 * time.Second

type Service struct {
	config config.UserControl
	f      servicemgr.ServiceFactory
}

func New(f servicemgr.ServiceFactory, conf config.Configuration) *Service {
	return &Service{
		config: conf.UserControl,
		f:      f,
	}
}

func (s Service) CheckAccess(telegramUserId int) (ok bool, token string, err error) {
	if !s.config.Enabled {
		ok = true
		token = s.config.DefaultToken
		return
	}

	req := rms_users.GetUserByTelegramIdRequest{TelegramUserId: int32(telegramUserId)}
	resp := &rms_users.User{}
	resp, err = s.f.NewUsers().GetUserByTelegramId(context.TODO(), &req, client.WithRequestTimeout(usersRequestTimeout))
	if err != nil {
		return
	}

	for _, perm := range resp.Perms {
		if perm == rms_users.Permissions_ListeningMusic {
			ok = true
			break
		}
	}

	if resp.Token != nil {
		token = *resp.Token
	}
	return
}
