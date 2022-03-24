package reflector

import (
	"fmt"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/enum/generictype"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/enum/threeflag"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/enum/typecategory"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/idgen"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/types"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/util"
	"reflect"
	"sort"
	"strings"
)

const (
	NATIVE_DIALECT = "golang"
)

// Reflector provides functions to build type and values from a Go value.
type Reflector struct {
	// Keep track of refs found during parsing.
	Schema *types.Schema
}

func NewReflector() *Reflector {
	r := &Reflector{}

	r.Reset()

	return r
}

func (r *Reflector) Reset() *Reflector {
	// Initialize state.
	idgen.Reset()

	r.Schema = &types.Schema{
		Root:     types.NewRootElement("Root", NATIVE_DIALECT),
		TypeRefs: types.NewRootElement("TypeRefs", NATIVE_DIALECT),
	}

	// Return *Reflector for chaining.
	return r
}

// DeriveSchema builds a reflector list of elements from the given interface.
func (r *Reflector) DeriveSchema(x interface{}) *types.Schema {
	if r.Schema == nil {
		r.Reset()
	}

	// Start recursive reflection.
	r.reflectTypeImpl(types.NewAncestorTypeRef(), r.Schema.Root.NewChild(""), reflect.ValueOf(x), nil)

	return r.Schema
}

// reflectTypeImpl is a recursive function to reflect Go values.
//
// Args:
// - typeList (TypeList): list of TypeElement found so far
// - ancestoreTypeRef (AncestorTypeRef): keeps track of TypeRef names seen so far, used for cycle detection
// - currentElem (*types.TypeElement): current TypeElement, must be initialized in caller!
// - v (reflect.Value): Value of current element
// - s (*reflect.StructField): pointer to StructField for current element if part of a struct
//
// Returns:
// - TypeList: list of TypeElement after reflection
func (r *Reflector) reflectTypeImpl(ancestorTypeRef types.AncestorTypeRef, currentElem *types.TypeElement, v reflect.Value, s *reflect.StructField) {
	// currentElem must be initialized in caller!!!
	if currentElem == nil {
		panic("currentElem cannot be nil")
	}

	// Create temporary list for named type refs.
	refList := types.NewTypeList()
	refList.Push(currentElem)

	// Capture native golang features.
	native := currentElem.NativeDefault()
	native.Options.AddKeyVal("Kind", v.Kind().String())

	// Get generic type for value.
	genericType := generictype.GenericTypeOf(v)
	currentElem.Type = genericType.String()
	currentElem.TypeCategory = genericType.Category().String()

	// ERROR CHECKING
	// Check for invalid types. These may panic on some operations so we exit quickly with minimal reflection.
	if genericType.Category() == typecategory.Invalid {
		currentElem.Error = types.InvalidKindErr

		if v == reflect.ValueOf(nil) {
			currentElem.Type = currentElem.Type + ":nil"
		} else {
			currentElem.Type = currentElem.Type + ":" + v.Kind().String()
		}

		return
	}

	// If parent is a root, the current element must be a struct or a Reference.
	if currentElem.Parent == nil {
		panic("parent is nil")
	} else if currentElem.Parent.Type == generictype.Root.String() {
		if genericType != generictype.Struct && genericType.Category() != typecategory.Reference {
			currentElem.Error = types.RootKindErr
			return
		}
	}

	// Capture Go-specific attributes common to all types.
	native.Options.AddBool("IsZero", v.IsZero())
	native.Options.AddBool("IsValid", v.IsValid())
	native.Options.AddThreeFlag("IsNil", threeflag.Undefined)
	native.Type = v.Kind().String()
	native.Options.AddKeyVal("Type.Name", v.Type().Name())
	native.Options.AddKeyVal("Type.Kind", v.Type().Kind().String())
	native.Options.AddKeyVal("Type.PkgPath", v.Type().PkgPath())

	// If type.Name differs from type.Kind, element is a TypeRef.
	if v.Type().Name() != v.Type().Kind().String() {
		currentElem.TypeRef = v.Type().Name()

		native.TypeRef = currentElem.TypeRef
		native.Options.AddKeyVal("TypeRef", currentElem.TypeRef)

		// Check for cyclical references.
		if ancestorTypeRef.Contains(currentElem.TypeRef) {
			currentElem.Error = types.CyclicalReferenceErr
			return
		}
		ancestorTypeRef.Add(currentElem.TypeRef)
	}

	// Capture attributes that differ by type.
	unhandledType := false
	switch genericType.Category() {
	case typecategory.Basic:
		// Basic types are already handled by the default operations above. Nothing else to do here.
	case typecategory.Known:
		// Known types are already handled by the default operations above. However, TypeRef should be removed.
		currentElem.TypeRef = ""
		native.TypeRef = ""
	case typecategory.Compound:
		switch genericType {
		// Compound types are reflected in their own functions. Capture ref list for processing below.
		case generictype.List:
			r.reflectTypeListImpl(ancestorTypeRef, currentElem, v, s)
		case generictype.Struct:
			r.reflectTypeStructImpl(ancestorTypeRef, currentElem, v, s)
		default:
			unhandledType = true
		}

	case typecategory.Reference:
		switch genericType {
		case generictype.Interface:
			r.reflectTypeInterfaceImpl(ancestorTypeRef, currentElem, v, s)
		case generictype.Pointer:
			r.reflectTypePointerImpl(ancestorTypeRef, currentElem, v, s)
		default:
			unhandledType = true
		}

	default:
		// This should never happen!!! Just break the chain.
		panic(fmt.Sprintf("unexpected type category %q", genericType.Category()))
	}

	if unhandledType {
		// This should never happen!!! Just break the chain.
		panic(fmt.Sprintf("unexpected type %q", genericType))
	}

	// If current element is ancestorTypeRef named type, add to typeRefs.
	r.addTypeRef(currentElem)
}

