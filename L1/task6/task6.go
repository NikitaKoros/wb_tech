package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go [condition|done|context-cancel|context-timeout|")
		fmt.Println("		channel-close|goexit|return|osexit|panic-recover|timer]")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "condition":
		withCondition()
	case "done":
		withDoneChannel()
	case "context-cancel":
		withContextCancel()
	case "context-timeout":
		withContextTimeout()
	case "channel-close":
		withChannelClose()
	case "goexit":
		withGoExit()
	case "return":
		withReturn()
	case "osexit":
		withOsExit()
	case "panic-recover":
		withPanicRecover()
	case "timer":
		withTimer()
	default:
		fmt.Println("Invalid working mode")
	}
}

func withCondition() {
	var stopAfter = 5
	for i := 0; i < stopAfter; i++ {
		fmt.Printf("Work: %d\n", i)
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("Stopped by condition")
}

func withDoneChannel() {
	doneCh := make(chan struct{})
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-doneCh:
				fmt.Println("Stopped by done channel")
				return
			case <-ticker.C:
				fmt.Println("Working...")
			}
		}
	}()

	time.Sleep(time.Second * 5)
	close(doneCh)
	wg.Wait()
}

func withContextCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Stopped by context cancel")
				return
			case <-ticker.C:
				fmt.Println("Working...")
			}
		}
	}(ctx)

	time.Sleep(time.Second * 5)
	cancel()

	wg.Wait()
}

func withContextTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Stopped by context cancel")
				return
			default:
				fmt.Println("Working...")
				time.Sleep(time.Second)
			}
		}
	}(ctx)

	wg.Wait()
}

func withChannelClose() {
	ch := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for val := range ch {
			fmt.Printf("Recieved value: %d\n", val)
		}
		fmt.Println("Stopped by channel close")
	}()

	ch <- 50
	ch <- 60
	close(ch)

	wg.Wait()
}

func withGoExit() {
	var wg sync.WaitGroup
	wg.Add(1)
	
	go func() {
		defer wg.Done()
		fmt.Println("Working...")
		time.Sleep(time.Second)
		defer fmt.Println("Just before exit")
		runtime.Goexit()
	}()
	
	wg.Wait()
}

func withReturn() {
	fmt.Println("Working...")
	time.Sleep(time.Second)
	fmt.Println("Stopped by return")
	// return
}

func withOsExit() {
	fmt.Println("Working...")
	time.Sleep(time.Second)
	defer fmt.Println("Won't print this")
	os.Exit(0)
}

func withPanicRecover() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered, worker stopped:", r)
		}
	}()
	for now := time.Now(); time.Since(now) < 3 * time.Second; {
		fmt.Println("Working...")
		time.Sleep(time.Second)
	}
	panic("force stop")
}

func withTimer() {
	timer := time.After(3 * time.Second)
	
	var wg sync.WaitGroup
	wg.Add(1)
	
	go func() {
		defer wg.Done()
		for {
			select {
			case <-timer:
				fmt.Println("Stopped by timer")
				return
			default:
				fmt.Println("Working...")
				time.Sleep(time.Second)
			}
		}
	}()
	
	wg.Wait()
}
