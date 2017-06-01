package main

import (
	"os"
	"syscall"
	"time"

	"github.com/docker/docker/pkg/signal"
)

func main() {
	sigmap := map[string]os.Signal{
		"TERM": syscall.SIGTERM,
		"QUIT": syscall.SIGQUIT,
		"INT":  os.Interrupt,
	}

	defer time.Sleep(10 * time.Millisecond)

	signal.Trap(func() {
		time.Sleep(10 * time.Millisecond)
		os.Exit(99)
	})

	go func() {
		p, err := os.FindProcess(os.Getpid())

		if err != nil {
			panic(err)
		}
		s := os.Getenv("SIGNAL_TYPE")
		switch s {
		case "TERM":
			for {
				p.Signal(sigmap[s])
			}
		case "QUIT":
			p.Signal(sigmap[s])
		case "INT":
			p.Signal(sigmap[s])
		}
	}()
	select {}
}
