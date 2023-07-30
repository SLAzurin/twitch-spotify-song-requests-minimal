package utils

import "strings"

func RawIRCUserToUsername(rawIRCUsername string) string {
	if len(rawIRCUsername) == 0 {
		return rawIRCUsername
	}
	if !strings.HasPrefix(rawIRCUsername, ":") || !strings.Contains(rawIRCUsername[1:], "!") {
		return rawIRCUsername
	}
	return rawIRCUsername[1:strings.Index(rawIRCUsername, "!")]
}
