package reflector

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/gitmann/shiny-reflector-golang/b9schema/enum/generictype"
)

var allTests = [][]TestCase{
	rootJSONTests,
	rootGoTests,

	// invalidTests,
	// specialTests,
	//basicTests,
	//arrayTests,
	//sliceTests,
	//mapTests,
	// structTests,
	//interfaceTests,
	//pointerTests,
	//jsonTests,
}

type TestCase struct {
	name  string
	value interface{}

	wantStrings []string
}

// *** All reflect types ***

// rootTests validate that the top-level element is either a struct or Reference.
var rootJSONTests = []TestCase{
	{name: "json-null", value: fromJSON([]byte(`null`)), wantStrings: []string{"Root.!invalid! ERROR:kind not supported"}},

	{name: "json-string", value: fromJSON([]byte(`"Hello"`)), wantStrings: []string{"Root.!string! ERROR:root type must be a struct"}},
	{name: "json-int", value: fromJSON([]byte(`123`)), wantStrings: []string{"Root.!float! ERROR:root type must be a struct"}},
	{name: "json-float", value: fromJSON([]byte(`234.345`)), wantStrings: []string{"Root.!float! ERROR:root type must be a struct"}},
	{name: "json-bool", value: fromJSON([]byte(`true`)), wantStrings: []string{"Root.!boolean! ERROR:root type must be a struct"}},

	{name: "json-list-empty", value: fromJSON([]byte(`[]`)), wantStrings: []string{"Root.![]! ERROR:root type must be a struct"}},
	{name: "json-list", value: fromJSON([]byte(`[1,2,3]`)), wantStrings: []string{"Root.![]! ERROR:root type must be a struct"}},

	{name: "json-object-empty", value: fromJSON([]byte(`{}`)), wantStrings: []string{"Root.!{}! ERROR:empty map not supported"}},

	{
		name:  "json-object",
		value: fromJSON([]byte(`{"key1":"Hello"}`)),
		wantStrings: []string{
			"Root.{}",
			"Root.{}.Key1:string",
		},
	},
}

var rootGoTests = []TestCase{
	{name: "golang-nil", value: nil, wantStrings: []string{"Root.!invalid! ERROR:kind not supported"}},

	{name: "golang-string", value: "Hello", wantStrings: []string{"Root.!string! ERROR:root type must be a struct"}},
	{name: "golang-int", value: 123, wantStrings: []string{"Root.!integer! ERROR:root type must be a struct"}},
	{name: "golang-float", value: 234.345, wantStrings: []string{"Root.!float! ERROR:root type must be a struct"}},
	{name: "golang-bool", value: true, wantStrings: []string{"Root.!boolean! ERROR:root type must be a struct"}},

	{name: "golang-array-0", value: [0]string{}, wantStrings: []string{"Root.![]! ERROR:root type must be a struct"}},
	{name: "golang-array-3", value: [3]string{}, wantStrings: []string{"Root.![]! ERROR:root type must be a struct"}},

	{name: "golang-slice-nil", value: func() interface{} { var s []string; return s }(), wantStrings: []string{"Root.![]! ERROR:root type must be a struct"}},
	{name: "golang-slice-0", value: []string{}, wantStrings: []string{"Root.![]! ERROR:root type must be a struct"}},
	{name: "golang-slice-3", value: make([]string, 3), wantStrings: []string{"Root.![]! ERROR:root type must be a struct"}},

	{name: "golang-struct-empty", value: func() interface{} { var s struct{}; return s }(), wantStrings: []string{"Root.!{}! ERROR:empty struct not supported"}},

	{
		name:  "golang-struct-noinit",
		value: func() interface{} { var s StringStruct; return s }(),
		wantStrings: []string{
			"TypeRefs.StringStruct:{}",
			"TypeRefs.StringStruct:{}.Value:string",
			"Root.{}:StringStruct",
			"Root.{}:StringStruct.Value:string",
		},
	},
	{
		name:  "golang-struct-init",
		value: StringStruct{},
		wantStrings: []string{
			"TypeRefs.StringStruct:{}",
			"TypeRefs.StringStruct:{}.Value:string",
			"Root.{}:StringStruct",
			"Root.{}:StringStruct.Value:string",
		},
	},
	{
		name:  "golang-struct-private",
		value: PrivateStruct{},
		wantStrings: []string{
			"TypeRefs.!PrivateStruct:{}! ERROR:struct has no exported fields",
			"Root.!{}:PrivateStruct! ERROR:struct has no exported fields",
		},
	},

	{
		name:  "golang-interface-struct-noinit",
		value: func() interface{} { var s interface{} = StringStruct{}; return s }(),
		wantStrings: []string{
			"TypeRefs.StringStruct:{}",
			"TypeRefs.StringStruct:{}.Value:string",
			"Root.{}:StringStruct",
			"Root.{}:StringStruct.Value:string",
		},
	},
	{
		name:  "golang-pointer-struct-noinit",
		value: func() interface{} { var s *StringStruct; return s }(),
		wantStrings: []string{
			"TypeRefs.StringStruct:{}",
			"TypeRefs.StringStruct:{}.Value:string",
			"Root.{}:StringStruct",
			"Root.{}:StringStruct.Value:string",
		},
	},
	{
		name:  "golang-pointer-struct-init",
		value: &StringStruct{},
		wantStrings: []string{
			"TypeRefs.StringStruct:{}",
			"TypeRefs.StringStruct:{}.Value:string",
			"Root.{}:StringStruct",
			"Root.{}:StringStruct.Value:string",
		},
	},
}

