package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	c := Default()
	if c.CSSTheme != DefaultCSSTheme {
		t.Errorf("CSSTheme default: want %q, got %q", DefaultCSSTheme, c.CSSTheme)
	}
	if c.DataPath != DefaultDataPath {
		t.Errorf("DataPath default: want %q, got %q", DefaultDataPath, c.DataPath)
	}
	if c.Port != DefaultPort {
		t.Errorf("Port default: want %d, got %d", DefaultPort, c.Port)
	}
	if c.Language != DefaultLanguage {
		t.Errorf("Language default: want %q, got %q", DefaultLanguage, c.Language)
	}
	if c.UpdateChannel != DefaultUpdateChannel {
		t.Errorf("UpdateChannel default: want %q, got %q", DefaultUpdateChannel, c.UpdateChannel)
	}
	// Zero-value optionals in defaults.
	if c.Hostname != "" || c.Cert != "" || c.Telemetry != nil || c.ProxyPort != 0 || c.Demo != nil {
		t.Fatalf("expected zero-value optional fields in default Config: %+v", c)
	}
}

func TestLoadFromEnvBoolsAndStrings(t *testing.T) {
	t.Setenv("FOUNDRY_COMPRESS_WEBSOCKET", "true")
	t.Setenv("FOUNDRY_HOT_RELOAD", "false")
	t.Setenv("FOUNDRY_HOSTNAME", "vtt.example.com")
	t.Setenv("FOUNDRY_TELEMETRY", "true")
	t.Setenv("FOUNDRY_PROXY_PORT", "443")
	t.Setenv("FOUNDRY_CSS_THEME", "fantasy")

	c := Default()
	if err := LoadFromEnv(&c); err != nil {
		t.Fatal(err)
	}

	if !c.CompressSocket {
		t.Error("CompressSocket should be true")
	}
	if c.HotReload {
		t.Error("HotReload should be false (value != 'true')")
	}
	if c.Hostname != "vtt.example.com" {
		t.Errorf("Hostname: got %q", c.Hostname)
	}
	if c.Telemetry == nil || !*c.Telemetry {
		t.Errorf("Telemetry: got %v, want true", c.Telemetry)
	}
	if c.ProxyPort != 443 {
		t.Errorf("ProxyPort: got %d, want 443", c.ProxyPort)
	}
	if c.CSSTheme != "fantasy" {
		t.Errorf("CSSTheme override: got %q", c.CSSTheme)
	}
}

func TestLoadFromEnvProxyPortClamped(t *testing.T) {
	t.Setenv("FOUNDRY_PROXY_PORT", "999999")
	c := Default()
	_ = LoadFromEnv(&c)
	if c.ProxyPort != MaxPort {
		t.Errorf("expected clamp to %d, got %d", MaxPort, c.ProxyPort)
	}

	t.Setenv("FOUNDRY_PROXY_PORT", "-7")
	c = Default()
	_ = LoadFromEnv(&c)
	if c.ProxyPort != MinPort {
		t.Errorf("expected clamp to %d, got %d", MinPort, c.ProxyPort)
	}
}

func TestLoadFromEnvDemoConfigValidates(t *testing.T) {
	demo := `{"worldName":"d","sourceZip":"/d.zip","resetSeconds":3600}`
	t.Setenv("FOUNDRY_DEMO_CONFIG", demo)

	c := Default()
	if err := LoadFromEnv(&c); err != nil {
		t.Fatal(err)
	}
	if string(c.Demo) != demo {
		t.Errorf("demo round-trip mismatch: got %s", c.Demo)
	}

	t.Setenv("FOUNDRY_DEMO_CONFIG", "{not json")
	c = Default()
	if err := LoadFromEnv(&c); err == nil {
		t.Error("expected demo parse error for invalid JSON")
	}
}

