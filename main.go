package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			continue
		}
		line := scanner.Text()
		if line == "" {
			continue
		}
		firstWord := cleanInput(line)[0]
		fmt.Println("Your command was: " + firstWord)
	}
}
