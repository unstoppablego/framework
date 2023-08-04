package cache

import (
	"context"
	"reflect"
	"sync"
	"time"

	cache2 "github.com/Code-Hex/go-generics-cache"
)

var cacheMap = make(map[string]any)
var lockrw sync.RWMutex

func getMapKey[K comparable, V any]() string {
	var k K
	var v V
	// logs.Info(reflect.TypeOf(k).String())
	// logs.Info(reflect.TypeOf(v).String())
	mapKey := reflect.TypeOf(k).String() + reflect.TypeOf(v).String()
	return mapKey
}

func getInstance[K comparable, V any]() *cache2.Cache[K, V] {
	mapKey := getMapKey[K, V]()

	// use simple cache algorithm without options.
	lockrw.Lock()
	defer lockrw.Unlock()
	if cc, ok := cacheMap[mapKey]; !ok {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c := cache2.NewContext[K, V](ctx)
		cacheMap[mapKey] = c
		return c
	} else {
		rc, _ := cc.(*cache2.Cache[K, V])
		return rc
	}
}

func Contains[K comparable, V any](key K) bool {
	return getInstance[K, V]().Contains(key)
}

func Delete[K comparable, V any](key K) {
	getInstance[K, V]().Delete(key)

}
func DeleteExpired[K comparable, V any]() {
	getInstance[K, V]().DeleteExpired()
}
func Get[K comparable, V any](key K) (value V, ok bool) {
	return getInstance[K, V]().Get(key)
}
func Keys[K comparable, V any]() []K {
	return getInstance[K, V]().Keys()
}
func Set[K comparable, V any](key K, val V, opts ...cache2.ItemOption) {
	getInstance[K, V]().Set(key, val, opts...)
}

func WithExpiration(exp time.Duration) cache2.ItemOption {
	return cache2.WithExpiration(exp)
}