// Invalid types for shiny schemas.
// - Invalid
// - Complex64
// - Complex128
// - Chan
// - Func
// - UnsafePointer
var invalidTests = []TestCase{
	{name: "nil", value: nil},
	{name: "complex", value: complex(123, 456)},
	{name: "complex", value: complex64(complex(123, 456))},
	{name: "complex", value: complex128(complex(123, 456))},
	{name: "chan", value: make(chan bool)},
	{name: "func", value: func() {}},
	{name: "unsafeptr", value: func() interface{} { s := "hello"; return unsafe.Pointer(&s) }()},
}

// Basic types for shiny schemas.
//Bool
//Int
//Int8
//Int16
//Int32
//Int64
//Uint
//Uint8
//Uint16
//Uint32
//Uint64
//Uintptr
//Float32
//Float64
//String
var basicTests = []TestCase{
	{name: "bool-var", value: func() interface{} { var s bool; return s }()},
	{name: "bool-value", value: true},

	{name: "int-var", value: func() interface{} { var s int; return s }()},
	{name: "int-value", value: int(123)},

	{name: "int8-var", value: func() interface{} { var s int8; return s }()},
	{name: "int8-value", value: int8(123)},

	{name: "int16-var", value: func() interface{} { var s int16; return s }()},
	{name: "int16-value", value: int16(123)},

	{name: "int32-var", value: func() interface{} { var s int32; return s }()},
	{name: "int32-value", value: int32(123)},

	{name: "int64-var", value: func() interface{} { var s int64; return s }()},
	{name: "int64-value", value: int64(123)},

	{name: "uint-var", value: func() interface{} { var s uint; return s }()},
	{name: "uint-value", value: uint(123)},

	{name: "uint8-var", value: func() interface{} { var s uint8; return s }()},
	{name: "uint8-value", value: uint8(123)},

	{name: "int16-var", value: func() interface{} { var s int16; return s }()},
	{name: "int16-value", value: int16(123)},

	{name: "int32-var", value: func() interface{} { var s int32; return s }()},
	{name: "int32-value", value: int32(123)},

	{name: "uint64-var", value: func() interface{} { var s uint64; return s }()},
	{name: "uint64-value", value: uint64(123)},

	{name: "uintptr-var", value: func() interface{} { var s uintptr; return s }()},
	{name: "uintptr-value", value: uintptr(123)},

	{name: "float32-var", value: func() interface{} { var s float32; return s }()},
	{name: "float32-value", value: float32(234.345)},

	{name: "float64-var", value: func() interface{} { var s float64; return s }()},
	{name: "float64-value", value: float64(234.345)},

	{name: "string-var", value: func() interface{} { var s string; return s }()},
	{name: "string-value", value: "hello"},
}

