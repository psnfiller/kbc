package main

import "strings"

func exampleClassify(in string) string {
	switch {
	case strings.Contains(in, "AMAZON"):
		return "Amazon"
	default:
		return defaultClass
	}
}
