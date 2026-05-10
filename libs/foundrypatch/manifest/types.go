package manifest

import "errors"

// SchemaVersion is the version of the manifest schema this package
// understands. Stored as a top-level `version:` field so older
// controllers can refuse manifests they don't understand.
const SchemaVersion = 1

// ActionType enumerates the verbs the applier supports.
type ActionType string

const (
	// ActionDownload writes the body of URL to Dest, verifying SHA256.
	ActionDownload ActionType = "download"
	// ActionZipOverlay extracts the zip at URL on top of Dest, with
	// zip-slip protection.
	ActionZipOverlay ActionType = "zip-overlay"
	// ActionFileReplace writes Content (inline string) to Dest.
	ActionFileReplace ActionType = "file-replace"
)

// File is the on-disk representation of patches/manifest.yaml.
type File struct {
	Version int     `yaml:"version"`
	Patches []Patch `yaml:"patches"`
}

// Patch is one logical hotfix.
type Patch struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description,omitempty"`
	DocURL      string   `yaml:"doc_url,omitempty"`
	Versions    string   `yaml:"versions"` // semver constraint, e.g. ">=0.7.3 <0.7.4"
	Actions     []Action `yaml:"actions"`
}

// Action is a single step within a patch.
type Action struct {
	Type    ActionType `yaml:"type"`
	URL     string     `yaml:"url,omitempty"`
	SHA256  string     `yaml:"sha256,omitempty"`
	Dest    string     `yaml:"dest"`
	Content string     `yaml:"content,omitempty"`
}

// Errors surfaced during validation.
var (
	ErrUnsupportedSchema = errors.New("manifest: unsupported schema version")
	ErrEmptyID           = errors.New("manifest: patch missing id")
	ErrEmptyVersions     = errors.New("manifest: patch missing versions constraint")
	ErrInvalidConstraint = errors.New("manifest: invalid versions constraint")
	ErrUnknownAction     = errors.New("manifest: unknown action type")
	ErrMissingDest       = errors.New("manifest: action missing dest")
	ErrDownloadNeedsURL  = errors.New("manifest: download/zip-overlay requires url+sha256")
)