// Special types from protobuf: https://developers.google.com/protocol-buffers/docs/reference/google.protobuf
var specialTests = []TestCase{
	// Duration
	// {name: "duration-var", value: func() interface{} { var s time.Duration; return s }()},
	// {name: "duration-value", value: time.Minute},

	// {name: "datetime-var", value: func() interface{} { var s time.Time; return s }()},
	{name: "datetime-value", value: time.Now()},
}

// Array tests.
var arrayTests = []TestCase{
	{name: "[0]string-var", value: func() interface{} { var s [0]string; return s }()},
	{name: "[0]string-nil", value: [0]string{}},

	{name: "[3]string-var", value: func() interface{} { var s [3]string; return s }()},
	{name: "[3]string-nil", value: [3]string{}},
	{name: "[3]string-value", value: [3]string{"hello", "hey", "hi"}},

	{name: "[2][3]string-var", value: func() interface{} { var s [2][3]string; return s }()},
	{name: "[2][3]string-nil", value: [2][3]string{}},
	{name: "[2][3]string-value", value: [2][3]string{[3]string{"hello", "hey", "hi"}}},
}

// Slice tests.
var sliceTests = []TestCase{
	{name: "[]string-var", value: func() interface{} { var s []string; return s }()},
	{name: "[]string-nil", value: ([]string)(nil)},
	{name: "[]string-empty", value: []string{}},
	{name: "[]string-value", value: []string{"hello", "hey", "hi"}},
	{name: "[][]string-value", value: [][]string{[]string{"hello", "hey", "hi"}}},
}

var mapTests = []TestCase{
	{name: "map[StringStruct]bool-nil", value: func() interface{} { var s map[StringStruct]bool; return s }()},
	{name: "map[string]bool-nil", value: func() interface{} { var s map[string]bool; return s }()},
	{name: "map[string]bool-empty", value: map[string]bool{}},
	{name: "map[string]bool-value", value: map[string]bool{"One": true, "two": false, "Three": false, "four": true}},
	{name: "map[string]interface-value", value: map[string]interface{}{"One": true, "two": false, "Three": false, "four": true}},
	{name: "map[string]map[string]bool-empty", value: map[string]map[string]bool{}},
}

var structTests = []TestCase{
	// {name: "struct-empty", value: func() interface{} { var g struct{}; return g }()},
	// {name: "PrivateStruct-nil", value: func() interface{} { var g PrivateStruct; return g }()},
	{name: "BasicStruct-nil", value: func() interface{} { var g BasicStruct; return g }()},
	// {name: "CompoundStruct-nil", value: func() interface{} { var g CompoundStruct; return g }()},
	// {name: "CycleTest-nil", value: func() interface{} { var g CycleTest; return g }()},
}

var interfaceTests = []TestCase{
	{name: "interface{}-var", value: func() interface{} { var g interface{}; return g }()},
	{name: "interface{}-nil", value: nil},
	{name: "interface{}-bool", value: true},
	{name: "interface{}-int", value: 123},
	{name: "interface{}-float", value: 234.345},
	{name: "interface{}-string", value: "hello"},
}

var pointerTests = []TestCase{
	{name: "*bool", value: func() interface{} { var g bool; return &g }()},
	{name: "*int", value: func() interface{} { var g int; return &g }()},
	{name: "*float", value: func() interface{} { var g float64; return &g }()},
	{name: "*string", value: func() interface{} { var g string; return &g }()},
	{name: "*array", value: func() interface{} { var g [0]string; return &g }()},
	{name: "*slice", value: func() interface{} { var g []string; return &g }()},
	{name: "*map", value: func() interface{} { var g map[string]string; return &g }()},
	{name: "*StringStruct-var", value: func() interface{} { var g *StringStruct; return g }()},
	{name: "**StringStruct-var", value: func() interface{} { var g *StringStruct; return &g }()},
}

