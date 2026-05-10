package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

// Default returns a Config with production defaults.
func Default() Config {
	return Config{
		StorageConfig: StorageConfig{DataPath: DefaultDataPath},
		ServerConfig: ServerConfig{
			Port:          DefaultPort,
			Language:      DefaultLanguage,
			CSSTheme:      DefaultCSSTheme,
			UpdateChannel: DefaultUpdateChannel,
		},
	}
}

// LoadFromEnv overlays environment variables onto c using confloader binders.
// Returns an error only for malformed inputs (e.g. invalid FOUNDRY_DEMO_CONFIG JSON).
func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		loadStorageFromEnv(&c.StorageConfig),
		loadServerFromEnv(&c.ServerConfig),
		loadNetworkFromEnv(&c.NetworkConfig),
		loadTLSFromEnv(&c.TLSConfig),
		loadFeaturesFromEnv(&c.FeatureConfig),
		func() error { return bindTelemetry(&c.Telemetry) },
		func() error { return bindDemo(&c.Demo) },
	)
}

func loadStorageFromEnv(c *StorageConfig) confloader.Binder {
	return func() error {
		return confloader.BindEnv(
			confloader.BindField(&c.TempDir, envTempDir, nil),
			confloader.BindField(&c.World, envWorld, nil),
		)
	}
}

func loadServerFromEnv(c *ServerConfig) confloader.Binder {
	return func() error {
		return confloader.BindEnv(
			confloader.BindField(&c.CSSTheme, envCSSTheme, nil),
			confloader.BindField(&c.Language, envLanguage, nil),
			confloader.BindField(&c.UpdateChannel, envUpdateChannel, nil),
			confloader.BindField(&c.Port, envPort, clampedPort),
			confloader.BindField(&c.ServiceConfig, envServiceConfig, nil),
			confloader.BindField(&c.AWSConfig, envAWSConfig, nil),
			confloader.BindField(&c.PasswordSalt, envPasswordSalt, nil),
		)
	}
}

func loadNetworkFromEnv(c *NetworkConfig) confloader.Binder {
	return func() error {
		return confloader.BindEnv(
			confloader.BindField(&c.Hostname, envHostname, nil),
			confloader.BindField(&c.LocalHostname, envLocalHostname, nil),
			confloader.BindField(&c.UnixSocket, envUnixSocket, nil),
			confloader.BindField(&c.Protocol, envProtocol, nil),
			confloader.BindField(&c.RoutePrefix, envRoutePrefix, nil),
			confloader.BindField(&c.ProxyPort, envProxyPort, clampedPort),
			confloader.BindField(&c.ProxySSL, envProxySSL, parseBoolTrue),
			confloader.BindField(&c.UPnP, envUPnP, parseBoolTrue),
			confloader.BindField(&c.UPnPLeaseDuration, envUPnPLeaseDuration, nil),
		)
	}
}

func loadTLSFromEnv(c *TLSConfig) confloader.Binder {
	return func() error {
		return confloader.BindEnv(
			confloader.BindField(&c.Cert, envSSLCert, nil),
			confloader.BindField(&c.Key, envSSLKey, nil),
		)
	}
}

func loadFeaturesFromEnv(c *FeatureConfig) confloader.Binder {
	return func() error {
		return confloader.BindEnv(
			confloader.BindField(&c.CompressSocket, envCompressWebsocket, parseBoolTrue),
			confloader.BindField(&c.CompressStatic, envMinifyStaticFiles, parseBoolTrue),
			confloader.BindField(&c.DeleteNEDB, envDeleteNEDB, parseBoolTrue),
			confloader.BindField(&c.HotReload, envHotReload, parseBoolTrue),
		)
	}
}

// ── private helpers ───────────────────────────────────────────────────────────

// parseBoolTrue returns true when the env value is "true", false otherwise.
func parseBoolTrue(v string) (bool, error) {
	return v == "true", nil
}

// clampedPort parses a port integer and clamps it to [MinPort, MaxPort].
func clampedPort(v string) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("config: invalid port %q: %w", v, err)
	}
	if n < MinPort {
		return MinPort, nil
	}
	if n > MaxPort {
		return MaxPort, nil
	}
	return n, nil
}

// bindTelemetry handles the tri-state FOUNDRY_TELEMETRY (true/false/unset).
// This is the one case where *bool is justified: absent, false, and true are
// three distinct states in Foundry's consent UX.
func bindTelemetry(ptr **bool) error {
	v, ok := os.LookupEnv(envTelemetry)
	if !ok {
		return nil
	}
	switch v {
	case "true":
		t := true
		*ptr = &t
	case "false":
		f := false
		*ptr = &f
	}
	return nil
}

// bindDemo validates and stores FOUNDRY_DEMO_CONFIG as raw JSON.
func bindDemo(ptr *json.RawMessage) error {
	v, ok := os.LookupEnv(envDemoConfig)
	if !ok || v == "" {
		return nil
	}
	var probe any
	if err := json.Unmarshal([]byte(v), &probe); err != nil {
		return fmt.Errorf("config: parse FOUNDRY_DEMO_CONFIG: %w", err)
	}
	*ptr = json.RawMessage(v)
	return nil
}
