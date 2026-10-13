package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	var trimetName = strings.TrimSpace(name)
	if trimetName == "" {
		return "Hello, World!"
	}
	return fmt.Sprintf("Hello, %s!", trimetName)
}
