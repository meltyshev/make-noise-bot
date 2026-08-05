package bot

import (
	"context"
	"sync"
	"time"
)

const convTTL = time.Hour

// conversation points the next plain message from a user in a chat to a
// command, carrying its intermediate state.
type conversation struct {
	Name    string
	State   any
	expires time.Time
}

type convKey struct {
	UserID int64
	ChatID int64
}

type convStore struct {
	mu      sync.Mutex
	entries map[convKey]conversation
}

func newConvStore() *convStore {
	return &convStore{entries: map[convKey]conversation{}}
}

func (s *convStore) Get(userID, chatID int64) (conversation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := convKey{UserID: userID, ChatID: chatID}
	entry, ok := s.entries[key]
	if !ok {
		return conversation{}, false
	}
	if time.Now().After(entry.expires) {
		delete(s.entries, key)
		return conversation{}, false
	}
	return entry, true
}

func (s *convStore) Set(userID, chatID int64, name string, state any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[convKey{UserID: userID, ChatID: chatID}] = conversation{
		Name:    name,
		State:   state,
		expires: time.Now().Add(convTTL),
	}
}

func (s *convStore) Delete(userID, chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, convKey{UserID: userID, ChatID: chatID})
}

func (s *convStore) StartJanitor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(convTTL)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				s.mu.Lock()
				for key, entry := range s.entries {
					if now.After(entry.expires) {
						delete(s.entries, key)
					}
				}
				s.mu.Unlock()
			}
		}
	}()
}
