package reflector

import (
	"reflect"
	"strconv"
	"strings"
)

// StructFieldTag stores attributes of a struct field tag.
//
// Tags are parsed as follows:
// - tag="-" --> ignored field, Ignore=true
// - tag="someString" --> alias only, Alias = "someString"
// - tag="someString,options" --> alias with options, Alias="someString", Options=remainder after the first comma
// - If tag = "-", Ignore is true -->
type StructFieldTag struct {
	Ignore  bool
	Alias   string
	Options *NativeOption
}

// NewStructFieldTag parses the contents of tag string to initialize a StructFieldTag.
// - Reference for common tags: https://zchee.github.io/golang-wiki/Well-known-struct-tags/
//
// Tags alwyas follow the pattern: <alias>,<comma-delimited options>
// - if tag string is "-", field is ignored
// - either <alias> or <options> can be omitted
// - if <options> is empty, the comma may be omitted
func NewStructFieldTag(tag string) *StructFieldTag {
	t := &StructFieldTag{
		Options: NewNativeOption(),
	}

	if s, err := strconv.Unquote(tag); err == nil {
		tag = strings.TrimSpace(s)
	}

	if tag == "" {
		// Empty tag.
		return nil
	}

	var rawOptions string
	if tag == "-" {
		// Ignored field.
		t.Ignore = true
	} else if strings.Contains(tag, ",") {
		//	GetName with options.
		tokens := strings.SplitN(tag, ",", 2)

		t.Alias = strings.TrimSpace(tokens[0])
		rawOptions = strings.TrimSpace(tokens[1])
	} else {
		// Just an alias.
		t.Alias = tag
	}

	if rawOptions != "" {
		// The raw option string is a comma-delimited list of option values.
		for _, opt := range strings.Split(rawOptions, ",") {
			opt = strings.TrimSpace(opt)
			if opt != "" {
				tokens := strings.SplitN(opt, "=", 2)
				if len(tokens) > 0 {
					var key, val string
					key = strings.TrimSpace(tokens[0])
					if len(tokens) > 1 {
						val = strings.TrimSpace(tokens[1])
					}

					if val != "" {
						t.Options.AddKeyVal(key, val)
					} else {
						t.Options.AddVal(key)
					}
				}
			}
		}
	}

	return t
}

// Equals returns true if two StructFieldTag structs have the same values.
func (s *StructFieldTag) Equals(other *StructFieldTag) bool {
	if s == nil && other == nil {
		// Both are nil so consider them equal.
		return true
	} else if s == nil || other == nil {
		// One is nil so these are different.
		return false
	}

	if s.Ignore != other.Ignore {
		return false
	}
	if s.Alias != other.Alias {
		return false
	}

	return s.Options.Equals(other.Options)
}

// Tags stores struct tags by tag name.
type Tags map[string]*StructFieldTag

// ParseTags
// Parsing code is derived from: go/src/reflect/type.go --> Lookup()
func ParseTags(tag reflect.StructTag) Tags {
	tags := make(Tags)

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		tags[name] = NewStructFieldTag(qvalue)
	}
	return tags
}
