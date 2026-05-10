package manifest

import "errors"

const SchemaVersion = 1

type ActionType string

const (
	ActionDownload    ActionType = "download"
	ActionZipOverlay  ActionType = "zip-overlay"
	ActionFileReplace ActionType = "file-replace"
)

type File struct {
	Version int     `yaml:"version"`
	Patches []Patch `yaml:"patches"`
}

type Patch struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description,omitempty"`
	DocURL      string   `yaml:"doc_url,omitempty"`
	Versions    string   `yaml:"versions"`
	Actions     []Action `yaml:"actions"`
}

type Action struct {
	Type    ActionType `yaml:"type"`
	URL     string     `yaml:"url,omitempty"`
	SHA256  string     `yaml:"sha256,omitempty"`
	Dest    string     `yaml:"dest"`
	Content string     `yaml:"content,omitempty"`
}

var (
	ErrUnsupportedSchema = errors.New("manifest: unsupported schema version")
	ErrEmptyID           = errors.New("manifest: patch missing id")
	ErrEmptyVersions     = errors.New("manifest: patch missing versions constraint")
	ErrInvalidConstraint = errors.New("manifest: invalid versions constraint")
	ErrUnknownAction     = errors.New("manifest: unknown action type")
	ErrMissingDest       = errors.New("manifest: action missing dest")
	ErrDownloadNeedsURL  = errors.New("manifest: download/zip-overlay requires url+sha256")
)
