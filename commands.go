package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mrasore98/pokedexcli/internal/pokecache"
	"github.com/mrasore98/pokedexcli/internal/responses"
)

var cache *pokecache.Cache

type apiNav struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(*apiNav, []string) error
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
		"explore": {
			name:        "explore",
			description: "Explore the area for pokemon. (Provide area name as argument).",
			callback:    commandExplore,
		},
	}

	// Add help command after so it can reference command registry
	registeredCommands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp(registeredCommands, nil),
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
		"explore": {
			Next: "https://pokeapi.co/api/v2/location-area",
		},
	}

	// Use the same pointer for the forward and backward map commands
	confMap["mapb"] = confMap["map"]

	return confMap
}

func runCommand(cmdRegistry map[string]cliCommand, confRegistry map[string]*apiNav, cmd string, params []string) error {
	if cliCmd, ok := cmdRegistry[cmd]; ok {
		if urls, ok := confRegistry[cmd]; ok {
			return cliCmd.callback(urls, params)
		}
	}

	fmt.Println("Unknown command")
	return nil
}

func commandExit(config *apiNav, params []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cmdRegistry map[string]cliCommand, params []string) func(*apiNav, []string) error {
	return func(config *apiNav, params []string) error {
		fmt.Println("Welcome to the Pokedex!")
		fmt.Print("Usage:\n\n")
		for _, registeredCmd := range cmdRegistry {
			fmt.Printf("%v: %v\n", registeredCmd.name, registeredCmd.description)
		}
		return nil
	}
}

func commandMap(config *apiNav, params []string) error {
	url := config.Next
	decResp, err := makeApiRequest(url, responses.ResponseModel{})
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

func commandMapB(config *apiNav, params []string) error {
	url := config.Previous
	if url == "" {
		fmt.Println("you're on the first page")
		return nil
	}
	decResp, err := makeApiRequest(url, responses.ResponseModel{})
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

func commandExplore(conf *apiNav, params []string) error {
	if len(params) < 1 {
		return fmt.Errorf("must provide area name to explore")
	}
	url := conf.Next + "/" + params[0]
	resp, err := makeApiRequest(url, responses.LocationAreaResponse{})
	if err != nil {
		return err
	}
	for _, encounter := range resp.PokemonEncounters {
		fmt.Println(encounter.Pokemon.Name)
	}
	return nil
}

func makeApiRequest[T any](url string, model T) (T, error) {
	var resBytes []byte
	var ok bool

	// First check for response in cache
	// fmt.Println("Making API Request to ", url)
	if resBytes, ok = cache.Get(url); !ok {
		// If not in cache, make Http request
		res, err := http.Get(url)
		if err != nil {
			return model, fmt.Errorf("could not get requested endpoint: %w", err)
		}
		defer res.Body.Close()
		if resBytes, err = io.ReadAll(res.Body); err != nil {
			return model, err
		}
		// Add Http response to cache
		cache.Add(url, resBytes)
	} else {
		fmt.Println("Used cache!!")
	}

	err := json.Unmarshal(resBytes, &model)
	if err != nil {
		return model, fmt.Errorf("could not decode response: %w", err)
	}
	return model, nil
}
