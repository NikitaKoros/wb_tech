package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	things := make([]interface{}, 0)

	things = append(things, 64)
	things = append(things, "cat")
	things = append(things, true)
	things = append(things, make(chan interface{}))

	random := rand.New(rand.NewSource(time.Now().Unix()))
	thing := things[random.Intn(len(things))]
	typeStr, err := identifyType(thing)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(thing)
	fmt.Println("It's a", typeStr)
}

func identifyType(v interface{}) (string, error) {
	switch v.(type) {
	case int:
		return "integer", nil
	case string:
		return "string", nil
	case bool:
		return "boolean", nil
	case chan interface{}:
		return "chan", nil
	default:
		return "", fmt.Errorf("failed to recognize value type")
	}
}
