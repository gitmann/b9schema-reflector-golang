package reflector

import (
	"reflect"
)

const (
	NATIVE_DIALECT = "golang"

	// Default options.
	PRINT_NATIVE  = true
	PATH_PREFIX   = true
	DEREFERENCE   = false
	PARSE_AS_JSON = true
)

type RenderOptions struct {
	// PrintNative includes native details in output if true.
	PrintNative bool

	// PathPrefix uses the path as the prefix for string output.
	PathPrefix bool

	// DeReference converts TypeRefs to their included types.
	// - If TyepRefs have a cyclical relationship, the last TypeRef is kept as a TypeRef.
	DeReference bool

	// ParseAsJSON applies JSON parsing rules.
	// - Any lowercase element names are converted to 1st character uppercase so that they are exported.
	//   - Original lowercase name is saved as the native "json" GetName.
	ParseAsJSON bool
}

func NewRenderOptions() *RenderOptions {
	opt := &RenderOptions{
		PrintNative: PRINT_NATIVE,
		PathPrefix:  PATH_PREFIX,
		DeReference: DEREFERENCE,
		ParseAsJSON: PARSE_AS_JSON,
	}
	return opt
}

// Reflector provides functions to build type and values from a Go value.
type Reflector struct {
	// Label is an optional label for a block of elements.
	Label string

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
		Root:     NewRootElement("Root"),
		TypeRefs: NewRootElement("TypeRefs"),
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
