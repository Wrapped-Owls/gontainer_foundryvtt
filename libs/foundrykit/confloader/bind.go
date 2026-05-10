package confloader

import (
	"fmt"
	"os"
)

type Binder func() error

func BindEnv(binders ...Binder) error {
	for _, b := range binders {
		if err := b(); err != nil {
			return err
		}
	}
	return nil
}

func BindField[T any](ptr *T, envKey string, parse func(string) (T, error)) Binder {
	return func() error {
		v, ok := os.LookupEnv(envKey)
		if !ok || v == "" {
			return nil
		}
		if parse == nil { // only valid when T is string, else it panics
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

func BindFieldPresent[T any](ptr *T, envKey string, parse func(string) (T, error)) Binder {
	return func() error {
		v, ok := os.LookupEnv(envKey)
		if !ok { // an explicitly-set empty string is valid; only an absent var skips
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
