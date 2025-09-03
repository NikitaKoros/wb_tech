package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	var (
		number int64
		i      int
		newBit int64
	)

	for {
		fmt.Fprint(w, "Number int64: ")
		w.Flush()
		
		line, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stdout, "Failed to read number from stdin: %v\n", err)
			w.Flush()
			os.Exit(1)
		}
		
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Fprintln(w, "Number is required")
			w.Flush()
			continue
		}

		number, err = strconv.ParseInt(line, 10, 64)
		if err != nil {
			fmt.Fprintln(w, "Failed to convert given string to int")
			w.Flush()
			continue
		}
		break
	}

	for {
		fmt.Fprint(w, "Bit place (1-indexed, LSB): ")
		w.Flush()
		
		line, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintf(w, "Failed to read bit place from stdin: %v\n", err)
			w.Flush()
			os.Exit(1)
		}
		
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Fprintln(w, "Bit place is required")
			w.Flush()
			continue
		}

		i, err = strconv.Atoi(line)
		if err != nil || i < 1 || i > 64 {
			fmt.Fprintln(w, "Bit place must be an integer between 1 and 64")
			w.Flush()
			continue
		}
		break
	}

	for {
		fmt.Fprint(w, "New bit (0/1): ")
		w.Flush()
		
		line, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintf(w, "Failed to read new bit from stdin: %v\n", err)
			w.Flush()
			os.Exit(1)
		}
		
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Fprintln(w, "New bit is required")
			w.Flush()
			continue
		}

		newBit, err = strconv.ParseInt(line, 10, 64)
		if err != nil || (newBit != 1 && newBit != 0) {
			fmt.Fprintln(w, "New bit must be 1 or 0")
			w.Flush()
			continue
		}
		break
	}

	fmt.Fprintf(w, "Number before: %d (binary: %b)\n", number, number)
	
	bitPos := i - 1
	mask := int64(1) << bitPos
	if newBit == 1 {
		number |= mask
	} else {
		number &= ^mask
	}
	
	fmt.Fprintf(w, "Number after: %d (binary: %b)\n", number, number)
}
