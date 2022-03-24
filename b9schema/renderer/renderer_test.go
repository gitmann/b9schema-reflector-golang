package renderer

import (
	"encoding/json"
	"fmt"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/reflector"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"
)

var allTests = [][]TestCase{
	rootJSONTests,
	rootGoTests,
	typeTests,
	listTests,
	compoundTests,
	referenceTests,
	cycleTests,
	jsonTagTests,

	// structTests,
	// pointerTests,
}

type TestCase struct {
	name  string
	value interface{}

	// Expected strings for reference and de-reference.
	refStrings     []string
	derefStrings   []string
	jsonStrings    []string
	openapiStrings []string
}

// *** All reflect types ***

// rootTests validate that the top-level element is either a struct or Reference.
var rootJSONTests = []TestCase{
	{
		name:         "json-null",
		value:        fromJSON([]byte(`null`)),
		refStrings:   []string{"Root.!invalid:nil! ERROR:kind not supported"},
		derefStrings: []string{"Root.!invalid:nil! ERROR:kind not supported"},
	},
	{
		name:         "json-string",
		value:        fromJSON([]byte(`"Hello"`)),
		refStrings:   []string{"Root.!string! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.!string! ERROR:root type must be a struct"},
	},
	{
		name:         "json-int",
		value:        fromJSON([]byte(`123`)),
		refStrings:   []string{"Root.!float! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.!float! ERROR:root type must be a struct"},
	},
	{
		name:         "json-float",
		value:        fromJSON([]byte(`234.345`)),
		refStrings:   []string{"Root.!float! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.!float! ERROR:root type must be a struct"},
	},
	{
		name:         "json-bool",
		value:        fromJSON([]byte(`true`)),
		refStrings:   []string{"Root.!boolean! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.!boolean! ERROR:root type must be a struct"},
	},
	{
		name:         "json-list-empty",
		value:        fromJSON([]byte(`[]`)),
		refStrings:   []string{"Root.![]! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.![]! ERROR:root type must be a struct"},
	},
	{
		name:         "json-list",
		value:        fromJSON([]byte(`[1,2,3]`)),
		refStrings:   []string{"Root.![]! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.![]! ERROR:root type must be a struct"},
	},
	{
		name:         "json-object-empty",
		value:        fromJSON([]byte(`{}`)),
		refStrings:   []string{"Root.!{}! ERROR:empty map not supported"},
		derefStrings: []string{"Root.!{}! ERROR:empty map not supported"},
	},
	{
		name:  "json-object",
		value: fromJSON([]byte(`{"key1":"Hello"}`)),
		refStrings: []string{
			"Root.{}",
			"Root.{}.Key1:string",
		},
		derefStrings: []string{
			"Root.{}",
			"Root.{}.Key1:string",
		},
	},
}

var rootGoTests = []TestCase{
	{
		name:         "golang-nil",
		value:        nil,
		refStrings:   []string{"Root.!invalid:nil! ERROR:kind not supported"},
		derefStrings: []string{"Root.!invalid:nil! ERROR:kind not supported"},
	},
	{
		name:         "golang-string",
		value:        "Hello",
		refStrings:   []string{"Root.!string! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.!string! ERROR:root type must be a struct"},
	},
	{
		name:         "golang-int",
		value:        123,
		refStrings:   []string{"Root.!integer! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.!integer! ERROR:root type must be a struct"},
	},
	{
		name:         "golang-float",
		value:        234.345,
		refStrings:   []string{"Root.!float! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.!float! ERROR:root type must be a struct"},
	},
	{
		name:         "golang-bool",
		value:        true,
		refStrings:   []string{"Root.!boolean! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.!boolean! ERROR:root type must be a struct"},
	},
	{
		name:         "golang-array-0",
		value:        [0]string{},
		refStrings:   []string{"Root.![]! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.![]! ERROR:root type must be a struct"},
	},
	{
		name:         "golang-array-3",
		value:        [3]string{},
		refStrings:   []string{"Root.![]! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.![]! ERROR:root type must be a struct"},
	},
	{
		name:         "golang-slice-nil",
		value:        func() interface{} { var s []string; return s }(),
		refStrings:   []string{"Root.![]! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.![]! ERROR:root type must be a struct"},
	},
	{
		name:         "golang-slice-0",
		value:        []string{},
		refStrings:   []string{"Root.![]! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.![]! ERROR:root type must be a struct"},
	},
	{
		name:         "golang-slice-3",
		value:        make([]string, 3),
		refStrings:   []string{"Root.![]! ERROR:root type must be a struct"},
		derefStrings: []string{"Root.![]! ERROR:root type must be a struct"},
	},
	{
		name: "golang-struct-empty", value: func() interface{} { var s struct{}; return s }(),
		refStrings:   []string{"Root.!{}! ERROR:empty struct not supported"},
		derefStrings: []string{"Root.!{}! ERROR:empty struct not supported"},
	},
	{
		name:  "golang-struct-noinit",
		value: func() interface{} { var s StringStruct; return s }(),
		refStrings: []string{
			`TypeRefs.StringStruct:{}`,
			`TypeRefs.StringStruct:{}.Value:string`,
			`Root.{}:StringStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Value:string`,
		},
	},
	{
		name:  "golang-struct-init",
		value: StringStruct{},
		refStrings: []string{
			`TypeRefs.StringStruct:{}`,
			`TypeRefs.StringStruct:{}.Value:string`,
			`Root.{}:StringStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Value:string`,
		},
	},
	{
		name:  "golang-struct-private",
		value: PrivateStruct{},
		refStrings: []string{
			`TypeRefs.!PrivateStruct:{}! ERROR:struct has no exported fields`,
			`Root.!{}:PrivateStruct! ERROR:struct has no exported fields`,
		},
		derefStrings: []string{
			`Root.!{}! ERROR:struct has no exported fields`,
		},
	},

	{
		name:  "golang-interface-struct-noinit",
		value: func() interface{} { var s interface{} = StringStruct{}; return s }(),
		refStrings: []string{
			`TypeRefs.StringStruct:{}`,
			`TypeRefs.StringStruct:{}.Value:string`,
			`Root.{}:StringStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Value:string`,
		},
	},
	{
		name:  "golang-pointer-struct-noinit",
		value: func() interface{} { var s *StringStruct; return s }(),
		refStrings: []string{
			`TypeRefs.StringStruct:{}`,
			`TypeRefs.StringStruct:{}.Value:string`,
			`Root.{}:StringStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Value:string`,
		},
	},
	{
		name:  "golang-pointer-struct-init",
		value: &StringStruct{},
		refStrings: []string{
			`TypeRefs.StringStruct:{}`,
			`TypeRefs.StringStruct:{}.Value:string`,
			`Root.{}:StringStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Value:string`,
		},
	},
}

