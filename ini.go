// Package ini provides functions for parsing INI configuration files.
package ini

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	sectionRegex = regexp.MustCompile(`^\[(.*)\]$`)
	assignRegex  = regexp.MustCompile(`^([^=]+)=(.*)$`)
)

// ErrSyntax is returned when there is a syntax error in an INI file.
type ErrSyntax struct {
	Line   int
	Source string // erroneous line contents, without leading or trailing whitespace
}

func (e ErrSyntax) Error() string {
	return fmt.Sprintf("invalid INI syntax on line %d: %s", e.Line, e.Source)
}

// A File represents a parsed INI file.
type File map[string]Section

// A Section represents a single section of an INI file.
type Section map[string]string

// Section returns a named Section. A Section will be created if one does not
// already exist for the given name.
func (f File) Section(name string) Section {
	section := f[name]
	if section == nil {
		section = make(Section)
		f[name] = section
	}
	return section
}

// Get looks up a value for a key in a section and returns that value, along
// with a boolean result similar to a map lookup.
func (f File) Get(section, key string) (value string, ok bool) {
	if s := f[section]; s != nil {
		value, ok = s[key]
	}
	return
}

// Read loads a File from a Reader.
func Read(r io.Reader) (File, error) {
	f := make(File)
	bufin, ok := r.(*bufio.Reader)
	if !ok {
		bufin = bufio.NewReader(r)
	}
	err := parseFile(bufin, f)
	return f, err
}

// Load reads an INI File from a file on disk.
func Load(path string) (File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Read(f)
}

func parseFile(r *bufio.Reader, file File) (err error) {
	section := ""
	lineNum := 0
	for done := false; !done; {
		var line string
		if line, err = r.ReadString('\n'); err != nil {
			if err == io.EOF {
				done = true
			} else {
				return
			}
		}
		lineNum++
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			// Skip blank lines
			continue
		}
		if line[0] == ';' || line[0] == '#' {
			// Skip comments
			continue
		}

		if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), strings.TrimSpace(val)
			file.Section(section)[key] = val
		} else if groups := sectionRegex.FindStringSubmatch(line); groups != nil {
			name := strings.TrimSpace(groups[1])
			section = name
			// Create the section if it does not exist
			file.Section(section)
		} else {
			return ErrSyntax{lineNum, line}
		}

	}
	return nil
}

//func (f File) load(r io.Reader) (err error) {
//	bufin, ok := r.(*bufio.Reader)
//	if !ok {
//		bufin = bufio.NewReader(r)
//	}
//	return parseFile(bufin, f)
//}
