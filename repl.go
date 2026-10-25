package main

import "strings"

func cleanInput(text string) []string {
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	var result []string
	for _, word := range strings.Split(text, " ") {
		if word != "" {
			result = append(result, word)
		}
	}
	return result
}
