package github.com/jamesBan/sensitive/sensitive

import (
	"sensitive/store"
	"sensitive/filter"
	"time"
	"sync"
)

type Manager struct {
	store store.Store
	filter filter.Filter
	version uint64
	locker sync.RWMutex
	interval time.Duration
}


func (m *Manager) GetFilter() (filter.Filter)  {
	return m.filter
}

func (m *Manager) GetStore()(store.Store) {
	return m.store
}

func (m *Manager) checkVersion()  {
	time.AfterFunc(m.interval, func() {
		if m.store.Version() > m.version {
			m.locker.Lock()
			m.filter.UpdateAll(m.store)
			m.locker.Unlock()

			m.version = m.store.Version()
		}

		m.checkVersion()
	})
}

func NewManager(store store.Store, filter filter.Filter, interval time.Duration)(*Manager) {
	manager := &Manager{
		store:store,
		filter:filter,
		version:0,
		interval: interval,
	}

	go func() {
		manager.checkVersion()
	}()

	return manager
}