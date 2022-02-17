package reflector

import (
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"
)

var testCases = []struct {
	name  string
	value interface{}
}{
	{name: "nil", value: nil},
	{name: "complex", value: complex(123, 456)},
	{name: "chan", value: make(chan bool)},
	{name: "func", value: func() {}},
	{name: "uintptr", value: uintptr(123)},
	{name: "unsafeptr", value: func() interface{} { s := "hello"; return unsafe.Pointer(&s) }()},
	{name: "map[GoodEntity]bool, nil", value: func() interface{} { var s map[GoodEntity]bool; return s }()},
	{name: "map[string]bool, nil", value: func() interface{} { var s map[string]bool; return s }()},
	{name: "map[string]bool, empty", value: map[string]bool{}},

	{name: "string, var", value: func() interface{} { var s string; return s }()},
	{name: "string, empty", value: ""},
	{name: "string, value", value: "hello"},

	{name: "int64, var", value: func() interface{} { var s int64; return s }()},
	{name: "int64, empty", value: int64(0)},
	{name: "int64, value", value: int64(123)},

	{name: "float64, var", value: func() interface{} { var s float64; return s }()},
	{name: "float64, empty", value: float64(0)},
	{name: "float64, value", value: float64(456.789)},

	{name: "bool, var", value: func() interface{} { var s bool; return s }()},
	{name: "bool, empty", value: false},
	{name: "bool, value", value: true},

	{name: "[]string, var", value: func() interface{} { var s []string; return s }()},
	{name: "[]string, nil", value: ([]string)(nil)},
	{name: "[]string, empty", value: []string{}},
	{name: "[]string, value", value: []string{"hello", "hey", "hi"}},

	{name: "[0]string, var", value: func() interface{} { var s [0]string; return s }()},
	{name: "[0]string, nil", value: [0]string{}},

	{name: "[3]string, var", value: func() interface{} { var s [3]string; return s }()},
	{name: "[3]string, nil", value: [3]string{}},
	{name: "[3]string, value", value: [3]string{"hello", "hey", "hi"}},

	{name: "map[string]string, var", value: func() interface{} { var s map[string]string; return s }()},
	{name: "map[string]string, nil", value: map[string]string{}},
	{name: "map[string]string, value", value: map[string]string{"hello": "one", "hey": "two", "hi": "three"}},

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

	{name: "CycleTest, empty", value: &CycleTest{}},

	{name: "*TypeElement, values", value: &TypeElement{
		ID:          123,
		ParentID:    234,
		Name:        "hello",
		Description: "blah",
		Label:       "something",
		Type:        "test",
		TypeRef:     "testRef",
	}},

	{name: "makeJSON, value", value: makeJSON(nil)},
}

func TestReflector_ReflectTypes(t *testing.T) {
	r := NewReflector()

	// Configure package-level settings.
	PrintNative = false
	NoRefs = true
	ParseAsJSON = true

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r.Reset()
			r.Label = test.name

			gotResult := r.ReflectTypes(test.value)
			fmt.Println(gotResult.BuildString("json-list"))

			t.Logf("TEST_OK %s", test.name)
		})
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

type CycleTest struct {
	Level  int
	CycleA *AStruct
	CycleB *BStruct
	CycleC *CStruct
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
		var x interface{}
		if err := json.Unmarshal(b, &x); err != nil {
			return fmt.Errorf("ERROR json.Unmarshal: %s", err)
		}

		// DEBUGXXXXX Print indented JSON string.
		if x != nil {
			if b, err := json.MarshalIndent(x, "", "  "); err == nil {
				fmt.Println(string(b))
			}
		}

		return x
	}
}
