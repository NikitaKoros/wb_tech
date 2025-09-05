package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	var str string
	for {
		fmt.Print("String to flip over: ")
		
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read from stdin:", err)
			continue
		}
		
		input = strings.TrimSuffix(input, "\n")
		input = strings.TrimSuffix(input, "\r")
		
		if input == "" {
			fmt.Println("Given string must not be empty")
			continue
		}
		str = input
		break
	}
	
	runes := []rune(str)
	for i := range len(runes) / 2 {
		runes[i], runes[len(runes)-1-i] = runes[len(runes)-1-i], runes[i]
	}
	
	fmt.Println("Flipped string:", string(runes))
}
