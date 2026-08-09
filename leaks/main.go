package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func f(n int) int {
	time.Sleep(30 * time.Millisecond)
	return n * n
}

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	c := make(chan int)
	go func() { c <- f(7) }()
	select {
	case v := <-c:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}

	// Keep the program alive so the HTTP server (and the leaked goroutine)
	// stay around and can be inspected via /debug/pprof/goroutineleak.
	select {}
}
