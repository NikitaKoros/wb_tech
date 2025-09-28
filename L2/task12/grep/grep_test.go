package greppkg

import (
	"bytes"
	"strings"
	"testing"
)

func TestFixedStringIgnoreCase(t *testing.T) {
	cfg := &Config{
		FixedString:    true,
		IgnoreRegister: true,
	}

	input := "Foo\nbar\nfoobar\n"
	g, err := NewGrepper(cfg, "foo")
	if err != nil {
		t.Fatalf("NewGrepper error: %v", err)
	}

	var out bytes.Buffer
	if err := g.Process(strings.NewReader(input), "", &out, false); err != nil {
		t.Fatalf("Process error: %v", err)
	}

	got := out.String()
	expected := "Foo\nfoobar\n"
	if got != expected {
		t.Fatalf("unexpected output\n got:\n%q\nwant:\n%q", got, expected)
	}
}

func TestRegexpMatching(t *testing.T) {
	cfg := &Config{}
	g, err := NewGrepper(cfg, "a.b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := []string{"abc", "axb", "ab"}
	var buf bytes.Buffer
	err = g.Process(strings.NewReader(strings.Join(lines, "\n")+"\n"), "", &buf, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	want := "axb\n"
	if got != want {
		t.Errorf("unexpected output\n got:\n%q\nwant:\n%q", got, want)
	}
}

func TestInvertFilter(t *testing.T) {
	cfg := &Config{
		FixedString:  true,
		InvertFilter: true,
	}

	input := "line1\nxline\nline2\n"
	g, err := NewGrepper(cfg, "x")
	if err != nil {
		t.Fatalf("NewGrepper error: %v", err)
	}

	var out bytes.Buffer
	if err := g.Process(strings.NewReader(input), "", &out, false); err != nil {
		t.Fatalf("Process error: %v", err)
	}

	got := out.String()
	expected := "line1\nline2\n"
	if got != expected {
		t.Fatalf("unexpected output\n got:\n%q\nwant:\n%q", got, expected)
	}
}

func TestCountFlagWithFilename(t *testing.T) {
	cfg := &Config{
		FixedString: true,
		Count:       true,
	}

	input := "1\n2\n1\n3\n"
	g, err := NewGrepper(cfg, "1")
	if err != nil {
		t.Fatalf("NewGrepper error: %v", err)
	}

	var out bytes.Buffer
	if err := g.Process(strings.NewReader(input), "file.txt", &out, true); err != nil {
		t.Fatalf("Process error: %v", err)
	}

	got := out.String()
	expected := "file.txt:2\n"
	if got != expected {
		t.Fatalf("unexpected output\n got:%q\nwant:%q", got, expected)
	}
}

func TestContextBeforeAfterAndGroupSeparator(t *testing.T) {
	lines := []string{
		"line0",
		"match",
		"line2",
		"line3",
		"line4",
		"match",
		"line6",
	}
	input := strings.Join(lines, "\n") + "\n"

	cfg := &Config{
		FixedString: true,
		Before:      1,
		After:       0,
	}

	g, err := NewGrepper(cfg, "match")
	if err != nil {
		t.Fatalf("NewGrepper error: %v", err)
	}

	var out bytes.Buffer
	if err := g.Process(strings.NewReader(input), "", &out, false); err != nil {
		t.Fatalf("Process error: %v", err)
	}

	got := out.String()
	expected := "line0\nmatch\n--\nline4\nmatch\n"
	if got != expected {
		t.Fatalf("unexpected output\n got:\n%q\nwant:\n%q", got, expected)
	}
}

func TestContextMerging(t *testing.T) {
	lines := []string{
		"l0",
		"match",
		"match",
		"l3",
		"l4",
	}
	input := strings.Join(lines, "\n") + "\n"

	cfg := &Config{
		FixedString: true,
		Before:      1,
		After:       1,
	}

	g, err := NewGrepper(cfg, "match")
	if err != nil {
		t.Fatalf("NewGrepper error: %v", err)
	}

	var out bytes.Buffer
	if err := g.Process(strings.NewReader(input), "", &out, false); err != nil {
		t.Fatalf("Process error: %v", err)
	}

	got := out.String()
	expected := "l0\nmatch\nmatch\nl3\n"
	if got != expected {
		t.Fatalf("unexpected output\n got:\n%q\nwant:\n%q", got, expected)
	}
}

func TestPrintNumberAndShowFilename(t *testing.T) {
	cfg := &Config{
		FixedString: true,
		PrintNumber: true,
	}

	input := "a\nb\nmatch\nc\n"
	g, err := NewGrepper(cfg, "match")
	if err != nil {
		t.Fatalf("NewGrepper error: %v", err)
	}

	var out bytes.Buffer
	if err := g.Process(strings.NewReader(input), "myfile.txt", &out, true); err != nil {
		t.Fatalf("Process error: %v", err)
	}

	got := out.String()
	expected := "myfile.txt:3:match\n"
	if got != expected {
		t.Fatalf("unexpected output\n got:%q\nwant:%q", got, expected)
	}
}

func TestNoMatchesProducesNoOutput(t *testing.T) {
	cfg := &Config{
		FixedString: true,
	}

	input := "one\ntwo\nthree\n"
	g, err := NewGrepper(cfg, "absent")
	if err != nil {
		t.Fatalf("NewGrepper error: %v", err)
	}

	var out bytes.Buffer
	if err := g.Process(strings.NewReader(input), "", &out, false); err != nil {
		t.Fatalf("Process error: %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("expected no output but got %q", out.String())
	}
}
