package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	llog "github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/genapi"
	"github.com/mediocregopher/lever"
	okq "github.com/mediocregopher/okq-go.v2"
)

var ga = genapi.GenAPI{
	Name: "okq-cat",
	LeverParams: []lever.Param{
		{
			Name:        "--src-okq",
			Description: `Source events from one or more queues in an instance. E.g. "127.0.0.1:4777,queue_0,queue_1"`,
		},
		{
			Name:        "--src-stdin",
			Description: "Source events from stdin. The stream should have originally come from a previous call with --dst-stdout",
			Flag:        true,
		},
		{
			Name:        "--dst-okq",
			Description: `Write events to one or more queues in an instance. If more than one queue is given events will be round-robin'd across them. E.g. "127.0.0.1:4777,queue_0,queue_1"`,
		},
		{
			Name:        "--dst-stdout",
			Description: "JSON encode events and write them, newline delimited, to stdout",
			Flag:        true,
		},
	},
}

type dst struct {
	ch   chan okq.Event
	done chan struct{}
}

func newDst() dst {
	return dst{
		ch:   make(chan okq.Event),
		done: make(chan struct{}),
	}
}

func main() {
	llog.Out = os.Stderr
	ga.CLIMode()

	var srcs []<-chan okq.Event
	srcs = append(srcs, srcOkq()...)
	srcs = append(srcs, srcStdin()...)
	if len(srcs) == 0 {
		llog.Fatal("no srcs specified")
	}

	var dsts []dst
	dsts = append(dsts, dstOkq()...)
	dsts = append(dsts, dstStdout()...)
	if len(dsts) == 0 {
		llog.Fatal("no dsts specified")
	}

	stopCh := make(chan struct{})

	var mainWG sync.WaitGroup
	mainCh := make(chan okq.Event)
	for i := range srcs {
		mainWG.Add(1)
		go func(i int) {
			defer mainWG.Done()
			for {
				e, ok := <-srcs[i]
				if !ok {
					return
				}
				mainCh <- e
			}
		}(i)
	}

	go func() {
		mainWG.Wait()
		close(stopCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		close(stopCh)
	}()

	countCh := make(chan bool)
	go func() {
		var c uint64
		tick := time.Tick(1 * time.Minute)
		for {
			select {
			case <-countCh:
				c++
			case <-tick:
				llog.Info("events processed last minute", llog.KV{"count": c})
				c = 0
			}
		}
	}()

	for {
		select {
		case e := <-mainCh:
			for i := range dsts {
				dsts[i].ch <- e
			}
			countCh <- true
		case <-stopCh:
			llog.Info("stopping")
			for i := range dsts {
				close(dsts[i].ch)
				<-dsts[i].done
			}
			llog.Flush()
			os.Exit(0)
		}
	}
}
