package reflector

import (
	"reflect"
)

var (
	// PrintNative includes native details in output if true.
	PrintNative = true

	// NoRefs expands TypeRefs in types if true.
	NoRefs = false

	// ParseAsJSON applies JSON parsing rules.
	// - Any lowercase element names are converted to 1st character uppercase so that they are exported.
	//   - Original lowercase name is saved as the native "json" Alias.
	ParseAsJSON = true
)

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
	r.typeResult.Types = r.reflectTypeImpl(parentID, "", r.typeResult.Types, NewAncestorList(), reflect.ValueOf(x), nil, nil)

	return r.typeResult
}
