package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run task3.go <number_of_workers>")
		os.Exit(1)
	}
	
	numWorkers, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Error: number of workers must be non-negative")
		os.Exit(1)
	}
	
	ch := make(chan int)
	
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := range numWorkers {
		go worker(i, ch, &wg)
	}
	
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for start := time.Now(); time.Since(start) < time.Second; {
		ch <- random.Intn(1000)
		time.Sleep(time.Millisecond * 100)
	}
	
	close(ch)
	wg.Wait()
}

func worker(id int, channel <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for message := range channel {
		fmt.Printf("Worker %d: %d\n", id, message)
	}
}