package task_1

import "fmt"

type Human struct {
	Name string
	Age int
}

func (h *Human) Greet() {
	fmt.Printf("Hello, my name is %s\n", h.Name)
}

func (h *Human) DescribeAge() {
	fmt.Printf("I am %d years old\n", h.Age)
}

type Action struct {
	Human
	Energy int
}

func (a *Action) Perform() {
    if a.Energy > 0 {
        a.Greet()
        a.Energy--
    }
}