// addTypeRef adds a TypeRef for the current element.
// - This function should only be called on an element with a TypeRef.
func (r *Reflector) addTypeRef(currentElem *types.TypeElement) {
	// Do nothing if the current element is not a TypeRef.
	if currentElem.NativeDefault().TypeRef == "" {
		return
	}

	// Skip if the TypeRef has already been captured.
	if r.Schema.TypeRefs.ChildByName(currentElem.NativeDefault().TypeRef, nil) != nil {
		return
	}

	// Skip if the TypeRef has a cyclical reference error.
	if currentElem.Error == types.CyclicalReferenceErr {
		return
	}

	refElem := currentElem.Copy()

	// The first element of a type ref does is not a type ref.
	refElem.Name = currentElem.NativeDefault().TypeRef
	refElem.TypeRef = ""
	refElem.NativeDefault().TypeRef = ""

	r.typeRefRecursion(refElem)

	r.Schema.TypeRefs.AddChild(refElem)
}

// typeRefRecursion is an internal recursive function to handle nested TypeRefs.
// - Recursively process elements.
// - If TypeRef is found, process TypeRef then remove its children.
func (r *Reflector) typeRefRecursion(currentElem *types.TypeElement) {
	if currentElem.NativeDefault().TypeRef != "" {
		// Add TypeRefs only if they are not cyclical errors.
		if currentElem.Error != types.CyclicalReferenceErr {
			r.addTypeRef(currentElem)
			currentElem.RemoveAllChildren()
		}

		currentElem.Error = ""

		return
	}

	// Keep current element and process children.
	for _, childElem := range currentElem.Children {
		r.typeRefRecursion(childElem)
	}
}

