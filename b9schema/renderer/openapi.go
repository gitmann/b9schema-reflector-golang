package renderer

import (
	"fmt"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/enum/generictype"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/enum/threeflag"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/types"
	"strings"
)

// OpenAPIRenderer provides a simple string renderer.
type OpenAPIRenderer struct {
	opt *Options
}

func NewOpenAPIRenderer(opt *Options) *OpenAPIRenderer {
	if opt == nil {
		opt = NewOptions()
	}

	opt.Prefix = "  "

	return &OpenAPIRenderer{opt: opt}
}

func (r *OpenAPIRenderer) ProcessResult(result *types.Schema) ([]string, error) {
	// Header
	return RenderSchema(result, r), nil
	// Footer
}

func (r *OpenAPIRenderer) DeReference() bool {
	return r.opt.DeReference
}

func (r *OpenAPIRenderer) Indent() int {
	return r.opt.Indent
}

func (r *OpenAPIRenderer) SetIndent(value int) {
	r.opt.Indent = value
}

func (r *OpenAPIRenderer) Prefix() string {
	if r.opt.Prefix == "" {
		return ""
	}
	return strings.Repeat(r.opt.Prefix, r.opt.Indent)
}

func (r *OpenAPIRenderer) Pre(t *types.TypeElement) []string {
	jsonType := t.GetNativeType("json")
	if jsonType.Include == threeflag.False {
		// Skip this element.
		return []string{}
	}

	// Special handling for root elements.
	if t.Type == generictype.Root.String() {
		r.SetIndent(r.Indent() + 1)

		if t.Name == "Root" {
			// Build an object named "root".
			return []string{`root:`}
		} else if t.Name == "TypeRefs" {
			// Store TypeRefs under the "definitions" key.
			return []string{`definitions:`}
		}
	}

	nativeType := t.NativeDefault()

	outLines := []string{}

	if jsonType.Name != "" {
		outLines = append(outLines, fmt.Sprintf("%s%s:", r.Prefix(), jsonType.Name))
		r.SetIndent(r.Indent() + 1)
	}

	if jsonType.TypeRef != "" {
		outLines = append(outLines, fmt.Sprintf(`%s$ref: '#/definitions/%s'`, r.Prefix(), jsonType.TypeRef))
	} else {
		switch t.Type {
		case generictype.Struct.String():
			outLines = append(outLines,
				r.Prefix()+"type: object",
				r.Prefix()+"properties:",
			)
			r.SetIndent(r.Indent() + 1)
		case generictype.List.String():
			outLines = append(outLines,
				r.Prefix()+"type: array",
				r.Prefix()+"items:",
			)
			r.SetIndent(r.Indent() + 1)
		case generictype.Boolean.String():
			outLines = append(outLines,
				r.Prefix()+"type: boolean",
			)
		case generictype.Integer.String():
			outLines = append(outLines,
				r.Prefix()+"type: integer",
			)
			if nativeType.Type == "int64" || nativeType.Type == "uint64" {
				outLines = append(outLines,
					r.Prefix()+"format: int64",
				)
			}
		case generictype.Float.String():
			outLines = append(outLines,
				r.Prefix()+"type: number",
			)
			if nativeType.Type == "float64" {
				outLines = append(outLines,
					r.Prefix()+"format: double",
				)
			}
		case generictype.String.String():
			outLines = append(outLines,
				r.Prefix()+"type: string",
			)
		case generictype.DateTime.String():
			outLines = append(outLines,
				r.Prefix()+"type: string",
				r.Prefix()+"format: date-time",
			)
		default:
			outLines = append(outLines,
				r.Prefix()+"type: "+t.Type,
			)
		}
	}

	if t.Error != "" {
		outLines = append(outLines,
			r.Prefix()+"error: "+t.Error,
		)
	}

	return outLines
}

func (r *OpenAPIRenderer) Post(t *types.TypeElement) []string {
	return []string{}
}

// Path is a function that builds a path string from a TypeElement.
func (r *OpenAPIRenderer) Path(t *types.TypeElement) []string {
	return []string{}
}
