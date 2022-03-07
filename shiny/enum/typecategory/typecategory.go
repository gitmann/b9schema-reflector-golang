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

	// The following types are Go-specific.
	Pointer   = TypeCategory{"pointer"}
	Interface = TypeCategory{"interface"}
)
