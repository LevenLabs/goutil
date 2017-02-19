package main

import (
	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/genapi"
)

var ga = genapi.GenAPI{
	Name:    "okq-llog",
	OkqInfo: &genapi.OkqInfo{},
}

func main() {
	ga.CLIMode()

	ss, err := ga.OkqInfo.Status()
	if err != nil {
		llog.Fatal("could not get queue statuses", llog.KV{"err": err})
	}

	for _, s := range ss {
		llog.Info("queue status", llog.KV{
			"name":       s.Name,
			"total":      s.Total,
			"processing": s.Processing,
			"consumers":  s.Consumers,
		})
	}
	llog.Flush()
}
