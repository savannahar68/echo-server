package core

// EvictFirst SimpleFirst whenever cache is full evict first key
func EvictFirst() {
	for k, _ := range store {
		delete(store, k)
		return
	}
}

func Evict() {
	EvictFirst()
}
