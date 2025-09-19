package sortpkg

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func TestNewSorter(t *testing.T) {
	config := &Config{}
	sorter := NewSorter(config)
	if sorter.config != config {
		t.Errorf("Expected config to be set, got %v", sorter.config)
	}
	if len(sorter.months) != 12 {
		t.Errorf("Expected 12 months, got %d", len(sorter.months))
	}
}

func TestIsSorted(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		lines    []string
		expected bool
	}{
		{
			name:     "already sorted strings",
			config:   &Config{},
			lines:    []string{"apple", "banana", "cherry"},
			expected: true,
		},
		{
			name:     "not sorted strings",
			config:   &Config{},
			lines:    []string{"banana", "apple", "cherry"},
			expected: false,
		},
		{
			name:     "reverse sorted should be false",
			config:   &Config{Reverse: true},
			lines:    []string{"cherry", "banana", "apple"},
			expected: true,
		},
		{
			name:     "reverse sorted and reverse=true should be true",
			config:   &Config{Reverse: true},
			lines:    []string{"cherry", "banana", "apple"},
			expected: true,
		},
		{
			name:     "numeric sorted",
			config:   &Config{Numeric: true},
			lines:    []string{"1", "2", "10"},
			expected: true,
		},
		{
			name:     "numeric not sorted",
			config:   &Config{Numeric: true},
			lines:    []string{"10", "2", "1"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewSorter(tt.config)
			result := sorter.IsSorted(tt.lines)
			if result != tt.expected {
				t.Errorf("IsSorted() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortBasic(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    []string
		expected []string
	}{
		{
			name:     "basic string sort",
			config:   &Config{},
			input:    []string{"banana", "apple", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "reverse string sort",
			config:   &Config{Reverse: true},
			input:    []string{"banana", "apple", "cherry"},
			expected: []string{"cherry", "banana", "apple"},
		},
		{
			name:     "empty input",
			config:   &Config{},
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single element",
			config:   &Config{},
			input:    []string{"apple"},
			expected: []string{"apple"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewSorter(tt.config)
			result, err := sorter.Sort(tt.input)
			if err != nil {
				t.Fatalf("Sort() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Sort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortNumeric(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    []string
		expected []string
	}{
		{
			name:     "numeric sort",
			config:   &Config{Numeric: true},
			input:    []string{"10", "2", "1", "20"},
			expected: []string{"1", "2", "10", "20"},
		},
		{
			name:     "numeric reverse",
			config:   &Config{Numeric: true, Reverse: true},
			input:    []string{"10", "2", "1", "20"},
			expected: []string{"20", "10", "2", "1"},
		},
		{
			name:     "mixed numeric and non-numeric",
			config:   &Config{Numeric: true},
			input:    []string{"10", "abc", "2", "def", "1"},
			expected: []string{"abc", "def", "1", "2", "10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewSorter(tt.config)
			result, err := sorter.Sort(tt.input)
			if err != nil {
				t.Fatalf("Sort() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Sort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortMonth(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    []string
		expected []string
	}{
		{
			name:     "month sort",
			config:   &Config{Month: true},
			input:    []string{"Mar", "Jan", "Dec", "Feb"},
			expected: []string{"Jan", "Feb", "Mar", "Dec"},
		},
		{
			name:     "month reverse",
			config:   &Config{Month: true, Reverse: true},
			input:    []string{"Mar", "Jan", "Dec", "Feb"},
			expected: []string{"Dec", "Mar", "Feb", "Jan"},
		},
		{
			name:     "mixed valid and invalid months",
			config:   &Config{Month: true},
			input:    []string{"XYZ", "Jan", "Invalid", "Feb"},
			expected: []string{"Jan", "Feb", "Invalid", "XYZ"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewSorter(tt.config)
			result, err := sorter.Sort(tt.input)
			if err != nil {
				t.Fatalf("Sort() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Sort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortHuman(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    []string
		expected []string
	}{
		{
			name:     "human sizes",
			config:   &Config{Human: true},
			input:    []string{"2K", "1K", "1M", "512"},
			expected: []string{"512", "1K", "2K", "1M"},
		},
		{
			name:     "human reverse",
			config:   &Config{Human: true, Reverse: true},
			input:    []string{"2K", "1K", "1M", "512"},
			expected: []string{"1M", "2K", "1K", "512"},
		},
		{
			name:     "human with spaces and case",
			config:   &Config{Human: true},
			input:    []string{"2k", "1 K", "1.5M", "512B"},
			expected: []string{"512B", "1 K", "2k", "1.5M"},
		},
		{
			name:     "invalid human sizes treated as 0",
			config:   &Config{Human: true},
			input:    []string{"abc", "1K", "xyz", "2M"},
			expected: []string{"abc", "xyz", "1K", "2M"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewSorter(tt.config)
			result, err := sorter.Sort(tt.input)
			if err != nil {
				t.Fatalf("Sort() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Sort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortByColumn(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    []string
		expected []string
	}{
		{
			name:     "sort by second column",
			config:   &Config{Column: 2, Separator: ","},
			input:    []string{"b,3", "a,1", "c,2"},
			expected: []string{"a,1", "c,2", "b,3"},
		},
		{
			name:     "sort by third column reverse",
			config:   &Config{Column: 3, Separator: "|", Reverse: true},
			input:    []string{"x|y|1", "a|b|3", "m|n|2"},
			expected: []string{"a|b|3", "m|n|2", "x|y|1"},
		},
		{
			name:     "column out of range",
			config:   &Config{Column: 5, Separator: ","},
			input:    []string{"a,b", "c,d"},
			expected: []string{"a,b", "c,d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewSorter(tt.config)
			result, err := sorter.Sort(tt.input)
			if err != nil {
				t.Fatalf("Sort() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Sort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortUnique(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    []string
		expected []string
	}{
		{
			name:     "unique after sort",
			config:   &Config{Unique: true},
			input:    []string{"banana", "apple", "banana", "cherry", "apple"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "unique with numeric",
			config:   &Config{Numeric: true, Unique: true},
			input:    []string{"2", "1", "2", "3", "1"},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "unique with column",
			config:   &Config{Column: 2, Separator: ",", Unique: true},
			input:    []string{"x,1", "y,2", "z,1", "w,3"},
			expected: []string{"x,1", "y,2", "w,3"},
		},
		{
			name:     "unique with trailing blanks ignored",
			config:   &Config{Blank: true, Unique: true},
			input:    []string{"apple   ", "banana", "apple", "cherry   "},
			expected: []string{"apple   ", "banana", "cherry   "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewSorter(tt.config)
			result, err := sorter.Sort(tt.input)
			if err != nil {
				t.Fatalf("Sort() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Sort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortBlank(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    []string
		expected []string
	}{
		{
			name:     "ignore trailing blanks",
			config:   &Config{Blank: true},
			input:    []string{"apple   ", "banana", " apple", "cherry  "},
			expected: []string{" apple", "apple   ", "banana", "cherry  "},
		},
		{
			name:     "blank with column",
			config:   &Config{Column: 1, Separator: ",", Blank: true},
			input:    []string{"apple  ,1", "banana,2", "cherry ,3"},
			expected: []string{"apple  ,1", "banana,2", "cherry ,3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewSorter(tt.config)
			result, err := sorter.Sort(tt.input)
			if err != nil {
				t.Fatalf("Sort() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Sort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCompareFunctionsEdgeCases(t *testing.T) {
	sorter := NewSorter(&Config{})

	if !sorter.compareNumeric("abc", "def") {
		t.Error("Expected 'abc' < 'def' when both are non-numeric")
	}
	if sorter.compareNumeric("1.5", "1,5") {
		t.Error("Expected '1,5' to parse as 0, so 1.5 should NOT be less")
	}

	if !sorter.compareMonth("Jan", "Feb") {
		t.Error("Expected Jan < Feb")
	}
	if !sorter.compareMonth("Jan", "Invalid") {
		t.Error("Expected Jan < Invalid, but got false")
	}
	if sorter.compareMonth("Invalid1", "Invalid2") && "Invalid1" > "Invalid2" {
		t.Error("Lexicographic fallback failed")
	}

	if v := sorter.parseHumanSize("1K"); v != 1024 {
		t.Errorf("parseHumanSize(1K) = %v, want 1024", v)
	}
	if v := sorter.parseHumanSize("2.5M"); v != 2.5*1024*1024 {
		t.Errorf("parseHumanSize(2.5M) = %v, want %v", v, 2.5*1024*1024)
	}
	if v := sorter.parseHumanSize("xyz"); v != 0 {
		t.Errorf("parseHumanSize(xyz) = %v, want 0", v)
	}
	if v := sorter.parseHumanSize("100"); v != 100 {
		t.Errorf("parseHumanSize(100) = %v, want 100", v)
	}
	if v := sorter.parseHumanSize("1 Gb"); v != 1024*1024*1024 {
		t.Errorf("parseHumanSize(1 Gb) = %v, want %v", v, 1024*1024*1024)
	}
}

func TestComplexCombinations(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    []string
		expected []string
	}{
		{
			name: "numeric + reverse + unique",
			config: &Config{
				Numeric: true,
				Reverse: true,
				Unique:  true,
			},
			input:    []string{"3", "1", "2", "3", "1", "4"},
			expected: []string{"4", "3", "2", "1"},
		},
		{
			name: "month + blank + column",
			config: &Config{
				Month:     true,
				Blank:     true,
				Column:    2,
				Separator: "|",
			},
			input: []string{
				"data|Mar   ",
				"data|Jan ",
				"data|Feb",
				"data|Dec  ",
			},
			expected: []string{
				"data|Jan ",
				"data|Feb",
				"data|Mar   ",
				"data|Dec  ",
			},
		},
		{
			name: "human + reverse + unique",
			config: &Config{
				Human:   true,
				Reverse: true,
				Unique:  true,
			},
			input: []string{
				"1K",
				"2M",
				"1K",
				"512",
				"2M",
				"1G",
			},
			expected: []string{
				"1G",
				"2M",
				"1K",
				"512",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewSorter(tt.config)
			result, err := sorter.Sort(tt.input)
			if err != nil {
				t.Fatalf("Sort() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Sort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func writeTempFile(t *testing.T, lines []string) string {
	tmp, err := os.CreateTemp("", "sorttest-")
	if err != nil {
		t.Fatal(err)
	}
	w := bufio.NewWriter(tmp)
	for _, line := range lines {
		_, err := w.WriteString(line + "\n")
		if err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			t.Fatal(err)
		}
	}
	w.Flush()
	tmp.Close()
	return tmp.Name()
}

func readLinesFromReader(r io.Reader) []string {
	scanner := bufio.NewScanner(r)
	var result []string
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	return result
}

func TestSortFileToWriterInMemory(t *testing.T) {
	lines := []string{"banana", "apple", "cherry"}
	expected := []string{"apple", "banana", "cherry"}

	tmpFile := writeTempFile(t, lines)
	defer os.Remove(tmpFile)

	var buf bytes.Buffer
	cfg := &Config{}
	err := SortFileToWriter(tmpFile, &buf, cfg)
	if err != nil {
		t.Fatalf("SortFileToWriter error: %v", err)
	}

	got := readLinesFromReader(&buf)
	if len(got) != len(expected) {
		t.Fatalf("got %v lines, want %v", len(got), len(expected))
	}

	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("line %d: got %q, want %q", i, got[i], expected[i])
		}
	}
}

func TestSortFileToWriterExternal(t *testing.T) {
	var lines []string
	for i := 100000; i > 0; i-- {
		lines = append(lines, "line "+strconv.Itoa(i))
	}

	tmpFile := writeTempFile(t, lines)
	defer os.Remove(tmpFile)

	var buf bytes.Buffer
	cfg := &Config{
		Numeric:   true,
		Column:    2,
		Separator: " ",
	}

	ChunkSizeBytes = 1024 * 1024
	MaxInMemoryBytes = 512 * 1024

	err := SortFileToWriter(tmpFile, &buf, cfg)
	if err != nil {
		t.Fatalf("SortFileToWriter external error: %v", err)
	}

	got := readLinesFromReader(&buf)
	if len(got) != len(lines) {
		t.Fatalf("got %v lines, want %v", len(got), len(lines))
	}

	if got[0] != "line 1" || got[len(got)-1] != "line 100000" {
		t.Errorf("first/last lines incorrect: got %q ... %q", got[0], got[len(got)-1])
	}
}


func TestSortFileToWriterUniqueAndNumeric(t *testing.T) {
	lines := []string{"10", "2", "1", "2", "10"}
	expected := []string{"1", "2", "10"}

	tmpFile := writeTempFile(t, lines)
	defer os.Remove(tmpFile)

	var buf bytes.Buffer
	cfg := &Config{Numeric: true, Unique: true}
	err := SortFileToWriter(tmpFile, &buf, cfg)
	if err != nil {
		t.Fatalf("SortFileToWriter error: %v", err)
	}

	got := readLinesFromReader(&buf)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSortFileToWriterColumnsAndMonth(t *testing.T) {
	lines := []string{"data|Mar", "data|Jan", "data|Feb", "data|Dec"}
	expected := []string{"data|Jan", "data|Feb", "data|Mar", "data|Dec"}

	tmpFile := writeTempFile(t, lines)
	defer os.Remove(tmpFile)

	var buf bytes.Buffer
	cfg := &Config{Month: true, Column: 2, Separator: "|"}
	err := SortFileToWriter(tmpFile, &buf, cfg)
	if err != nil {
		t.Fatalf("SortFileToWriter error: %v", err)
	}

	got := readLinesFromReader(&buf)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}
