package core

import "time"

type Obj struct {
	Value     interface{}
	ExpiresAt int64
}

var store map[string]*Obj

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, expires int64) *Obj {
	var expiresAt int64 = -1
	if expires > 0 {
		expiresAt = time.Now().UnixMilli() + expires
	}
	return &Obj{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

func Put(key string, obj *Obj) {
	store[key] = obj
}

func Get(key string) *Obj {
	return store[key]
}
