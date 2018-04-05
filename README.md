# Funnel Queue

[![GoDoc](https://godoc.org/github.com/ssgreg/funnelqueue?status.svg)](https://godoc.org/github.com/ssgreg/funnelqueue)
[![Build Status](https://travis-ci.org/ssgreg/funnelqueue.svg?branch=master)](https://travis-ci.org/ssgreg/funnelqueue)
[![Go Report Status](https://goreportcard.com/badge/github.com/ssgreg/funnelqueue)](https://goreportcard.com/report/github.com/ssgreg/funnelqueue)

Package `funnelqueue` implements a FIFO, lock-free, concurrent, multi-producer/single-consumer, linked-list-based queue.

## Example

```go
package main

import (
    "fmt"
    "math/rand"
    "runtime"
    "sync"
    "github.com/ssgreg/funnelqueue"
)

func main() {
    n := 10
    wg := sync.WaitGroup{}
    wg.Add(n + 1)
    // make 100 random numbers (per 10 in each 10 goroutines)
    q := funnelqueue.New()
    for i := 0; i < n; i++ {
        go func(i int) {
            defer wg.Done()
            for j := 0; j < n; j++ {
                q.Push(rand.Int() % 64)
            }
        }(i)
    }
    // read these numbers concurrently
    go func() {
        runtime.Gosched()
        defer wg.Done()
        for {
            v := q.Pop()
            if v == nil {
                break
            }
            fmt.Print(v, " ")
        }
    }()
    wg.Wait()
}
```
