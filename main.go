package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
	"log"
)

var (
	routines = flag.Int("conn", 10, "Number of concurrent connections")
	host     = flag.String("host", "localhost", "Host to connect to")
	port     = flag.Int("port", 1965, "Port number")
	path     = flag.String("path", "", "Path to require")
	timeout  = flag.Int("timeout", 10000, "Request timeout in ms")
	duration = flag.Int("duration", 10, "Duration of the tests in seconds")
)

type Stats struct {
	Errors   int
	Requests int
	Duration time.Duration
}

func (s *Stats) Add(t Stats) {
        s.Errors += t.Errors
	s.Requests += t.Requests
	s.Duration += t.Duration
}

func requester(stats chan<- Stats, quit <-chan interface{}) {
	errors := 0
	requests := 0
	duration := time.Second * 0

	hst := fmt.Sprintf("%s:%d", *host, *port)

	req := fmt.Sprintf("gemini://%s:%d/%s\r\n", *host, *port, *path)

outer:
	for {
		select {
		case <-quit:
			break outer
		default:
			// continue
		}

		start := time.Now()
		_, err := request(hst, req)
		elapsed := time.Since(start)

		requests++
		duration += elapsed

		if err != nil {
			errors++
			log.Println(err)
		}
	}

	stats <- Stats{errors, requests, duration}
}

func main() {
	flag.Parse()

	stats := make(chan Stats, *routines)
	quit := make(chan interface{}, 1)

	signs := make(chan os.Signal, 1)
	signal.Notify(signs, os.Interrupt)

	grp := sync.WaitGroup{}

	for i := 0; i < *routines; i++ {
		grp.Add(1)
		go func() {
			requester(stats, quit)
			grp.Done()
		}()
	}

	go func() {
		select {
		case <-signs:
			log.Println("catched SIGINT")
		case <-time.After(time.Duration(*timeout) * time.Millisecond):
			log.Println("time expired")
		}
		close(quit)
	}()

        grp.Wait()

	cumulative := Stats{}
	for i := 0; i < *routines; i++ {
		cumulative.Add(<-stats)
	}

	fmt.Println("Total requests:", cumulative.Requests)
	fmt.Println("Total errors:", cumulative.Errors)
	fmt.Println("Total time:", cumulative.Duration)

	fmt.Println("Request mean time:", cumulative.Duration / time.Duration(cumulative.Requests))
}
