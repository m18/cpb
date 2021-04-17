package check

// StringMapsAreEqual checks if two maps contain the same data.
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

// StringToSimpleTypeMapsAreEqual checks if two maps contain the same data.
func StringToSimpleTypeMapsAreEqual(x, y map[string]interface{}) bool {
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

// StringSetsAreEqual checks if two sets containf the same items.
func StringSetsAreEqual(x map[string]struct{}, y map[string]struct{}) bool {
	if len(x) != len(y) {
		return false
	}
	for k := range x {
		if _, ok := y[k]; !ok {
			return false
		}
	}
	return true
}
