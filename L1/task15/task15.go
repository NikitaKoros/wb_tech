package main

import "strings"

var justString string

func someFunc() {
  v := createHugeString(1 << 10)
  justString = strings.Clone(v[:100])
}

func createHugeString(size int) string {
	return strings.Repeat("a", size)
}

func main() {
  someFunc()
}