// Check all types from reflect package.
type BoolTypes struct {
	Bool bool
}

type IntegerTypes struct {
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Uintptr uintptr
}

type FloatTypes struct {
	Float32 float32
	Float64 float64
}

type StringTypes struct {
	String string
}

type InvalidTypes struct {
	Complex64  complex64
	Complex128 complex128

	Chan          chan int
	Func          func()
	UnsafePointer unsafe.Pointer
}

type CompoundTypes struct {
	Array0 [0]string
	Array3 [3]string

	Interface  interface{}
	Map        map[int]int
	Ptr        *StringStruct
	PrivatePtr *PrivateStruct
	Slice      []interface{}
	Struct     struct{}
}

// Special types from protobuf: https://developers.google.com/protocol-buffers/docs/reference/google.protobuf
type SpecialTypes struct {
	DateTime time.Time
}

var typeTests = []TestCase{
	{
		name:  "boolean",
		value: BoolTypes{},
		refStrings: []string{
			`TypeRefs.BoolTypes:{}`,
			`TypeRefs.BoolTypes:{}.Bool:boolean`,
			`Root.{}:BoolTypes`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Bool:boolean`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  BoolTypes:`,
			`    type: object`,
			`    properties:`,
			`      Bool:`,
			`        type: boolean`,
			`root:`,
			`  $ref: '#/definitions/BoolTypes'`,
		},
	},
	{
		name:  "integer",
		value: IntegerTypes{},
		refStrings: []string{
			`TypeRefs.IntegerTypes:{}`,
			`TypeRefs.IntegerTypes:{}.Int:integer`,
			`TypeRefs.IntegerTypes:{}.Int16:integer`,
			`TypeRefs.IntegerTypes:{}.Int32:integer`,
			`TypeRefs.IntegerTypes:{}.Int64:integer`,
			`TypeRefs.IntegerTypes:{}.Int8:integer`,
			`TypeRefs.IntegerTypes:{}.Uint:integer`,
			`TypeRefs.IntegerTypes:{}.Uint16:integer`,
			`TypeRefs.IntegerTypes:{}.Uint32:integer`,
			`TypeRefs.IntegerTypes:{}.Uint64:integer`,
			`TypeRefs.IntegerTypes:{}.Uint8:integer`,
			`TypeRefs.IntegerTypes:{}.Uintptr:integer`,
			`Root.{}:IntegerTypes`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Int:integer`,
			`Root.{}.Int16:integer`,
			`Root.{}.Int32:integer`,
			`Root.{}.Int64:integer`,
			`Root.{}.Int8:integer`,
			`Root.{}.Uint:integer`,
			`Root.{}.Uint16:integer`,
			`Root.{}.Uint32:integer`,
			`Root.{}.Uint64:integer`,
			`Root.{}.Uint8:integer`,
			`Root.{}.Uintptr:integer`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  IntegerTypes:`,
			`    type: object`,
			`    properties:`,
			`      Int:`,
			`        type: integer`,
			`      Int16:`,
			`        type: integer`,
			`      Int32:`,
			`        type: integer`,
			`      Int64:`,
			`        type: integer`,
			`        format: int64`,
			`      Int8:`,
			`        type: integer`,
			`      Uint:`,
			`        type: integer`,
			`      Uint16:`,
			`        type: integer`,
			`      Uint32:`,
			`        type: integer`,
			`      Uint64:`,
			`        type: integer`,
			`        format: int64`,
			`      Uint8:`,
			`        type: integer`,
			`      Uintptr:`,
			`        type: integer`,
			`root:`,
			`  $ref: '#/definitions/IntegerTypes'`,
		},
	},
	{
		name:  `float`,
		value: FloatTypes{},
		refStrings: []string{
			`TypeRefs.FloatTypes:{}`,
			`TypeRefs.FloatTypes:{}.Float32:float`,
			`TypeRefs.FloatTypes:{}.Float64:float`,
			`Root.{}:FloatTypes`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Float32:float`,
			`Root.{}.Float64:float`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  FloatTypes:`,
			`    type: object`,
			`    properties:`,
			`      Float32:`,
			`        type: number`,
			`      Float64:`,
			`        type: number`,
			`        format: double`,
			`root:`,
			`  $ref: '#/definitions/FloatTypes'`,
		},
	},
	{
		name:  "string",
		value: StringTypes{},
		refStrings: []string{
			`TypeRefs.StringTypes:{}`,
			`TypeRefs.StringTypes:{}.String:string`,
			`Root.{}:StringTypes`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.String:string`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  StringTypes:`,
			`    type: object`,
			`    properties:`,
			`      String:`,
			`        type: string`,
			`root:`,
			`  $ref: '#/definitions/StringTypes'`,
		},
	},
	{
		name:  "invalid",
		value: InvalidTypes{},
		refStrings: []string{
			`TypeRefs.InvalidTypes:{}`,
			`TypeRefs.InvalidTypes:{}.!Chan:invalid:chan! ERROR:kind not supported`,
			`TypeRefs.InvalidTypes:{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
			`TypeRefs.InvalidTypes:{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
			`TypeRefs.InvalidTypes:{}.!Func:invalid:func! ERROR:kind not supported`,
			`TypeRefs.InvalidTypes:{}."!UnsafePointer:invalid:unsafe.Pointer!" ERROR:kind not supported`,
			`Root.{}:InvalidTypes`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.!Chan:invalid:chan! ERROR:kind not supported`,
			`Root.{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
			`Root.{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
			`Root.{}.!Func:invalid:func! ERROR:kind not supported`,
			`Root.{}."!UnsafePointer:invalid:unsafe.Pointer!" ERROR:kind not supported`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  InvalidTypes:`,
			`    type: object`,
			`    properties:`,
			`      Chan:`,
			`        type: invalid:chan`,
			`        error: kind not supported`,
			`      Complex128:`,
			`        type: invalid:complex128`,
			`        error: kind not supported`,
			`      Complex64:`,
			`        type: invalid:complex64`,
			`        error: kind not supported`,
			`      Func:`,
			`        type: invalid:func`,
			`        error: kind not supported`,
			`      UnsafePointer:`,
			`        type: invalid:unsafe.Pointer`,
			`        error: kind not supported`,
			`root:`,
			`  $ref: '#/definitions/InvalidTypes'`,
		},
	},
	{
		name:  "compound",
		value: CompoundTypes{},
		refStrings: []string{
			`TypeRefs.CompoundTypes:{}`,
			`TypeRefs.CompoundTypes:{}.Array0:[]`,
			`TypeRefs.CompoundTypes:{}.Array0:[].string`,
			`TypeRefs.CompoundTypes:{}.Array3:[]`,
			`TypeRefs.CompoundTypes:{}.Array3:[].string`,
			`TypeRefs.CompoundTypes:{}.!Interface:invalid! ERROR:interface element is nil`,
			`TypeRefs.CompoundTypes:{}.!Map:{}! ERROR:map key type must be string`,
			`TypeRefs.CompoundTypes:{}.PrivatePtr:{}:PrivateStruct`,
			`TypeRefs.CompoundTypes:{}.Ptr:{}:StringStruct`,
			`TypeRefs.CompoundTypes:{}.Slice:[]`,
			`TypeRefs.CompoundTypes:{}.Slice:[].!invalid! ERROR:interface element is nil`,
			`TypeRefs.CompoundTypes:{}.!Struct:{}! ERROR:empty struct not supported`,
			`TypeRefs.!PrivateStruct:{}! ERROR:struct has no exported fields`,
			`TypeRefs.StringStruct:{}`,
			`TypeRefs.StringStruct:{}.Value:string`,
			`Root.{}:CompoundTypes`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Array0:[]`,
			`Root.{}.Array0:[].string`,
			`Root.{}.Array3:[]`,
			`Root.{}.Array3:[].string`,
			`Root.{}.!Interface:invalid! ERROR:interface element is nil`,
			`Root.{}.!Map:{}! ERROR:map key type must be string`,
			`Root.{}.!PrivatePtr:{}! ERROR:struct has no exported fields`,
			`Root.{}.Ptr:{}`,
			`Root.{}.Ptr:{}.Value:string`,
			`Root.{}.Slice:[]`,
			`Root.{}.Slice:[].!invalid! ERROR:interface element is nil`,
			`Root.{}.!Struct:{}! ERROR:empty struct not supported`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  CompoundTypes:`,
			`    type: object`,
			`    properties:`,
			`      Array0:`,
			`        type: array`,
			`        items:`,
			`          type: string`,
			`      Array3:`,
			`        type: array`,
			`        items:`,
			`          type: string`,
			`      Interface:`,
			`        type: invalid`,
			`        error: interface element is nil`,
			`      Map:`,
			`        type: object`,
			`        properties:`,
			`          error: map key type must be string`,
			`      PrivatePtr:`,
			`        $ref: '#/definitions/PrivateStruct'`,
			`      Ptr:`,
			`        $ref: '#/definitions/StringStruct'`,
			`      Slice:`,
			`        type: array`,
			`        items:`,
			`          type: invalid`,
			`          error: interface element is nil`,
			`      Struct:`,
			`        type: object`,
			`        properties:`,
			`          error: empty struct not supported`,
			`  PrivateStruct:`,
			`    type: object`,
			`    properties:`,
			`      error: struct has no exported fields`,
			`  StringStruct:`,
			`    type: object`,
			`    properties:`,
			`      Value:`,
			`        type: string`,
			`root:`,
			`  $ref: '#/definitions/CompoundTypes'`,
		},
	},
	{
		name:  "special",
		value: SpecialTypes{},
		refStrings: []string{
			`TypeRefs.SpecialTypes:{}`,
			`TypeRefs.SpecialTypes:{}.DateTime:datetime`,
			`Root.{}:SpecialTypes`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.DateTime:datetime`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  SpecialTypes:`,
			`    type: object`,
			`    properties:`,
			`      DateTime:`,
			`        type: string`,
			`        format: date-time`,
			`root:`,
			`  $ref: '#/definitions/SpecialTypes'`,
		},
	},
}

type ArrayStruct struct {
	Array0   [0]string
	Array3   [3]string
	Array2_3 [2][3]string
}

type SliceStruct struct {
	Slice  []string
	Array2 [][]string
}

var jsonArrayTest = `
{
	"Array0": [],
	"Array3": ["a","b","c"],
	"Array2_3": [
		[1,2,3],
		[2,3,4]
	]
}
`

// Array tests.
var listTests = []TestCase{
	{
		name:  "arrays",
		value: &ArrayStruct{},
		refStrings: []string{
			`TypeRefs.ArrayStruct:{}`,
			`TypeRefs.ArrayStruct:{}.Array0:[]`,
			`TypeRefs.ArrayStruct:{}.Array0:[].string`,
			`TypeRefs.ArrayStruct:{}.Array2_3:[]`,
			`TypeRefs.ArrayStruct:{}.Array2_3:[].[]`,
			`TypeRefs.ArrayStruct:{}.Array2_3:[].[].string`,
			`TypeRefs.ArrayStruct:{}.Array3:[]`,
			`TypeRefs.ArrayStruct:{}.Array3:[].string`,
			`Root.{}:ArrayStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Array0:[]`,
			`Root.{}.Array0:[].string`,
			`Root.{}.Array2_3:[]`,
			`Root.{}.Array2_3:[].[]`,
			`Root.{}.Array2_3:[].[].string`,
			`Root.{}.Array3:[]`,
			`Root.{}.Array3:[].string`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  ArrayStruct:`,
			`    type: object`,
			`    properties:`,
			`      Array0:`,
			`        type: array`,
			`        items:`,
			`          type: string`,
			`      Array2_3:`,
			`        type: array`,
			`        items:`,
			`          type: array`,
			`          items:`,
			`            type: string`,
			`      Array3:`,
			`        type: array`,
			`        items:`,
			`          type: string`,
			`root:`,
			`  $ref: '#/definitions/ArrayStruct'`,
		},
	},
	{
		name:  "json-array",
		value: fromJSON([]byte(jsonArrayTest)),
		refStrings: []string{
			`Root.{}`,
			`Root.{}.Array0:[]`,
			`Root.{}.Array0:[].!invalid! ERROR:interface element is nil`,
			`Root.{}.Array2_3:[]`,
			`Root.{}.Array2_3:[].[]`,
			`Root.{}.Array2_3:[].[].float`,
			`Root.{}.Array3:[]`,
			`Root.{}.Array3:[].string`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Array0:[]`,
			`Root.{}.Array0:[].!invalid! ERROR:interface element is nil`,
			`Root.{}.Array2_3:[]`,
			`Root.{}.Array2_3:[].[]`,
			`Root.{}.Array2_3:[].[].float`,
			`Root.{}.Array3:[]`,
			`Root.{}.Array3:[].string`,
		},
		openapiStrings: []string{
			`root:`,
			`  type: object`,
			`  properties:`,
			`    Array0:`,
			`      type: array`,
			`      items:`,
			`        type: invalid`,
			`        error: interface element is nil`,
			`    Array2_3:`,
			`      type: array`,
			`      items:`,
			`        type: array`,
			`        items:`,
			`          type: number`,
			`          format: double`,
			`    Array3:`,
			`      type: array`,
			`      items:`,
			`        type: string`,
		},
	},
	{
		name:  "slices",
		value: &SliceStruct{},
		refStrings: []string{
			`TypeRefs.SliceStruct:{}`,
			`TypeRefs.SliceStruct:{}.Array2:[]`,
			`TypeRefs.SliceStruct:{}.Array2:[].[]`,
			`TypeRefs.SliceStruct:{}.Array2:[].[].string`,
			`TypeRefs.SliceStruct:{}.Slice:[]`,
			`TypeRefs.SliceStruct:{}.Slice:[].string`,
			`Root.{}:SliceStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Array2:[]`,
			`Root.{}.Array2:[].[]`,
			`Root.{}.Array2:[].[].string`,
			`Root.{}.Slice:[]`,
			`Root.{}.Slice:[].string`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  SliceStruct:`,
			`    type: object`,
			`    properties:`,
			`      Array2:`,
			`        type: array`,
			`        items:`,
			`          type: array`,
			`          items:`,
			`            type: string`,
			`      Slice:`,
			`        type: array`,
			`        items:`,
			`          type: string`,
			`root:`,
			`  $ref: '#/definitions/SliceStruct'`,
		},
	},
}

type MapTestsStruct struct {
	MapOK struct {
		StringVal string
		IntVal    float64
		FloatVal  float32
		BoolVal   bool
		ListVal   []float64
		MapVal    struct {
			Key1 string
			Key2 struct {
				DeepKey1 string
				DeepKey2 float64
			}
		}
	}
}

var jsonMapTests = `
{
	"MapOK": {
		"StringVal": "Hello",
		"IntVal": 123,
		"FloatVal": 234.345,
		"BoolVal": true,
		"ListVal": [2,3,4,5],
		"MapVal": {
			"Key1": "Hey",
			"Key2": {
				"DeepKey1": "Hi",
				"DeepKey2": 234
			}
		}
	}
}
`

var compoundTests = []TestCase{
	{
		name:  "golang-map",
		value: MapTestsStruct{},
		refStrings: []string{
			`TypeRefs.MapTestsStruct:{}`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.BoolVal:boolean`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.FloatVal:float`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.IntVal:float`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.ListVal:[]`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.ListVal:[].float`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key1:string`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.StringVal:string`,
			`Root.{}:MapTestsStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.MapOK:{}`,
			`Root.{}.MapOK:{}.BoolVal:boolean`,
			`Root.{}.MapOK:{}.FloatVal:float`,
			`Root.{}.MapOK:{}.IntVal:float`,
			`Root.{}.MapOK:{}.ListVal:[]`,
			`Root.{}.MapOK:{}.ListVal:[].float`,
			`Root.{}.MapOK:{}.MapVal:{}`,
			`Root.{}.MapOK:{}.MapVal:{}.Key1:string`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
			`Root.{}.MapOK:{}.StringVal:string`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  MapTestsStruct:`,
			`    type: object`,
			`    properties:`,
			`      MapOK:`,
			`        type: object`,
			`        properties:`,
			`          BoolVal:`,
			`            type: boolean`,
			`          FloatVal:`,
			`            type: number`,
			`          IntVal:`,
			`            type: number`,
			`            format: double`,
			`          ListVal:`,
			`            type: array`,
			`            items:`,
			`              type: number`,
			`              format: double`,
			`          MapVal:`,
			`            type: object`,
			`            properties:`,
			`              Key1:`,
			`                type: string`,
			`              Key2:`,
			`                type: object`,
			`                properties:`,
			`                  DeepKey1:`,
			`                    type: string`,
			`                  DeepKey2:`,
			`                    type: number`,
			`                    format: double`,
			`          StringVal:`,
			`            type: string`,
			`root:`,
			`  $ref: '#/definitions/MapTestsStruct'`,
		},
	},
	{
		name:  "json-map",
		value: fromJSON([]byte(jsonMapTests)),
		refStrings: []string{
			`Root.{}`,
			`Root.{}.MapOK:{}`,
			`Root.{}.MapOK:{}.BoolVal:boolean`,
			`Root.{}.MapOK:{}.FloatVal:float`,
			`Root.{}.MapOK:{}.IntVal:float`,
			`Root.{}.MapOK:{}.ListVal:[]`,
			`Root.{}.MapOK:{}.ListVal:[].float`,
			`Root.{}.MapOK:{}.MapVal:{}`,
			`Root.{}.MapOK:{}.MapVal:{}.Key1:string`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
			`Root.{}.MapOK:{}.StringVal:string`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.MapOK:{}`,
			`Root.{}.MapOK:{}.BoolVal:boolean`,
			`Root.{}.MapOK:{}.FloatVal:float`,
			`Root.{}.MapOK:{}.IntVal:float`,
			`Root.{}.MapOK:{}.ListVal:[]`,
			`Root.{}.MapOK:{}.ListVal:[].float`,
			`Root.{}.MapOK:{}.MapVal:{}`,
			`Root.{}.MapOK:{}.MapVal:{}.Key1:string`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
			`Root.{}.MapOK:{}.StringVal:string`,
		},
		openapiStrings: []string{
			`root:`,
			`  type: object`,
			`  properties:`,
			`    MapOK:`,
			`      type: object`,
			`      properties:`,
			`        BoolVal:`,
			`          type: boolean`,
			`        FloatVal:`,
			`          type: number`,
			`          format: double`,
			`        IntVal:`,
			`          type: number`,
			`          format: double`,
			`        ListVal:`,
			`          type: array`,
			`          items:`,
			`            type: number`,
			`            format: double`,
			`        MapVal:`,
			`          type: object`,
			`          properties:`,
			`            Key1:`,
			`              type: string`,
			`            Key2:`,
			`              type: object`,
			`              properties:`,
			`                DeepKey1:`,
			`                  type: string`,
			`                DeepKey2:`,
			`                  type: number`,
			`                  format: double`,
			`        StringVal:`,
			`          type: string`,
		},
	},
}

type ReferenceTestsStruct struct {
	InterfaceVal interface{}
	PtrVal       *BasicStruct
	PtrPtrVal    **BasicStruct
}

var referenceTests = []TestCase{
	{
		name:  "reference-tests-empty",
		value: ReferenceTestsStruct{},
		refStrings: []string{
			`TypeRefs.BasicStruct:{}`,
			`TypeRefs.BasicStruct:{}.BoolVal:boolean`,
			`TypeRefs.BasicStruct:{}.Float64Val:float`,
			`TypeRefs.BasicStruct:{}.IntVal:integer`,
			`TypeRefs.BasicStruct:{}.StringVal:string`,
			`TypeRefs.ReferenceTestsStruct:{}`,
			`TypeRefs.ReferenceTestsStruct:{}.!InterfaceVal:invalid! ERROR:interface element is nil`,
			`TypeRefs.ReferenceTestsStruct:{}.PtrPtrVal:{}:BasicStruct`,
			`TypeRefs.ReferenceTestsStruct:{}.PtrVal:{}:BasicStruct`,
			`Root.{}:ReferenceTestsStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.!InterfaceVal:invalid! ERROR:interface element is nil`,
			`Root.{}.PtrPtrVal:{}`,
			`Root.{}.PtrPtrVal:{}.BoolVal:boolean`,
			`Root.{}.PtrPtrVal:{}.Float64Val:float`,
			`Root.{}.PtrPtrVal:{}.IntVal:integer`,
			`Root.{}.PtrPtrVal:{}.StringVal:string`,
			`Root.{}.PtrVal:{}`,
			`Root.{}.PtrVal:{}.BoolVal:boolean`,
			`Root.{}.PtrVal:{}.Float64Val:float`,
			`Root.{}.PtrVal:{}.IntVal:integer`,
			`Root.{}.PtrVal:{}.StringVal:string`,
		},
	},
	{
		name:  "reference-tests-init",
		value: ReferenceTestsStruct{InterfaceVal: &BasicStruct{}},
		refStrings: []string{
			`TypeRefs.BasicStruct:{}`,
			`TypeRefs.BasicStruct:{}.BoolVal:boolean`,
			`TypeRefs.BasicStruct:{}.Float64Val:float`,
			`TypeRefs.BasicStruct:{}.IntVal:integer`,
			`TypeRefs.BasicStruct:{}.StringVal:string`,
			`TypeRefs.ReferenceTestsStruct:{}`,
			`TypeRefs.ReferenceTestsStruct:{}.InterfaceVal:{}:BasicStruct`,
			`TypeRefs.ReferenceTestsStruct:{}.PtrPtrVal:{}:BasicStruct`,
			`TypeRefs.ReferenceTestsStruct:{}.PtrVal:{}:BasicStruct`,
			`Root.{}:ReferenceTestsStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.InterfaceVal:{}`,
			`Root.{}.InterfaceVal:{}.BoolVal:boolean`,
			`Root.{}.InterfaceVal:{}.Float64Val:float`,
			`Root.{}.InterfaceVal:{}.IntVal:integer`,
			`Root.{}.InterfaceVal:{}.StringVal:string`,
			`Root.{}.PtrPtrVal:{}`,
			`Root.{}.PtrPtrVal:{}.BoolVal:boolean`,
			`Root.{}.PtrPtrVal:{}.Float64Val:float`,
			`Root.{}.PtrPtrVal:{}.IntVal:integer`,
			`Root.{}.PtrPtrVal:{}.StringVal:string`,
			`Root.{}.PtrVal:{}`,
			`Root.{}.PtrVal:{}.BoolVal:boolean`,
			`Root.{}.PtrVal:{}.Float64Val:float`,
			`Root.{}.PtrVal:{}.IntVal:integer`,
			`Root.{}.PtrVal:{}.StringVal:string`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  BasicStruct:`,
			`    type: object`,
			`    properties:`,
			`      BoolVal:`,
			`        type: boolean`,
			`      Float64Val:`,
			`        type: number`,
			`        format: double`,
			`      IntVal:`,
			`        type: integer`,
			`      StringVal:`,
			`        type: string`,
			`  ReferenceTestsStruct:`,
			`    type: object`,
			`    properties:`,
			`      InterfaceVal:`,
			`        $ref: '#/definitions/BasicStruct'`,
			`      PtrPtrVal:`,
			`        $ref: '#/definitions/BasicStruct'`,
			`      PtrVal:`,
			`        $ref: '#/definitions/BasicStruct'`,
			`root:`,
			`  $ref: '#/definitions/ReferenceTestsStruct'`,
		},
	},
}

// Test cyclical relationships:
// A --> B --> C --> A
type AStruct struct {
	AName  string   `json:"aName,omitempty"`
	AChild *BStruct `json:"aChild"`
}

type BStruct struct {
	BName  string   `json:"bName"`
	BChild *CStruct `json:"bChild"`
}

type CStruct struct {
	CName  string   `json:"cName"`
	CChild *AStruct `json:"cChild"`
}

type BadType interface{}

type CycleTest struct {
	Level  int      `json:"-"`
	CycleA AStruct  `json:"cycleA"`
	CycleB *BStruct `json:"cycleB"`
	CycleC struct {
		C CStruct `json:"c"`
	}
}

var cycleTests = []TestCase{
	{
		name:  "cycle-test",
		value: &CycleTest{},
		refStrings: []string{
			`TypeRefs.AStruct:{}`,
			`TypeRefs.AStruct:{}.AChild:{}:BStruct`,
			`TypeRefs.AStruct:{}.AName:string`,
			`TypeRefs.BStruct:{}`,
			`TypeRefs.BStruct:{}.BChild:{}:CStruct`,
			`TypeRefs.BStruct:{}.BName:string`,
			`TypeRefs.CStruct:{}`,
			`TypeRefs.CStruct:{}.CChild:{}:AStruct`,
			`TypeRefs.CStruct:{}.CName:string`,
			`TypeRefs.CycleTest:{}`,
			`TypeRefs.CycleTest:{}.CycleA:{}:AStruct`,
			`TypeRefs.CycleTest:{}.CycleB:{}:BStruct`,
			`TypeRefs.CycleTest:{}.CycleC:{}`,
			`TypeRefs.CycleTest:{}.CycleC:{}.C:{}:CStruct`,
			`TypeRefs.CycleTest:{}.Level:integer`,
			`Root.{}:CycleTest`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.CycleA:{}`,
			`Root.{}.CycleA:{}.AChild:{}`,
			`Root.{}.CycleA:{}.AChild:{}.BChild:{}`,
			`Root.{}.CycleA:{}.AChild:{}.BChild:{}.!CChild:{}:AStruct! ERROR:cyclical reference`,
			`Root.{}.CycleA:{}.AChild:{}.BChild:{}.CName:string`,
			`Root.{}.CycleA:{}.AChild:{}.BName:string`,
			`Root.{}.CycleA:{}.AName:string`,
			`Root.{}.CycleB:{}`,
			`Root.{}.CycleB:{}.BChild:{}`,
			`Root.{}.CycleB:{}.BChild:{}.CChild:{}`,
			`Root.{}.CycleB:{}.BChild:{}.CChild:{}.!AChild:{}:BStruct! ERROR:cyclical reference`,
			`Root.{}.CycleB:{}.BChild:{}.CChild:{}.AName:string`,
			`Root.{}.CycleB:{}.BChild:{}.CName:string`,
			`Root.{}.CycleB:{}.BName:string`,
			`Root.{}.CycleC:{}`,
			`Root.{}.CycleC:{}.C:{}`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}.!BChild:{}:CStruct! ERROR:cyclical reference`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}.BName:string`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}.AName:string`,
			`Root.{}.CycleC:{}.C:{}.CName:string`,
			`Root.{}.Level:integer`,
		},
		jsonStrings: []string{
			`definitions.cycleA:{}`,
			`definitions.cycleA:{}.aChild:{}:BStruct`,
			`definitions.cycleA:{}.aName:string`,
			`definitions.aChild:{}`,
			`definitions.aChild:{}.bChild:{}:CStruct`,
			`definitions.aChild:{}.bName:string`,
			`definitions.bChild:{}`,
			`definitions.bChild:{}.cChild:{}:AStruct`,
			`definitions.bChild:{}.cName:string`,
			`definitions.CycleTest:{}`,
			`definitions.CycleTest:{}.cycleA:{}:AStruct`,
			`definitions.CycleTest:{}.cycleB:{}:BStruct`,
			`definitions.CycleTest:{}.CycleC:{}`,
			`definitions.CycleTest:{}.CycleC:{}.c:{}:CStruct`,
			`$.{}:CycleTest`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  cycleA:`,
			`    type: object`,
			`    properties:`,
			`      aChild:`,
			`        $ref: '#/definitions/BStruct'`,
			`      aName:`,
			`        type: string`,
			`  aChild:`,
			`    type: object`,
			`    properties:`,
			`      bChild:`,
			`        $ref: '#/definitions/CStruct'`,
			`      bName:`,
			`        type: string`,
			`  bChild:`,
			`    type: object`,
			`    properties:`,
			`      cChild:`,
			`        $ref: '#/definitions/AStruct'`,
			`      cName:`,
			`        type: string`,
			`  CycleTest:`,
			`    type: object`,
			`    properties:`,
			`      cycleA:`,
			`        $ref: '#/definitions/AStruct'`,
			`      cycleB:`,
			`        $ref: '#/definitions/BStruct'`,
			`      CycleC:`,
			`        type: object`,
			`        properties:`,
			`          c:`,
			`            $ref: '#/definitions/CStruct'`,
			`root:`,
			`  $ref: '#/definitions/CycleTest'`,
		},
	},
}

