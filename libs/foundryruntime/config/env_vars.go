package config

// Env var names for all Foundry runtime options.
// All constants are package-private; callers use LoadFromEnv.
const (
	envCSSTheme          = "FOUNDRY_CSS_THEME"
	envLanguage          = "FOUNDRY_LANGUAGE"
	envPort              = "FOUNDRY_PORT"
	envCompressWebsocket = "FOUNDRY_COMPRESS_WEBSOCKET"
	envMinifyStaticFiles = "FOUNDRY_MINIFY_STATIC_FILES"
	envDeleteNEDB        = "FOUNDRY_DELETE_NEDB"
	envHotReload         = "FOUNDRY_HOT_RELOAD"
	envProxySSL          = "FOUNDRY_PROXY_SSL"
	envUPnP              = "FOUNDRY_UPNP"
	envAWSConfig         = "FOUNDRY_AWS_CONFIG"
	envHostname          = "FOUNDRY_HOSTNAME"
	envLocalHostname     = "FOUNDRY_LOCAL_HOSTNAME"
	envPasswordSalt      = "FOUNDRY_PASSWORD_SALT"
	envProtocol          = "FOUNDRY_PROTOCOL"
	envRoutePrefix       = "FOUNDRY_ROUTE_PREFIX"
	envServiceConfig     = "FOUNDRY_SERVICE_CONFIG"
	envSSLCert           = "FOUNDRY_SSL_CERT"
	envSSLKey            = "FOUNDRY_SSL_KEY"
	envTempDir           = "FOUNDRY_TEMP_DIR"
	envUnixSocket        = "FOUNDRY_UNIX_SOCKET"
	envUPnPLeaseDuration = "FOUNDRY_UPNP_LEASE_DURATION"
	envWorld             = "FOUNDRY_WORLD"
	envProxyPort         = "FOUNDRY_PROXY_PORT"
	envTelemetry         = "FOUNDRY_TELEMETRY"
	envDemoConfig        = "FOUNDRY_DEMO_CONFIG"
	envUpdateChannel     = "FOUNDRY_UPDATE_CHANNEL"
)
