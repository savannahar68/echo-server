package core

import (
	"fmt"
	"time"
)

func expireSample() float32 {
	var limit = 20
	var expiredCount = 0

	// range gets object in random order based on hash
	for key, obj := range store {
		if obj.ExpiresAt != -1 {
			limit--
			if obj.ExpiresAt <= time.Now().UnixMilli() {
				delete(store, key)
				expiredCount++
			}
		}
		//once we have limit = 20 we have some our expiration set
		if limit == 20 {
			break
		}
	}

	return float32(expiredCount) / float32(limit)
}

func DeleteExpiredKeys() {
	for {
		frac := expireSample()

		if frac < 0.25 {
			break
		}
	}
	fmt.Println("deleted expired but undeleted keys. total number of keys is ", len(store))
}
