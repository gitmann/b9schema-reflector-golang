package renderer

import (
	"github.com/gitmann/b9schema-reflector-golang/b9schema/lib/types"
)

type Renderer interface {
	// ProcessResult starts the render process on a Schema and returns a slice of strings.
	ProcessResult(result *types.Schema) ([]string, error)

	// DeReference returns true if schema references should be replaced with inline types.
	DeReference() bool

	// Indent returns the current indent value.
	Indent() int

	// SetIndent sets the indent to a given value.
	SetIndent(value int)

	// Prefix returns a prefix string with the current indent.
	Prefix() string

	// Pre and Post return strings before/after a type element's children are processed.
	Pre(t *types.TypeElement) []string
	Post(t *types.TypeElement) []string

	// Path is a function that builds a path string from a TypeElement.
	Path(t *types.TypeElement) []string
}
