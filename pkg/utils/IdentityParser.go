package utils

import "strings"

func IdentityParser(identity string) *map[string]string {
	r := make(map[string]string)
	if !strings.Contains(identity, ";") {
		v := strings.Split(identity, "=")
		r[v[0]] = v[1]
		return &r
	}

	for _, group := range strings.Split(identity, ";") {
		v := strings.Split(group, "=")
		r[v[0]] = v[1]
	}
	return &r
}

func GetPermissionLevel(identityMap *map[string]string, selfBotUsername string) int {
	/* (
	0 = viewer,
	1 = sub,
	2 = founder,
	3 = vip,
	4 = mod,
	5 = broadcaster
	6 = actualgod
	) */
	if v, ok := (*identityMap)["display-name"]; ok && strings.EqualFold(v, selfBotUsername) {
		return 6
	}
	if v, ok := (*identityMap)["badges"]; ok {
		switch {
		case strings.Contains(v, "broadcaster/"):
			return 5
		}
	}
	if v, ok := (*identityMap)["mod"]; ok && v == "1" {
		return 4
	}
	if v, ok := (*identityMap)["vip"]; ok && v == "1" {
		return 3
	}
	if v, ok := (*identityMap)["badges"]; ok {
		switch {
		case strings.Contains(v, "founder/"):
			return 2
		}
	}
	if v, ok := (*identityMap)["subscriber"]; ok && v == "1" {
		return 1
	}

	return 0
}
