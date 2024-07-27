package main

import (
	"context"
	"os"

	"github.com/denzelpenzel/magic-chain/internal/app"
	"github.com/denzelpenzel/magic-chain/internal/client"
	"github.com/denzelpenzel/magic-chain/internal/config"
	"github.com/denzelpenzel/magic-chain/internal/logging"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var (
	cliFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     "env-file",
			EnvVars:  []string{"env-file"},
			Value:    "config.env",
			Usage:    "Set the path to config file",
			Category: "Sources Configuration",
		},
		&cli.StringFlag{
			Name:  "data-dir",
			Value: "tmp",
			Usage: "Set the output dirname",
		},
	}
)

const version = "0.0.1"

// Application driver
func main() {
	ctx := context.Background() // Create context
	logger := logging.WithContext(ctx)

	app := cli.NewApp()
	app.Name = "magic-chain"
	app.Version = version
	app.Description = "A monitoring service that allows for EVM compatible " +
		"blockchains to be continuously assessed for real-time txs"
	app.Action = RunMagicChain
	app.Flags = cliFlags

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal("Error running application", zap.Error(err))
	}
}

// RunMagicChain app entry point
func RunMagicChain(c *cli.Context) error {
	cfg := config.NewConfig(c)
	ctx := context.Background()

	// Create a new logger
	logging.New(cfg.Environment)
	logger := logging.WithContext(ctx)

	bundle, err := client.NewBundle(ctx, cfg.ClientConfig)
	if err != nil {
		logger.Fatal("Error creating client bundle", zap.Error(err))
		return err
	}

	ctx = app.InitContext(ctx, bundle)

	logger.Info("Staring magic-chain application",
		zap.String("version", version),
		zap.String("data-dir", cfg.DataDir),
	)

	magicChain, shutDown, err := app.NewMagicChainApp(ctx, cfg)
	if err != nil {
		logger.Fatal("Error creating application", zap.Error(err))
		return err
	}

	logger.Info("Starting application")
	if err := magicChain.Start(); err != nil {
		logger.Fatal("Error starting application", zap.Error(err))
		return err
	}

	magicChain.ListenForShutdown(shutDown)

	logger.Debug("Waiting for all application threads to end")
	logger.Info("Successful app shutdown")

	return nil
}
