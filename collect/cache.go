package collect

import (
	"sync"

	"github.com/danclive/nson-go"
)

var _cache map[string]nson.Value
var _tick map[string]nson.Value
var _rwlock sync.RWMutex

func initCache() {
	_cache = make(map[string]nson.Value)
	_tick = make(map[string]nson.Value)
}

func CacheGet(key string) nson.Value {
	_rwlock.RLock()
	defer _rwlock.RUnlock()

	if v, ok := _cache[key]; ok {
		return v
	}

	return nil
}

func CacheSet(key string, value nson.Value) {
	_rwlock.Lock()
	defer _rwlock.Unlock()

	_cache[key] = value
	_tick[key] = value
}

func CacheDel(key string) {
	_rwlock.Lock()
	defer _rwlock.Unlock()

	delete(_cache, key)
	delete(_tick, key)
}

func CacheTick() map[string]nson.Value {
	_rwlock.Lock()
	defer _rwlock.Unlock()

	tmp := _tick
	_tick = make(map[string]nson.Value)

	return tmp
}
