package check

// StringMapsAreEqual checks whether two maps contain the same data
func StringMapsAreEqual(x, y map[string]string) bool {
	if len(x) != len(y) {
		return false
	}
	for k, v := range x {
		if y[k] != v {
			return false
		}
	}
	return true
}

// StringSlicesAreEqual checks whether two slices contain the same data in the same order
func StringSlicesAreEqual(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	for i, s := range x {
		if y[i] != s {
			return false
		}
	}
	return true
}
