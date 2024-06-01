package session

import (
	"sync"
	"time"
)

type Manager struct {
	Users map[int64]*User
	mu    sync.Mutex
}

type User struct {
	ID      int64
	Name    string
	Request string
	Action  string
	IDList  []string
	timer   *time.Timer
}

func New() *Manager {
	return &Manager{
		Users: make(map[int64]*User),
		mu:    sync.Mutex{},
	}
}

func (m *Manager) timer(id int64) *time.Timer {
	return time.AfterFunc(time.Hour, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.Users, id)
	})
}

func (m *Manager) AddUser(id int64, username string) *User {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Users[id] = &User{
		ID:      id,
		Name:    username,
		Request: "",
		Action:  "default",
		IDList:  nil,
		timer:   m.timer(id),
	}
	return m.Users[id]
}

func (m *Manager) GetUser(id int64) *User {
	user, ok := m.Users[id]
	if !ok {
		return nil
	}
	return user
}

func (u *User) updateTimer() {
	if !u.timer.Stop() {
		<-u.timer.C
	}
	u.timer.Reset(time.Hour)
}
