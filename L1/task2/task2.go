package task2

import (
	"fmt"
	"sync"
)

type result struct {
	num int
	sqr int
}

func main() {
	numbers := []int{2, 4, 6, 8, 10}

	ch := make(chan result)
	var wg sync.WaitGroup
	
	for _, n := range numbers {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			ch <- result{num: x, sqr: x*x}
		}(n)
	}
	
	go func() {
		wg.Wait()
		close(ch)
	}()
	
	for r := range ch {
		fmt.Printf("%d^2 = %d\n", r.num, r.sqr)
	}
}
