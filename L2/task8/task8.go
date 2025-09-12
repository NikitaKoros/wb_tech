// Package implements a program that prints time using NTP-server
package main

import (
	"fmt"
	"log"

	"github.com/beevik/ntp"
)

const (
	defaultNTPServer = "time.google.com"
)

func main() {
	time, err := ntp.Time(defaultNTPServer)
	if err != nil {
		log.Fatalf("Failed to get time from NTP-server: %v\n", err)
	}
	
	fmt.Println(time.Format("2006-01-02 15:04:05.000 -0700 MST"))
}
