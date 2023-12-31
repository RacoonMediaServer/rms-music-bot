package chatting

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
	"sync"
	"time"
)

const resetStateInterval = 12 * time.Hour

type userChat struct {
	mu            sync.Mutex
	accessService connectivity.AccessService
	state         chatState

	checkAccessTime time.Time
	prevCommand     string

	record *model.Chat
	token  string
}

func newUserChat(interlayer connectivity.Interlayer, chatRecord *model.Chat) *userChat {
	return &userChat{
		accessService: interlayer.AccessService,
		record:        chatRecord,
		prevCommand:   "search",
	}
}

func (c *userChat) requestState() (chatState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != stateAccessGranted || time.Since(c.checkAccessTime) > resetStateInterval {
		granted, token, err := c.accessService.CheckAccess(c.record.UserID)
		if err != nil {
			return c.state, err
		}
		if granted {
			c.token = token
			c.state = stateAccessGranted
			c.checkAccessTime = time.Now()
		} else if c.state == stateAccessGranted {
			// TODO: как навсегда забанить пользователя ?
			c.state = stateNoAccess
		}
	}

	return c.state, nil
}

func (c *userChat) loadPrevCommand() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.prevCommand
}

func (c *userChat) savePrevCommand(command string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.prevCommand = command
}

func (c *userChat) getToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.token
}
