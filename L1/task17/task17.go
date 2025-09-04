package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	SLICE_SIZE = 20
	MAX_NUM    = 100
)

func main() {
	slice := make([]int, 0, SLICE_SIZE)

	random := rand.New(rand.NewSource(time.Now().Unix()))
	for _ = range SLICE_SIZE {
		slice = append(slice, random.Intn(MAX_NUM))
	}

	randomElem := slice[random.Intn(SLICE_SIZE)]
	sorted := quickSort(slice)
	
	fmt.Println("Sorted slice:", sorted)
	fmt.Println("Random element:", randomElem)
	idx := binarySearch(sorted, randomElem)
	fmt.Println("Found index in slice:", idx)
	fmt.Println("Found element:", sorted[idx])
}

func binarySearch(numbers []int, val int) int {
	low := 0
	high := len(numbers)
	for low <= high {
		mid := (high + low) / 2
		if numbers[mid] == val {
			return mid
		} else if numbers[mid] < val {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1
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
