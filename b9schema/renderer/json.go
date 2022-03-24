package renderer

import (
	"fmt"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/enum/generictype"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/enum/threeflag"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/enum/typecategory"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/types"
	"strings"
)

// JSONRenderer provides a simple string renderer.
type JSONRenderer struct {
	opt *Options
}

func NewJSONRenderer(opt *Options) *JSONRenderer {
	if opt == nil {
		opt = NewOptions()
	}

	return &JSONRenderer{opt: opt}
}

func (r *JSONRenderer) ProcessResult(result *types.Schema) ([]string, error) {
	// Header
	return RenderSchema(result, r), nil
	// Footer
}

func (r *JSONRenderer) DeReference() bool {
	return r.opt.DeReference
}

func (r *JSONRenderer) Indent() int {
	return r.opt.Indent
}

func (r *JSONRenderer) SetIndent(value int) {
	r.opt.Indent = value
}

func (r *JSONRenderer) Prefix() string {
	if r.opt.Prefix == "" {
		return ""
	}
	return strings.Repeat(r.opt.Prefix, r.opt.Indent)
}

func (r *JSONRenderer) Pre(t *types.TypeElement) []string {
	jsonType := t.GetNativeType("json")
	if jsonType.Include == threeflag.False {
		// Skip this element.
		return []string{}
	}

	if jsonType.Type == generictype.Root.String() {
		return []string{}
	}

	path := r.Path(t)

	// Add quotes around any path strings that contain "."
	newPath := []string{}
	for _, p := range path {
		if strings.Contains(p, ".") {
			p = fmt.Sprintf("%q", p)
		}
		newPath = append(newPath, p)
	}
	out := strings.Join(newPath, ".")

	if t.Error != "" {
		out += " ERROR:" + t.Error
	}

	return []string{out}
}

func (r *JSONRenderer) Post(t *types.TypeElement) []string {
	return []string{}
}

// Path is a function that builds a path string from a TypeElement.
// Format is: [<Name>:]<Type>[:<TypeRef>]
// - If Name is set, prefix with "Name", otherwise "-"
// - If TypeRef is set, suffix with "TypeRef", otherwise "-"
// - If Error is set, wrap entire string with "!"
func (r *JSONRenderer) Path(t *types.TypeElement) []string {
	// Check root.
	if t.Type == generictype.Root.String() {
		switch t.Name {
		case "Root":
			// JSON Path root is "$"
			return []string{"$"}
		case "TypeRefs":
			return []string{"definitions"}
		default:
			return []string{t.Name}
		}
	}

	jsonType := t.GetNativeType("json")
	if jsonType.Include == threeflag.False {
		// Skip this element.
		return []string{}
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
	refPart := ""
	if !r.DeReference() {
		refPart = jsonType.TypeRef
	} else if r.DeReference() && t.Error == types.CyclicalReferenceErr {
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

	// Wrap type in "!" if current element is an error.
	if t.Error != "" {
		path = fmt.Sprintf("!%s!", path)
	}

	return append(r.Path(t.Parent), path)
}