// reflectTypeInterfaceImpl refects on interface types
// Interface is a special case which is either:
// - nil -- nil has no discernable type and is an error
// - a wrapper around another type -- ignore the interface and continue reflection with the wrapped type
func (r *Reflector) reflectTypeInterfaceImpl(ancestorTypeRef types.AncestorTypeRef, currentElem *types.TypeElement, v reflect.Value, s *reflect.StructField) {
	if v.IsZero() {
		// nil is an invalid element because its type cannot be determined
		currentElem.Type = "invalid"
		currentElem.Error = types.NilInterfaceErr
		return
	}

	// Interface is nullable.
	currentElem.Nullable = true

	// Non-Zero interface is just an extra layer of abstraction around ancestorTypeRef real type.
	// Reuse the current element in order to "skip" the interface element.
	r.reflectTypeImpl(ancestorTypeRef.Copy(), currentElem, v.Elem(), nil)
}

// reflectTypePointerImpl refects on pointer types
func (r *Reflector) reflectTypePointerImpl(ancestorTypeRef types.AncestorTypeRef, currentElem *types.TypeElement, v reflect.Value, s *reflect.StructField) {
	// Pointer is a memory address pointing to some other type element.
	currentElem.NativeDefault().Options.AddBool("IsNil", v.IsNil())

	if currentElem.Error == "" {
		// Get target of pointer.
		var targetValue reflect.Value

		if v.IsNil() {
			// Create ancestorTypeRef new value if pointer is nil.
			targetValue = reflect.New(v.Type().Elem()).Elem()
		} else {
			// Use existing value for valid pointer.
			targetValue = v.Elem()
		}

		// Pointer is nullable.
		currentElem.Nullable = true

		r.reflectTypeImpl(ancestorTypeRef.Copy(), currentElem, targetValue, nil)
	}
}

// reflectTypeListImpl refects on list types: Slice, Array
// Array and Slice represent lists of elements.
// - 1st element of list will be used to determine element type
// - If list is empty, ancestorTypeRef one-element list will be created to use for typing.
func (r *Reflector) reflectTypeListImpl(ancestorTypeRef types.AncestorTypeRef, currentElem *types.TypeElement, v reflect.Value, s *reflect.StructField) {
	// Value for next reflect iteration.
	var targetValue reflect.Value

	// Keep track of whether list has elements.
	listHasElements := false

	// Count number of elements.
	currentElem.Native[NATIVE_DIALECT].Options.AddKeyVal("Len", fmt.Sprintf("%d", v.Len()))

	switch v.Kind() {
	case reflect.Array:
		if currentElem.Error == "" {
			//	Get kind of underlying elements.
			currentElem.Native[NATIVE_DIALECT].Options.AddKeyVal("Len", fmt.Sprintf("%d", v.Len()))
			if v.Len() == 0 {
				targetValue = reflect.New(v.Type().Elem()).Elem()
			} else {
				listHasElements = true
			}
		}

	case reflect.Slice:
		currentElem.NativeDefault().Options.AddBool("IsNil", v.IsNil())

		if currentElem.Error == "" {
			//	Get kind of underlying elements.
			if v.IsNil() || v.Len() == 0 {
				targetValue = reflect.MakeSlice(v.Type(), 1, 1).Index(0)
			} else {
				listHasElements = true
			}
		}

	default:
		// All other types should be handled above.
		panic(fmt.Sprintf("value.Kind %q is not a List type", v.Kind()))
	}

	if listHasElements {
		// Check all slice elements to verify that they are all the same kind.
		kindsFound := map[string]int{}
		childElem := []*types.TypeElement{}

		for i := 0; i < v.Len(); i++ {
			nextElem := currentElem.NewChild("")
			childElem = append(childElem, nextElem)

			targetValue = v.Index(i)
			r.reflectTypeImpl(ancestorTypeRef.Copy(), nextElem, targetValue, nil)

			kindsFound[nextElem.Type]++
			if len(kindsFound) > 1 {
				// If multiple types found, set error and exit.
				currentElem.Error = types.SliceMultiTypeErr

				// Build a string with type:count elements.
				out := []string{}
				for k, v := range kindsFound {
					out = append(out, fmt.Sprintf("%s:%d", k, v))
				}
				currentElem.NativeDefault().Error = fmt.Sprintf("%s: %s", types.SliceMultiTypeErr, strings.Join(out, ","))
				return
			}
		}

		// All list elements have same type. Add first element as child of current element.
		currentElem.AddChild(childElem[0])

		// Remove extra child elements.
		if len(childElem) > 1 {
			for i := 1; i < len(childElem); i++ {
				currentElem.RemoveChild(childElem[i])
			}
		}
	} else {
		// Iterate using target value.
		nextElem := currentElem.NewChild("")
		r.reflectTypeImpl(ancestorTypeRef.Copy(), nextElem, targetValue, nil)
	}
}

