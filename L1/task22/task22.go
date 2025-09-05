package main

import (
	"fmt"
	"math/big"
)

func main() {
	a := big.NewInt(2_000_000)
	b := big.NewInt(3_000_000)
	
	calculateWithBigInt(a, b)
}

func calculateWithBigInt(a, b *big.Int) {
	sum := new(big.Int).Add(a, b)
	fmt.Printf("Summary: %s + %s = %s\n", a, b, sum)
	
	diff := new(big.Int).Sub(a, b)
	fmt.Printf("Subtraction: %s - %s = %s\n", a, b, diff)
	
	mul := new(big.Int).Mul(a, b)
	fmt.Printf("Multiplication: %s * %s = %s\n", a, b, mul)
	
	zero := big.NewInt(0)
	if b.Cmp(zero) == 0 {
		fmt.Println("Devision: error - division by zero")
	} else {
		div := new(big.Int).Div(a, b)
		fmt.Printf("Division: %s / %s = %s\n", a, b, div)
	}
}