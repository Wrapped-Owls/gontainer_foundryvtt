// Package probe extracts a FoundryVTT version string from either a
// filename or the package.json inside a release artefact (zip entry or
// folder file). It is used by the zip and folder source strategies to
// determine what a candidate would install without performing a full
// extraction.
package probe
