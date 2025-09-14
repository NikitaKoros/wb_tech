// Package for unpacking strings utility
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	w.WriteString("Write <exit> to exit program\n")
	for {
		w.WriteString("Input string to unpack: ")
		w.Flush()

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
			return
		}
		str := scanner.Text()

		if str == "exit" {
			w.WriteString("Exiting program\n")
			return
		}

		unpackedStr, err := Unpack(str)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unpacking failed: %v\n\n", err)
			continue
		}

		w.WriteString("Unpacked string: ")
		w.WriteString(unpackedStr)
		w.WriteString("\n\n")
	}
}

func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	var unpacked strings.Builder
	runes := []rune(input)
	length := len(runes)

	var lastRune rune
	var escaping bool
	var hasLiteral bool

	flushLast := func(count int) {
		if lastRune != 0 {
			if count > 0 {
				unpacked.WriteString(strings.Repeat(string(lastRune), count))
			}
			lastRune = 0
		}
	}

	for i := 0; i < length; i++ {
		ch := runes[i]

		switch {
		case escaping:
			flushLast(1)
			lastRune = ch
			hasLiteral = true
			escaping = false

		case ch == '\\':
			flushLast(1)
			escaping = true

		case isDigit(ch):
			if lastRune == 0 {
				return "", fmt.Errorf("invalid string: starts with digit at position %d", i)
			}

			j := i
			for j < length && isDigit(runes[j]) {
				j++
			}

			count, err := strconv.Atoi(string(runes[i:j]))
			if err != nil {
				return "", fmt.Errorf("invalid number at position %d: %w", i, err)
			}

			flushLast(count)
			i = j - 1

		default:
			flushLast(1)
			lastRune = ch
			hasLiteral = true
		}
	}

	if escaping {
		return "", fmt.Errorf("invalid string: unfinished escape at end")
	}
	flushLast(1)

	if !hasLiteral {
		return "", fmt.Errorf("invalid string: no literals found")
	}

	return unpacked.String(), nil
}


func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
