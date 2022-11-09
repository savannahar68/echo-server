package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/savannahar68/echo-server/config"
)

func DumpAllAOF() {
	fp, err := os.OpenFile(config.AOFFile, os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	fmt.Println("rewriting AOF file at ", config.AOFFile)
	for k, obj := range store {
		dumpKey(fp, k, obj)
	}
	fmt.Println("rewriting AOF file done at file ", config.AOFFile)
}

func dumpKey(fp *os.File, key string, obj *Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	fp.Write(Encode(tokens, false))
}
