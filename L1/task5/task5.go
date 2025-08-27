package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const DURATION_IN_SECONDS = 5

func main() {
	messageCh := make(chan int)

	timer := time.After(time.Second * DURATION_IN_SECONDS)
	var wg sync.WaitGroup
	wg.Add(2)

	go produce(timer, messageCh, &wg)
	go consume(messageCh, &wg)

	wg.Wait()
	fmt.Println("Program finished")
}

func produce(timer <-chan time.Time, ch chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	for {
		select {
		case <-timer:
			fmt.Println("Time limit reached, producing stopped")
			close(ch)
			return
		default:
			ch <- r.Intn(100)
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func consume(ch <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for val := range ch {
		fmt.Printf("Recieved value: %d\n", val)
	}

	fmt.Println("Consuming stopped")
}
