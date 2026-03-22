package service

import "sync"

type memCache struct {
	mu sync.Mutex
	users map[int]struct{}
}

func newCache() *memCache {
	return &memCache{
		users: make(map[int]struct{}),
	}
}

func (m *memCache) SetUser(userID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[userID] = struct{}{}
}

func (m *memCache) GetCountUsers() int {
	return len(m.users)
}