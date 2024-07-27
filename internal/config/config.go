package config

import (
	"log"
	"os"
	"strconv"

	"github.com/denzelpenzel/magic-chain/internal/core"
	"github.com/denzelpenzel/magic-chain/internal/logging"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type SystemConfig struct {
	L1PollInterval int
}

// Config app level config defined
type Config struct {
	Environment  core.Env
	DataDir      string
	ClientConfig *core.ClientConfig
	SystemConfig *SystemConfig
}

func NewConfig(c *cli.Context) *Config {
	envFile := c.String("env-file")
	dataDir := c.String("data-dir")

	if err := godotenv.Load(envFile); err != nil {
		logging.NoContext().Warn("config file not found for file: %s", zap.Any("file", envFile))
	}

	return &Config{
		Environment: core.Env(getEnvStr("ENV")),
		DataDir:     dataDir,

		ClientConfig: &core.ClientConfig{
			L1RpcEndpoint: getEnvStr("L1_RPC_ENDPOINT"),
			NumOfRetries:  getEnvInt("NUM_OF_RETRIES"),
		},

		SystemConfig: &SystemConfig{
			L1PollInterval: getEnvInt("L1_POLL_INTERVAL"),
		},
	}
}

// IsProduction Returns true if the env is production
func (cfg *Config) IsProduction() bool {
	return cfg.Environment == core.Production
}

// IsDevelopment Returns true if the env is development
func (cfg *Config) IsDevelopment() bool {
	return cfg.Environment == core.Development
}

// IsLocal Returns true if the env is local
func (cfg *Config) IsLocal() bool {
	return cfg.Environment == core.Local
}

// getEnvStr Reads env var from process environment, panics if not found
func getEnvStr(key string) string {
	envVar, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("could not find env var given key: %s", key)
	}
	return envVar
}

// getEnvInt ... Reads env vars and converts to int
func getEnvInt(key string) int {
	val := getEnvStr(key)
	intRep, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("env val is not int; got: %s=%s; err: %s", key, val, err.Error())
	}
	return intRep
}
