package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Printf("Input string: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read string from stdin: %v", err)
	}
	
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\t")
	
	reversed := reverseWords(input)
	fmt.Println("Reversed sring:", reversed)
}

func reverse(r []rune, start, end int) {
	for start < end {
		r[start], r[end] = r[end], r[start]
		start++
		end--
	}
}

func reverseWords(s string) string {
	runes := []rune(s)
	n := len(runes)
	
	reverse(runes, 0, n - 1)
	
	start := 0
	for i := 0; i <= n; i++ {
		if i == n || runes[i] == ' ' {
			reverse(runes, start, i-1)
			start = i+1
		}
	}
	
	return string(runes)
}
