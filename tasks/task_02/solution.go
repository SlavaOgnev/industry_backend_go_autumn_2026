package main

func rotateRunes(s string, shift int) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}
	shift = shift % len(runes)

	if shift < 0 {
		newshift := len(runes) + shift
		return string(runes[newshift:]) + string(runes[:newshift])
	}
	return string(runes[shift:]) + string(runes[:shift])

}
