package threeflag

// ThreeFlag implements a 3-value flag: "undefined", "true", "false"
// Uses slugs from: https://threedots.tech/post/safer-enums-in-go/
type ThreeFlag struct {
	slug string
}

func (i ThreeFlag) String() string {
	return i.slug
}

var (
	Undefined = ThreeFlag{"undefined"}
	False     = ThreeFlag{"false"}
	True      = ThreeFlag{"true"}
)
