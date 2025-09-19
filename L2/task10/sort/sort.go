package sortpkg

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// MaxInMemoryBytes — maximum file size (in bytes) that will be sorted in-memory.
// If the file is larger — external sort is used.
var MaxInMemoryBytes int64 = 100 * 1024 * 1024 // 100 MB

// ChunkSizeBytes — size of a chunk (in bytes) for external sort.
var ChunkSizeBytes int64 = 10 * 1024 * 1024 // 10 MB

// Config holds sorting configuration options
type Config struct {
	Column    int    // column number to sort by (1-indexed)
	Numeric   bool   // numeric sort
	Reverse   bool   // reverse sort order
	Unique    bool   // output only unique lines
	Month     bool   // sort by month names
	Blank     bool   // ignore trailing blanks
	Check     bool   // check if input is already sorted
	Human     bool   // human-readable sizes (e.g., 1K, 2M)
	Separator string // field separator
}

// Sorter implements sorting functionality
type Sorter struct {
	config *Config
	months map[string]int
}

func NewSorter(config *Config) *Sorter {
	months := map[string]int{
		"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4,
		"May": 5, "Jun": 6, "Jul": 7, "Aug": 8,
		"Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	}

	return &Sorter{
		config: config,
		months: months,
	}
}

// IsSorted checks whether the input lines are already sorted
func (s *Sorter) IsSorted(lines []string) bool {
	for i := 1; i < len(lines); i++ {
		if s.less(lines[i], lines[i-1]) {
			return false
		}
	}
	return true
}

// Sort performs sorting on the input string slice
func (s *Sorter) Sort(lines []string) ([]string, error) {
	if len(lines) == 0 {
		return lines, nil
	}

	result := make([]string, len(lines))
	copy(result, lines)

	sort.SliceStable(result, func(i, j int) bool {
		return s.less(result[i], result[j])
	})

	if s.config.Unique {
		unique := make([]string, 0, len(result))
		var prevKey string
		first := true
		for _, line := range result {
			key := s.getKey(line)
			if first || key != prevKey {
				unique = append(unique, line)
				prevKey = key
				first = false
			}
		}
		result = unique
	}

	return result, nil
}

// less compares two strings according to configuration
func (s *Sorter) less(a, b string) bool {
	var keyA, keyB string

	if s.config.Blank {
		a = strings.TrimRight(a, " \t")
		b = strings.TrimRight(b, " \t")
	}

	if s.config.Column > 0 {
		keyA = s.extractColumn(a, s.config.Column)
		keyB = s.extractColumn(b, s.config.Column)
	} else {
		keyA = a
		keyB = b
	}

	var result bool

	switch {
	case s.config.Human:
		result = s.compareHuman(keyA, keyB)
	case s.config.Numeric:
		result = s.compareNumeric(keyA, keyB)
	case s.config.Month:
		result = s.compareMonth(keyA, keyB)
	default:
		result = keyA < keyB
	}

	if s.config.Reverse {
		result = !result
	}

	return result
}

// extractColumn extracts the specified column from a line
func (s *Sorter) extractColumn(line string, column int) string {
	fields := strings.Split(line, s.config.Separator)
	if column > len(fields) || column < 1 {
		return ""
	}
	return fields[column-1]
}

// getKey returns the sorting key for a line according to config
func (s *Sorter) getKey(line string) string {
	if s.config.Blank {
		line = strings.TrimRight(line, " \t")
	}

	if s.config.Column > 0 {
		return s.extractColumn(line, s.config.Column)
	}

	return line
}

// compareNumeric compares strings as numbers
func (s *Sorter) compareNumeric(a, b string) bool {
	numA, errA := strconv.ParseFloat(strings.TrimSpace(a), 64)
	numB, errB := strconv.ParseFloat(strings.TrimSpace(b), 64)

	if errA != nil {
		numA = 0
	}
	if errB != nil {
		numB = 0
	}

	if numA != numB {
		return numA < numB
	}

	return a < b
}

// compareMonth compares strings as month names
func (s *Sorter) compareMonth(a, b string) bool {
	monthA, okA := s.months[strings.TrimSpace(a)]
	monthB, okB := s.months[strings.TrimSpace(b)]

	if okA && okB {
		return monthA < monthB
	}

	if okA && !okB {
		return true
	}
	if !okA && okB {
		return false
	}

	return a < b
}

// compareHuman compares strings with human-readable sizes
func (s *Sorter) compareHuman(a, b string) bool {
	sizeA := s.parseHumanSize(strings.TrimSpace(a))
	sizeB := s.parseHumanSize(strings.TrimSpace(b))
	return sizeA < sizeB
}

// parseHumanSize parses human-readable size into bytes
func (s *Sorter) parseHumanSize(str string) float64 {
	re := regexp.MustCompile(`^\s*(\d+(?:\.\d+)?)\s*([KMGTPEZYkmgtpezy]?)[Bb]?\s*$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(str))

	if len(matches) != 3 {
		if num, err := strconv.ParseFloat(str, 64); err == nil {
			return num
		}
		return 0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	suffix := strings.ToUpper(matches[2])
	multipliers := map[string]float64{
		"":  1,
		"K": 1024,
		"M": 1024 * 1024,
		"G": 1024 * 1024 * 1024,
		"T": 1024 * 1024 * 1024 * 1024,
		"P": 1024 * 1024 * 1024 * 1024 * 1024,
		"E": 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
		"Z": 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
		"Y": 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
	}

	multiplier, exists := multipliers[suffix]
	if !exists {
		multiplier = 1
	}

	return value * multiplier
}

// ---------------------- External sort section ----------------------

// SortFileToWriter: sorts inputPath and writes to w.
// If file size is less than or equal to MaxInMemoryBytes, in-memory sorting is used.
// Otherwise — external sort (splitting into chunks and then k-way merge).
func SortFileToWriter(inputPath string, w io.Writer, cfg *Config) error {
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("stat input file: %w", err)
	}

	sorter := NewSorter(cfg)

	if cfg.Check {
		f, err := os.Open(inputPath)
		if err != nil {
			return err
		}
		defer f.Close()
		reader := bufio.NewReader(f)
		var prev string
		first := true
		for {
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}
			if line != "" {
				line = strings.TrimSuffix(line, "\n")
				if first {
					prev = line
					first = false
				} else {
					if sorter.less(line, prev) {
						return fmt.Errorf("sort: %s: disorder detected", inputPath)
					}
					prev = line
				}
			}
			if err == io.EOF {
				break
			}
		}
		return nil
	}

	if info.Size() <= MaxInMemoryBytes {
		// in-memory sorting
		lines := []string{}
		f, err := os.Open(inputPath)
		if err != nil {
			return err
		}
		defer f.Close()

		reader := bufio.NewReader(f)
		for {
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}
			if line != "" {
				lines = append(lines, strings.TrimSuffix(line, "\n"))
			}
			if err == io.EOF {
				break
			}
		}

		sorted, err := sorter.Sort(lines)
		if err != nil {
			return err
		}

		var prevKey string
		first := true
		bufw := bufio.NewWriter(w)
		for _, line := range sorted {
			if cfg.Unique {
				key := sorter.getKey(line)
				if first || key != prevKey {
					fmt.Fprintln(bufw, line)
					prevKey = key
					first = false
				}
			} else {
				fmt.Fprintln(bufw, line)
			}
		}
		return bufw.Flush()
	}

	// external sort
	return externalSortToWriter(inputPath, w, sorter)
}

// externalSortToWriter — splits the file into chunks, sorts them, and performs k-way merge, writing the result to w.
func externalSortToWriter(inputPath string, w io.Writer, sorter *Sorter) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	var tempFiles []string
	var chunk []string
	var chunkBytes int64

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			cleanupTempFiles(tempFiles)
			return err
		}

		if line != "" {
			plain := strings.TrimSuffix(line, "\n")
			chunk = append(chunk, plain)
			chunkBytes += int64(len(plain)) + 1
		}

		if chunkBytes >= ChunkSizeBytes || (err == io.EOF && len(chunk) > 0) {
			tmpName, werr := writeSortedChunkToTemp(chunk, sorter)
			if werr != nil {
				cleanupTempFiles(tempFiles)
				return werr
			}
			tempFiles = append(tempFiles, tmpName)
			chunk = nil
			chunkBytes = 0
		}

		if err == io.EOF {
			break
		}
	}

	if len(tempFiles) == 0 {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return err
		}
		allLines := []string{}
		r := bufio.NewReader(f)
		for {
			line, err := r.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}
			if line != "" {
				allLines = append(allLines, strings.TrimSuffix(line, "\n"))
			}
			if err == io.EOF {
				break
			}
		}
		sorted, err := sorter.Sort(allLines)
		if err != nil {
			return err
		}
		bufw := bufio.NewWriter(w)
		for _, line := range sorted {
			fmt.Fprintln(bufw, line)
		}
		return bufw.Flush()
	}

	if err := mergeTempFilesToWriter(tempFiles, w, sorter); err != nil {
		cleanupTempFiles(tempFiles)
		return err
	}

	cleanupTempFiles(tempFiles)
	return nil
}

// writeSortedChunkToTemp — sorts the chunk and writes it to a temporary file, returning the file name.
func writeSortedChunkToTemp(chunk []string, sorter *Sorter) (string, error) {
	sorted, err := sorter.Sort(chunk)
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", "sortchunk-")
	if err != nil {
		return "", err
	}
	writer := bufio.NewWriter(tmpFile)
	for _, line := range sorted {
		if _, err := fmt.Fprintln(writer, line); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return "", err
		}
	}
	if err := writer.Flush(); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	return tmpFile.Name(), nil
}

// mergeTempFilesToWriter — performs k-way merge of temporary files and writes the result to w.
func mergeTempFilesToWriter(tempFiles []string, w io.Writer, sorter *Sorter) error {
	type readerWrap struct {
		file   *os.File
		reader *bufio.Reader
	}

	readers := make([]readerWrap, 0, len(tempFiles))
	for _, fn := range tempFiles {
		f, err := os.Open(fn)
		if err != nil {
			for _, rw := range readers {
				rw.file.Close()
			}
			return err
		}
		readers = append(readers, readerWrap{file: f, reader: bufio.NewReader(f)})
	}

	h := &lineHeap{items: []*lineItem{}, sorter: sorter}
	heap.Init(h)

	for idx := range readers {
		line, err := readLineFromReader(readers[idx].reader)
		if err != nil && err != io.EOF {
			for _, rw := range readers {
				rw.file.Close()
			}
			return err
		}
		if line != "" {
			heap.Push(h, &lineItem{
				line:  line,
				index: idx,
			})
		}
	}

	bufw := bufio.NewWriter(w)
	var prevKey string
	firstOut := true

	for h.Len() > 0 {
		item := heap.Pop(h).(*lineItem)
		if sorter.config.Unique {
			key := sorter.getKey(item.line)
			if firstOut || key != prevKey {
				if _, err := fmt.Fprintln(bufw, item.line); err != nil {
					for _, rw := range readers {
						rw.file.Close()
					}
					return err
				}
				prevKey = key
				firstOut = false
			}
		} else {
			if _, err := fmt.Fprintln(bufw, item.line); err != nil {
				for _, rw := range readers {
					rw.file.Close()
				}
				return err
			}
		}

		idx := item.index
		nextLine, err := readLineFromReader(readers[idx].reader)
		if err != nil && err != io.EOF {
			for _, rw := range readers {
				rw.file.Close()
			}
			return err
		}
		if nextLine != "" {
			heap.Push(h, &lineItem{
				line:  nextLine,
				index: idx,
			})
		}
	}

	for _, rw := range readers {
		rw.file.Close()
	}

	if err := bufw.Flush(); err != nil {
		return err
	}
	return nil
}

// readLineFromReader reads a line (without trailing '\n') from bufio.Reader.
func readLineFromReader(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	line = strings.TrimSuffix(line, "\n")
	if line != "" {
		return line, nil
	}
	if err != nil {
		return "", err
	}
	return "", nil
}

func cleanupTempFiles(files []string) {
	for _, fn := range files {
		_ = os.Remove(fn)
	}
}

// ---------------------- heap for k-way merge ----------------------

type lineItem struct {
	line  string
	index int // index of the reader in the readers array
}

type lineHeap struct {
	items  []*lineItem
	sorter *Sorter
}

func (h lineHeap) Len() int { return len(h.items) }
func (h lineHeap) Less(i, j int) bool {
	return h.sorter.less(h.items[i].line, h.items[j].line)
}
func (h lineHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }

func (h *lineHeap) Push(x interface{}) {
	h.items = append(h.items, x.(*lineItem))
}

func (h *lineHeap) Pop() interface{} {
	n := len(h.items)
	if n == 0 {
		return nil
	}
	it := h.items[n-1]
	h.items = h.items[:n-1]
	return it
}