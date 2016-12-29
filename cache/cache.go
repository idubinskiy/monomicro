package cache

import "time"

// Interface represents the minimum implementation of a key-value cache with expiry.
//
// All methods should be safe for concurrent access.
type Interface interface {
	Set(key string, value []byte, timeout time.Duration) error
	Get(key string) (value []byte, err error)
	Delete(key string) error
}
