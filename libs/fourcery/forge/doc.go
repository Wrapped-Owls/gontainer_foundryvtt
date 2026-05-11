// Package forge orchestrates fourcery's install pipeline. It scans the
// install root for existing candidates, consults the registry's
// enumerated sources, picks the right Source via the Resolver, and
// materialises the chosen Source into a versioned subdirectory.
//
// The public API is a Builder that produces a Forge:
//
//	f, err := forge.New("/foundry").
//	    WithSources(sources...).
//	    WithObserver(forge.SlogObserver{Logger: logger}).
//	    Build()
//	plan, err := f.Resolve(ctx, "14.361.2")
//	inst, err := f.Acquire(ctx, plan)
package forge
