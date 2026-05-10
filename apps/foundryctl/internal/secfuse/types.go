package secfuse

const DefaultSecretPath = "/run/secrets/config.json"

var KnownKeys = map[string]string{
	"foundry_admin_key":     "FOUNDRY_ADMIN_KEY",
	"foundry_license_key":   "FOUNDRY_LICENSE_KEY",
	"foundry_password":      "FOUNDRY_PASSWORD",
	"foundry_password_salt": "FOUNDRY_PASSWORD_SALT",
	"foundry_service_key":   "FOUNDRY_SERVICE_KEY",
	"foundry_username":      "FOUNDRY_USERNAME",
}

type Result struct {
	SourcePath string
	Applied    []string
	Unknown    []string
}
