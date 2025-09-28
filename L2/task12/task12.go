package main

import (
	"flag"
	"fmt"
	greppkg "grep-util/grep"
	"log"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... [TEMPLATE] [FILE]...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Filter lines by given template.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	var (
		afterFlag          = flag.Int("A", 0, "display N lines after found line that contains template")
		beforeFlag         = flag.Int("B", 0, "display N lines before found line that contains template")
		aroundFlag         = flag.Int("C", 0, "display N lines before and after found line that contains template")
		countFlag          = flag.Bool("c", false, "print only number of lines containing template")
		ignoreRegisterFlag = flag.Bool("i", false, "ignore register")
		invertFilterFlag   = flag.Bool("v", false, "displays lines which do not contain template")
		fixedStringFlag    = flag.Bool("F", false, "process template as a fixed string and not a regexp")
		printNumberFlag    = flag.Bool("n", false, "print number of the line before it")
		helpFlag           = flag.Bool("help", false, "display this help and exit")
	)

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	if *aroundFlag > 0 {
		*afterFlag = *aroundFlag
		*beforeFlag = *aroundFlag
	}

	config := &greppkg.Config{
		After:          *afterFlag,
		Before:         *beforeFlag,
		Count:          *countFlag,
		IgnoreRegister: *ignoreRegisterFlag,
		InvertFilter:   *invertFilterFlag,
		FixedString:    *fixedStringFlag,
		PrintNumber:    *printNumberFlag,
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Template must not be empty")
		flag.Usage()
		os.Exit(2)
	}

	template := args[0]
	files := args[1:]
	grepper, err := greppkg.NewGrepper(config, template)
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		if err := handleInput(grepper, "", false); err != nil {
			log.Fatal(err)
		}
	} else {
		showFilename := len(files) > 1
		for _, f := range files {
			if err := handleInput(grepper, f, showFilename); err != nil {
				log.Println("error:", err)
			}
		}
	}
}

func handleInput(g *greppkg.Grepper, filename string, showFilename bool) error {
	var file *os.File

	if filename != "" {
		newFile, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer newFile.Close()

		file = newFile
	} else {
		file = os.Stdin
	}

	err := g.Process(file, filename, os.Stdout, showFilename)
	if err != nil {
		return err
	}

	return nil
}
