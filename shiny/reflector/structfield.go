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
	Options string

	// If parsing fails, Raw holds the original string.
	Raw string
}

// NewStructFieldTag parses the contents of tag string to initialize a StructFieldTag.
func NewStructFieldTag(tag string) *StructFieldTag {
	t := &StructFieldTag{}

	if s, err := strconv.Unquote(tag); err == nil {
		tag = strings.TrimSpace(s)
	}

	if tag == "" {
		// Empty tag.
		return nil
	}

	if tag == "-" {
		// Ignored field.
		t.Ignore = true
	} else if strings.Contains(tag, ",") {
		//	Alias with options.
		tokens := strings.SplitN(tag, ",", 2)

		t.Alias = strings.TrimSpace(tokens[0])
		t.Options = strings.TrimSpace(tokens[1])
	} else {
		// Just an alias.
		t.Alias = tag
	}

	return t
}

// AsMap renders the StructFieldTag as a map of strings.
func (t *StructFieldTag) AsMap() map[string]string {
	m := map[string]string{}
	if t.Ignore {
		m["Ignore"] = "true"
	}
	if t.Alias != "" {
		m["Alias"] = t.Alias
	}
	if t.Options != "" {
		m["Options"] = t.Options
	}
	if t.Raw != "" {
		m["Raw"] = t.Raw
	}
	if len(m) > 0 {
		return m
	}
	return nil
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
