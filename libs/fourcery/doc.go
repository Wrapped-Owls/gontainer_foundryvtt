// Package fourcery is the unified acquisition + source-handling library
// for FoundryVTT installations: it knows how to obtain bytes (URL, auth
// session, local zip, local folder) and how to normalise them into a
// canonical install tree.
//
// Subpackages:
//
//   - archive: zip detection and extraction with Linux/Node layout
//     normalisation.
//   - auth:    cookie-based credential session to foundryvtt.com.
//   - release: presigned-URL fetch using an authenticated session.
//   - source:  the Source strategy interface and concrete kinds (url,
//     session, zip, folder), plus a Registry factory.
//   - forge:   Builder + orchestrator that resolves a Plan from the
//     registry output and a list of installed candidates, then
//     materialises the chosen source into a versioned install dir.
package fourcery
