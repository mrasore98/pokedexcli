package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mrasore98/pokedexcli/internal/pokecache"
)

var cache *pokecache.Cache

type responseModel struct {
	Next     string              `json:"next"`
	Previous string              `json:"previous"`
	Results  []map[string]string `json:"results"`
}

type apiNav struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(*apiNav) error
}

func commandRegistry(cmdCache *pokecache.Cache) map[string]cliCommand {
	// Update the cache which will be used by commands
	cache = cmdCache
	registeredCommands := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Display 20 locations in the Pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display previous 20 locations in the Pokemon world",
			callback:    commandMapB,
		},
	}

	// Add help command after so it can reference command registry
	registeredCommands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp(registeredCommands),
	}

	return registeredCommands
}

func configRegistry() map[string]*apiNav {
	confMap := map[string]*apiNav{
		"exit": {},
		"help": {},
		"map": {
			Next: "https://pokeapi.co/api/v2/location-area",
		},
	}

	// Use the same pointer for the forward and backward map commands
	confMap["mapb"] = confMap["map"]

	return confMap
}

func runCommand(cmd string, cmdRegistry map[string]cliCommand, confRegistry map[string]*apiNav) error {
	if cliCmd, ok := cmdRegistry[cmd]; ok {
		if urls, ok := confRegistry[cmd]; ok {
			return cliCmd.callback(urls)
		}
	}

	fmt.Println("Unknown command")
	return nil
}

func commandExit(config *apiNav) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cmdRegistry map[string]cliCommand) func(*apiNav) error {
	return func(config *apiNav) error {
		fmt.Println("Welcome to the Pokedex!")
		fmt.Print("Usage:\n\n")
		for _, registeredCmd := range cmdRegistry {
			fmt.Printf("%v: %v\n", registeredCmd.name, registeredCmd.description)
		}
		return nil
	}
}

func commandMap(config *apiNav) error {
	url := config.Next
	decResp, err := makeApiRequest(url)
	if err != nil {
		return err
	}
	config.Next = decResp.Next
	config.Previous = decResp.Previous

	for _, result := range decResp.Results {
		fmt.Println(result["name"])
	}
	return nil
}

func commandMapB(config *apiNav) error {
	url := config.Previous
	if url == "" {
		fmt.Println("you're on the first page")
		return nil
	}
	decResp, err := makeApiRequest(url)
	if err != nil {
		return err
	}
	config.Next = decResp.Next
	config.Previous = decResp.Previous

	for _, result := range decResp.Results {
		fmt.Println(result["name"])
	}
	return nil
}

func makeApiRequest(url string) (responseModel, error) {
	var resBytes []byte
	var ok bool
	decResp := responseModel{}

	// First check for response in cache
	if resBytes, ok = cache.Get(url); !ok {
		// If not in cache, make Http request
		res, err := http.Get(url)
		if err != nil {
			return responseModel{}, fmt.Errorf("could not get requested endpoint: %w", err)
		}
		defer res.Body.Close()
		if resBytes, err = io.ReadAll(res.Body); err != nil {
			return responseModel{}, err
		}
		// Add Http response to cache
		cache.Add(url, resBytes)
	} else {
		fmt.Println("Used cache!!")
	}

	err := json.Unmarshal(resBytes, &decResp)
	if err != nil {
		return responseModel{}, fmt.Errorf("could not decode response: %w", err)
		return responseModel{}, nil
	}
	return decResp, nil
}
