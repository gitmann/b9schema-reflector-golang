package reflector

import (
	"fmt"
	"github.com/gitmann/shiny-reflector-golang/v2/shiny/util"
	"reflect"
	"sort"
)

func (r *Reflector) reflectValueImpl(parentID int, name string, typeList TypeList, v reflect.Value, s *reflect.StructField) TypeList {
	currentElem := NewTypeElement(r.nextID(), parentID, name)

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
	currentElem = r.reflectTypeImpl(v.Type(), currentElem)

	native["IsZero"] = util.ValueIfTrue(v.IsZero(), "z", "-")
	native["IsValid"] = util.ValueIfTrue(v.IsValid(), "v", "-")
	native["IsNil"] = "-"

	// Handle un-exported fields.
	var vValue interface{}
	if s != nil && s.PkgPath != "" {
		native["PkgPath"] = s.PkgPath
	} else {
		vValue = v.Interface()
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
			typeList = r.reflectValueImpl(parentID, name, typeList, v.Elem(), nil)
		}

	// Pointer is an intermediate type that has no value and points to one child.
	case reflect.Ptr:
		currentElem.Type = "pointer"
		native["IsNil"] = util.ValueIfTrue(v.IsNil(), "n", "-")

		// Get target of pointer.
		var targetValue reflect.Value
		if v.IsNil() {
			// Create a new value if pointer is nil.
			targetValue = reflect.New(v.Type().Elem()).Elem()
		} else {
			// Use existing value for valid pointer.
			targetValue = v.Elem()
		}
		refList = r.reflectValueImpl(currentElem.ID, "", refList, targetValue, nil)

	// Array and Slice represent lists of elements.
	// - 1st element of list will be used to determine element type
	// - If list is empty, a one-element list will be created to use for typing.
	case reflect.Array:
		currentElem.Type = "list"

		//	Get kind of underlying elements.
		var targetValue reflect.Value
		if v.Len() > 0 {
			targetValue = v.Index(0)
		} else {
			targetValue = reflect.New(v.Type().Elem()).Elem()
		}
		refList = r.reflectValueImpl(currentElem.ID, "", refList, targetValue, nil)

	case reflect.Slice:
		currentElem.Type = "list"
		native["IsNil"] = util.ValueIfTrue(v.IsNil(), "n", "-")

		//	Get kind of underlying elements.
		var sliceElem reflect.Value
		if v.IsNil() || v.Len() == 0 {
			sliceElem = reflect.MakeSlice(v.Type(), 1, 1).Index(0)
		} else {
			sliceElem = v.Index(0)
		}

		refList = r.reflectValueImpl(currentElem.ID, "", refList, sliceElem, nil)

	// Struct and Map represent key-value pairs.
	// - Struct keys are field names which are always strings.
	// - Map keys can be any comprable Go type.
	case reflect.Struct:
		currentElem.Type = "struct"

		for i := 0; i < v.NumField(); i++ {
			structField := v.Type().Field(i)
			targetValue := v.Field(i)

			refList = r.reflectValueImpl(currentElem.ID, structField.Name, refList, targetValue, &structField)
		}

	case reflect.Map:
		currentElem.Type = "struct"
		native["IsNil"] = util.ValueIfTrue(v.IsNil(), "n", "-")

		// Map key must be a string.
		if v.Type().Key().Kind() != reflect.String {
			currentElem.Err = fmt.Errorf("map key type %q not supported", v.Type().Key())
		}

		// Empty map not allowed.
		if v.Len() == 0 {
			currentElem.Err = fmt.Errorf("empty map not supported")
		}

		// Iterate through map by keys in sorted order.
		type mapKey struct {
			Name  string
			Value reflect.Value
		}
		keys := []*mapKey{}
		for _, k := range v.MapKeys() {
			newKey := &mapKey{
				Name:  k.Interface().(string),
				Value: k,
			}
			keys = append(keys, newKey)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i].Name < keys[j].Name })

		for _, k := range keys {
			mapKeyName := k.Name
			mapValue := v.MapIndex(k.Value)

			refList = r.reflectValueImpl(currentElem.ID, mapKeyName, refList, mapValue, nil)
		}

	default:
		// All other types should be handled in the unsupported check above.
		panic(fmt.Sprintf("value.Kind %q not supported", v.Kind()))
	}

	// If current element is a named type, add to typeRefs. Otherwise, add to typeList.
	if currentElem.TypeRef != "" {
		r.typeResult.TypeRefs[currentElem.TypeRef] = refList
	} else if len(refList) > 1 {
		// Only add refList if it is a compound type.
		typeList = append(typeList, refList[1:]...)
	}

	return typeList
}

func (r *Reflector) reflectTypeImpl(t reflect.Type, currentElem *TypeElement) *TypeElement {
	// Initialize nil variables.
	if currentElem == nil {
		currentElem = NewTypeElement(0, 0, "")
	}

	native := currentElem.Native["golang"]
	if native == nil {
		native = make(NativeType)
		currentElem.Native["golang"] = native
	}

	// Type Name is the name of a type if any.
	if t.Name() != t.Kind().String() {
		currentElem.TypeRef = t.Name()
	}

	// Capture native features.
	native["Kind"] = t.Kind().String()
	native["Type"] = t.String()
	native["Type.Name"] = t.Name()
	native["Type.Kind"] = t.Kind().String()

	return currentElem
}
