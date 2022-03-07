package util

import (
	"fmt"
	"strings"
)

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
// Line format is: <prefix><indent><text>
//
// Args:
// - txt -- the block of text to indent
// - prefixes -- strings that appears before the indent on each line.
//   - if length=0, prefix is empty
//   - if length=1, same prefix is used on every line
//   - if length=2, 1st line prefix and all other line prefix
//   - any other length is invalid
// - indents -- indent strings after prefix but before text.
//   - if length=0, indent is empty
//   - if length=1, same indent is used on every line
//   - if length=2, 1st line indent and all other line indent
//   - any other length is invalid
//
// Returns:
// - if error, string with format: ERROR: <error message>
// - indented string if no errors
func BlockIndent(txt string, prefixes []string, indents []string) string {
	// Exit quickly if input is empty.
	if txt == "" {
		return txt
	}

	// Verify prefixes.
	var prefix, otherPrefix string
	switch len(prefixes) {
	case 0:
		//	Empty prefix on each line.
	case 1:
		//	Same prefix on all lines.
		prefix = prefixes[0]
		otherPrefix = prefix
	case 2:
		//	Different prefix for 1st line.
		prefix = prefixes[0]
		otherPrefix = prefixes[1]
	default:
		return fmt.Sprintf("ERROR: invalid prefixes len=%d", len(prefixes))
	}

	// Verify indents.
	var indent, otherIndent string
	switch len(indents) {
	case 0:
		//	Empty indent on each line.
	case 1:
		//	Same indent on all lines.
		indent = indents[0]
		otherIndent = indent
	case 2:
		//	Different indent for 1st line.
		indent = indents[0]
		otherIndent = indents[1]
	default:
		return fmt.Sprintf("ERROR: invalid indents len=%d", len(indents))
	}

	// Iterate over each line of input text.
	lines := strings.Split(txt, "\n")
	out := make([]string, len(lines))

	for i, line := range lines {
		out[i] = prefix + indent + line

		prefix = otherPrefix
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
