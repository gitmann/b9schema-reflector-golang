package util

// ValueIfTrue converts a boolean into strings for true and false.
// - if boolTest is true, return trueValue else return falseValue
func ValueIfTrue(boolTest bool, trueValue, falseValue string) string {
	if boolTest {
		return trueValue
	}
	return falseValue
}
