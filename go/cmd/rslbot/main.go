package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Use anonymous function to handle logger.Sync error and ensure it runs
	defer func() {
		// Sync can sometimes fail with "invalid argument" on certain platforms
		// when writing to stderr - this is a known issue with zap
		_ = logger.Sync()
	}()

	// Execute root command and handle errors without immediate exit
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", zap.Error(err))
		// Set exit code but don't exit immediately to allow deferred functions to run
		defer os.Exit(1)
	}
}
