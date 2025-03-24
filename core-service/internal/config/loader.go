package config

import (
	"embed"
)

//go:embed registry.json
var configFS embed.FS

// QueryParamDefinition defines a single query parameter's metadata
type QueryParamDefinition struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Format      string      `json:"format,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Example     string      `json:"example,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Max         int         `json:"max,omitempty"`
	Min         int         `json:"min,omitempty"`
}

type QueryParamConfig struct {
	Required []string `json:"required"`
	Optional []string `json:"optional"`
}

// APIResource defines a single API path
type APIResource struct {
	Methods     []string                    `json:"methods"`
	Subhandler  string                      `json:"subhandler"`
	QueryParams map[string]QueryParamConfig `json:"queryParams"`
}

// APIRegistry holds the dynamically loaded API groups
type APIRegistry map[string]map[string]map[string]APIResource

var Registry APIRegistry

var (
	PublicRoutes = []string{
		"/apis/core/v1/auth/login",
		"/apis/core/v1/auth/register",
		"/apis/core/v1/tenants/onboard",
		"/openapi/v2",
	}
)

// func init() {
// 	// Load the API registry from the embedded file
// 	cwd, _ := os.Getwd()
// 	log.Printf("Current working directory: %s", cwd)
// 	data, err := configFS.ReadFile("registry.json")
// 	if err != nil {
// 		log.Fatalf("Failed to read embedded registry.json: %v", err)
// 	}

// 	if err := json.Unmarshal(data, &Registry); err != nil {
// 		log.Fatalf("Failed to parse registry.json: %v", err)
// 	}

// 	log.Println("Static configuration loaded successfully")
// }
