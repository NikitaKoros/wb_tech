package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	NUMS_AMOUNT = 50
	MAX_NUM     = 100
)

func main() {
	numbers := make([]int, NUMS_AMOUNT)

	random := rand.New(rand.NewSource(time.Now().Unix()))
	for i := range NUMS_AMOUNT {
		numbers[i] = random.Intn(MAX_NUM)
	}
	
	var wg sync.WaitGroup
	wg.Add(2)
	
	inputCh := make(chan int)
	outputCh := make(chan int)
	
	go func() {
		defer wg.Done()
		for val := range inputCh {
			outputCh <- 2 * val
		}
		close(outputCh)
	}()
	
	go func() {
		defer wg.Done()
		for val := range outputCh {
			fmt.Println(val)
		}
	}()
	
	for _, num := range numbers {
		inputCh <- num
	}
	
	close(inputCh)
	
	wg.Wait()
}
