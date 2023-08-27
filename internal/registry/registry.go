package registry

import (
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

const cleanInterval = 1 * time.Minute

type Registry interface {
	Add(value interface{}, keepDuration time.Duration) (key string)
	Get(key string) (interface{}, bool)
}

type registry struct {
	mu      sync.RWMutex
	storage map[string]*element
}

type element struct {
	expired time.Time
	content interface{}
}

func New() Registry {
	r := registry{
		storage: map[string]*element{},
	}
	go r.cleaner()
	return &r
}

func (r *registry) cleaner() {
	t := time.NewTicker(cleanInterval)
	defer t.Stop()
	for {
		now := <-t.C
		r.mu.Lock()
		var keys []string
		for k, v := range r.storage {
			if now.After(v.expired) {
				keys = append(keys, k)
			}
		}
		for _, k := range keys {
			delete(r.storage, k)
		}
		r.mu.Unlock()
	}
}

func (r *registry) Add(value interface{}, keepDuration time.Duration) (key string) {
	key = uuid.NewV4().String()

	r.mu.Lock()
	defer r.mu.Unlock()

	item := element{
		expired: time.Now().Add(keepDuration),
		content: value,
	}
	r.storage[key] = &item
	return
}

func (r *registry) Get(key string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.storage[key]
	if !ok {
		return nil, false
	}
	return item.content, true
}
