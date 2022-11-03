package core

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"
)

func EvalAndRespond(cmd RedisCmd, c io.ReadWriter) error {
	switch cmd.Cmd {
	case "Ping":
		return evalPing(cmd.Args, c)
	case "GET":
		return evalGET(cmd.Args, c)
	case "SET":
		return evalSET(cmd.Args, c)
	case "TTL":
		return evalTTL(cmd.Args, c)
	default:
		return evalPing(cmd.Args, c)
	}
}

func evalPing(args []string, c io.ReadWriter) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for PING command")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err
}

func evalGET(args []string, c io.ReadWriter) error {
	if len(args) == 0 {
		return errors.New("(error) wrong number of arguments for GET command")
	}

	obj := Get(args[0])
	if obj == nil {
		_, err := c.Write([]byte("+nil\r\n"))
		return err
	}

	if obj.ExpiresAt != -1 && time.Now().UnixMilli() > obj.ExpiresAt {
		return errors.New("(error) Key Expired")
	}

	_, err := c.Write(Encode(obj.Value, false))
	return err
}

func evalSET(args []string, c io.ReadWriter) error {
	if len(args) <= 1 {
		return errors.New("(error) wrong number of arguments for SET command")
	}

	var key, value string
	// Don't expire
	var exDurationMs int64 = -1
	key, value = args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return errors.New("(error) Invalid syntax error")
			}

			exDurationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return errors.New("ERR value is not an integer or out of range")
			}
			exDurationMs = exDurationSec * 1000
		default:
			return errors.New("(error) Invalid syntax")
		}
	}
	// putting key and value in hash table
	Put(key, NewObj(value, exDurationMs))
	_, err := c.Write([]byte("+OK\r\n"))
	return err
}

func evalTTL(args []string, c io.ReadWriter) error {
	if len(args) == 0 {
		return errors.New("(error) wrong number of arguments for TTL command")
	}

	obj := Get(args[0])
	if obj == nil {
		return errors.New("(error) Key doesn't exist")
	}

	if obj.ExpiresAt == -1 {
		_, err := c.Write([]byte(":-1\r\n"))
		return err
	}

	ttlDifference := obj.ExpiresAt - time.Now().UnixMilli()

	if ttlDifference < 0 {
		_, err := c.Write([]byte(":-2\r\n"))
		return err
	}

	_, err := c.Write([]byte(fmt.Sprintf(":%d\r\n", int(math.Round(float64(ttlDifference/1000))))))
	return err
}
