package configs

import (
	"flag"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Debug       bool
	ConfigFile  string
	Environment string
}

func (c Config) String() string {
	return fmt.Sprintf("[ Debug: %v, ConfigFile: %v, Environment: %v ]", c.Debug, c.ConfigFile, c.Environment)
}

func InitApp() *Config {
	//Param
	debug := flag.Bool("debug", false, "sets log level to debug")
	environment := flag.String("env", "", "environment")
	configFile := flag.String("config", "config.toml", "toml configuration file")
	flag.Parse()
	cfg := &Config{Debug: *debug, ConfigFile: *configFile, Environment: *environment}
	// Logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Debug().Msgf("Configuration: %s", cfg)
	return cfg
}