func TestWriteConfigRoundTrip(t *testing.T) {
	c := Default()

	var buf bytes.Buffer
	if err := WriteConfig(&buf, c); err != nil {
		t.Fatal(err)
	}

	s := buf.String()
	if !strings.Contains(s, `"port": 30000`) {
		t.Errorf("expected indented port in output:\n%s", s)
	}

	var rt map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rt); err != nil {
		t.Fatal(err)
	}
	if rt["dataPath"] != DefaultDataPath {
		t.Errorf("round-trip lost dataPath: got %v", rt["dataPath"])
	}
	// Optional fields absent in defaults must not appear in JSON.
	if _, ok := rt["hostname"]; ok {
		t.Error("hostname should be absent when empty")
	}
	if _, ok := rt["proxyPort"]; ok {
		t.Error("proxyPort should be absent when 0")
	}
}

func TestWriteConfigFieldMapping(t *testing.T) {
	t.Setenv("FOUNDRY_HOSTNAME", "vtt.example.com")
	t.Setenv("FOUNDRY_SSL_CERT", "/etc/ssl/cert.pem")
	t.Setenv("FOUNDRY_TELEMETRY", "false")
	t.Setenv("FOUNDRY_CSS_THEME", "light")
	t.Setenv("FOUNDRY_PROXY_PORT", "8443")

	c := Default()
	if err := LoadFromEnv(&c); err != nil {
		t.Fatal(err)
	}

	if c.Hostname != "vtt.example.com" {
		t.Errorf("Hostname: %q", c.Hostname)
	}
	if c.Cert != "/etc/ssl/cert.pem" {
		t.Errorf("TLS Cert: %q", c.Cert)
	}
	if c.Telemetry == nil || *c.Telemetry {
		t.Errorf("Telemetry false: got %v", c.Telemetry)
	}
	if c.CSSTheme != "light" {
		t.Errorf("CSSTheme: %q", c.CSSTheme)
	}
	if c.ProxyPort != 8443 {
		t.Errorf("ProxyPort: %d", c.ProxyPort)
	}

	// Verify JSON output uses correct wire field names.
	var buf bytes.Buffer
	if err := WriteConfig(&buf, c); err != nil {
		t.Fatal(err)
	}
	var rt map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rt); err != nil {
		t.Fatal(err)
	}
	if rt["hostname"] != "vtt.example.com" {
		t.Errorf("JSON hostname: %v", rt["hostname"])
	}
	if rt["sslCert"] != "/etc/ssl/cert.pem" {
		t.Errorf("JSON sslCert: %v", rt["sslCert"])
	}
	if rt["telemetry"] != false {
		t.Errorf("JSON telemetry: %v", rt["telemetry"])
	}
}

func TestTelemetryOmittedWhenUnset(t *testing.T) {
	c := Default()
	if c.Telemetry != nil {
		t.Errorf("Telemetry should be nil when unset, got %v", c.Telemetry)
	}
	var buf bytes.Buffer
	if err := WriteConfig(&buf, c); err != nil {
		t.Fatal(err)
	}
	var rt map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rt); err != nil {
		t.Fatal(err)
	}
	if _, ok := rt["telemetry"]; ok {
		t.Error("telemetry key must not appear when unset")
	}
}

func TestHashAdminKey(t *testing.T) {
	got, err := HashAdminKey("hunter2", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != pbkdfKeyLen*2 {
		t.Fatalf("expected %d hex chars, got %d", pbkdfKeyLen*2, len(got))
	}
	got2, _ := HashAdminKey("hunter2", "")
	if got != got2 {
		t.Errorf("hash unstable")
	}
	got3, _ := HashAdminKey("hunter2", "salty")
	if got == got3 {
		t.Errorf("salt had no effect")
	}
}

func TestHashAdminKeyTrimsAndRejectsEmpty(t *testing.T) {
	a, _ := HashAdminKey("  abc  ", "")
	b, _ := HashAdminKey("abc", "")
	if a != b {
		t.Errorf("expected whitespace trim parity")
	}
	if _, err := HashAdminKey("   ", ""); err == nil {
		t.Errorf("empty key should error")
	}
}
