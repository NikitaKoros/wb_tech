package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	SLICE_SIZE = 20
	MAX_NUM = 100
)

func main() {
	slc := make([]int, 0, SLICE_SIZE)
	
	random := rand.New(rand.NewSource(time.Now().Unix()))
	for _ = range SLICE_SIZE {
		slc = append(slc, random.Intn(MAX_NUM))
	}
	
	sorted := quickSort(slc)
	fmt.Println("Unsorted slice:", slc)
	fmt.Println("Sorted slice:", sorted)
}

func quickSort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}
	
	mid := len(arr) / 2
	pivot := arr[mid]
	
	left := make([]int, 0, mid)
	right := make([]int, 0, mid)
	for i, val := range arr {
		if i == mid {
			continue
		}
		if val <= pivot {
			left = append(left, val)
		} else {
			right = append(right, val)
		}
	}
	
	sortedLeft := quickSort(left)
	sortedRight := quickSort(right)
	
	combinedArr := make([]int, 0, len(arr))
	combinedArr = append(combinedArr, sortedLeft...)
	combinedArr = append(combinedArr, pivot)
	combinedArr = append(combinedArr, sortedRight...)
	
	return combinedArr
}