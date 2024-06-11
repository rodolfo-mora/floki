package floki

import "sync"

type MemoryStore struct {
	Users map[string]User
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (m *MemoryStore) UserExists(user string) bool {
	_, e := m.Users[user]
	return e
}

func (m *MemoryStore) Save(user string, groups []string) error {
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()
	m.Users[user] = User{Email: user, SSOGroups: groups}
	return nil
}

func (m *MemoryStore) GetSSOGroups(user string) []string {
	return m.Users[user].SSOGroups
}
