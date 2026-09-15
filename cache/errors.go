package cache

import (
	"github.com/redis/go-redis/v9"
)

// ErrNotFound is returned by read operations, when key or sorted set member does
// not exist.
//
// It is an alias of the driver's sentinel on purpose: consumers stop importing
// redis driver just to tell "no such key" from a real failure, while code, which
// already checks redis.Nil, keeps working.
var ErrNotFound = redis.Nil
