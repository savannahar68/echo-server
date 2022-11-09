package core

type Obj struct {
	// First 4 bits will have type and last 4 bits will have encoding
	TypeEncoding uint8
	Value        interface{}
	// TODO: change data type
	ExpiresAt int64
}

// Type is stored in first 4 bits so left shift by 4
var OBJ_TYPE_STRING uint8 = 0 << 4

var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8
