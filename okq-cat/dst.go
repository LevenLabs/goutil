package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	llog "github.com/levenlabs/go-llog"
	okq "github.com/mediocregopher/okq-go.v2"
)

func dstOkq() []dst {
	dsts, _ := ga.ParamStrs("--dst-okq")
	chs := make([]dst, len(dsts))

	for i := range dsts {
		kv := llog.KV{"dst-okq": dsts[i]}
		parts := strings.Split(dsts[i], ",")
		if len(parts) < 2 {
			llog.Fatal("malformed --dst-okq", llog.KV{"dst-okq": dsts[i]})
		}

		addr := parts[0]
		queues := parts[1:]
		kv["addr"] = addr
		kv["queues"] = queues
		llog.Info("connecting to --dst-okq", kv)

		cl, err := okq.New(addr, 1)
		if err != nil {
			llog.Fatal("error connecting to --dst-okq", kv, llog.ErrKV(err))
		}

		nextQueue := make(chan string)
		go func() {
			for {
				for _, q := range queues {
					nextQueue <- q
				}
			}
		}()

		d := newDst()
		go func() {
			for {
				e, ok := <-d.ch
				if !ok {
					cl.Close()
					close(d.done)
					return
				}
				e.Queue = <-nextQueue
				if err := cl.PushEvent(e, okq.Normal); err != nil {
					llog.Fatal("error writing to okq", kv, llog.ErrKV(err))
				}
			}
		}()

		chs[i] = d
	}
	return chs
}

func dstStdout() []dst {
	if !ga.ParamFlag("--dst-stdout") {
		return nil
	}

	d := newDst()
	go func() {
		for {
			e, ok := <-d.ch
			if !ok {
				os.Stdout.Sync()
				close(d.done)
				return
			}
			b, err := json.Marshal(e)
			if err != nil {
				llog.Fatal("could not json marshal", llog.ErrKV(err))
			}
			fmt.Println(string(b))
		}
	}()

	return []dst{d}
}
