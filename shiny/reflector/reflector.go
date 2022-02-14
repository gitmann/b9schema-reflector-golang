package reflector

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const (
	PRINT_NATIVE = true
)

// TypeElement holds type information about an element.
type TypeElement struct {
	// Unique identifier for an element.
	ID int

	// Identifier for parent of an element.
	ParentID int

	// Optional Name and Description of element.
	// - Name applies to struct/map types with string keys.
	Name        string
	Description string

	// Label is an optional label for a block of elements.
	Label string

	// Generic type of element.
	Type string

	// TypeRef holds the name of a type (e.g. struct)
	TypeRef string

	// Native type features by language or implementation name.
	Native map[string]NativeType

	// Types of child elements.
	// - KeyType is only applicable to map types.
	ValueTypes []*TypeElement

	// Capture error if element cannot reflect.
	Err error
}

func NewTypeElement(ID, ParentID int, name, label string) *TypeElement {
	return &TypeElement{
		ID:       ID,
		ParentID: ParentID,
		Name:     name,
		Label:    label,
		Native:   make(map[string]NativeType),
	}
}

// String builds a string representation of the TypeElement.
// Default is to build a CSV representation:
// <id>,<parentID>,<type>,<name>,<err>
//
// Implementation specifics are output on multiple, indented lines in JSON format.
func (t *TypeElement) String() string {
	// typeString is the type followed by optional type ref.
	typeString := t.Type
	if t.TypeRef != "" {
		typeString += ":" + t.TypeRef
	}

	out := fmt.Sprintf("%d,%d,%q,%s", t.ID, t.ParentID, t.Name, typeString)
	if t.Err != nil {
		out += "," + t.Err.Error()
	} else if len(t.Native) > 0 {
		if PRINT_NATIVE {
			nativeLines := []string{out}

			// Print native fields with fixed length for keys.
			for language, features := range t.Native {
				nativeLines = append(nativeLines, fmt.Sprintf("  Native: %q", language))

				// Collect key/value pairs and max key length
				nativeKeyVal := [][]string{}
				keyLen := 0

				for k, v := range features {
					if v == "" {
						// Skip empty values.
						continue
					}

					if len(k) > keyLen {
						keyLen = len(k)
					}
					nativeKeyVal = append(nativeKeyVal, []string{k, v})
				}

				// Sort by key name. This sorts uppercase before lowercase.
				sort.Slice(nativeKeyVal, func(i, j int) bool { return nativeKeyVal[i][0] < nativeKeyVal[j][0] })

				// Add lines using key length.
				for _, line := range nativeKeyVal {
					newLine := fmt.Sprintf("    %-*s: %s", keyLen, line[0], line[1])
					nativeLines = append(nativeLines, newLine)
				}
			}

			// Construct output string.
			out = strings.Join(nativeLines, "\n")
		}
	}

	return out
}

// NativeType holds key-value attributes specific to one language or implementation.
type NativeType map[string]string

// TypeList holds a slice of TypeElements.
type TypeList []*TypeElement

// TypeResult is the result of parsing types.
type TypeResult struct {
	// Types is a list of types in the order found.
	Types TypeList

	// TypeRefs holds a map of named types by name.
	TypeRefs map[string]TypeList
}

// SortedTypeNames returns an alphabetically sorted list of type names.
func (t *TypeResult) SortedTypeNames() []string {
	names := make([]string, len(t.TypeRefs))

	i := 0
	for k, _ := range t.TypeRefs {
		names[i] = k
		i++
	}
	sort.Strings(names)
	return names
}

// Reflector provides functions to build type and values from a Go value.
type Reflector struct {
	// Label is an optional label for a block of elements.
	Label string

	// Keep track of the last ID assigned.
	lastID int

	// Keep track of refs found during parsing.
	typeResult *TypeResult
}

func NewReflector() *Reflector {
	r := &Reflector{}
	r.Reset()

	return r
}

func (r *Reflector) Reset() *Reflector {
	// Initialize state.
	r.lastID = 0

	r.typeResult = &TypeResult{
		Types:    make(TypeList, 0),
		TypeRefs: make(map[string]TypeList),
	}

	// Return *Reflector for chaining.
	return r
}

func (r *Reflector) nextID() int {
	r.lastID++
	return r.lastID
}

// ReflectTypes builds a reflector list of elements from the given interface.
func (r *Reflector) ReflectTypes(x interface{}) *TypeResult {
	if r.typeResult == nil {
		r.Reset()
	}

	// Reset parentID to root.
	parentID := 0

	// Start recursive reflection.
	r.typeResult.Types = r.reflectTypeImpl(parentID, "", r.typeResult.Types, reflect.ValueOf(x), nil)

	return r.typeResult
}
