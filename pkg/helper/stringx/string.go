package stringx

// StringCoalesce returns the first non-empty string value
func StringCoalesce(str ...string) string {
	for _, s := range str {
		if s != "" {
			return s
		}
	}
	return ""
}
