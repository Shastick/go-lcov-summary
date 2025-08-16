// Package lcov provides functionality to parse and summarize LCOV files.
package lcov

import (
	"io"
)

// Summarize processes LCOV data from an io.Reader and returns summary information.
// This function is the main public API for the lcov package.
func Summarize(reader io.Reader) (*Summary, error) {
	parser := NewParser(reader)
	return parser.Parse()
}
