package core

import (
	"errors"
	"io"
)

func EvalAndRespond(cmd RedisCmd, c io.ReadWriter) error {
	switch cmd.Cmd {
	case "Ping":
		return evalPing(cmd.Args, c)
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
