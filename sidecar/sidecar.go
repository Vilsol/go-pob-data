package sidecar

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

// Confidence grades how a file was identified; FS serves only entries at or above its threshold.
type Confidence int

const (
	Low Confidence = iota + 1
	Medium
	High
)

func (c Confidence) String() string {
	switch c {
	case High:
		return "high"
	case Medium:
		return "medium"
	case Low:
		return "low"
	}
	return "none"
}

func ParseConfidence(s string) (Confidence, error) {
	for _, c := range []Confidence{Low, Medium, High} {
		if c.String() == s {
			return c, nil
		}
	}
	return 0, fmt.Errorf("unknown confidence %q (low, medium, high)", s)
}

// Entry is one identified file; Path is the lower-case game path the bundle loader would serve it under.
type Entry struct {
	Path       string
	Bundle     string
	Offset     int64
	Size       int64
	Confidence Confidence
	Method     string
}

// Span is a stretch of a bundle that holds a file nothing identified.
type Span struct {
	Bundle string
	Offset int64
	Size   int64
	Note   string
}

const (
	entryHeader = "path\tbundle\toffset\tsize\tconfidence\tmethod"
	spanHeader  = "bundle\toffset\tsize\tnote"
)

// WriteEntries writes entries as TSV sorted by path, so two sidecars (or a sidecar and a real index) diff by line.
func WriteEntries(w io.Writer, es []Entry) error {
	es = slices.Clone(es)
	slices.SortFunc(es, func(a, b Entry) int { return strings.Compare(a.Path, b.Path) })
	bw := bufio.NewWriter(w)
	fmt.Fprintln(bw, entryHeader)
	for _, e := range es {
		fmt.Fprintf(bw, "%s\t%s\t%d\t%d\t%s\t%s\n", e.Path, e.Bundle, e.Offset, e.Size, e.Confidence, clean(e.Method))
	}
	return bw.Flush()
}

// WriteSpans writes the unnamed spans sorted by bundle and offset.
func WriteSpans(w io.Writer, ss []Span) error {
	ss = slices.Clone(ss)
	slices.SortFunc(ss, func(a, b Span) int {
		if c := strings.Compare(a.Bundle, b.Bundle); c != 0 {
			return c
		}
		return int(a.Offset - b.Offset)
	})
	bw := bufio.NewWriter(w)
	fmt.Fprintln(bw, spanHeader)
	for _, s := range ss {
		fmt.Fprintf(bw, "%s\t%d\t%d\t%s\n", s.Bundle, s.Offset, s.Size, clean(s.Note))
	}
	return bw.Flush()
}

func clean(s string) string { return strings.NewReplacer("\t", " ", "\n", " ").Replace(s) }

// ReadEntries reads what WriteEntries wrote.
func ReadEntries(r io.Reader) ([]Entry, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	var out []Entry
	line := 0
	for sc.Scan() {
		line++
		if line == 1 {
			if sc.Text() != entryHeader {
				return nil, fmt.Errorf("not a sidecar index: header %q", sc.Text())
			}
			continue
		}
		f := strings.Split(sc.Text(), "\t")
		if len(f) != 6 {
			return nil, fmt.Errorf("line %d: %d fields, want 6", line, len(f))
		}
		off, err1 := strconv.ParseInt(f[2], 10, 64)
		size, err2 := strconv.ParseInt(f[3], 10, 64)
		c, err3 := ParseConfidence(f[4])
		if err1 != nil || err2 != nil || err3 != nil || off < 0 || size < 0 {
			return nil, fmt.Errorf("line %d: bad offset, size or confidence", line)
		}
		out = append(out, Entry{Path: f[0], Bundle: f[1], Offset: off, Size: size, Confidence: c, Method: f[5]})
	}
	if line == 0 {
		return nil, fmt.Errorf("empty sidecar index")
	}
	return out, sc.Err()
}
