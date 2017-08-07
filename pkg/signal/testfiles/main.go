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
		multiple := os.Getenv("IF_MULTIPLE")
		switch s {
		case "TERM", "INT":
			if multiple == "1" {
				for {
					p.Signal(sigmap[s])
				}
			} else {
				p.Signal(sigmap[s])
			}
		case "QUIT":
			p.Signal(sigmap[s])
		}
	}()
	time.Sleep(2 * time.Second)
}
