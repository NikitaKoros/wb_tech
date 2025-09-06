package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	MIN_LENGTH = 5
	MAX_LENGTH = 15
)
var random = rand.New(rand.NewSource(time.Now().Unix()))

func main() {
	strLength := random.Intn(MAX_LENGTH - MIN_LENGTH) + MIN_LENGTH
	randStr := randomString(strLength)
	
	fmt.Println("Generated string:", randStr)
	allUnique := checkUniqueLetters(randStr)
	fmt.Println("String has all unique letters:",  allUnique)
}

func checkUniqueLetters(str string) bool {
	bytes := []byte(str)
	uniqueLetters := make(map[byte]int)
	for _, b := range bytes {
		if uniqueLetters[b] == 1 {
			return false
		}
		uniqueLetters[b]++
	}
	return true
}

func randomString(n int) string {
	if n <= 0 {
		return ""
	}
	
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	
	
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}
	
	return string(b)
}