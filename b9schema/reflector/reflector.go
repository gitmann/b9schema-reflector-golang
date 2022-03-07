package reflector

import (
	"reflect"
)

const (
	NATIVE_DIALECT = "golang"
)

var (
	// PrintNative includes native details in output if true.
	PrintNative = true

	// PathPrefix uses the path as the prefix for string output.
	PathPrefix = true

	// DeReference converts TypeRefs to their included types.
	// - If TyepRefs have a cyclical relationship, the last TypeRef is kept as a TypeRef.
	DeReference = false

	// ParseAsJSON applies JSON parsing rules.
	// - Any lowercase element names are converted to 1st character uppercase so that they are exported.
	//   - Original lowercase name is saved as the native "json" GetName.
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
	resetID()

	r.typeResult = &TypeResult{
		Root: NewRootElement(),
		Refs: NewTypeRefs(),
	}

	// Return *Reflector for chaining.
	return r
}

// ReflectTypes builds a reflector list of elements from the given interface.
func (r *Reflector) ReflectTypes(x interface{}) *TypeResult {
	if r.typeResult == nil {
		r.Reset()
	}

	// Start recursive reflection.
	r.reflectTypeImpl(NewAncestorTypeRef(), r.typeResult.Root.NewChild(""), reflect.ValueOf(x), nil)

	return r.typeResult
}
