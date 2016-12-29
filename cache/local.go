package cache

import (
	"sync"
	"time"
)

// Local implements cache.Interface as an in-memory cache.
//
// No guarantees are made about performance or efficiency.
type Local struct {
	cache         map[string][]byte
	cacheMutex    *sync.RWMutex
	timeouts      map[string]*time.Timer
	timeoutsMutex *sync.Mutex
}

// NewLocal initializes and returns a Local in-memory cache.
func NewLocal() *Local {
	return &Local{
		cache:         map[string][]byte{},
		cacheMutex:    &sync.RWMutex{},
		timeouts:      map[string]*time.Timer{},
		timeoutsMutex: &sync.Mutex{},
	}
}

// Set sets the value of key in the cache to value.
//
// A timeout of 0 (or less) means the key will never expire.
//
// A positive timeout will cause the key to be deleted once the timeout expires.
// If a timeout was set on the key previously, it will be re-set to the new timeout.
func (l *Local) Set(key string, value []byte, timeout time.Duration) error {
	l.cacheMutex.Lock()
	l.cache[key] = value
	l.cacheMutex.Unlock()

	l.timeoutsMutex.Lock()
	if timer, ok := l.timeouts[key]; ok {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		delete(l.timeouts, key)
	}

	if timeout > 0 {
		timer := time.AfterFunc(timeout, func() {
			l.Delete(key)
		})
		l.timeouts[key] = timer
	}
	l.timeoutsMutex.Unlock()

	return nil
}

// Get returns the cached value of key. If the key does not exist in the cache,
// the returned value will be an empty slice.
func (l *Local) Get(key string) (value []byte, err error) {
	l.cacheMutex.RLock()
	value = l.cache[key]
	l.cacheMutex.RUnlock()

	return value, nil
}

// Delete removes key from the cache.
func (l *Local) Delete(key string) error {
	l.cacheMutex.Lock()
	delete(l.cache, key)
	l.cacheMutex.Unlock()

	l.timeoutsMutex.Lock()
	delete(l.timeouts, key)
	l.timeoutsMutex.Unlock()

	return nil
}
