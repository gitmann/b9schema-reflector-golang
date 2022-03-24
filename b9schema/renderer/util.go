package renderer

import (
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/types"
)

// RenderStrings builds a string representation of a type result using the given pre, path, and post functions.
func RenderSchema(schema *types.Schema, r Renderer) []string {
	// Build output outLines.
	out := []string{}

	// Print type refs.
	if !r.DeReference() {
		if len(schema.TypeRefs.Children) > 0 {
			rendered := RenderType(schema.TypeRefs, r)
			for _, r := range rendered {
				if r != "" {
					out = append(out, r)
				}
			}
		}
	}

	//	Print types.
	if len(schema.Root.Children) > 0 {
		rendered := RenderType(schema.Root, r)
		for _, r := range rendered {
			if r != "" {
				out = append(out, r)
			}
		}
	}

	//	Return strings.
	return out
}

// RenderType builds strings for a TypeElement and its children.
func RenderType(t *types.TypeElement, r Renderer) []string {
	// Capture initial indent and restore on exit.
	originalIndent := r.Indent()

	out := []string{}

	// Process element with preFunc.
	out = appendStrings(out, r.Pre(t))

	// Process children.
	if !r.DeReference() && t.TypeRef != "" {
		// Skip children.
	} else {
		// Always process children in alphabetical order.
		typeRefMap := t.ChildMap()
		typeRefKeys := t.ChildKeys(typeRefMap)

		// Capture indent before children.
		childIndent := r.Indent()

		for _, childName := range typeRefKeys {
			// Reset indent before each child.
			r.SetIndent(childIndent)
			out = appendStrings(out, RenderType(typeRefMap[childName], r))
		}
	}

	// Restore original indent.
	r.SetIndent(originalIndent)

	// Process element with postFunc.
	out = appendStrings(out, r.Post(t))

	// Restore original indent.
	r.SetIndent(originalIndent)

	return out
}

// appendStrings adds non-empty strings from in to out and returns a new slice.
func appendStrings(out []string, in []string) []string {
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
