package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"

	llog "github.com/levenlabs/go-llog"
	okq "github.com/mediocregopher/okq-go.v2"
)

func srcOkq() []<-chan okq.Event {
	srcs, _ := ga.ParamStrs("--src-okq")
	chs := make([]<-chan okq.Event, len(srcs))

	for i := range srcs {
		kv := llog.KV{"src-okq": srcs[i]}
		parts := strings.Split(srcs[i], ",")
		if len(parts) < 2 {
			llog.Fatal("malformed --src-okq", llog.KV{"src-okq": srcs[i]})
		}

		addr := parts[0]
		queues := parts[1:]
		kv["addr"] = addr
		kv["queues"] = queues
		llog.Info("connecting to --src-okq", kv)

		cl, err := okq.New(addr, 1)
		if err != nil {
			llog.Fatal("error connecting to --src-okq", kv, llog.ErrKV(err))
		}

		ch := make(chan okq.Event)
		fn := func(e okq.Event) bool {
			ch <- e
			return true
		}
		go func() {
			errCh := cl.Consumer(fn, nil, queues...)
			llog.Fatal("error consuming", kv, llog.ErrKV(<-errCh))
		}()

		chs[i] = ch
	}

	return chs
}

func srcStdin() []<-chan okq.Event {
	if !ga.ParamFlag("--src-stdin") {
		return nil
	}

	ch := make(chan okq.Event)
	go func() {
		buf := bufio.NewReader(os.Stdin)
		for {
			line, err := buf.ReadString('\n')
			if err == io.EOF {
				close(ch)
				return
			} else if err != nil {
				llog.Fatal("error reading from stdin", llog.ErrKV(err))
			}

			var e okq.Event
			if err := json.Unmarshal([]byte(line), &e); err != nil {
				llog.Fatal("error unmarshalling json from stdin", llog.ErrKV(err))
			}
			ch <- e
		}
	}()

	return []<-chan okq.Event{ch}
}
