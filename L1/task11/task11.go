package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	MAX_NUM = 100
	SET_A_SIZE = 20
	SET_B_SIZE = 15
)

func main() {
	setA := getSet(SET_A_SIZE, MAX_NUM)
	setB := getSet(SET_B_SIZE, MAX_NUM)

	allElems := make(map[int]int)
	for _, val := range setA {
		allElems[val]++
	}

	for _, val := range setB {
		allElems[val]++
	}

	sharedElems := make([]int, 0)
	for key, val := range allElems {
		if val == 2 {
			sharedElems = append(sharedElems, key)
		}
	}

	fmt.Println("Set A: ", setA)
	fmt.Println("Set B: ", setB)
	fmt.Println("Intersection of sets: ", sharedElems)
}

func getSet(count, max int) []int {
	random := rand.New(rand.NewSource(time.Now().Unix()))

	set := make(map[int]struct{})
	for len(set) < count {
		val := random.Intn(max)
		set[val] = struct{}{}
	}

	keys := make([]int, 0, count)
	for key, _ := range set {
		keys = append(keys, key)
	}

	return keys
}
