package helpers

import "strings"

// ToBool convert string from config to bool
func ToBool(s string) bool {
	ls := strings.ToLower(s)
	for _, t := range []string{"1", "on", "y", "yes", "t", "true", "enable", "enabled"} {
		if t == ls {
			return true
		}
	}
	return false
}