var jsonTests = []TestCase{
	{name: "json-null", value: fromJSON([]byte(`null`))},
	{name: "json-string", value: fromJSON([]byte(`"hello"`))},
	{name: "json-int", value: fromJSON([]byte(`123`))},
	{name: "json-float", value: fromJSON([]byte(`234.345`))},
	{name: "json-bool", value: fromJSON([]byte(`true`))},
	{name: "json-list", value: fromJSON([]byte(`[123,234,345]`))},
	{name: "json-list-mixed", value: fromJSON([]byte(`["hello",123,234.345,true]`))},
	{name: "json-object", value: fromJSON([]byte(`{"key1":"val1","key2":123,"key3":false}`))},
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
	Int8Val    int8
	Int16Val   int16
	Int32Val   int32
	Int64Val   int64
	UintVal    uint
	Uint8Val   uint8
	Uint16Val  uint16
	Uint32Val  uint32
	Uint64Val  uint64
	UintptrVal uintptr
	Float32Val float32
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

// Test cyclical relationships:
// A --> B --> C --> A
type AStruct struct {
	AName  string
	AChild *BStruct
}

type BStruct struct {
	BName  string
	BChild *CStruct
}

type CStruct struct {
	CName  string
	CChild *AStruct
}

type BadType interface{}

type CycleTest struct {
	Level  int
	BadVal BadType
	CycleA AStruct
	CycleB *BStruct
	CycleC struct {
		C CStruct
	}
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
		return fmt.Errorf("ERROR json.Unmarshal: %s", err)
	}

	// DEBUGXXXXX Print indented JSON string.
	if out != nil {
		if b, err := json.MarshalIndent(out, "", "  "); err == nil {
			fmt.Println(string(b))
		}
	}

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

// pathRender renders a PathString from a TypeElement.
func pathRender(t *TypeElement) string {
	return ""
}

// preRender renders a string from a TypeElement before Children are processed.
func preRender(t *TypeElement, pathFunc PathStringRenderer) string {
	if t.Type == generictype.Root.String() {
		return ""
	}

	path := t.RenderPath(pathFunc)
	out := path.String()
	if t.Error != "" {
		out += " ERROR:" + t.Error
	}

	return out
}

// postRender renders a string from a TypeElement after Children are processed.
func postRender(t *TypeElement, pathFunc PathStringRenderer) string {
	return ""
}

func runTests(t *testing.T, testCases []TestCase) {
	r := NewReflector()

	// Configure package-level settings.
	PrintNative = true
	PathPrefix = false
	DeReference = false
	ParseAsJSON = true

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r.Reset()
			//r.Label = test.name

			gotResult := r.ReflectTypes(test.value)

			// if b, err := json.MarshalIndent(gotResult, "", "  "); err != nil {
			// 	t.Errorf("TEST_FAIL %s: json.Marshal err=%s", test.name, err)
			// } else {
			// 	fmt.Println(string(b))
			// }

			gotStrings := gotResult.RenderStrings(preRender, postRender, nil)
			if !reflect.DeepEqual(gotStrings, test.wantStrings) {
				t.Errorf("TEST_FAIL %s", test.name)

				maxLen := len(gotStrings)
				if len(test.wantStrings) > maxLen {
					maxLen = len(test.wantStrings)
				}

				for i := 0; i < maxLen; i++ {
					got := ""
					if i < len(gotStrings) {
						got = gotStrings[i]
					}

					want := ""
					if i < len(test.wantStrings) {
						want = test.wantStrings[i]
					}

					t.Logf("%05d got =%s", i, got)
					t.Logf("      want=%s", want)
				}

			} else {
				t.Logf("TEST_OK %s", test.name)
			}
		})
	}
}

func TestReflector_AllTests(t *testing.T) {
	for _, testCases := range allTests {
		runTests(t, testCases)
	}
}
