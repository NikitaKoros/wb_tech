package main

import (
	"fmt"
	"sync"
)

const GOROUTINES_AMOUNT = 50

type Counter struct {
	count int
	mu sync.Mutex
}

func (c *Counter) Add() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func NewCounter() *Counter {
	return &Counter{
		count: 0,
		mu: sync.Mutex{},
	}
}

func main() {
	counter := NewCounter()
	
	var wg sync.WaitGroup
	wg.Add(GOROUTINES_AMOUNT)
	
	for _ = range GOROUTINES_AMOUNT {
		go func() {
			defer wg.Done()
			counter.Add()
		}()
	}
	
	wg.Wait()
	fmt.Println("Resulting counter value:", counter.count)
}