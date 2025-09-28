package greppkg

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Config holds grep configuration options
type Config struct {
	After          int
	Before         int
	Count          bool
	IgnoreRegister bool
	InvertFilter   bool
	FixedString    bool
	PrintNumber    bool
}

// Grepper implements grep functionality
type Grepper struct {
	Config  *Config
	Pattern string

	re           *regexp.Regexp
	lowerPattern string
}

// NewGrepper creates a Grepper with the given config and search pattern.
// Returns an error if config or pattern is invalid.
func NewGrepper(cfg *Config, pattern string) (*Grepper, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config for new grepper cannot be empty")
	}

	if pattern == "" {
		return nil, fmt.Errorf("pattern for new grepper cannot be empty")
	}

	lowPattern := strings.ToLower(pattern)

	g := &Grepper{
		Config:       cfg,
		Pattern:      pattern,
		lowerPattern: lowPattern,
	}

	if g.Config.FixedString {
		return g, nil
	}

	regexPattern := pattern
	if g.Config.IgnoreRegister {
		regexPattern = "(?i)" + regexPattern
	}

	compiledRegex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regexp %s: %w", regexPattern, err)
	}

	g.re = compiledRegex
	return g, nil
}

// Process reads input, finds matching lines, and writes results to w.
// If showFilename is true and filename is not empty, prefixes lines with filename.
func (g *Grepper) Process(r io.Reader, filename string, w io.Writer, showFilename bool) error {
	br := bufio.NewReader(r)

	var lines []string
	for {
		line, err := br.ReadString('\n')
		if errors.Is(err, io.EOF) {
			if len(line) > 0 {
				lines = append(lines, strings.TrimRight(line, "\r\n"))
			}
			break
		}
		if err != nil {
			return err
		}
		lines = append(lines, strings.TrimRight(line, "\r\n"))
	}

	var matchedLineIndices []int
	for i, line := range lines {
		isMatch := g.isMatch(line)
		if g.Config.InvertFilter {
			isMatch = !isMatch
		}
		if isMatch {
			matchedLineIndices = append(matchedLineIndices, i)
		}
	}

	if g.Config.Count {
		count := len(matchedLineIndices)
		if showFilename && filename != "" {
			fmt.Fprintf(w, "%s:%d\n", filename, count)
		} else {
			fmt.Fprintf(w, "%d\n", count)
		}
		return nil
	}

	if len(matchedLineIndices) == 0 {
		return nil
	}

	type interval struct {
		start int
		end   int
	}

	var intervalsWithContext []interval
	for _, idx := range matchedLineIndices {
		startIdx := max(idx-g.Config.Before, 0)

		endIdx := idx + g.Config.After
		if endIdx >= len(lines) {
			endIdx = len(lines) - 1
		}

		intervalsWithContext = append(intervalsWithContext, interval{startIdx, endIdx})
	}

	var mergedIntervals []interval
	for _, it := range intervalsWithContext {
		if len(mergedIntervals) == 0 {
			mergedIntervals = append(mergedIntervals, it)
			continue
		}
		last := &mergedIntervals[len(mergedIntervals)-1]
		if it.start <= last.end+1 {
			if it.end > last.end {
				last.end = it.end
			}
		} else {
			mergedIntervals = append(mergedIntervals, it)
		}
	}

	firstGroup := true
	printGroupSeparator := g.Config.Before > 0 || g.Config.After > 0
	
	for _, grp := range mergedIntervals {
		if !firstGroup {
			if printGroupSeparator {
				fmt.Fprintln(w, "--")
			}
		}
		firstGroup = false
		for i := grp.start; i <= grp.end; i++ {
			var prefix string
			if showFilename && filename != "" {
				prefix = filename + ":"
			}
			if g.Config.PrintNumber {
				prefix = fmt.Sprintf("%s%d:", prefix, i+1)
			}
			if prefix != "" {
				fmt.Fprint(w, prefix)
			}
			fmt.Fprintln(w, lines[i])
		}
	}

	return nil
}

// isMatch returns true if line matches the pattern based on current settings.
func (g *Grepper) isMatch(line string) bool {

	if g.Config.FixedString {
		if g.Config.IgnoreRegister {
			return strings.Contains(strings.ToLower(line), g.lowerPattern)
		}
		return strings.Contains(line, g.Pattern)
	}

	return g.re.MatchString(line)
}
