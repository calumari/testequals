package testequals

import (
	"fmt"
	"strings"
)

// MismatchError describes a single comparison failure. Path segments are
// formatted using dot notation for object keys and [index] for array indices
// (e.g. .user.address[0].city). Message holds a human‑readable description.
type MismatchError struct {
	Path    []string
	Message string
}

func (e *MismatchError) Error() string {
	if len(e.Path) == 0 {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", strings.Join(e.Path, ""), e.Message)
}

// MultiError aggregates multiple mismatches produced when CollectAll is
// enabled. It implements error and unwraps to the first mismatch for
// compatibility with errors.Is / errors.As.
type MultiError struct{ Mismatches []*MismatchError }

func (e *MultiError) Error() string {
	switch len(e.Mismatches) {
	case 0:
		return "no mismatches"
	case 1:
		return e.Mismatches[0].Error()
	default:
		return fmt.Sprintf("%d mismatches (first: %s)", len(e.Mismatches), e.Mismatches[0].Error())
	}
}

func (e *MultiError) Unwrap() error {
	if len(e.Mismatches) == 0 {
		return nil
	}
	return e.Mismatches[0]
}

func mismatch(path []string, msg string) *MismatchError {
	return &MismatchError{Path: append([]string{}, path...), Message: msg}
}

func keySeg(k string) string {
	return "." + k
}

func indexSeg(i int) string {
	return fmt.Sprintf("[%d]", i)
}
