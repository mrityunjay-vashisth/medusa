package config

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
)
