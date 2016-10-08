package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/levenlabs/go-llog"
	"github.com/mediocregopher/lever"
)

func main() {
	l := lever.New("stdin-llog", &lever.Opts{
		HelpHeader:         "Usage: stdin-llog [options] <message>\n",
		DisallowConfigFile: true,
	})
	l.Add(lever.Param{
		Name:        "--key",
		Description: "The key that contains the value(s) from stdin",
		Default:     "value",
	})
	l.Parse()

	argv := l.ParamRest()
	if len(argv) != 1 {
		fmt.Print(l.Help())
		os.Exit(0)
	}

	message := argv[0]
	key, _ := l.ParamStr("--key")
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		v := s.Text()
		if v != "" {
			kv := llog.KV{}
			kv[key] = v
			llog.Info(message, kv)
		}
	}
	if err := s.Err(); err != nil {
		llog.Error("error reading stdin", llog.ErrKV(err))
	}

	llog.Flush()
}
