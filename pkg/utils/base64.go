package utils

// Base64LineBreaker break any string with maximum 64 bytes per line
func Base64LineBreaker(s string) string {
	var (
		lineMax = 64
		start   int
		max     = len(s)
		breaked string
	)

	for {
		end := start + lineMax
		if end >= max {
			end = max
			breaked += string(s[start:end]) + "\n"
			break
		}
		breaked += string(s[start:end]) + "\n"
		start = end
	}

	return breaked
}
