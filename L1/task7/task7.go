package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	people := make(map[string]int)
	
	names := []string{
		"Tom", "Andy", "Ron", "Ann", "Jerry", "Larry", "Bran", "Leslie",
	}
	
	random := rand.New(rand.NewSource(time.Now().Unix()))
	
	var mu sync.Mutex
	numOfPeople := 20
	
	var wg sync.WaitGroup
	wg.Add(numOfPeople)
	
	for i := range numOfPeople {
		go func() {
			mu.Lock()
			defer mu.Unlock()
			defer wg.Done()
			people[names[random.Intn(len(names))]] = i
		}()
	}
	
	wg.Wait()
	fmt.Println(people)
}