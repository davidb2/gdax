package gdax

// stringFilter filters a slice of strings with a specified function.
// Source: https://gobyexample.com/collection-functions
func stringFilter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// stringMap maps a function on each string in the provided slice.
func stringMap(vs []string, f func(string) string) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		vsf = append(vsf, f(v))
	}
	return vsf
}

// notEmpty returns true if the provided string is not empty.
func notEmpty(s string) bool {
	return s != ""
}
