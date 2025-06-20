package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mrasore98/pokedexcli/internal/pokecache"
)

func main() {
	const prompt string = "Pokedex > "
	cache := pokecache.NewCache(30 * time.Second)
	cmdRegistry := commandRegistry(cache)
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
		var params []string = nil
		if len(clean) > 1 {
			params = clean[1:]
		}
		if err := runCommand(cmdRegistry, configRegistry, cmd, params); err != nil {
			fmt.Println(err)
			// os.Exit(1)
		}
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
