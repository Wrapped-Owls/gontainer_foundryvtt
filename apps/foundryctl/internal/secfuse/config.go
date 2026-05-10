package secfuse

import "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"

const envSecretFilePath = "FOUNDRY_SECRET_FILE"

type Config struct {
	Path string
}

func Default() Config {
	return Config{Path: DefaultSecretPath}
}

func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		confloader.BindField(&c.Path, envSecretFilePath, nil),
	)
}
