package util

func Contains[T comparable](slice []T, given T) bool {
	for _, current := range slice {
		if current == given {
			return true
		}
	}

	return false
}
