package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	sortpkg "sort-utility/sort"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... [FILE]...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Sort lines of text files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	args := os.Args[1:]
	processdArgs := make([]string, 0)
	for _, arg := range args {
		if len(arg) > 1 && arg[0] == '-' && arg[1] != '-' {
			for _, c := range arg[1:] {
				processdArgs = append(processdArgs, "-"+string(c))
			}
		} else {
			processdArgs = append(processdArgs, arg)
		}
	}
	os.Args = append([]string{os.Args[0]}, processdArgs...)

	var (
		columnFlag    = flag.Int("k", 0, "sort by column N (1-indexed)")
		numericFlag   = flag.Bool("n", false, "sort numerically")
		reverseFlag   = flag.Bool("r", false, "reverse sort order")
		uniqueFlag    = flag.Bool("u", false, "output only unique lines")
		monthFlag     = flag.Bool("M", false, "sort by month names")
		blankFlag     = flag.Bool("b", false, "ignore trailing blanks")
		checkFlag     = flag.Bool("c", false, "check if input is sorted")
		humanFlag     = flag.Bool("h", false, "sort by human-readable sizes")
		separatorFlag = flag.String("t", "\t", "field separator")
		helpFlag      = flag.Bool("help", false, "display this help and exit")
	)

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	config := &sortpkg.Config{
		Column:    *columnFlag,
		Numeric:   *numericFlag,
		Reverse:   *reverseFlag,
		Unique:    *uniqueFlag,
		Month:     *monthFlag,
		Blank:     *blankFlag,
		Check:     *checkFlag,
		Human:     *humanFlag,
		Separator: *separatorFlag,
	}

	files := flag.Args()
	switch len(files) {
	case 0:
		if err := handleInput("", config, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "sort: %v\n", err)
			os.Exit(1)
		}
	default:
		for i, file := range files {
			if len(files) > 1 {
				fmt.Printf("==> %s <==\n", file)
			}
			if err := handleInput(file, config, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "sort: %s: %v\n", file, err)
				os.Exit(1)
			}
			if i != len(files)-1 {
				fmt.Println()
			}
		}
	}
}

func handleInput(filepath string, cfg *sortpkg.Config, w io.Writer) error {
	if filepath != "" {
		return sortpkg.SortFileToWriter(filepath, w, cfg)
	}

	tmpFile, err := os.CreateTemp("", "sortstdin-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	_, err = io.Copy(tmpFile, os.Stdin)
	if err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()
	return sortpkg.SortFileToWriter(tmpFile.Name(), w, cfg)
}
