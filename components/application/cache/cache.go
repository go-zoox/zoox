package cache

import (
	"sync"

	extcache "github.com/go-zoox/cache"
)

var (
	once     sync.Once
	instance extcache.Cache
)

// Set sets the global application cache instance.
func Set(c extcache.Cache) {
	if c == nil {
		panic("application cache cannot be nil")
	}

	once.Do(func() {
		instance = c
	})
}

// Get returns the global application cache instance.
func Get() extcache.Cache {
	if instance == nil {
		panic("application cache is not initialized")
	}

	return instance
}
