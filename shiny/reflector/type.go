package reflector

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/gitmann/shiny-reflector-golang/shiny/util"
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

// Copy makes a copy of a TypeElement.
func (t *TypeElement) Copy() *TypeElement {
	n := &TypeElement{
		ID:          t.ID,
		ParentID:    t.ParentID,
		Name:        t.Name,
		Description: t.Description,
		Label:       t.Label,
		Type:        t.Type,
		TypeRef:     t.TypeRef,
		Native:      make(map[string]NativeType),
		Err:         t.Err,
	}

	for langName, nativeMap := range t.Native {
		n.Native[langName] = make(NativeType)
		for k, v := range nativeMap {
			n.Native[langName][k] = v
		}
	}

	return n
}

// Alias returns the alias for the given lang or Name.
func (t *TypeElement) Alias(lang string) string {
	if t.Native != nil {
		if t.Native[lang] != nil {
			if a := t.Native[lang]["Alias"]; a != "" {
				return a
			}
		}
	}
	return t.Name
}

// SetAlias sets the Alias for the native language implementation.
func (t *TypeElement) SetAlias(lang, alias string) {
	if t.Native == nil {
		t.Native = make(map[string]NativeType)
	}
	if a := t.Native[lang]; a == nil {
		t.Native[lang] = make(NativeType)
	}
	t.Native[lang]["Alias"] = alias
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
		if PrintNative {
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

func (r *Reflector) reflectTypeImpl(parentID int, name string, typeList TypeList, a AncestorList, v reflect.Value, s *reflect.StructField, parentErr error) TypeList {
	currentElem := NewTypeElement(r.nextID(), parentID, name, r.Label)
	if parentErr != nil {
		currentElem.Err = parentErr
	}

	// Append current element to master type list.
	typeList = append(typeList, currentElem)

	// Create temporary list for named type refs.
	refList := TypeList{currentElem}

	// Capture native golang features.
	native := make(NativeType)
	currentElem.Native["golang"] = native

	// Check for unsupported types. These do not support many functions without panic.
	switch v.Kind() {
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		currentElem.Type = "invalid"
		currentElem.Err = fmt.Errorf("value.Kind %q not supported", v.Kind())

		native["Kind"] = v.Kind().String()

		return typeList
	}

	// Capture native features from type.
	currentElem = r.reflectGolangType(v.Type(), currentElem)

	// Check cyclical references.
	if a.Has(currentElem.TypeRef) {
		currentElem.Err = fmt.Errorf("cyclical reference: %s", currentElem.TypeRef)
		return typeList
	}
	a.Add(currentElem.TypeRef)

	native["IsZero"] = util.ValueIfTrue(v.IsZero(), "z", "-")
	native["IsValid"] = util.ValueIfTrue(v.IsValid(), "v", "-")
	native["IsNil"] = "-"

	// Handle un-exported struct fields.
	var vValue interface{}
	if s != nil && s.PkgPath != "" {
		native["PkgPath"] = s.PkgPath
	} else {
		vValue = v.Interface()
	}

	// Parse struct tags.
	if s != nil {
		tags := ParseTags(s.Tag)
		if len(tags) > 0 {
			for tagName, tagVal := range tags {
				tagMap := tagVal.AsMap()
				if tagMap != nil {
					tempNative := currentElem.Native[tagName]
					if tempNative == nil {
						// Set new native block.
						currentElem.Native[tagName] = tagMap
					} else {
						// Copy from tagMap into existing block.
						for k, v := range tagMap {
							tempNative[k] = v
						}
					}
				}
			}
		}
	}

	// Implement special JSON parsing rules.
	if ParseAsJSON {
		exportName := util.Capitalize(name)
		if exportName != name {
			// Use exportName as element name and save unexported name as JSON Alias.
			currentElem.Name = exportName

		}
	}

	// TODO: ignore vValue for now
	_ = vValue

	// Get features that vary by value.Kind
	switch v.Kind() {
	// string, integer, float, and boolean are simple types with no children.
	case reflect.String:
		currentElem.Type = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		currentElem.Type = "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		currentElem.Type = "integer"
	case reflect.Float32, reflect.Float64:
		currentElem.Type = "float"
	case reflect.Bool:
		currentElem.Type = "boolean"

	case reflect.Interface:
		currentElem.Type = "interface"

		if v.IsZero() {
			// Zero value for interface means nil.
			currentElem.Type = "invalid"
			currentElem.Err = fmt.Errorf("interface element is nil")
		} else {
			// Non-Zero interface is just an extra layer of abstraction around a real type.
			// Remove interface from typeList and reflect child element.
			typeList = typeList[:len(typeList)-1]
			typeList = r.reflectTypeImpl(parentID, name, typeList, a.Copy(), v.Elem(), nil, parentErr)
		}

	// Pointer is an intermediate type that has no value and points to one child.
	case reflect.Ptr:
		currentElem.Type = "pointer"
		native["IsNil"] = util.ValueIfTrue(v.IsNil(), "n", "-")

		if currentElem.Err == nil {
			// Get target of pointer.
			var targetValue reflect.Value
			if v.IsNil() {
				// Create a new value if pointer is nil.
				targetValue = reflect.New(v.Type().Elem()).Elem()
			} else {
				// Use existing value for valid pointer.
				targetValue = v.Elem()
			}
			refList = r.reflectTypeImpl(currentElem.ID, "", refList, a.Copy(), targetValue, nil, nil)
		}

	// Array and Slice represent lists of elements.
	// - 1st element of list will be used to determine element type
	// - If list is empty, a one-element list will be created to use for typing.
	case reflect.Array:
		currentElem.Type = "list"

		if currentElem.Err == nil {
			//	Get kind of underlying elements.
			var targetValue reflect.Value
			if v.Len() > 0 {
				targetValue = v.Index(0)
			} else {
				targetValue = reflect.New(v.Type().Elem()).Elem()
			}
			refList = r.reflectTypeImpl(currentElem.ID, "", refList, a.Copy(), targetValue, nil, nil)
		}

	case reflect.Slice:
		currentElem.Type = "list"
		native["IsNil"] = util.ValueIfTrue(v.IsNil(), "n", "-")

		if currentElem.Err == nil {
			//	Get kind of underlying elements.
			var sliceElem reflect.Value
			if v.IsNil() || v.Len() == 0 {
				sliceElem = reflect.MakeSlice(v.Type(), 1, 1).Index(0)
			} else {
				sliceElem = v.Index(0)
			}

			refList = r.reflectTypeImpl(currentElem.ID, "", refList, a.Copy(), sliceElem, nil, nil)
		}

	// Struct and Map represent key-value pairs.
	// - Struct keys are field names which are always strings.
	// - Map keys can be any comprable Go type.
	case reflect.Struct:
		currentElem.Type = "struct"

		if currentElem.Err == nil {
			for i := 0; i < v.NumField(); i++ {
				structField := v.Type().Field(i)
				targetValue := v.Field(i)

				refList = r.reflectTypeImpl(currentElem.ID, structField.Name, refList, a.Copy(), targetValue, &structField, nil)
			}
		}

	case reflect.Map:
		currentElem.Type = "struct"
		native["IsNil"] = util.ValueIfTrue(v.IsNil(), "n", "-")

		if currentElem.Err == nil {
			// Map key must be a string.
			if v.Type().Key().Kind() != reflect.String {
				currentElem.Err = fmt.Errorf("map key type %q not supported", v.Type().Key())
			}

			// Empty map not allowed.
			if v.Len() == 0 {
				currentElem.Err = fmt.Errorf("empty map not supported")
			}

			// Iterate through map by keys in sorted order.
			// - ExportName must be unique.
			//   - If ParseAsJSON is true, ExportName is the capitalized Name
			type mapKey struct {
				Name       string
				ExportName string
				Value      reflect.Value
			}
			keys := []*mapKey{}
			for _, k := range v.MapKeys() {
				newKey := &mapKey{
					Name:  k.Interface().(string),
					Value: k,
				}
				if ParseAsJSON {
					newKey.ExportName = util.Capitalize(newKey.Name)
				} else {
					newKey.ExportName = newKey.Name
				}

				keys = append(keys, newKey)
			}
			sort.Slice(keys, func(i, j int) bool {
				if keys[i].ExportName == keys[j].ExportName {
					return keys[i].Name < keys[j].Name
				}
				return keys[i].ExportName < keys[j].ExportName
			})

			uniqKeys := map[string]int{}
			for _, k := range keys {
				mapKeyName := k.Name
				mapValue := v.MapIndex(k.Value)

				var duplicateErr error
				if uniqKeys[k.ExportName] > 0 {
					duplicateErr = fmt.Errorf("duplicate map key %q (%q)", k.ExportName, k.Name)
				}
				uniqKeys[k.ExportName]++

				refList = r.reflectTypeImpl(currentElem.ID, mapKeyName, refList, a.Copy(), mapValue, nil, duplicateErr)
			}
		}

	default:
		// All other types should be handled in the unsupported check above.
		panic(fmt.Sprintf("value.Kind %q not supported", v.Kind()))
	}

	// If current element is a named type, add to typeRefs.
	if currentElem.TypeRef != "" {
		if r.typeResult.TypeRefs[currentElem.TypeRef] == nil {
			// Copy all elements with new IDs and no errors.
			// - Stop at the 2nd element that is a TypeRef.
			newList := make(TypeList, 0)
			for i, listItem := range refList {
				newItem := listItem.Copy()
				newItem.ID = r.nextID()
				newItem.Err = nil
				if i == 0 {
					newItem.Name = ""
				}

				newList = append(newList, newItem)

				//	Stop on the 2nd element with a RefType
				if i > 0 && newItem.TypeRef != "" {
					break
				}
			}

			r.typeResult.TypeRefs[currentElem.TypeRef] = newList
		}
	}

	if len(refList) > 1 {
		if NoRefs || currentElem.TypeRef == "" {
			// Only add refList if it is a compound type.
			typeList = append(typeList, refList[1:]...)
		}
	}

	return typeList
}

// reflectGolangType parses native features for Go.
func (r *Reflector) reflectGolangType(t reflect.Type, currentElem *TypeElement) *TypeElement {
	// Initialize nil variables.
	if currentElem == nil {
		currentElem = NewTypeElement(0, 0, "", "")
	}

	native := currentElem.Native["golang"]
	if native == nil {
		native = make(NativeType)
		currentElem.Native["golang"] = native
	}

	// Type AName is the name of a type if any.
	if t.Name() != t.Kind().String() {
		currentElem.TypeRef = t.Name()
	}

	// Capture native features.
	native["Kind"] = t.Kind().String()
	native["Type"] = t.String()
	native["Type.AName"] = t.Name()
	native["Type.Kind"] = t.Kind().String()

	return currentElem
}

// AncestorList keeps track of type references that are ancestors of the current element.
// - Stores a count of references found.
// - If count > 1, a cyclical reference exists.
type AncestorList map[string]int

// NewAncestorList initializes a new ancestor list.
func NewAncestorList() AncestorList {
	return make(AncestorList)
}

// Copy makes a copy of the ancestor list.
func (a AncestorList) Copy() AncestorList {
	n := make(AncestorList)
	for k, v := range a {
		n[k] = v
	}
	return n
}

// Has returns true if the key exists in ancestor list.
func (a AncestorList) Has(key string) bool {
	return a[key] > 0
}

// Add adds a reference count to the ancestor list.
func (a AncestorList) Add(key string) int {
	if key == "" {
		return 0
	}

	a[key]++
	return a[key]
}
