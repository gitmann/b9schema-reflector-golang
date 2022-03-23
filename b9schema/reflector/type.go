package reflector

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/gitmann/shiny-reflector-golang/b9schema/enum/generictype"
	"github.com/gitmann/shiny-reflector-golang/b9schema/enum/threeflag"
	"github.com/gitmann/shiny-reflector-golang/b9schema/enum/typecategory"

	"github.com/gitmann/shiny-reflector-golang/b9schema/util"
)

// TypeElement holds type information about an element.
// - TypeElement should be cross-platform and only use basic types.
type TypeElement struct {
	// Unique identifier for an element.
	ID int `json:"-"`

	// Optional Name and Description of element.
	// - Name applies to struct/map types with string keys.
	Name        string `json:",omitempty"`
	Description string `json:",omitempty"`

	// Nullable indicates that a field should accept null in addition to values.
	Nullable bool `json:",omitempty"`

	// Generic type of element.
	TypeCategory string `json:",omitempty"`
	Type         string `json:",omitempty"`

	// TypeRef holds the name of a type (e.g. struct)
	TypeRef string `json:",omitempty"`

	// NativeDialect is the name of the dialect that was the source for the schema.
	NativeDialect string `json:"-"`

	// Native type features by dialect name.
	Native map[string]*NativeType `json:"-"`

	// Capture error if element cannot reflect.
	Error string `json:",omitempty"`

	// Pointers to Parent and Children.
	Parent   *TypeElement   `json:"-"`
	Children []*TypeElement `json:",omitempty"`
}

// NewRootElement creates a new type element that is a root of a tree.
// - Root elements do not have parents and do not produce output.
func NewRootElement(name string) *TypeElement {
	r := NewTypeElement(name)

	r.Type = generictype.Root.String()
	r.TypeCategory = generictype.Root.Category().String()

	return r
}

// NewTypeElement creates a new type element without a Parent or Children.
func NewTypeElement(name string) *TypeElement {
	t := &TypeElement{
		ID: nextID(),

		Parent:   nil,
		Children: []*TypeElement{},

		Name: name,

		NativeDialect: NATIVE_DIALECT,
		Native:        map[string]*NativeType{},
	}
	t.Native[NATIVE_DIALECT] = NewNativeType(NATIVE_DIALECT)

	return t
}

// NewChild creates a new type element that is a child of the current one.
func (t *TypeElement) NewChild(name string) *TypeElement {
	childElem := NewTypeElement(name)
	t.AddChild(childElem)

	return childElem
}

// AddChild adds a child element to the current element.
// - Sets Parent on the child element.
func (t *TypeElement) AddChild(childElem *TypeElement) {
	// Ignore nil.
	if childElem == nil {
		return
	}

	if childElem.Parent != nil {
		childElem.Parent.RemoveChild(childElem)
	}

	childElem.Parent = t
	t.Children = append(t.Children, childElem)
}

// ChildMap returns a map of Children name --> *TypeElement
// - Output map can be passed to ChildKeys, ContainsChild, ChildByName for reuse.
func (t *TypeElement) ChildMap() map[string]*TypeElement {
	out := map[string]*TypeElement{}
	for _, childElem := range t.Children {
		out[childElem.Name] = childElem
	}
	return out
}

// ChildKeys returns a sorted list of child names.
func (t *TypeElement) ChildKeys(m map[string]*TypeElement) []string {
	if len(m) == 0 {
		m = t.ChildMap()
	}

	out := make([]string, len(m))
	if len(m) > 0 {
		i := 0
		for k := range m {
			out[i] = k
			i++
		}

		sort.Strings(out)
	}

	return out
}

// ContainsChild returns true if a child with the given name exist.
func (t *TypeElement) ContainsChild(name string, m map[string]*TypeElement) bool {
	c := t.ChildByName(name, m)
	return c != nil
}

// ChildByName gets the child with the given element name.
// - Returns nil if child does not exist.
func (t *TypeElement) ChildByName(name string, m map[string]*TypeElement) *TypeElement {
	if len(m) == 0 {
		m = t.ChildMap()
	}
	return m[name]
}

