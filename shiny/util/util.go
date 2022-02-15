package util

import "strings"

// ValueIfTrue converts a boolean into strings for true and false.
// - if boolTest is true, return trueValue else return falseValue
func ValueIfTrue(boolTest bool, trueValue, falseValue string) string {
	if boolTest {
		return trueValue
	}
	return falseValue
}

// BlockIndent indents each line in a block of text.
//
// Args:
// - txt -- the block of text to indent
// - prefixes -- list of prefixes
//   - if prefixes is length 1, all lines use the same prefix
//   - if prefixes is length 2, first line uses prefix[0] and all other lines use prefix[1]
//   - any other length for prefixes is invalid
func BlockIndent(txt string, prefixes []string) string {
	var firstIndent, otherIndent string
	switch len(prefixes) {
	case 1:
		firstIndent = prefixes[0]
		otherIndent = firstIndent
	case 2:
		firstIndent = prefixes[0]
		otherIndent = prefixes[1]
	default:
		// Invalid prefixes length.
		return ""
	}

	indent := firstIndent

	lines := strings.Split(txt, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = indent + line
		indent = otherIndent
	}
	return strings.Join(out, "\n")
}

// Capitalize returns string with its first letter in uppercase.
func Capitalize(s string) string {
	// Capitalize entire string if it is short.
	if len(s) < 2 {
		return strings.ToUpper(s)
	}

	// Build a new string with 1st letter uppercase.
	return strings.ToUpper(s[0:1]) + s[1:]
}
