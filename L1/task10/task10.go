package main

import (
	"fmt"
	"math"
	"sort"
)

func main() {
	temps := []float64{-25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5}

	tempRanges := make(map[int][]float64)

	for _, temp := range temps {
		rng := int(math.Trunc(temp/10) * 10)
		tempRanges[rng] = append(tempRanges[rng], temp)
	}

	keys := make([]int, 0, len(tempRanges))
	for k, _ := range tempRanges {
		keys = append(keys, k)
	}

	sort.Ints(keys)
	for _, key := range keys {
		fmt.Print(key, ":{")
		for i, temp := range tempRanges[key] {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Printf("%.1f", temp)
		}
		fmt.Println("}")
	}
}
