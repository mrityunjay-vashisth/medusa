package apiserver

import (
	"encoding/json"
	"log"
	"os"
)

// APIResource defines a single API path
type APIResource struct {
	Methods    []string `json:"methods"`
	Subhandler string   `json:"subhandler"`
}

// APIRegistry holds the dynamically loaded API groups
type APIRegistry map[string]map[string]map[string]APIResource

// LoadRegistry loads API definitions from `apiserver/registry.json`
func LoadRegistry() (APIRegistry, error) {
	// Directly reference the relative path
	registryPath := "../internal/apiserver/registry.json"

	// Read the file
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, err
	}

	var registry APIRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}

	log.Println("Loaded API Registry from", registryPath)
	return registry, nil
}
