package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	LENGTH = 15
	MAX_NUM = 50
)

func main() {
	numbers := make([]int, 0, LENGTH)
	random := rand.New(rand.NewSource(time.Now().Unix()))
	
	for _ = range LENGTH {
		numbers = append(numbers, random.Intn(MAX_NUM))
	}
	
	idx := random.Intn(LENGTH)
	
	fmt.Println("Slice before:", numbers)
	fmt.Printf("Deleting number %d on index %d\n", numbers[idx], idx)
	
	copy(numbers[idx:], numbers[idx+1:])
	numbers = numbers[:len(numbers) - 1]
	
	fmt.Println("Slice after deletion:", numbers)
}