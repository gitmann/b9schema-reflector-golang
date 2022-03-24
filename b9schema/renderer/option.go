package renderer

type Options struct {
	// DeReference converts TypeRefs to their included types.
	// - If TyepRefs have a cyclical relationship, the last TypeRef is kept as a TypeRef.
	DeReference bool

	// Prefix is a string used as a prefix for indented lines.
	Prefix string

	// Indent is used for rendering where indent matters.
	Indent int
}

func NewOptions() *Options {
	opt := &Options{}
	return opt
}
