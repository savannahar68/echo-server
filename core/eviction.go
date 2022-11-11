package core

import (
	"github.com/savannahar68/echo-server/config"
)

// EvictFirst SimpleFirst whenever cache is full evict first key
func EvictFirst() {
	for k, _ := range store {
		delete(store, k)
		return
	}
}

func Evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		EvictFirst()
	case "allkeys-random":
		EvictAllRandomKeys()
	case "allkeys-lru":
		EvicAllKeysLru()
	}
}

func EvictAllRandomKeys() {
	evictCount := int64(config.EvictionRatio * float64(config.KeysLimit))
	for k, _ := range store {
		Del(k)
		evictCount--
		if evictCount <= 0 {
			break
		}
	}
}

func getIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	if c > lastAccessedAt {
		return c - lastAccessedAt
	}
	return (0x00FFFFF - lastAccessedAt) + c
}

func populateEvictionPool() {
	sampleSize := 5
	for k := range store {
		ePool.Push(k, store[k].LastAccessedAt)
		sampleSize--
		if sampleSize == 0 {
			break
		}
	}
}

func EvicAllKeysLru() {
	populateEvictionPool()
	evictCount := int16(config.EvictionRatio * float64(config.KeysLimit))
	for i := 0; i < int(evictCount) && len(ePool.pool) > 0; i++ {
		item := ePool.Pop()
		if item == nil {
			return
		}
		Del(item.key)
	}
}