// RemoveAllChildren removes all children from the current element.
func (t *TypeElement) RemoveAllChildren() {
	for _, childElem := range t.Children {
		childElem.Parent = nil
	}

	t.Children = []*TypeElement{}
}

// RemoveChild removes the given child from the Children list.
// - Uses ID for matching.
// - Sets Parent on child to nil.
func (t *TypeElement) RemoveChild(childElem *TypeElement) {
	if childElem == nil {
		return
	}

	// Copy all children except the given one.
	newChildren := []*TypeElement{}
	for _, elem := range t.Children {
		if elem.ID != childElem.ID {
			newChildren = append(newChildren, elem)
		} else {
			childElem.Parent = nil
		}
	}

	t.Children = newChildren
}

// Copy makes a copy of a TypeElement and its Children.
// - The copied element has no Parent.
func (t *TypeElement) Copy() *TypeElement {
	n := &TypeElement{
		ID: nextID(),

		Parent:   nil,
		Children: []*TypeElement{},

		Name:        t.Name,
		Description: t.Description,

		Type:         t.Type,
		TypeCategory: t.TypeCategory,

		TypeRef: t.TypeRef,

		NativeDialect: t.NativeDialect,
		Native:        make(map[string]*NativeType),

		Error: t.Error,
	}

	// Copy Children with new element as parent.
	for _, childElem := range t.Children {
		newChild := childElem.Copy()
		n.AddChild(newChild)
	}

	for dialect, native := range t.Native {
		n.Native[dialect] = native.Copy()
	}

	return n
}

// ParentID returns the ID of the parent of the current element.
func (t *TypeElement) ParentID() int {
	if t.Parent != nil {
		return t.Parent.ID
	}

	// Return -1 if no parent.
	return -1
}

// ChildPaths returns a list of paths for all Children of the element.
// - If ChildPaths are equal, then the types are equal.
func (t *TypeElement) ChildPaths() []*PathList {
	//TODO: implement this!
	panic("not implemented")
}

// RenderPath returns the PathList to the current element built from all ancestor path lists.
func (t *TypeElement) RenderPath(renderFunc PathStringRenderer, opt *RenderOptions) *PathList {
	var p *PathList

	if t.Parent == nil {
		// This is a root element. Return a new PathList.
		p := NewPathList()
		p.Push(t.RenderPathString(renderFunc, opt))
		return p
	}

	// Get the parent's path list and append the current element's path.
	p = t.Parent.RenderPath(renderFunc, opt)
	p.Push(t.RenderPathString(renderFunc, opt))

	return p
}

// GetName returns the alias for the given lang or Name.
func (t *TypeElement) GetName(lang string) string {
	if t.Native != nil {
		if t.Native[lang] != nil {
			if a := t.Native[lang].Name; a != "" {
				return a
			}
		}
	}
	return t.Name
}

// SetName sets the GetName for the native dialect.
func (t *TypeElement) SetName(dialect, alias string) {
	if t.Native == nil {
		t.Native = make(map[string]*NativeType)
	}
	if a := t.Native[dialect]; a == nil {
		t.Native[dialect] = NewNativeType(dialect)
	}
	t.Native[dialect].Name = alias
}

// NativeDefault returns the native element for the NativeDialect.
func (t *TypeElement) NativeDefault() *NativeType {
	return t.Native[t.NativeDialect]
}

// ElementStringRenderer is a function that builds a string from a TypeElement.
type ElementStringRenderer func(t *TypeElement, pathFunc PathStringRenderer, opt *RenderOptions) string

// PathStringRenderer is a function that builds a path string from a TypeElement.
type PathStringRenderer func(t *TypeElement, opt *RenderOptions) string

