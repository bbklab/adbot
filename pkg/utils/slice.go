package utils

// SliceContains check if a slice has given element
func SliceContains(slice []string, ele string) bool {
	for _, val := range slice {
		if val == ele {
			return true
		}
	}
	return false
}

// SliceUniq check if the given slice has duplicated elements
func SliceUniq(slice []string) bool {
	if len(slice) < 2 {
		return true
	}

	ele := make(map[string]int)
	for _, val := range slice {
		if _, ok := ele[val]; ok {
			return false
		}
		ele[val] = 1
	}
	return true
}

// MakeUniq make uniq on the given string slice
func MakeUniq(slice []string) []string {
	m := map[string]bool{}
	for _, val := range slice {
		if _, ok := m[val]; !ok {
			slice[len(m)] = val
			m[val] = true
		}
	}
	return slice[:len(m)]
}
