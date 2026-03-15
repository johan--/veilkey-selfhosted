package config

import "fmt"

type Config struct {
	TargetUID    uint
	TargetCgroup string
	Format       string
	Once         bool
	OnlySuspicious bool
	EnforceKill    bool
}

func Default() Config {
	return Config{
		Format: "text",
	}
}

func (c Config) Validate() error {
	switch c.Format {
	case "text", "json":
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", c.Format)
	}
}
