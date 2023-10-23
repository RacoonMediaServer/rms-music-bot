package access

import (
	"context"
	"errors"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	rms_users "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-users"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/golang/protobuf/ptypes/empty"
	"go-micro.dev/v4/client"
	"math/rand"
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

func (s Service) GetAdminUserId() ([]int, error) {
	users, err := s.f.NewUsers().GetAdminUsers(context.TODO(), &empty.Empty{}, client.WithRequestTimeout(usersRequestTimeout))
	if err != nil {
		return nil, err
	}

	result := make([]int, 0, len(users.Users))
	for _, u := range users.Users {
		if u.TelegramUserID != nil {
			result = append(result, int(*u.TelegramUserID))
		}
	}

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	if len(result) == 0 {
		return nil, errors.New("no any admin user found")
	}

	return result, nil
}
