package config

import (
	"embed"
	"encoding/json"
	"log"
	"os"
)

//go:embed registry.json
var configFS embed.FS

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
	DatabaseNames = struct {
		CoreDB string
		AuthDB string
	}{
		CoreDB: "coredb",
		AuthDB: "authdb",
	}

	CollectionNames = struct {
		OnboardingRequests string
		OnboardedTenants   string
		UserData           string
		Sessions           string
	}{
		OnboardingRequests: "onboarding_requests",
		OnboardedTenants:   "onboarded_tenants",
		UserData:           "user_data",
		Sessions:           "session_store",
	}

	PublicRoutes = []string{
		"/apis/core/v1/auth/login",
		"/apis/core/v1/auth/register",
		"/apis/core/v1/tenants/onboard",
		"/openapi/v2",
	}
)

func init() {
	// Load the API registry from the embedded file
	cwd, _ := os.Getwd()
	log.Printf("Current working directory: %s", cwd)
	data, err := configFS.ReadFile("registry.json")
	if err != nil {
		log.Fatalf("Failed to read embedded registry.json: %v", err)
	}

	if err := json.Unmarshal(data, &Registry); err != nil {
		log.Fatalf("Failed to parse registry.json: %v", err)
	}

	log.Println("Static configuration loaded successfully")
}
