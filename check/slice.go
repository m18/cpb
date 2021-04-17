package check

type slicesInterface interface {
	LenX() int
	LenY() int
	AreEqual(i int) bool
}

func slicesAreEqual(s slicesInterface) bool {
	if s.LenX() != s.LenY() {
		return false
	}
	for i := 0; i < s.LenX(); i++ {
		if !s.AreEqual(i) {
			return false
		}
	}
	return true
}

type byteSlices struct {
	x, y []byte
}

func (s *byteSlices) LenX() int           { return len(s.x) }
func (s *byteSlices) LenY() int           { return len(s.y) }
func (s *byteSlices) AreEqual(i int) bool { return s.x[i] == s.y[i] }

type stringSlices struct {
	x, y []string
}

func (s *stringSlices) LenX() int           { return len(s.x) }
func (s *stringSlices) LenY() int           { return len(s.y) }
func (s *stringSlices) AreEqual(i int) bool { return s.x[i] == s.y[i] }

// ByteSlicesAreEqual checks if two slices contain the same items in the same order.
func ByteSlicesAreEqual(x, y []byte) bool {
	b := &byteSlices{x, y}
	return slicesAreEqual(b)
}

// StringSlicesAreEqual checks if two slices contain the same items in the same order.
func StringSlicesAreEqual(x, y []string) bool {
	b := &stringSlices{x, y}
	return slicesAreEqual(b)
}
