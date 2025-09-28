package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	anagrams := FindAnagrams(words)

	for k, v := range anagrams {
		fmt.Printf("%q: %v\n", k, v)
	}
}

func FindAnagrams(words []string) map[string][]string {
	anagrams := make(map[string][]string)
	result := make(map[string][]string)
	for _, word := range words {
		wordLower := strings.ToLower(word)
		sortedWord := sortString(wordLower)
		anagrams[sortedWord] = append(anagrams[sortedWord], wordLower)
	}

	for _, group := range anagrams {
		if len(group) > 1 {
			sort.Strings(group)
			result[group[0]] = append(result[group[0]], group...)
		}
	}

	return result
}

func sortString(s string) string {
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	return string(runes)
}
