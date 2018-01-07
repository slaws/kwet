package lib

import "regexp"

func ContainsString(list []string, str string) bool {
	for _, b := range list {
		if b == str {
			return true
		}
	}
	return false
}

func MatchStringInList(list []string, str string) bool {
	for _, b := range list {
		match, _ := regexp.MatchString(b, str)
		if match {
			return true
		}
	}
	return false
}

func HasKey(m map[string]interface{}, str string) bool {
	for k, _ := range m {
		if k == str {
			return true
		}
	}
	return false
}
