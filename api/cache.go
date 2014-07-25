package api

import (
	"encoding/json"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/juju/errgo"
	"net/http"
	"strings"
)

const defaultExpiration = 300 // 5 minutes

type cache interface {
	set(string, interface{}) ([]byte, error)
	get(string, func() (interface{}, error)) (interface{}, error)
	clear() error
	render(http.ResponseWriter, int, string, func() (interface{}, error)) error
}

type memcacheCache struct {
	mc *memcache.Client
}

func (m *memcacheCache) set(key string, obj interface{}) ([]byte, error) {
	value, err := json.Marshal(obj)
	if err != nil {
		return value, err
	}
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: defaultExpiration,
	}
	if err := m.mc.Set(item); err != nil {
		return value, err
	}
	return value, nil
}

func (m *memcacheCache) get(key string, fn func() (interface{}, error)) (interface{}, error) {
	it, err := m.mc.Get(key)
	if err == nil {
		var obj interface{}
		if err := json.Unmarshal(it.Value, obj); err != nil {
			return obj, errgo.Mask(err)
		}
		return obj, nil
	} else if err != memcache.ErrCacheMiss {
		return nil, errgo.Mask(err)
	}
	obj, err := fn()
	if err != nil {
		return obj, err
	}
	if _, err := m.set(key, obj); err != nil {
		return obj, err
	}
	return obj, nil
}

func (m *memcacheCache) render(w http.ResponseWriter, status int, key string, fn func() (interface{}, error)) error {

	var write = func(value []byte) error {
		return writeBody(w, value, status, "application/json")
	}

	it, err := m.mc.Get(key)
	if err == nil {
		return write(it.Value)
	} else if err != memcache.ErrCacheMiss {
		return errgo.Mask(err)
	}
	obj, err := fn()
	if err != nil {
		return err
	}
	value, err := m.set(key, obj)
	if err != nil {
		return err
	}
	return write(value)

}

func (m *memcacheCache) clear() error {
	return errgo.Mask(m.mc.DeleteAll())
}

// NewCache creates a new Cache instance
func newCache(config *appConfig) cache {
	mc := memcache.New(strings.Split(config.MemcacheHost, ",")...) // will be from config
	return &memcacheCache{mc}
}
