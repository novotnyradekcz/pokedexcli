package main

import "strings"

func cleanInput(text string) []string {
	// make sure to to trim space inbetween words as well
	// split the text into words
	//
	text = strings.TrimSpace(text)
	var result []string
	for _, word := range strings.Split(text, " ") {
		if word != "" {
			result = append(result, word)
		}
	}
	return result
}
