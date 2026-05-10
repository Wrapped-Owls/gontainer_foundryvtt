package confloader

import (
	"fmt"
	"os"
)

// Binder is a single env-var→field binding created by BindField or BindRequired.
type Binder func() error

// BindEnv runs all supplied binders and returns the first non-nil error.
func BindEnv(binders ...Binder) error {
	for _, b := range binders {
		if err := b(); err != nil {
			return err
		}
	}
	return nil
}

// BindField returns a Binder that reads envKey from the environment,
// calls parse(value) to convert it to T, and stores the result into *ptr.
// If parse is nil, the value is assigned directly when T is string; for
// other T a nil parse panics at binder-construction time.
// If the env var is unset or empty, the field is left unchanged.
func BindField[T any](ptr *T, envKey string, parse func(string) (T, error)) Binder {
	return func() error {
		v, ok := os.LookupEnv(envKey)
		if !ok || v == "" {
			return nil
		}
		if parse == nil {
			// Only works when T is string; checked at compile time by the
			// constraint below — at runtime we use any-cast.
			if s, ok := any(ptr).(*string); ok {
				*s = v
				return nil
			}
			panic(fmt.Sprintf("confloader: BindField: nil parse for non-string type %T", *ptr))
		}
		result, err := parse(v)
		if err != nil {
			return fmt.Errorf("confloader: env %s=%q: %w", envKey, v, err)
		}
		*ptr = result
		return nil
	}
}

// BindFieldPresent is like BindField but treats an explicitly-set empty
// string as a valid value (only skips when the env var is absent entirely).
func BindFieldPresent[T any](ptr *T, envKey string, parse func(string) (T, error)) Binder {
	return func() error {
		v, ok := os.LookupEnv(envKey)
		if !ok {
			return nil
		}
		if parse == nil {
			if s, ok := any(ptr).(*string); ok {
				*s = v
				return nil
			}
			panic(
				fmt.Sprintf("confloader: BindFieldPresent: nil parse for non-string type %T", *ptr),
			)
		}
		result, err := parse(v)
		if err != nil {
			return fmt.Errorf("confloader: env %s=%q: %w", envKey, v, err)
		}
		*ptr = result
		return nil
	}
}

// BindRequired is like BindField but returns an error when the env var is
// absent or empty.
func BindRequired[T any](ptr *T, envKey string, parse func(string) (T, error)) Binder {
	return func() error {
		v, ok := os.LookupEnv(envKey)
		if !ok || v == "" {
			return fmt.Errorf("confloader: required env var %s is not set", envKey)
		}
		if parse == nil {
			if s, ok := any(ptr).(*string); ok {
				*s = v
				return nil
			}
			panic(fmt.Sprintf("confloader: BindRequired: nil parse for non-string type %T", *ptr))
		}
		result, err := parse(v)
		if err != nil {
			return fmt.Errorf("confloader: env %s=%q: %w", envKey, v, err)
		}
		*ptr = result
		return nil
	}
}
