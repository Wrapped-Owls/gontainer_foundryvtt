// Package ledger tracks which patches have been applied to a Foundry
// install. The ledger is a JSON file at <installRoot>/.foundry-patches.json
// listing applied patches keyed by id + content hash, so the applier
// can skip patches whose definition is unchanged from a previous run.
package ledger
