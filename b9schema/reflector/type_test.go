package reflector

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/gitmann/b9schema-reflector-golang/b9schema/enum/generictype"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/enum/threeflag"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/enum/typecategory"
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
	refStrings   []string
	derefStrings []string
	jsonStrings  []string
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
	},
	{
		name:  "integer",
		value: IntegerTypes{},
		refStrings: []string{
			`TypeRefs.IntegerTypes:{}`,
			`TypeRefs.IntegerTypes:{}.Int:integer`,
			`TypeRefs.IntegerTypes:{}.Int8:integer`,
			`TypeRefs.IntegerTypes:{}.Int16:integer`,
			`TypeRefs.IntegerTypes:{}.Int32:integer`,
			`TypeRefs.IntegerTypes:{}.Int64:integer`,
			`TypeRefs.IntegerTypes:{}.Uint:integer`,
			`TypeRefs.IntegerTypes:{}.Uint8:integer`,
			`TypeRefs.IntegerTypes:{}.Uint16:integer`,
			`TypeRefs.IntegerTypes:{}.Uint32:integer`,
			`TypeRefs.IntegerTypes:{}.Uint64:integer`,
			`TypeRefs.IntegerTypes:{}.Uintptr:integer`,
			`Root.{}:IntegerTypes`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Int:integer`,
			`Root.{}.Int8:integer`,
			`Root.{}.Int16:integer`,
			`Root.{}.Int32:integer`,
			`Root.{}.Int64:integer`,
			`Root.{}.Uint:integer`,
			`Root.{}.Uint8:integer`,
			`Root.{}.Uint16:integer`,
			`Root.{}.Uint32:integer`,
			`Root.{}.Uint64:integer`,
			`Root.{}.Uintptr:integer`,
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
	},
	{
		name:  "invalid",
		value: InvalidTypes{},
		refStrings: []string{
			`TypeRefs.InvalidTypes:{}`,
			`TypeRefs.InvalidTypes:{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
			`TypeRefs.InvalidTypes:{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
			`TypeRefs.InvalidTypes:{}.!Chan:invalid:chan! ERROR:kind not supported`,
			`TypeRefs.InvalidTypes:{}.!Func:invalid:func! ERROR:kind not supported`,
			`TypeRefs.InvalidTypes:{}."!UnsafePointer:invalid:unsafe.Pointer!" ERROR:kind not supported`,
			`Root.{}:InvalidTypes`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
			`Root.{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
			`Root.{}.!Chan:invalid:chan! ERROR:kind not supported`,
			`Root.{}.!Func:invalid:func! ERROR:kind not supported`,
			`Root.{}."!UnsafePointer:invalid:unsafe.Pointer!" ERROR:kind not supported`,
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
			`TypeRefs.CompoundTypes:{}.Ptr:{}:StringStruct`,
			`TypeRefs.CompoundTypes:{}.PrivatePtr:{}:PrivateStruct`,
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
			`Root.{}.Ptr:{}`,
			`Root.{}.Ptr:{}.Value:string`,
			`Root.{}.!PrivatePtr:{}! ERROR:struct has no exported fields`,
			`Root.{}.Slice:[]`,
			`Root.{}.Slice:[].!invalid! ERROR:interface element is nil`,
			`Root.{}.!Struct:{}! ERROR:empty struct not supported`,
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
			`TypeRefs.ArrayStruct:{}.Array3:[]`,
			`TypeRefs.ArrayStruct:{}.Array3:[].string`,
			`TypeRefs.ArrayStruct:{}.Array2_3:[]`,
			`TypeRefs.ArrayStruct:{}.Array2_3:[].[]`,
			`TypeRefs.ArrayStruct:{}.Array2_3:[].[].string`,
			`Root.{}:ArrayStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Array0:[]`,
			`Root.{}.Array0:[].string`,
			`Root.{}.Array3:[]`,
			`Root.{}.Array3:[].string`,
			`Root.{}.Array2_3:[]`,
			`Root.{}.Array2_3:[].[]`,
			`Root.{}.Array2_3:[].[].string`,
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
	},
	{
		name:  "slices",
		value: &SliceStruct{},
		refStrings: []string{
			`TypeRefs.SliceStruct:{}`,
			`TypeRefs.SliceStruct:{}.Slice:[]`,
			`TypeRefs.SliceStruct:{}.Slice:[].string`,
			`TypeRefs.SliceStruct:{}.Array2:[]`,
			`TypeRefs.SliceStruct:{}.Array2:[].[]`,
			`TypeRefs.SliceStruct:{}.Array2:[].[].string`,
			`Root.{}:SliceStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Slice:[]`,
			`Root.{}.Slice:[].string`,
			`Root.{}.Array2:[]`,
			`Root.{}.Array2:[].[]`,
			`Root.{}.Array2:[].[].string`,
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
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.StringVal:string`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.IntVal:float`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.FloatVal:float`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.BoolVal:boolean`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.ListVal:[]`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.ListVal:[].float`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key1:string`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`TypeRefs.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
			`Root.{}:MapTestsStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.MapOK:{}`,
			`Root.{}.MapOK:{}.StringVal:string`,
			`Root.{}.MapOK:{}.IntVal:float`,
			`Root.{}.MapOK:{}.FloatVal:float`,
			`Root.{}.MapOK:{}.BoolVal:boolean`,
			`Root.{}.MapOK:{}.ListVal:[]`,
			`Root.{}.MapOK:{}.ListVal:[].float`,
			`Root.{}.MapOK:{}.MapVal:{}`,
			`Root.{}.MapOK:{}.MapVal:{}.Key1:string`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
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
			`TypeRefs.BasicStruct:{}.IntVal:integer`,
			`TypeRefs.BasicStruct:{}.Float64Val:float`,
			`TypeRefs.BasicStruct:{}.StringVal:string`,
			`TypeRefs.ReferenceTestsStruct:{}`,
			`TypeRefs.ReferenceTestsStruct:{}.!InterfaceVal:invalid! ERROR:interface element is nil`,
			`TypeRefs.ReferenceTestsStruct:{}.PtrVal:{}:BasicStruct`,
			`TypeRefs.ReferenceTestsStruct:{}.PtrPtrVal:{}:BasicStruct`,
			`Root.{}:ReferenceTestsStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.!InterfaceVal:invalid! ERROR:interface element is nil`,
			`Root.{}.PtrVal:{}`,
			`Root.{}.PtrVal:{}.BoolVal:boolean`,
			`Root.{}.PtrVal:{}.IntVal:integer`,
			`Root.{}.PtrVal:{}.Float64Val:float`,
			`Root.{}.PtrVal:{}.StringVal:string`,
			`Root.{}.PtrPtrVal:{}`,
			`Root.{}.PtrPtrVal:{}.BoolVal:boolean`,
			`Root.{}.PtrPtrVal:{}.IntVal:integer`,
			`Root.{}.PtrPtrVal:{}.Float64Val:float`,
			`Root.{}.PtrPtrVal:{}.StringVal:string`,
		},
	},
	{
		name:  "reference-tests-init",
		value: ReferenceTestsStruct{InterfaceVal: &BasicStruct{}},
		refStrings: []string{
			`TypeRefs.BasicStruct:{}`,
			`TypeRefs.BasicStruct:{}.BoolVal:boolean`,
			`TypeRefs.BasicStruct:{}.IntVal:integer`,
			`TypeRefs.BasicStruct:{}.Float64Val:float`,
			`TypeRefs.BasicStruct:{}.StringVal:string`,
			`TypeRefs.ReferenceTestsStruct:{}`,
			`TypeRefs.ReferenceTestsStruct:{}.InterfaceVal:{}:BasicStruct`,
			`TypeRefs.ReferenceTestsStruct:{}.PtrVal:{}:BasicStruct`,
			`TypeRefs.ReferenceTestsStruct:{}.PtrPtrVal:{}:BasicStruct`,
			`Root.{}:ReferenceTestsStruct`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.InterfaceVal:{}`,
			`Root.{}.InterfaceVal:{}.BoolVal:boolean`,
			`Root.{}.InterfaceVal:{}.IntVal:integer`,
			`Root.{}.InterfaceVal:{}.Float64Val:float`,
			`Root.{}.InterfaceVal:{}.StringVal:string`,
			`Root.{}.PtrVal:{}`,
			`Root.{}.PtrVal:{}.BoolVal:boolean`,
			`Root.{}.PtrVal:{}.IntVal:integer`,
			`Root.{}.PtrVal:{}.Float64Val:float`,
			`Root.{}.PtrVal:{}.StringVal:string`,
			`Root.{}.PtrPtrVal:{}`,
			`Root.{}.PtrPtrVal:{}.BoolVal:boolean`,
			`Root.{}.PtrPtrVal:{}.IntVal:integer`,
			`Root.{}.PtrPtrVal:{}.Float64Val:float`,
			`Root.{}.PtrPtrVal:{}.StringVal:string`,
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
			`TypeRefs.AStruct:{}.AName:string`,
			`TypeRefs.AStruct:{}.AChild:{}:BStruct`,
			`TypeRefs.BStruct:{}`,
			`TypeRefs.BStruct:{}.BName:string`,
			`TypeRefs.BStruct:{}.BChild:{}:CStruct`,
			`TypeRefs.CStruct:{}`,
			`TypeRefs.CStruct:{}.CName:string`,
			`TypeRefs.CStruct:{}.CChild:{}:AStruct`,
			`TypeRefs.CycleTest:{}`,
			`TypeRefs.CycleTest:{}.Level:integer`,
			`TypeRefs.CycleTest:{}.CycleA:{}:AStruct`,
			`TypeRefs.CycleTest:{}.CycleB:{}:BStruct`,
			`TypeRefs.CycleTest:{}.CycleC:{}`,
			`TypeRefs.CycleTest:{}.CycleC:{}.C:{}:CStruct`,
			`Root.{}:CycleTest`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.Level:integer`,
			`Root.{}.CycleA:{}`,
			`Root.{}.CycleA:{}.AName:string`,
			`Root.{}.CycleA:{}.AChild:{}`,
			`Root.{}.CycleA:{}.AChild:{}.BName:string`,
			`Root.{}.CycleA:{}.AChild:{}.BChild:{}`,
			`Root.{}.CycleA:{}.AChild:{}.BChild:{}.CName:string`,
			`Root.{}.CycleA:{}.AChild:{}.BChild:{}.!CChild:{}:AStruct! ERROR:cyclical reference`,
			`Root.{}.CycleB:{}`,
			`Root.{}.CycleB:{}.BName:string`,
			`Root.{}.CycleB:{}.BChild:{}`,
			`Root.{}.CycleB:{}.BChild:{}.CName:string`,
			`Root.{}.CycleB:{}.BChild:{}.CChild:{}`,
			`Root.{}.CycleB:{}.BChild:{}.CChild:{}.AName:string`,
			`Root.{}.CycleB:{}.BChild:{}.CChild:{}.!AChild:{}:BStruct! ERROR:cyclical reference`,
			`Root.{}.CycleC:{}`,
			`Root.{}.CycleC:{}.C:{}`,
			`Root.{}.CycleC:{}.C:{}.CName:string`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}.AName:string`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}.BName:string`,
			`Root.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}.!BChild:{}:CStruct! ERROR:cyclical reference`,
		},
		jsonStrings: []string{
			`$.{}`,
			`$.{}.cycleA:{}`,
			`$.{}.cycleA:{}.aName:string`,
			`$.{}.cycleA:{}.aChild:{}`,
			`$.{}.cycleA:{}.aChild:{}.bName:string`,
			`$.{}.cycleA:{}.aChild:{}.bChild:{}`,
			`$.{}.cycleA:{}.aChild:{}.bChild:{}.cName:string`,
			`$.{}.cycleA:{}.aChild:{}.bChild:{}.!cChild:{}:AStruct! ERROR:cyclical reference`,
			`$.{}.cycleB:{}`,
			`$.{}.cycleB:{}.bName:string`,
			`$.{}.cycleB:{}.bChild:{}`,
			`$.{}.cycleB:{}.bChild:{}.cName:string`,
			`$.{}.cycleB:{}.bChild:{}.cChild:{}`,
			`$.{}.cycleB:{}.bChild:{}.cChild:{}.aName:string`,
			`$.{}.cycleB:{}.bChild:{}.cChild:{}.!aChild:{}:BStruct! ERROR:cyclical reference`,
			`$.{}.CycleC:{}`,
			`$.{}.CycleC:{}.c:{}`,
			`$.{}.CycleC:{}.c:{}.cName:string`,
			`$.{}.CycleC:{}.c:{}.cChild:{}`,
			`$.{}.CycleC:{}.c:{}.cChild:{}.aName:string`,
			`$.{}.CycleC:{}.c:{}.cChild:{}.aChild:{}`,
			`$.{}.CycleC:{}.c:{}.cChild:{}.aChild:{}.bName:string`,
			`$.{}.CycleC:{}.c:{}.cChild:{}.aChild:{}.!bChild:{}:CStruct! ERROR:cyclical reference`,
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
			`TypeRefs.JSONTagTests:{}.NoTag:string`,
			`TypeRefs.JSONTagTests:{}.ExcludeTag:string`,
			`TypeRefs.JSONTagTests:{}.RenameOne:string`,
			`TypeRefs.JSONTagTests:{}.RenameTwo:string`,
			`Root.{}:JSONTagTests`,
		},
		derefStrings: []string{
			`Root.{}`,
			`Root.{}.NoTag:string`,
			`Root.{}.ExcludeTag:string`,
			`Root.{}.RenameOne:string`,
			`Root.{}.RenameTwo:string`,
		},
		jsonStrings: []string{
			`$.{}`,
			`$.{}.NoTag:string`,
			`$.{}.renameOne:string`,
			`$.{}.something:string`,
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

func jsonPathRender(t *TypeElement, opt *RenderOptions) string {
	// Check root.
	if t.Type == generictype.Root.String() {
		// JSON Path root is "$"
		return "$"
	}

	jsonType := t.GetNativeType("json")
	if jsonType.Include == threeflag.False {
		// Skip this element.
		return ""
	}

	namePart := jsonType.Name
	if namePart != "" {
		namePart += ":"
	}

	// Type.
	var typePart string
	if t.TypeCategory == typecategory.Invalid.String() {
		typePart = t.Type
	} else {
		typePart = generictype.PathDefaultOfType(jsonType.Type)
	}

	// Add TypeRef suffix if set but not if de-referencing.
	// NOTE: json never uses references!
	refPart := ""
	if opt.DeReference && t.Error == CyclicalReferenceErr {
		// Keep reference if it's a cyclical error.
		refPart = jsonType.TypeRef
	}
	if refPart != "" {
		refPart = ":" + refPart
	}

	// Build path.
	path := namePart + typePart + refPart

	// Wrap type in "!" if current element is an error.
	if t.Error != "" {
		path = fmt.Sprintf("!%s!", path)
	}

	return path
}

// jsonPreRender renders a string using the "json" dialect.
func jsonPreRender(t *TypeElement, pathFunc PathStringRenderer, opt *RenderOptions) string {
	jsonType := t.GetNativeType("json")
	if jsonType.Include == threeflag.False {
		// Skip this element.
		return ""
	}

	if jsonType.Type == generictype.Root.String() {
		return ""
	}

	path := t.RenderPath(pathFunc, opt)
	out := path.String()
	if t.Error != "" {
		out += " ERROR:" + t.Error
	}

	return out
}

// preRender renders a string from a TypeElement before Children are processed.
func preRender(t *TypeElement, pathFunc PathStringRenderer, opt *RenderOptions) string {
	if t.Type == generictype.Root.String() {
		return ""
	}

	path := t.RenderPath(pathFunc, opt)
	out := path.String()
	if t.Error != "" {
		out += " ERROR:" + t.Error
	}

	return out
}

// postRender renders a string from a TypeElement after Children are processed.
func postRender(t *TypeElement, pathFunc PathStringRenderer, opt *RenderOptions) string {
	return ""
}

func compareStrings(t *testing.T, testName string, gotStrings, wantStrings []string) {
	if !reflect.DeepEqual(gotStrings, wantStrings) {
		t.Errorf("TEST_FAIL %s", testName)

		maxLen := len(gotStrings)
		if len(wantStrings) > maxLen {
			maxLen = len(wantStrings)
		}

		for i := 0; i < maxLen; i++ {
			got := ""
			if i < len(gotStrings) {
				got = gotStrings[i]
			}

			want := ""
			if i < len(wantStrings) {
				want = wantStrings[i]
			}

			var flag string
			if got != want {
				flag = ">"
			}

			t.Logf("%05d got =%s", i, got)
			t.Logf("%5s want=%s", flag, want)
		}

	} else {
		t.Logf("TEST_OK %s", testName)
	}
}

func runTests(t *testing.T, testCases []TestCase) {
	r := NewReflector()

	for _, test := range testCases {
		r.Reset()
		//r.Label = test.name

		gotResult := r.ReflectTypes(test.value)

		// if b, err := json.MarshalIndent(gotResult, "", "  "); err != nil {
		// 	t.Errorf("TEST_FAIL %s: json.Marshal err=%s", test.name, err)
		// } else {
		// 	fmt.Println(string(b))
		// }

		opt := NewRenderOptions()
		for i := 0; i < 2; i++ {
			opt.DeReference = i == 1

			gotStrings := gotResult.RenderStrings(preRender, postRender, nil, opt)

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
			opt.DeReference = true
			gotStrings := gotResult.RenderStrings(jsonPreRender, postRender, jsonPathRender, opt)
			wantStrings := test.jsonStrings

			testName := fmt.Sprintf("%s: dialect=json", test.name)
			compareStrings(t, testName, gotStrings, wantStrings)
		}
	}
}

func TestReflector_AllTests(t *testing.T) {
	for _, testCases := range allTests {
		runTests(t, testCases)
	}
}
