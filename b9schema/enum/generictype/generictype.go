package generictype

import (
	"fmt"
	"github.com/gitmann/shiny-reflector-golang/b9schema/enum/typecategory"
	"reflect"
)

// GenericType defines generic types for shiny schemas.
// Uses slugs from: https://threedots.tech/post/safer-enums-in-go/
type GenericType struct {
	slug        string
	pathDefault string
	cat         typecategory.TypeCategory
	kinds       map[reflect.Kind]interface{}
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
	slug:        "invalid",
	pathDefault: "!invalid!",
	cat:         typecategory.Invalid,
	kinds: map[reflect.Kind]interface{}{
		reflect.Invalid:       nil,
		reflect.Complex64:     nil,
		reflect.Complex128:    nil,
		reflect.Chan:          nil,
		reflect.Func:          nil,
		reflect.UnsafePointer: nil,
	},
}

// Basic types.
var Boolean = &GenericType{
	slug: "boolean",
	cat:  typecategory.Basic,
	kinds: map[reflect.Kind]interface{}{
		reflect.Bool: nil,
	},
}

var Integer = &GenericType{
	slug: "integer",
	cat:  typecategory.Basic,
	kinds: map[reflect.Kind]interface{}{
		reflect.Int:     nil,
		reflect.Int8:    nil,
		reflect.Int16:   nil,
		reflect.Int32:   nil,
		reflect.Int64:   nil,
		reflect.Uint:    nil,
		reflect.Uint8:   nil,
		reflect.Uint16:  nil,
		reflect.Uint32:  nil,
		reflect.Uint64:  nil,
		reflect.Uintptr: nil,
	},
}

var Float = &GenericType{
	slug: "float",
	cat:  typecategory.Basic,
	kinds: map[reflect.Kind]interface{}{
		reflect.Float32: nil,
		reflect.Float64: nil,
	},
}

var String = &GenericType{
	slug: "string",
	cat:  typecategory.Basic,
	kinds: map[reflect.Kind]interface{}{
		reflect.String: nil,
	},
}

// Compound types.
var List = &GenericType{
	slug: "list",
	cat:  typecategory.Compound,
	kinds: map[reflect.Kind]interface{}{
		reflect.Array: nil,
		reflect.Slice: nil,
	},
}

var Struct = &GenericType{
	slug: "struct",
	cat:  typecategory.Compound,
	kinds: map[reflect.Kind]interface{}{
		reflect.Map:    nil,
		reflect.Struct: nil,
	},
}

// Other types.
var Interface = &GenericType{
	slug: "interface",
	cat:  typecategory.Interface,
	kinds: map[reflect.Kind]interface{}{
		reflect.Interface: nil,
	},
}

var Pointer = &GenericType{
	slug: "pointer",
	cat:  typecategory.Pointer,
	kinds: map[reflect.Kind]interface{}{
		reflect.Ptr: nil,
	},
}

// genericTypeLookup provides fast mapping from reflect.Kind to GenericType.
var genericTypeLookup map[reflect.Kind]*GenericType

// init() initializes the genericTypeLookup map.
func init() {
	genericTypeLookup = map[reflect.Kind]*GenericType{}
	pathDefaultLookup = map[string]string{}

	// mapTypes is a utility function to create map entries for the given GenericType.
	mapTypes := func(t *GenericType) {
		for k, _ := range t.kinds {
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

	mapTypes(Interface)
	mapTypes(Pointer)
}

// GenericTypeOf returns the GenericType of the given reflect.Value.
func GenericTypeOf(v reflect.Value) *GenericType {
	if t := genericTypeLookup[v.Kind()]; t != nil {
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
