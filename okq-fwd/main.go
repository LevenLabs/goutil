package main

import (
	"sort"
	"sync"
	"time"

	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/genapi"
	"github.com/levenlabs/golib/radixutil"
	"github.com/mediocregopher/lever"
	"github.com/mediocregopher/okq-go.v2"
	"github.com/mediocregopher/radix.v2/pool"
)

var ga = genapi.GenAPI{
	Name: "okq-fwd",
	LeverParams: []lever.Param{
		{
			Name:        "--src-okq-addr",
			Description: "Address of okq instance to pull jobs from",
			Default:     "127.0.0.1:4777",
		},
		{
			Name:        "--src-okq-pool-size",
			Description: "Number of connections to make to --src-okq-addr",
			Default:     "10",
		},
		{
			Name:        "--dst-okq-addr",
			Description: "(Required) Address of okq instance to forward jobs to",
		},
		{
			Name:        "--dst-okq-pool-size",
			Description: "Number of connections to make to --dst-okq-addr",
			Default:     "10",
		},
	},
}

const notifyTimeout = 30 * time.Second

// copied pretty much directly from genapi. Once genapi.v2 (or whatever it ends
// up being called) is usable we won't have to do this
func okqClient(addr string, size int) *okq.Client {
	kv := llog.KV{
		"addr":     addr,
		"poolSize": size,
	}

	df := radixutil.SRVDialFunc(ga.SRVClient, notifyTimeout)

	llog.Info("connecting to okq", kv)
	p, err := pool.NewCustom("tcp", addr, size, df)
	if err != nil {
		llog.Fatal("error connecting to okq", kv, llog.ErrKV(err))
	}

	return okq.NewWithOpts(okq.Opts{
		RedisPool:     p,
		NotifyTimeout: notifyTimeout,
	})
}

var src, dst *okq.Client
var queueNames []string
var queueNamesL sync.RWMutex

func getQueueNames() ([]string, error) {
	qq, err := src.Status()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(qq))
	for i, q := range qq {
		names[i] = q.Name
	}
	sort.Strings(names)
	return names, nil
}

// Returns true if the given list of names is different than the active one
func queueNamesChanged(names []string) bool {
	queueNamesL.RLock()
	curNames := queueNames
	queueNamesL.RUnlock()

	if len(curNames) != len(names) {
		return true
	}
	for i := range curNames {
		if curNames[i] != names[i] {
			return true
		}
	}
	return false
}

func main() {
	ga.CLIMode()
	srcAddr, _ := ga.ParamStr("--src-okq-addr")
	srcPoolSize, _ := ga.ParamInt("--src-okq-pool-size")
	dstAddr, _ := ga.ParamStr("--dst-okq-addr")
	dstPoolSize, _ := ga.ParamInt("--dst-okq-pool-size")

	if dstAddr == "" {
		llog.Fatal("--dst-okq-addr is required")
	}

	src, dst = okqClient(srcAddr, srcPoolSize), okqClient(dstAddr, dstPoolSize)

	// Do an initial call to populate queue names before anything else needs it,
	// then set up a spinner to continuously update it after
	var err error
	if queueNames, err = getQueueNames(); err != nil {
		llog.Fatal("could not get queue names", llog.ErrKV(err))
	}
	go func() {
		tick := time.Tick(5 * time.Second)
		for {
			if names, err := getQueueNames(); err != nil {
				llog.Warn("could not get queue names", llog.ErrKV(err))
			} else {
				queueNamesL.Lock()
				queueNames = names
				queueNamesL.Unlock()
			}
			<-tick
		}
	}()

	for i := 0; i < srcPoolSize; i++ {
		go forwarder(i)
	}

	select {}
}

func forwarder(i int) {
	kv := llog.KV{"i": i}
	llog.Info("starting forwarder", kv)
	fn := func(e okq.Event) bool {
		if err := dst.PushEvent(e, okq.Normal); err != nil {
			ekv := llog.KV{"id": e.ID, "queue": e.Queue}
			llog.Warn("error forwarding event", kv, ekv, llog.ErrKV(err))
			return false
		}
		return true
	}

	tick := time.Tick(5 * time.Second)
outer:
	for {
		queueNamesL.RLock()
		names := queueNames
		queueNamesL.RUnlock()
		if len(names) == 0 {
			<-tick
			continue
		}

		stopCh := make(chan bool)
		errCh := src.Consumer(fn, stopCh, names...)

		for {
			select {
			case <-tick:
				if queueNamesChanged(names) {
					close(stopCh)
					<-errCh
					continue outer
				}
			case err := <-errCh:
				llog.Warn("consumer error", kv, llog.ErrKV(err))
				continue outer
			}
		}
	}
}
