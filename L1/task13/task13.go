package main

import "fmt"

func main() {
	a, b := 15, 18
	fmt.Println("a =", a, ", b =", b)
	a = a ^ b
	b = a ^ b
	a = a ^ b
	fmt.Println("a =", a, ", b =", b)
}