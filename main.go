package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	const prompt string = "Pokedex > "
	cmdRegistry := commandRegistry()
	configRegistry := configRegistry()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(prompt)
		if ok := scanner.Scan(); !ok {
			fmt.Println("encountered an error scanning")
			os.Exit(1)
		}
		input := scanner.Text()
		clean := cleanInput(input)
		cmd := clean[0]
		runCommand(cmd, cmdRegistry, configRegistry)
	}
}

// Split the user input into words based on whitespace.
// Alse lowercases the input and trims surrounding whitespace.
func cleanInput(text string) []string {
	text = strings.Trim(text, " ")
	text = strings.ToLower(text)
	words := strings.Split(text, " ")
	return words
}
