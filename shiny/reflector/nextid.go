package reflector

var lastID int

// nextID is used internally to generate the next element ID
func nextID() int {
	lastID++
	return lastID
}

// resetID resets the ID counter to its initial state.
func resetID() {
	lastID = 0
}
