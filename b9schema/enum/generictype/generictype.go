package generictype

import (
	"fmt"
	"reflect"

	"github.com/gitmann/shiny-reflector-golang/b9schema/enum/typecategory"
)

// GenericType defines generic types for shiny schemas.
// Uses slugs from: https://threedots.tech/post/safer-enums-in-go/
type GenericType struct {
	slug        string
	pathDefault string
	cat         typecategory.TypeCategory
	kinds       []string
}

// String returns GenericType as a string.
func (t *GenericType) String() string {
	return t.slug
}

// Category returns the TypeCategory for the GenericType.
func (t *GenericType) Category() typecategory.TypeCategory {
	return t.cat
}

// PathDefault returns the default path string for the GenericType.
func (t *GenericType) PathDefault() string {
	if t.pathDefault != "" {
		return t.pathDefault
	}
	return t.slug
}

// Invalid types are not allowed in shiny schemas.
var Invalid = &GenericType{
	slug: "invalid",
	cat:  typecategory.Invalid,
	kinds: []string{
		reflect.Invalid.String(),
		reflect.Complex64.String(),
		reflect.Complex128.String(),
		reflect.Chan.String(),
		reflect.Func.String(),
		reflect.UnsafePointer.String(),
	},
}

// Basic types.
var Boolean = &GenericType{
	slug: "boolean",
	cat:  typecategory.Basic,
	kinds: []string{
		reflect.Bool.String(),
	},
}

var Integer = &GenericType{
	slug: "integer",
	cat:  typecategory.Basic,
	kinds: []string{
		reflect.Int.String(),
		reflect.Int8.String(),
		reflect.Int16.String(),
		reflect.Int32.String(),
		reflect.Int64.String(),
		reflect.Uint.String(),
		reflect.Uint8.String(),
		reflect.Uint16.String(),
		reflect.Uint32.String(),
		reflect.Uint64.String(),
		reflect.Uintptr.String(),
	},
}

var Float = &GenericType{
	slug: "float",
	cat:  typecategory.Basic,
	kinds: []string{
		reflect.Float32.String(),
		reflect.Float64.String(),
	},
}

var String = &GenericType{
	slug: "string",
	cat:  typecategory.Basic,
	kinds: []string{
		reflect.String.String(),
	},
}

// Compound types.
var List = &GenericType{
	slug:        "list",
	pathDefault: "[]",
	cat:         typecategory.Compound,
	kinds: []string{
		reflect.Array.String(),
		reflect.Slice.String(),
	},
}

var Struct = &GenericType{
	slug:        "struct",
	pathDefault: "{}",
	cat:         typecategory.Compound,
	kinds: []string{
		reflect.Map.String(),
		reflect.Struct.String(),
	},
}

// Known types map Go standard types to b9schema types.
// - kinds is a list of "PkgPath.Type"
// These are a subset of protobuf well-known types:
// https://developers.google.com/protocol-buffers/docs/reference/google.protobuf

var DateTime = &GenericType{
	slug: "datetime",
	cat:  typecategory.Known,
	kinds: []string{
		"time.Time",
	},
}

// Reference types.
var Interface = &GenericType{
	slug:        "interface",
	pathDefault: "{?}",
	cat:         typecategory.Reference,
	kinds: []string{
		reflect.Interface.String(),
	},
}

var Pointer = &GenericType{
	slug:        "pointer",
	pathDefault: "*",
	cat:         typecategory.Reference,
	kinds: []string{
		reflect.Ptr.String(),
	},
}

// Internal types.
// These have no meaning outside of a b9schema.

// Root is at the top of any type tree.
var Root = &GenericType{
	slug:        "root",
	pathDefault: "$",
	cat:         typecategory.Internal,
	kinds:       []string{},
}

// genericTypeLookup provides fast mapping from reflect.Kind to GenericType.
var genericTypeLookup map[string]*GenericType

// init() initializes the genericTypeLookup map.
func init() {
	genericTypeLookup = map[string]*GenericType{}
	pathDefaultLookup = map[string]string{}

	// mapTypes is a utility function to create map entries for the given GenericType.
	mapTypes := func(t *GenericType) {
		for _, k := range t.kinds {
			// Panic if duplicate type mappings exist.
			if genericTypeLookup[k] != nil {
				panic(fmt.Sprintf("duplicate GenericType mapping for %q", t))
			}
			genericTypeLookup[k] = t
		}

		pathDefaultLookup[t.String()] = t.PathDefault()
	}

	mapTypes(Invalid)

	mapTypes(Boolean)
	mapTypes(Integer)
	mapTypes(Float)
	mapTypes(String)

	mapTypes(List)
	mapTypes(Struct)

	mapTypes(DateTime)

	mapTypes(Interface)
	mapTypes(Pointer)
}

// GenericTypeOf returns the GenericType of the given reflect.Value.
func GenericTypeOf(v reflect.Value) *GenericType {
	if t := genericTypeLookup[v.Kind().String()]; t != nil {
		if t == Invalid {
			// Return invalid types immediately.
			return t
		}

		// Look for special types.
		if v.Type().PkgPath() != "" {
			fullPath := fmt.Sprintf("%s.%s", v.Type().PkgPath(), v.Type().Name())
			if specialType := genericTypeLookup[fullPath]; specialType != nil {
				return specialType
			}
		}

		// Not a special type.
		return t
	}

	return Invalid
}

// pathDefaultLookup provides fast mapping from genericType.String to pathDefault
var pathDefaultLookup map[string]string

// PathDefaultOfType returns the path default for a given generic type string.
func PathDefaultOfType(typeString string) string {
	if p := pathDefaultLookup[typeString]; p != "" {
		return p
	}
	return Invalid.PathDefault()
}
