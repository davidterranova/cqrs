package eventsourcing

import (
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/rs/zerolog/log"
)

type Cache[K comparable, V any] interface {
	Add(key K, value V) bool
	Get(key K) (V, bool)
}

type CacheOption struct {
	Disabled bool
	Size     int
	TTL      time.Duration
}

// NewCache returns a new cache
// Disabled parameter set to true or Size parameter below 0 turns cache off.
// Size parameter set to 0 makes cache of unlimited size, e.g. turns LRU mechanism off.
// Providing 0 TTL turns expiring off.
func NewCache[K comparable, V any](option CacheOption) Cache[K, V] {
	if option.Disabled || option.Size < 0 {
		log.Info().Msg("cache disabled")
		return &noopCache[K, V]{}
	}

	levent := log.Info()
	if option.Size == 0 {
		levent.Str("size", "unlimited")
	} else {
		levent.Int("size", option.Size)
	}
	if option.TTL == 0 {
		levent.Str("ttl", "unlimited")
	} else {
		levent.Dur("ttl", option.TTL)
	}
	levent.Msg("cache enabled")

	return NewCacheLogger[K, V](
		lru.NewLRU[K, V](option.Size, nil, option.TTL),
	)
}

type noopCache[K any, V any] struct{}

func (c *noopCache[K, V]) Add(key K, value V) bool {
	return false
}

func (c *noopCache[K, V]) Get(key K) (V, bool) {
	var v V
	return v, false
}

type cacheLogger[K comparable, V any] struct {
	cache Cache[K, V]
}

func NewCacheLogger[K comparable, V any](cache Cache[K, V]) Cache[K, V] {
	return &cacheLogger[K, V]{
		cache: cache,
	}
}

func (c *cacheLogger[K, V]) Add(key K, value V) bool {
	log.Info().
		Str("key", fmt.Sprintf("%v", key)).
		Msg("cache add")
	return c.cache.Add(key, value)
}

func (c *cacheLogger[K, V]) Get(key K) (V, bool) {
	v, ok := c.cache.Get(key)
	log.Info().
		Str("key", fmt.Sprintf("%v", key)).
		Bool("ok", ok).
		Msg("cache get")

	return v, ok
}
