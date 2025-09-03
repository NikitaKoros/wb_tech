package main

import "fmt"

func main() {
	strings := []string{"cat", "dog", "bird", "fish", "ant",
		"cat", "dog", "ant", "bee", "rat",
		"bee", "rat", "cat", "frog", "egg",
		"dog", "ant", "bee", "pig", "owl"}
	
	set := make(map[string]struct{})
	
	for _, str := range strings {
		set[str] = struct{}{}
	}
	
	keys := make([]string, 0, len(set))
	for key, _ := range set {
		keys = append(keys, key)
	}
	fmt.Println("Set of strings: ", keys)
}
