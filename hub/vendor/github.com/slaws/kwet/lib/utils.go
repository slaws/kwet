package lib

func ContainsString(list []string, str string) bool {
	for _, b := range list {
		if b == str {
			return true
		}
	}
	return false
}
