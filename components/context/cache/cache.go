package cache

import (
	"time"

	"github.com/go-zoox/kv"
	"github.com/go-zoox/kv/typing"
)

type Config = typing.Config

// Cache ...
type Cache interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}, ttl time.Duration) error
	Del(key string) error
}

type cache struct {
	core kv.KV
}

func New(cfg *Config) Cache {
	if cfg == nil {
		cfg = &typing.Config{
			Engine: "memory",
		}
	}

	core, err := kv.New(cfg)
	if err != nil {
		panic(err)
	}

	return &cache{
		core: core,
	}
}

// Get ...
func (c *cache) Get(key string, value interface{}) error {
	return c.core.Get(key, value)
}

// Set ...
func (c *cache) Set(key string, value interface{}, ttl time.Duration) error {
	return c.core.Set(key, value, ttl)
}

// Del ...
func (c *cache) Del(key string) error {
	return c.core.Delete(key)
}
