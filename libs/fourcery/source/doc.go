// Package source defines the Source strategy interface and concrete
// implementations for the four ways fourcery can obtain a Foundry
// install: presigned URL, authenticated session, local zip, local
// folder. The Registry factory turns app config + filesystem state
// into a flat list of ready-to-use Source values.
package source
