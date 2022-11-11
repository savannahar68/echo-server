package core

import (
	"time"

	"github.com/savannahar68/echo-server/config"
)

var store map[string]*Obj
var expires map[*Obj]uint64

func init() {
	store = make(map[string]*Obj)
	expires = make(map[*Obj]uint64)
}

func SetExpiry(obj *Obj, exDurationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli() + exDurationMs)
}

func NewObj(value interface{}, exDurationMs int64, oType uint8, oEnc uint8) *Obj {
	obj := &Obj{
		TypeEncoding:   oType | oEnc,
		Value:          value,
		LastAccessedAt: getCurrentClock(),
	}
	if exDurationMs > 0 {
		SetExpiry(obj, exDurationMs)
	}
	return obj
}

func getCurrentClock() uint32 {
	return uint32(time.Now().UnixMilli()) & 0xFFFFF
}

func Put(key string, obj *Obj) {
	if len(store) >= config.KeysLimit {
		Evict()
	}
	obj.LastAccessedAt = getCurrentClock()
	store[key] = obj
	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++
}

func Get(key string) *Obj {
	v := store[key]
	if v != nil {
		if HasExpired(v) {
			Del(key)
			return nil
		}
	}
	v.LastAccessedAt = getCurrentClock()
	return v
}

func Del(key string) bool {
	if obj, ok := store[key]; ok {
		delete(store, key)
		delete(expires, obj)
		KeyspaceStat[0]["keys"]--
		return true
	}
	return false
}