// RenderPathString returns the element's string for the PathList.
// Format is: [<Name>:]<Type>[:<TypeRef>]
// - If Name is set, prefix with "Name", otherwise "-"
// - If TypeRef is set, suffix with "TypeRef", otherwise "-"
// - If Error is set, wrap entire string with "!"
func (t *TypeElement) RenderPathString(renderFunc PathStringRenderer, opt *RenderOptions) string {
	if renderFunc != nil {
		return renderFunc(t, opt)
	}

	if t.Type == generictype.Root.String() {
		return t.Name
	}

	namePart := t.Name
	if namePart != "" {
		namePart += ":"
	}

	// Type.
	var typePart string
	if t.TypeCategory == typecategory.Invalid.String() {
		typePart = t.Type
	} else {
		typePart = generictype.PathDefaultOfType(t.Type)
	}

	// Add TypeRef suffix if set but not if de-referencing.
	refPart := ""
	if !opt.DeReference {
		refPart = t.NativeDefault().TypeRef
	} else if opt.DeReference && t.Error == CyclicalReferenceErr {
		// Keep reference if it's a cyclical error.
		refPart = t.NativeDefault().TypeRef
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

// IsBasicType returns true if the element is a basic type.
func (t *TypeElement) IsBasicType() bool {
	switch t.Type {
	case "string", "integer", "float", "boolean":
		return true
	}
	return false
}

// IsExported returns true if the element Name starts with an uppercase letter.
func (t *TypeElement) IsExported() bool {
	if t.Name == "" {
		return false
	}

	r := []rune(t.Name)
	return unicode.IsUpper(r[0])
}

func (t *TypeElement) RenderStrings(preFunc, postFunc ElementStringRenderer, pathFunc PathStringRenderer, opt *RenderOptions) []string {
	out := []string{}

	// Process element with preFunc.
	if preFunc != nil {
		if s := preFunc(t, pathFunc, opt); s != "" {
			out = append(out, s)
		}
	}

	// Process children.
	if !opt.DeReference && t.TypeRef != "" {
		// Skip children.
	} else {
		for _, childElem := range t.Children {
			rendered := childElem.RenderStrings(preFunc, postFunc, pathFunc, opt)
			for _, r := range rendered {
				if r != "" {
					out = append(out, r)
				}
			}
		}
	}

	// Process element with postFunc.
	if postFunc != nil {
		if s := postFunc(t, pathFunc, opt); s != "" {
			out = append(out, s)
		}
	}

	return out
}

// PathList keeps a list of path string elements that form a unique identifier for a TypeElement.
// - PathList behaves like a stack with Push/Pop operators.
type PathList struct {
	paths []string
}

func NewPathList() *PathList {
	return &PathList{paths: make([]string, 0)}
}

func (p *PathList) Len() int {
	return len(p.paths)
}

func (p *PathList) Push(elem string) {
	// Ignore empty elements.
	if elem == "" {
		return
	}
	p.paths = append(p.paths, elem)
}

func (p *PathList) Pop() string {
	if len(p.paths) == 0 {
		return ""
	}

	elem := p.paths[len(p.paths)-1]
	p.paths = p.paths[:len(p.paths)-1]
	return elem
}

func (p *PathList) Copy() *PathList {
	n := &PathList{paths: make([]string, len(p.paths))}
	copy(n.paths, p.paths)
	return n
}

func (p *PathList) String() string {
	out := make([]string, len(p.paths))
	for i, s := range p.paths {
		if strings.Contains(s, ".") {
			s = fmt.Sprintf("%q", s)
		}
		out[i] = s
	}
	return strings.Join(out, ".")
}

// NativeOption stores options as key-value pairs but returns a list of strings.
// - Value-only entries are unique by value.
// - Values with keys are unique by key.
type NativeOption struct {
	optionMap map[string]string
}

func NewNativeOption() *NativeOption {
	return &NativeOption{
		optionMap: make(map[string]string),
	}
}

// Equals returns true if both NativeOption struct have the same values.
func (n *NativeOption) Equals(other *NativeOption) bool {
	if n == nil && other == nil {
		// Both are nil so equal.
		return true
	}

	// Treat nil option map as zero-length.
	var thisLen, otherLen int
	if n != nil && n.optionMap != nil {
		thisLen = len(n.optionMap)
	}
	if other != nil && other.optionMap != nil {
		otherLen = len(other.optionMap)
	}

	if thisLen != otherLen {
		return false
	} else if thisLen == 0 {
		return true
	}

	return reflect.DeepEqual(n.optionMap, n.optionMap)
}

// AsList returns options as a slice of strings.
func (n *NativeOption) AsList() []string {
	// Return empty slice if no options are set.
	if len(n.optionMap) == 0 {
		return make([]string, 0)
	}

	s := make([]string, len(n.optionMap))
	i := 0
	for k, v := range n.optionMap {
		if v == "" {
			//	Value only
			s[i] = k
		} else {
			// Key-Value pair
			s[i] = fmt.Sprintf("%s=%s", k, v)
		}
		i++
	}

	// Sort slice for output.
	sort.Strings(s)
	return s
}

// AddVal adds an option value string.
func (n *NativeOption) AddVal(val string) {
	// Ignore if value is empty.
	if val == "" {
		return
	}
	n.optionMap[val] = ""
}

// Delete removes an entry from the option map.
// - key will match either key-value pairs or value-only settings.
func (n *NativeOption) Delete(key string) {
	// Ignore empty key.
	if key == "" {
		return
	}
	delete(n.optionMap, key)
}

// AddKeyVal adds an option string key=val
func (n *NativeOption) AddKeyVal(key, val string) {
	// Ignore if key is empty.
	if key == "" {
		return
	}

	// If value is empty, delete key.
	if val == "" {
		n.Delete(key)
		return
	}

	// Set value.
	n.optionMap[key] = val
}

// AddBool adds a boolean as an option string.
// - key is required
//   - if key is empty, nothing is added
// - val is boolean value
func (n *NativeOption) AddBool(key string, val bool) {
	// Ignore if key is missing.
	if key == "" {
		return
	}

	n.optionMap[key] = fmt.Sprintf("%t", val)
}

// AddThreeFlag adds a ThreeFlag value as a string.
// - key is required
//   - if key is empty, nothing is added
// - val is ThreeFlag value
func (n *NativeOption) AddThreeFlag(key string, val threeflag.ThreeFlag) {
	// Ignore if key is missing.
	if key == "" {
		return
	}

	n.optionMap[key] = val.String()
}

// UpdateFrom updates with values from another NativeOption.
func (n *NativeOption) UpdateFrom(other *NativeOption) {
	for k, v := range other.optionMap {
		n.optionMap[k] = v
	}
}

// Copy makes a copy of the NativeOption.
func (n *NativeOption) Copy() *NativeOption {
	c := NewNativeOption()

	for k, v := range n.optionMap {
		c.optionMap[k] = v
	}

	return c
}

// NativeType holds key-value attributes specific to one dialect.
// - A dialect is the name of a language (e.g. golang) or implementation (e.g. json-schema)
type NativeType struct {
	// Name of language of dialect represented by NativeType.
	Dialect string

	// Name of element if different from generic Name.
	Name string

	// Native type of element if different from the generic Type.
	Type string

	// TypeRef holds the native name of a type if different from the generic TypeRef.
	TypeRef string

	// Include indicates whether an element should be included in output for a dialect.
	// Include has three value values:
	// - "" (empty string) means value is not set
	// - "yes" = include element in output
	// - "no" = exclude element from output
	Include threeflag.ThreeFlag

	// Options contains a list of strings representing dialect-specific options.
	// - Format is one of:
	//   - "value"
	//   - "key=value"
	Options *NativeOption

	// Capture error if element cannot reflect.
	Error string
}

// NewNativeType initializes a new NativeType with default settings.
func NewNativeType(dialect string) *NativeType {
	n := &NativeType{
		// Default to the native dialect.
		Dialect: dialect,

		// Include fields by default.
		Include: threeflag.True,

		// Empty options list.
		Options: NewNativeOption(),
	}

	return n
}

// UpdateFromTag sets NativeType fields from a StructFieldTag.
func (n *NativeType) UpdateFromTag(t *StructFieldTag) {
	if t == nil {
		return
	}

	if t.Ignore {
		n.Include = threeflag.False
	}

	if t.Alias != "" {
		n.Name = t.Alias
	}

	n.Options.UpdateFrom(t.Options)
}

// AsMap returns a map[string]string representation of the NativeType struct.
func (n *NativeType) AsMap() map[string]string {
	m := map[string]string{}

	if n.Include != threeflag.Undefined {
		m["Include"] = n.Include.String()
	}
	if n.Name != "" {
		m["Name"] = n.Name
	}
	if n.Type != "" {
		m["Type"] = n.Type
	}
	if n.TypeRef != "" {
		m["TypeRef"] = n.TypeRef
	}
	if n.Error != "" {
		m["Error"] = n.Error
	}

	for i, s := range n.Options.AsList() {
		k := fmt.Sprintf("Options[%03d]", i)
		m[k] = s
	}

	return m
}

// Copy makes a copy of a NativeType.
func (n *NativeType) Copy() *NativeType {
	c := &NativeType{
		Dialect: n.Dialect,
		Name:    n.Name,
		Type:    n.Type,
		TypeRef: n.TypeRef,
		Include: n.Include,
		Options: n.Options.Copy(),
		Error:   n.Error,
	}

	return c
}

// TypeList holds a slice of TypeElements.
// - Behavior is similar to a stack with Push/Pop methods to add/remove elements from the end
type TypeList struct {
	types []*TypeElement
}

func NewTypeList() *TypeList {
	// Initialize an empty TypeList.
	return &TypeList{
		types: make([]*TypeElement, 0),
	}
}

// Len returns the number of elements in the TypeList.
func (typeList *TypeList) Len() int {
	return len(typeList.types)
}

// Push adds an element to the list.
func (typeList *TypeList) Push(elem *TypeElement) {
	typeList.types = append(typeList.types, elem)
}

// Pop removes the last element from the list an returns it.
// - Returns nil is list is empty.
func (typeList *TypeList) Pop() *TypeElement {
	if len(typeList.types) > 0 {
		lastElem := typeList.types[len(typeList.types)-1]
		typeList.types = typeList.types[:len(typeList.types)-1]

		return lastElem
	}

	// Empty list.
	return nil
}

// Copy makes a copy of the current TypeList.
// - Parent is set if parentElem is not nil.
func (typeList *TypeList) Copy(parentElem *TypeElement) *TypeList {
	c := NewTypeList()

	// Copy all elements to new list.
	for _, elem := range typeList.types {
		newElem := elem.Copy()
		c.Push(newElem)

		if parentElem != nil {
			parentElem.AddChild(newElem)
		}
	}

	return c
}

// Elements returns the internal slice of TypeElements.
func (typeList *TypeList) Elements() []*TypeElement {
	return typeList.types
}

// TypeResult is the result of parsing types.
type TypeResult struct {
	// Root is a list of types in the order found.
	Root *TypeElement

	// TypeRefs holds a map of named types by name.
	TypeRefs *TypeElement
}

// // String builds a default string representation of a type result.
// // - Each line starts with 3 comma-delimited values: <prefix>,<id>,<parent>
// // - Each parent level has all prefixes with the same width.
// func (typeResult *TypeResult) String(formatName string) string {
// 	// Keep track of max prefix length by parent ID.
// 	maxParentLen := map[string]int{}

// 	// Keep map from ID --> ParentID. Every parent indent must be larger than its parent indent.
// 	idMap := map[string]string{}

// 	// Keep lines parts in a struct.
// 	type lineTokens struct {
// 		prefix string
// 		id     string
// 		parent string
// 		other  string
// 	}

// 	// Iterate through strings to determine max lengths for each parent level.
// 	outputLines := []*lineTokens{}
// 	for _, line := range typeResult.BuildStrings(formatName) {
// 		// Split line into prefix,id,parent,other
// 		tokens := strings.SplitN(line, ",", 4)

// 		if len(tokens) != 4 {
// 			// Some other line type. Just add everything in other.
// 			outputLines = append(outputLines, &lineTokens{other: line})
// 		} else {
// 			newLine := &lineTokens{
// 				prefix: tokens[0],
// 				id:     tokens[1],
// 				parent: tokens[2],
// 				other:  tokens[3],
// 			}
// 			outputLines = append(outputLines, newLine)

// 			// Update map from id --> parent
// 			idMap[newLine.id] = newLine.parent

// 			// Update max length for parent level.
// 			if len(newLine.prefix) > maxParentLen[newLine.parent] {
// 				maxParentLen[newLine.parent] = len(newLine.prefix)
// 			}
// 		}
// 	}

// 	// Make a pass to ensure that each parent indent is at least 2 larger than its parent.
// 	for parent, indent := range maxParentLen {
// 		parentParent := idMap[parent]
// 		if parentParent != "" {
// 			if maxParentLen[parentParent]+2 > indent {
// 				maxParentLen[parent] = maxParentLen[parentParent] + 2
// 			}
// 		}
// 	}

// 	// Build output using lengths.
// 	out := []string{}
// 	for _, line := range outputLines {
// 		var newLine string
// 		if line.prefix == "" && line.id == "" && line.parent == "" {
// 			newLine = line.other
// 		} else {
// 			newLine = fmt.Sprintf("%-*s >>> %s,%s,%s",
// 				maxParentLen[line.parent], line.prefix, line.id, line.parent, line.other)
// 		}
// 		out = append(out, newLine)
// 	}

// 	return strings.Join(out, "\n")
// }

// // BuildString builds a string representation of a type result using the given formatName.
// func (typeResult *TypeResult) BuildStrings(formatName string) []string {
// 	// Set formatting options.
// 	printHeaders := true
// 	printTypeRefs := true

// 	// Build output outLines.
// 	outLines := []string{}

// 	// Print type refs.
// 	if printTypeRefs {
// 		refNames := typeResult.TypeRefs.Keys()
// 		for _, typeName := range refNames {
// 			if printHeaders {
// 				outLines = append(outLines, fmt.Sprintf("*** TypeRef: %s", typeName))
// 			}
// 			outLines = append(outLines, typeResult.TypeRefs.Get(typeName).BuildStrings(formatName)...)
// 		}
// 	}

// 	//	Print types.
// 	if printHeaders {
// 		outLines = append(outLines, "*** Types")
// 	}
// 	outLines = append(outLines, typeResult.Root.BuildStrings(formatName)...)

// 	//	Return final string.
// 	return outLines
// }

// RenderStrings builds a string representation of a type result using the given pre, path, and post functions.
func (typeResult *TypeResult) RenderStrings(preFunc, postFunc ElementStringRenderer, pathFunc PathStringRenderer, opt *RenderOptions) []string {
	if opt == nil {
		opt = NewRenderOptions()
	}

	// Build output outLines.
	out := []string{}

	// Print type refs.
	if !opt.DeReference {
		typeRefMap := typeResult.TypeRefs.ChildMap()
		typeRefKeys := typeResult.TypeRefs.ChildKeys(typeRefMap)

		for _, childName := range typeRefKeys {
			rendered := typeRefMap[childName].RenderStrings(preFunc, postFunc, pathFunc, opt)
			for _, r := range rendered {
				if r != "" {
					out = append(out, r)
				}
			}
		}
	}

	//	Print types.
	rendered := typeResult.Root.RenderStrings(preFunc, postFunc, pathFunc, opt)
	for _, r := range rendered {
		if r != "" {
			out = append(out, r)
		}
	}

	//	Return final string.
	return out
}

// reflectTypeImpl is a recursive function to reflect Go values.
//
// Args:
// - typeList (TypeList): list of TypeElement found so far
// - ancestoreTypeRef (AncestorTypeRef): keeps track of TypeRef names seen so far, used for cycle detection
// - currentElem (*TypeElement): current TypeElement, must be initialized in caller!
// - v (reflect.Value): Value of current element
// - s (*reflect.StructField): pointer to StructField for current element if part of a struct
//
// Returns:
// - TypeList: list of TypeElement after reflection
func (r *Reflector) reflectTypeImpl(ancestorTypeRef AncestorTypeRef, currentElem *TypeElement, v reflect.Value, s *reflect.StructField) {
	// currentElem must be initialized in caller!!!
	if currentElem == nil {
		panic("currentElem cannot be nil")
	}

	// Create temporary list for named type refs.
	refList := NewTypeList()
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
		currentElem.Error = InvalidKindErr

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
			currentElem.Error = RootKindErr
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
			currentElem.Error = CyclicalReferenceErr
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

	// Add reference list if DeReference or current element is not a TypeRef.
	// - 1st element is already added so only add when there is more than 1 element in reference list.
	// - When DeReference is true, keep the last TypeRef if the last element is a TypeRef. This indicates a cyclical relationship.
	//if refList.Len() > 1 {
	//	if DeReference || currentElem.NativeDefault().TypeRef == "" {
	//		// Remove all TypeRefs except the last one when DeReference is true.
	//		var lastElem *TypeElement
	//		var lastTypeRef string
	//
	//		// Only add refList if it is ancestorTypeRef compound type.
	//		for _, newElem := range refList.Elements()[1:] {
	//			lastElem = newElem
	//			lastTypeRef = lastElem.NativeDefault().TypeRef
	//
	//			if DeReference {
	//				newElem.TypeRef = ""
	//			}
	//
	//			typeList.Push(newElem)
	//		}
	//
	//		// Restore the TypeRef on the last element
	//		lastElem.TypeRef = lastTypeRef
	//	}
	//}
}

// addTypeRef adds a TypeRef for the current element.
// - This function should only be called on an element with a TypeRef.
func (r *Reflector) addTypeRef(currentElem *TypeElement) {
	// Do nothing if the current element is not a TypeRef.
	if currentElem.NativeDefault().TypeRef == "" {
		return
	}

	// Skip if the TypeRef has already been captured.
	if r.typeResult.TypeRefs.ChildByName(currentElem.NativeDefault().TypeRef, nil) != nil {
		return
	}

	// Skip if the TypeRef has a cyclical reference error.
	if currentElem.Error == CyclicalReferenceErr {
		return
	}

	refElem := currentElem.Copy()

	// The first element of a type ref does is not a type ref.
	refElem.Name = currentElem.NativeDefault().TypeRef
	refElem.TypeRef = ""
	refElem.NativeDefault().TypeRef = ""

	r.typeRefRecursion(refElem)

	r.typeResult.TypeRefs.AddChild(refElem)
}

// typeRefRecursion is an internal recursive function to handle nested TypeRefs.
// - Recursively process elements.
// - If TypeRef is found, process TypeRef then remove its children.
func (r *Reflector) typeRefRecursion(currentElem *TypeElement) {
	if currentElem.NativeDefault().TypeRef != "" {
		// Add TypeRefs only if they are not cyclical errors.
		if currentElem.Error != CyclicalReferenceErr {
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
func (r *Reflector) reflectTypeInterfaceImpl(ancestorTypeRef AncestorTypeRef, currentElem *TypeElement, v reflect.Value, s *reflect.StructField) {
	if v.IsZero() {
		// nil is an invalid element because its type cannot be determined
		currentElem.Type = "invalid"
		currentElem.Error = NilInterfaceErr
		return
	}

	// Interface is nullable.
	currentElem.Nullable = true

	// Non-Zero interface is just an extra layer of abstraction around ancestorTypeRef real type.
	// Reuse the current element in order to "skip" the interface element.
	r.reflectTypeImpl(ancestorTypeRef.Copy(), currentElem, v.Elem(), nil)
}

// reflectTypePointerImpl refects on pointer types
func (r *Reflector) reflectTypePointerImpl(ancestorTypeRef AncestorTypeRef, currentElem *TypeElement, v reflect.Value, s *reflect.StructField) {
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
func (r *Reflector) reflectTypeListImpl(ancestorTypeRef AncestorTypeRef, currentElem *TypeElement, v reflect.Value, s *reflect.StructField) {
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
		childElem := []*TypeElement{}

		for i := 0; i < v.Len(); i++ {
			nextElem := currentElem.NewChild("")
			childElem = append(childElem, nextElem)

			targetValue = v.Index(i)
			r.reflectTypeImpl(ancestorTypeRef.Copy(), nextElem, targetValue, nil)

			kindsFound[nextElem.Type]++
			if len(kindsFound) > 1 {
				// If multiple types found, set error and exit.
				currentElem.Error = SliceMultiTypeErr

				// Build a string with type:count elements.
				out := []string{}
				for k, v := range kindsFound {
					out = append(out, fmt.Sprintf("%s:%d", k, v))
				}
				currentElem.NativeDefault().Error = fmt.Sprintf("%s: %s", SliceMultiTypeErr, strings.Join(out, ","))
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
func (r *Reflector) reflectTypeStructImpl(ancestorTypeRef AncestorTypeRef, currentElem *TypeElement, v reflect.Value, s *reflect.StructField) {
	switch v.Kind() {
	case reflect.Struct:
		if currentElem.Error == "" {
			if v.NumField() == 0 {
				currentElem.Error = EmptyStructErr
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
				if s != nil {
					tags := ParseTags(s.Tag)
					if len(tags) > 0 {
						for tagName, tagVal := range tags {
							tempNative := nextElem.Native[tagName]
							if tempNative == nil {
								tempNative = NewNativeType(tagName)
								nextElem.Native[tagName] = tempNative
							}
							tempNative.UpdateFromTag(tagVal)
						}
					}
				}

				r.reflectTypeImpl(ancestorTypeRef.Copy(), nextElem, targetValue, &structField)
			}

			if exportedFields == 0 {
				currentElem.Error = NoExportedFieldsErr
				return
			}
		}

	case reflect.Map:
		currentElem.Native[currentElem.NativeDialect].Options.AddBool("IsNil", v.IsNil())

		if currentElem.Error == "" {
			// Map key must be ancestorTypeRef string.
			if v.Type().Key().Kind() != reflect.String {
				currentElem.Error = MapKeyTypeErr
				currentElem.NativeDefault().Error = fmt.Sprintf("map key type must be string not %q", v.Type().Key())
				return
			}

			// Empty map not allowed.
			if v.Len() == 0 {
				currentElem.Error = EmptyMapErr
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
					nextElem.Error = DuplicateMapKeyErr
					nextElem.NativeDefault().Error = fmt.Sprintf("duplicate map key %q (%q)", k.ExportName, k.Name)
				}
				uniqKeys[k.ExportName]++

				r.reflectTypeImpl(ancestorTypeRef.Copy(), nextElem, mapValue, nil)
			}
		}
	}
}

// AncestorTypeRef keeps track of type references that are ancestors of the current element.
// - Stores a count of references found.
// - If count > 1, a cyclical reference exists.
type AncestorTypeRef map[string]int

// NewAncestorTypeRef initializes a new ancestor list.
func NewAncestorTypeRef() AncestorTypeRef {
	return make(AncestorTypeRef)
}

// Copy makes a copy of the ancestor list.
func (a AncestorTypeRef) Copy() AncestorTypeRef {
	n := make(AncestorTypeRef)
	for k, v := range a {
		n[k] = v
	}
	return n
}

// Contains returns true if the key exists in ancestor list.
func (a AncestorTypeRef) Contains(key string) bool {
	return a[key] > 0
}

// Add adds a reference count to the ancestor list.
func (a AncestorTypeRef) Add(key string) int {
	if key == "" {
		return 0
	}

	a[key]++
	return a[key]
}
