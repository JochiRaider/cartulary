package reference_data

// ValidOperatorBundleName is the Core/Reference Data scalar shared by command
// parsing and the confined incoming storage capability. It admits ASCII only.
func ValidOperatorBundleName(name string) bool {
	if len(name) < 1 || len(name) > 128 {
		return false
	}
	alphanumeric := func(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' }
	for i := range len(name) {
		c := name[i]
		if !alphanumeric(c) && (i == 0 || (c != '.' && c != '_' && c != '+' && c != '-')) {
			return false
		}
	}
	return true
}
