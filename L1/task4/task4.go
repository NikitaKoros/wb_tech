package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const NUM_WORKERS = 5

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)

	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(NUM_WORKERS)
	
	for i := range NUM_WORKERS {
		go worker(i, stopCh, &wg)
	}
	
	fmt.Println("Press Ctrl+C to stop program")
	<-sigCh
	
	close(stopCh)
	wg.Wait()
	
	fmt.Println("\nAll workers have been shut down")
}

func worker(id int, stop chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-stop:
			fmt.Printf("\nWorker %d has stopped", id)
			return
		default:
			fmt.Printf("Worker %d is doing something...\n", id)
			time.Sleep(time.Millisecond * 100)
		}
	}
}