// reflectTypeStructImpl reflects on struct types: Struct, Map
// Struct and Map represent key-value pairs.
// - Struct keys are field names which are always strings.
// - Map keys can be any comprable Go type.
func (r *Reflector) reflectTypeStructImpl(ancestorTypeRef types.AncestorTypeRef, currentElem *types.TypeElement, v reflect.Value, s *reflect.StructField) {
	switch v.Kind() {
	case reflect.Struct:
		if currentElem.Error == "" {
			if v.NumField() == 0 {
				currentElem.Error = types.EmptyStructErr
				return
			}

			// Count exported fields.
			exportedFields := 0

			for i := 0; i < v.NumField(); i++ {
				structField := v.Type().Field(i)
				targetValue := v.Field(i)

				// Skip un-exported fields.
				if structField.PkgPath != "" {
					continue
				}
				exportedFields++

				nextElem := currentElem.NewChild(structField.Name)

				// Parse struct tags.
				tags := types.ParseTags(structField.Tag)
				if len(tags) > 0 {
					for tagName, tagVal := range tags {
						tempNative := nextElem.Native[tagName]
						if tempNative == nil {
							tempNative = types.NewNativeType(tagName)
							nextElem.Native[tagName] = tempNative
						}
						tempNative.UpdateFromTag(tagVal)
					}
				}

				r.reflectTypeImpl(ancestorTypeRef.Copy(), nextElem, targetValue, &structField)
			}

			if exportedFields == 0 {
				currentElem.Error = types.NoExportedFieldsErr
				return
			}
		}

	case reflect.Map:
		currentElem.Native[currentElem.NativeDialect].Options.AddBool("IsNil", v.IsNil())

		if currentElem.Error == "" {
			// Map key must be ancestorTypeRef string.
			if v.Type().Key().Kind() != reflect.String {
				currentElem.Error = types.MapKeyTypeErr
				currentElem.NativeDefault().Error = fmt.Sprintf("map key type must be string not %q", v.Type().Key())
				return
			}

			// Empty map not allowed.
			if v.Len() == 0 {
				currentElem.Error = types.EmptyMapErr
				return
			}

			// Iterate through map by keys in sorted order.
			// - Assume that all map keys are exported fields which means they must be capitalized.
			//   - Name is the original name of the map field.
			//   - ExportName is the capitalized name fo the map field.
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
				newKey.ExportName = util.Capitalize(newKey.Name)

				keys = append(keys, newKey)
			}

			// Sort by ExportName then Name.
			sort.Slice(keys, func(i, j int) bool {
				if keys[i].ExportName != keys[j].ExportName {
					return keys[i].ExportName < keys[j].ExportName
				}
				return keys[i].Name < keys[j].Name
			})

			uniqKeys := map[string]int{}
			for _, k := range keys {
				mapValue := v.MapIndex(k.Value)

				nextElem := currentElem.NewChild(k.ExportName)
				if k.ExportName != k.Name {
					// Use original Name for native defaults.
					nextElem.NativeDefault().Name = k.Name
				}

				// Check for duplicate ExportName
				if uniqKeys[k.ExportName] > 0 {
					nextElem.Error = types.DuplicateMapKeyErr
					nextElem.NativeDefault().Error = fmt.Sprintf("duplicate map key %q (%q)", k.ExportName, k.Name)
				}
				uniqKeys[k.ExportName]++

				r.reflectTypeImpl(ancestorTypeRef.Copy(), nextElem, mapValue, nil)
			}
		}
	}
}