type JSONTagTests struct {
	NoTag      string
	ExcludeTag string `json:"-"`
	RenameOne  string `json:"renameOne"`
	RenameTwo  string `json:"something"`
}

var jsonTagTests = []TestCase{
	{
		name:  "json-tags",
		value: JSONTagTests{},
		refStrings: []string{
			`TypeRefs.JSONTagTests:{}`,
			`TypeRefs.JSONTagTests:{}.ExcludeTag:string`,
			`TypeRefs.JSONTagTests:{}.NoTag:string`,
			`TypeRefs.JSONTagTests:{}.RenameOne:string`,
			`TypeRefs.JSONTagTests:{}.RenameTwo:string`,
			`Root.{}:JSONTagTests`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.ExcludeTag:string`,
			`Root.{}.NoTag:string`,
			`Root.{}.RenameOne:string`,
			`Root.{}.RenameTwo:string`,
		},
		jsonStrings: []string{
			`definitions.JSONTagTests:{}`,
			`definitions.JSONTagTests:{}.NoTag:string`,
			`definitions.JSONTagTests:{}.renameOne:string`,
			`definitions.JSONTagTests:{}.something:string`,
			`$.{}:JSONTagTests`,
		},
		openapiStrings: []string{
			`definitions:`,
			`  JSONTagTests:`,
			`    type: object`,
			`    properties:`,
			`      NoTag:`,
			`        type: string`,
			`      renameOne:`,
			`        type: string`,
			`      something:`,
			`        type: string`,
			`root:`,
			`  $ref: '#/definitions/JSONTagTests'`,
		},
	},
}

var structTests = []TestCase{
	// {name: "struct-empty", value: func() interface{} { var g struct{}; return g }()},
	// {name: "PrivateStruct-nil", value: func() interface{} { var g PrivateStruct; return g }()},
	{name: "BasicStruct-nil", value: func() interface{} { var g BasicStruct; return g }()},
	// {name: "CompoundStruct-nil", value: func() interface{} { var g CompoundStruct; return g }()},
	// {name: "CycleTest-nil", value: func() interface{} { var g CycleTest; return g }()},
}

//
//{name: "makeJSON, value", value: makeJSON(nil)},

var testCases = []TestCase{
	{name: "GoodEntity, var", value: func() interface{} { var g GoodEntity; return g }()},
	{name: "GoodEntity, empty", value: GoodEntity{}},
	{name: "GoodEntity, values", value: GoodEntity{
		Message: "hello",
		IntVal:  123,
		Same:    true,
		secret:  "shh",
	}},

	{name: "map[string]bool, values", value: map[string]bool{"trueVal": true, "falseVal": false}},

	{name: "[]*MainStruct, nil", value: []*MainStruct{}},
	{name: "[0]*MainStruct, nil", value: [0]*MainStruct{}},
	{name: "[1]*MainStruct, nil", value: [1]*MainStruct{}},

	{name: "*GoodEntity, var", value: func() interface{} { var g *GoodEntity; return g }()},
	{name: "*GoodEntity, empty", value: &GoodEntity{}},
	{name: "*GoodEntity, values", value: &GoodEntity{
		Message: "hello",
		IntVal:  123,
		Same:    true,
		secret:  "shh",
	}},

	{name: "*OtherEntity, var", value: func() interface{} { var g *OtherEntity; return g }()},
	{name: "*OtherEntity, empty", value: &OtherEntity{}},
	{name: "*OtherEntity, values", value: &OtherEntity{
		Status:   "ok",
		IntVal:   123,
		FloatVal: 234.345,
		Same:     true,
		MapVal:   make(map[string]int64),
		Good:     GoodEntity{},
	}},

	{name: "NamedEntity, empty", value: &NamedEntity{}},
}

// StringStruct has one string field.
type StringStruct struct {
	Value string
}

// Private Struct only has private fields.
type PrivateStruct struct {
	boolVal    bool
	intVal     int
	float64Val float64
	stringVal  string
}

// BasicStruct has one field for each basic type.
type BasicStruct struct {
	BoolVal    bool
	IntVal     int
	Float64Val float64
	StringVal  string
}

// CompoundStruct has fields with compound types.
type CompoundStruct struct {
	//	Array
	ZeroArrayVal  [0]string
	ThreeArrayVal [3]string

	//	Slice
	SliceVal []string

	//	Map
	MapVal map[string]string

	//	Struct
	StructVal        StringStruct
	PrivateStructVal PrivateStruct
}

/*
Only consider basic types:
- string, int, float, bool
- slices, arrays
- structs, maps

*/
type MainStruct struct {
	StringVal string `json:"stringVal,omitempty"`
	IntVal    int    `json:"intVal" datastore:",noindex"`
	FloatVal  float64
	BoolVal   bool

	SliceVal []int

	InterfaceVal interface{}

	StructPtr *GoodEntity
	StructVal OtherEntity

	StringPtr *string

	// Test duplicate JSON keys when capitalized.
	DuplicateOne string
	DuplicateTwo string `json:"duplicateOne"`

	privateVal string
}

// define a struct for data storage
type GoodEntity struct {
	Message string
	IntVal  int64
	Same    bool

	secret string
}

// Test named and un-named types.
type SimpleString string
type SimpleInt int64
type SimpleFloat float64
type SimpleBool bool
type SimpleInterface interface{}
type SimpleSlice []string
type SimpleMap map[string]int64
type SimpleStruct GoodEntity
type SimpleStructSlice []GoodEntity
type SimplePtr *GoodEntity
type SimplePtrSlice []*GoodEntity

type NamedEntity struct {
	NamedString SimpleString `json:"namedString,omitempty"`
	RealString  string

	NamedInt SimpleInt
	RealInt  int64

	NamedFloat SimpleFloat
	RealFloat  float64

	NamedBool SimpleBool
	RealBool  bool

	NamedInterface SimpleInterface
	RealInterface  interface{}

	NamedSlice SimpleSlice
	RealSlice  []string

	NamedMap SimpleMap
	RealMap  map[string]int64

	NamedStruct SimpleStruct
	RealStruct  GoodEntity

	NamedStructSlice SimpleStructSlice
	RealStructSlice  []GoodEntity

	NamedPtr SimplePtr
	RealPtr  *GoodEntity

	NamedPtrSlice SimplePtrSlice
	RealPtrSlice  []*GoodEntity
}

// define a different struct to test mismatched structs
type OtherEntity struct {
	Status   string
	IntVal   int64
	FloatVal float64
	Same     bool
	Simple   SimpleInt

	MapNil map[string]int64
	MapVal map[string]int64

	Good         GoodEntity
	GoodPtr      *GoodEntity
	GoodSlice    []GoodEntity
	GoodPtrSlice []*GoodEntity

	AnonStruct struct {
		FieldOne   string
		FieldTwo   int32
		FieldThree float32
	}
}

// fromJSON converts a JSON string into an interface.
func fromJSON(in []byte) interface{} {
	var out interface{}

	if err := json.Unmarshal(in, &out); err != nil {
		err = fmt.Errorf("ERROR json.Unmarshal: %s\n%s", err, string(in))
		fmt.Println(err)
		return err
	}

	// // DEBUGXXXXX Print indented JSON string.
	// if out != nil {
	// 	if b, err := json.MarshalIndent(out, "", "  "); err == nil {
	// 		fmt.Println(string(b))
	// 	}
	// }

	return out
}

// makeJSON converts an interface to JSON.
func makeJSON(x interface{}) interface{} {
	var s = "hey"

	x = &MainStruct{
		StringVal: "hello",
		IntVal:    123,
		FloatVal:  234.345,
		BoolVal:   true,
		SliceVal:  []int{1, 2, 3},
		StructPtr: &GoodEntity{
			Message: "hi",
			IntVal:  234,
			Same:    true,
			secret:  "eyes only",
		},
		StructVal: OtherEntity{
			Status:   "ok",
			IntVal:   789,
			FloatVal: 789.123,
			Same:     true,
			MapVal:   map[string]int64{"one": 234, "two": 345, "three": 456},
			Good: GoodEntity{
				Message: "",
				IntVal:  0,
				Same:    false,
				secret:  "",
			},
			GoodPtr: &GoodEntity{
				Message: "hi",
				IntVal:  234,
				Same:    true,
				secret:  "eyes only",
			},
			GoodSlice:    []GoodEntity{},
			GoodPtrSlice: []*GoodEntity{},
		},
		StringPtr: &s,

		DuplicateOne: "one",
		DuplicateTwo: "two",

		privateVal: "shh",
	}

	if b, err := json.Marshal(x); err != nil {
		return fmt.Errorf("ERROR json.Marshal: %s", err)
	} else {
		return fromJSON(b)
	}
}

func compareStrings(t *testing.T, testName string, gotStrings, wantStrings []string) {
	// Split strings into lines.
	gotLines := []string{}
	for _, line := range gotStrings {
		lines := strings.Split(line, "\n")
		gotLines = append(gotLines, lines...)
	}
	wantLines := []string{}
	for _, line := range wantStrings {
		lines := strings.Split(line, "\n")
		wantLines = append(wantLines, lines...)
	}

	if !reflect.DeepEqual(gotLines, wantLines) {
		t.Errorf("TEST_FAIL %s", testName)

		maxLen := len(gotLines)
		if len(wantLines) > maxLen {
			maxLen = len(wantLines)
		}

		type diffStruct struct {
			got, want string
		}
		diff := []*diffStruct{}

		for i := 0; i < maxLen; i++ {
			newDiff := &diffStruct{}

			if i < len(gotLines) {
				newDiff.got = gotLines[i]
			}

			if i < len(wantLines) {
				newDiff.want = wantLines[i]
			}

			diff = append(diff, newDiff)
		}

		// Dump got and want lines.
		outLines := []string{}

		outLines = append(outLines, "***** GOT:")
		for i, newDiff := range diff {
			flag := " "
			if newDiff.got != newDiff.want {
				flag = ">"
			}

			outLines = append(outLines, fmt.Sprintf("%05d%s| %s", i, flag, newDiff.got))
		}

		outLines = append(outLines, "***** WANT:")
		for i, newDiff := range diff {
			flag := " "
			if newDiff.got != newDiff.want {
				flag = ">"
			}

			outLines = append(outLines, fmt.Sprintf("%05d%s| %s", i, flag, newDiff.want))
		}

		t.Errorf("TEST_FAIL %s\n%s", testName, strings.Join(outLines, "\n"))
	} else {
		t.Logf("TEST_OK %s", testName)
	}
}

func runTests(t *testing.T, testCases []TestCase) {
	r := reflector.NewReflector()

	for _, test := range testCases {
		r.Reset()
		//r.Label = test.name

		gotResult := r.DeriveSchema(test.value)

		// if b, err := json.MarshalIndent(gotResult, "", "  "); err != nil {
		// 	t.Errorf("TEST_FAIL %s: json.Marshal err=%s", test.name, err)
		// } else {
		// 	fmt.Println(string(b))
		// }

		for i := 0; i < 2; i++ {
			opt := NewOptions()
			opt.DeReference = i == 1

			r := NewSimpleRenderer(opt)
			gotStrings, _ := r.ProcessResult(gotResult)

			var wantStrings []string
			if opt.DeReference {
				wantStrings = test.derefStrings
			} else {
				wantStrings = test.refStrings
			}

			testName := fmt.Sprintf("%s: deref=%t", test.name, opt.DeReference)
			compareStrings(t, testName, gotStrings, wantStrings)
		}

		// Test json dialect.
		if len(test.jsonStrings) > 0 {
			opt := NewOptions()
			opt.DeReference = false

			r := NewJSONRenderer(opt)
			gotStrings, _ := r.ProcessResult(gotResult)
			wantStrings := test.jsonStrings

			testName := fmt.Sprintf("%s: dialect=json", test.name)
			compareStrings(t, testName, gotStrings, wantStrings)
		}

		// Test OpenAPI schema.
		if len(test.openapiStrings) > 0 {
			opt := NewOptions()
			opt.DeReference = false
			opt.Indent = 0

			r := NewOpenAPIRenderer(opt)
			gotStrings, _ := r.ProcessResult(gotResult)
			wantStrings := test.openapiStrings

			testName := fmt.Sprintf("%s: dialect=openapi", test.name)
			compareStrings(t, testName, gotStrings, wantStrings)
		}
	}
}

func TestReflector_AllTests(t *testing.T) {
	for _, testCases := range allTests {
		runTests(t, testCases)
	}
}
