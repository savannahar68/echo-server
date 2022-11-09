package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_OK []byte = []byte("+OK\r\n")
var RESP_QUEUED []byte = []byte("+QUEUED\r\n")
var RESP_ZERO []byte = []byte(":0\r\n")
var RESP_ONE []byte = []byte(":1\r\n")
var RESP_MINUS_1 []byte = []byte(":-1\r\n")
var RESP_MINUS_2 []byte = []byte(":-2\r\n")

func EvalAndRespond(cmds RedisCmds, c io.ReadWriter) {
	var response []byte
	buf := bytes.NewBuffer(response)
	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "Ping":
			buf.Write(evalPing(cmd.Args))
		case "GET":
			buf.Write(evalGET(cmd.Args))
		case "SET":
			buf.Write(evalSET(cmd.Args))
		case "TTL":
			buf.Write(evalTTL(cmd.Args))
		case "DEL":
			buf.Write(evalDEL(cmd.Args))
		case "EXPIRE":
			buf.Write(evalEXPIRE(cmd.Args))
		case "BGREWRITEAOF":
			buf.Write(evalBGREWRITEAOF(cmd.Args))
		case "INCR":
			buf.Write(evalINCR(cmd.Args))
		default:
			buf.Write(evalPing(cmd.Args))
		}
	}
	c.Write(buf.Bytes())
}

func evalPing(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return Encode(errors.New("ERR wrong number of arguments for PING command"), false)
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	return b
}

func evalGET(args []string) []byte {
	if len(args) == 0 {
		return Encode(errors.New("(error) wrong number of arguments for GET command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return []byte("+nil\r\n")
	}

	if obj.ExpiresAt != -1 && time.Now().UnixMilli() > obj.ExpiresAt {
		delete(store, args[0])
		return Encode(errors.New("(error) Key Expired"), false)
	}

	return Encode(obj.Value, false)
}

func evalSET(args []string) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("(error) wrong number of arguments for SET command"), false)
	}

	var key, value string
	// Don't expire
	var exDurationMs int64 = -1
	key, value = args[0], args[1]
	oType, oEnc := DeduceTypeEncoding(value)

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return Encode(errors.New("(error) Invalid syntax error"), false)
			}

			exDurationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Encode(errors.New("ERR value is not an integer or out of range"), false)
			}
			exDurationMs = exDurationSec * 1000
		default:
			return Encode(errors.New("(error) Invalid syntax"), false)
		}
	}
	// putting key and value in hash table
	Put(key, NewObj(value, exDurationMs, oType, oEnc))
	return []byte("+OK\r\n")
}

func evalTTL(args []string) []byte {
	if len(args) == 0 {
		return Encode(errors.New("(error) wrong number of arguments for TTL command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return Encode(errors.New("(error) Key doesn't exist"), false)
	}

	if obj.ExpiresAt == -1 {
		return []byte(":-1\r\n")
	}

	ttlDifference := obj.ExpiresAt - time.Now().UnixMilli()

	if ttlDifference < 0 {
		return []byte(":-2\r\n")
	}

	return []byte(fmt.Sprintf(":%d\r\n", int(math.Round(float64(ttlDifference/1000)))))
}

func evalDEL(args []string) []byte {
	if len(args) == 0 {
		return Encode(errors.New("(error) wrong number of arguments for DEL command"), false)
	}

	var countDeleted int = 0

	for _, key := range args {
		if ok := Del(key); ok {
			countDeleted++
		}
	}
	return Encode(countDeleted, false)
}

func evalEXPIRE(args []string) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("(error) wrong number of arguments for EXPIRE command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return []byte(":0\r\n")
	}

	if obj.ExpiresAt != -1 && obj.ExpiresAt < 0 {
		return Encode(0, false)
	}

	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
	}

	obj.ExpiresAt = time.Now().UnixMilli() + exDurationSec*1000
	return Encode(1, false)
}

// TODO: make it asnyc by forking process
func evalBGREWRITEAOF(args []string) []byte {
	DumpAllAOF()
	return RESP_OK
}

func evalINCR(args []string) []byte {
	if len(args) == 0 {
		return Encode(errors.New("(error) wrong number of arguments for INCR command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		obj = NewObj("0", -1, OBJ_TYPE_STRING, OBJ_ENCODING_INT)
		Put(args[0], obj)
	}

	if err := AssertType(obj.TypeEncoding, OBJ_TYPE_STRING); err != nil {
		return Encode(err, false)
	}

	if err := AssertEncoding(obj.TypeEncoding, OBJ_ENCODING_INT); err != nil {
		return Encode(err, false)
	}

	i, _ := strconv.ParseInt(obj.Value.(string), 10, 64)
	i++
	obj.Value = strconv.FormatInt(i, 10)

	return Encode(i, false)
}
