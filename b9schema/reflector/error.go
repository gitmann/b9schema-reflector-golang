package reflector

// Errors for type reflection.
const (
	InvalidKindErr       = "kind not supported"
	CyclicalReferenceErr = "cyclical reference"
	NilInterfaceErr      = "interface element is nil"
	EmptyStructErr       = "empty struct not supported"
	EmptyMapErr          = "empty map not supported"
	NoExportedFieldsErr  = "struct has no exported fields"
	MapKeyTypeErr        = "map key type must be string"
	SliceMultiTypeErr    = "slice contains multiple kinds"
	DuplicateMapKeyErr   = "duplicate map key"
)
