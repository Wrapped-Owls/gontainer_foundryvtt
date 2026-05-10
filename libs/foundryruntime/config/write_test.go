package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteConfigProducesValidJSON(t *testing.T) {
	c := Default()
	var buf bytes.Buffer
	if err := WriteConfig(&buf, c); err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("WriteConfig output is not valid JSON: %v", err)
	}
}

func TestWriteConfigContainsPort(t *testing.T) {
	c := Default()
	var buf bytes.Buffer
	if err := WriteConfig(&buf, c); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"port": 30000`) {
		t.Errorf("expected indented port in output:\n%s", buf.String())
	}
}

func TestWriteConfigOmitsEmptyOptionals(t *testing.T) {
	c := Default()
	var buf bytes.Buffer
	if err := WriteConfig(&buf, c); err != nil {
		t.Fatal(err)
	}
	var rt map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rt); err != nil {
		t.Fatal(err)
	}
	if _, ok := rt["hostname"]; ok {
		t.Error("hostname should be absent when empty")
	}
	if _, ok := rt["proxyPort"]; ok {
		t.Error("proxyPort should be absent when 0")
	}
	if _, ok := rt["telemetry"]; ok {
		t.Error("telemetry should be absent when nil")
	}
}

func TestWriteConfigDataPath(t *testing.T) {
	c := Default()
	var buf bytes.Buffer
	if err := WriteConfig(&buf, c); err != nil {
		t.Fatal(err)
	}
	var rt map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rt); err != nil {
		t.Fatal(err)
	}
	if rt["dataPath"] != DefaultDataPath {
		t.Errorf("dataPath = %v, want %v", rt["dataPath"], DefaultDataPath)
	}
}
