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
	// Try the file at the root first, then fall back to the original path
	data, err := os.ReadFile("/registry.json")
	if err != nil {
		// Fall back to the original path
		data, err = os.ReadFile("../internal/apiserver/registry.json")
		if err != nil {
			return nil, err
		}
	}

	var registry APIRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}

	log.Println("Loaded API Registry successfully")
	return registry, nil
}
