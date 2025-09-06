package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Time before sleep:", time.Now())
	MySleep(5*time.Second)
	fmt.Println("Time after sleep:", time.Now())
}

func MySleep(duration time.Duration) {
	if duration == 0 {
		return
	}
	
	now := time.Now()
	doneCh := make(chan struct{})
	go func() {
		for time.Since(now) < duration {}
		close(doneCh)
	}()

	<-doneCh
}
