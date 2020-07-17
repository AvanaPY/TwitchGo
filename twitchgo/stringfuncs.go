package twitchgo

import (
	"strings"
)

func LimitStringLength(s string, length int, fill string) string {
	if len(s) > length {
		s = s[:length-len(fill)] + fill
	}
	return s
}

func CreateTwitchURL(channel string) string {
	return "https://www.twitch.tv/" + strings.ToLower(channel)
}
