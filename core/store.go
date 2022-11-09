package core

import (
	"time"

	"github.com/savannahar68/echo-server/config"
)

var store map[string]*Obj

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, expires int64, oType uint8, oEnc uint8) *Obj {
	var expiresAt int64 = -1
	if expires > 0 {
		expiresAt = time.Now().UnixMilli() + expires
	}
	return &Obj{
		TypeEncoding: oType | oEnc,
		Value:        value,
		ExpiresAt:    expiresAt,
	}
}

func Put(key string, obj *Obj) {
	if len(store) >= config.KeysLimit {
		Evict()
	}
	store[key] = obj
}

func Get(key string) *Obj {
	return store[key]
}

func Del(key string) bool {
	if _, ok := store[key]; ok {
		delete(store, key)
		return true
	}
	return false
}
