package urlcache

import (
	"sync"
	"time"
)

// UrlCache is responsible for providing cache-like functional.
// It will say to users of cache whether certain url should be parsed.
// Such clarifications are based on TTL.
// e.g.:
//		UrlCache.Set(url) // Adds url to cache with defined TTL
//		UrlCache.ShouldParse(url) // Will return false(do not parse) until TTL is expired

type UrlCacher interface {
	// Sets url to cache with defined TTL inside implementation.
	// If url already exists then Set will overwrite existing url
	Set(url string)
	// Returns true if url can be parsed.
	// Otherwise returns false. Url is still present in cache.
	// After ShouldParse returns true url is automatically deleted from cache.
	ShouldParse(url string) bool
}

type UrlCache struct {
	mu    *sync.RWMutex
	cache map[string]time.Time
	ttl   time.Duration
}

func NewUrlCache(ttl time.Duration) UrlCacher {
	return &UrlCache{
		mu:    new(sync.RWMutex),
		cache: make(map[string]time.Time, 0),
		ttl:   ttl,
	}
}

func (u *UrlCache) Set(url string) {
	expiresAt := time.Now().Add(u.ttl)
	u.mu.Lock()
	u.cache[url] = expiresAt
	u.mu.Unlock()
}

func (u *UrlCache) ShouldParse(url string) bool {

	u.mu.RLock()
	itemExpiresAt := u.cache[url].Unix()
	u.mu.RUnlock()

	curr := time.Now().Unix()

	// Not yet expired
	// --------CURR---ITEMEXP--->t
	if curr < itemExpiresAt {
		return false
	}

	defer func() {
		// Invalidate url in cache
		u.mu.Lock()
		delete(u.cache, url)
		u.mu.Unlock()
	}()

	return true
}
