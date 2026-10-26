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
		if cmd, ok := commands[firstWord]; ok {
			if err := cmd.callback(); err != nil {
				fmt.Println(err)
			}
			continue
		}
		fmt.Println("Unknown command:", firstWord)
	}
}
