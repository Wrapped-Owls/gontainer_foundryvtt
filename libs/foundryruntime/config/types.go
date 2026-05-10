package config

import "encoding/json"

const (
	DefaultCSSTheme      = "dark"
	DefaultDataPath      = "/data"
	DefaultPort          = 30000
	DefaultLanguage      = "en.core"
	DefaultUpdateChannel = "stable"
	MinPort              = 1
	MaxPort              = 65535
)

type Config struct {
	StorageConfig
	ServerConfig
	NetworkConfig
	TLSConfig
	FeatureConfig
	Demo json.RawMessage `json:"demo,omitempty"`
}

type StorageConfig struct {
	DataPath string `json:"dataPath"`
	TempDir  string `json:"tempDir,omitempty"`
	World    string `json:"world,omitempty"`
}

type ServerConfig struct {
	Port          int    `json:"port"`
	Language      string `json:"language"`
	CSSTheme      string `json:"cssTheme"`
	UpdateChannel string `json:"updateChannel"`
	ServiceConfig string `json:"serviceConfig,omitempty"`
	AWSConfig     string `json:"awsConfig,omitempty"`
	PasswordSalt  string `json:"passwordSalt,omitempty"`
}

type NetworkConfig struct {
	Hostname          string `json:"hostname,omitempty"`
	LocalHostname     string `json:"localHostname,omitempty"`
	UnixSocket        string `json:"unixSocket,omitempty"`
	Protocol          string `json:"protocol,omitempty"`
	RoutePrefix       string `json:"routePrefix,omitempty"`
	ProxyPort         int    `json:"proxyPort,omitempty"`
	ProxySSL          bool   `json:"proxySSL"`
	UPnP              bool   `json:"upnp"`
	UPnPLeaseDuration string `json:"upnpLeaseDuration,omitempty"`
}

type TLSConfig struct {
	Cert string `json:"sslCert,omitempty"`
	Key  string `json:"sslKey,omitempty"`
}

type FeatureConfig struct {
	CompressSocket bool  `json:"compressSocket"`
	CompressStatic bool  `json:"compressStatic"`
	DeleteNEDB     bool  `json:"deleteNEDB"`
	HotReload      bool  `json:"hotReload"`
	Fullscreen     bool  `json:"fullscreen"`
	Telemetry      *bool `json:"telemetry,omitempty"`
}
