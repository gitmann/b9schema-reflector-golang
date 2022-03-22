package typecategory

// TypeCategory groups GenericType values.
// Uses slugs from: https://threedots.tech/post/safer-enums-in-go/
type TypeCategory struct {
	slug string
}

func (t TypeCategory) String() string {
	return t.slug
}

var (
	Invalid  = TypeCategory{"invalid"}
	Basic    = TypeCategory{"basic"}
	Compound = TypeCategory{"compound"}
	Known    = TypeCategory{"known"}

	// The following types are wrappers or pointers around other types.
	Reference = TypeCategory{"reference"}

	// The following types are for internal use only.
	Internal = TypeCategory{"internal"}
)
