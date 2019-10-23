package model

func option(a, b interface{}) interface{} {
	if a == nil {
		return b
	}
	return a
}
