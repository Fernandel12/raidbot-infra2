package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"raidbot.app/go/pkg/rbapi"
)

var (
	// API Server flags
	bindAddr           string
	dbURN              string
	redisURL           string
	corsAllowedOrigins string
	requestTimeout     time.Duration
	shutdownTimeout    time.Duration
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	RunE:  runAPIServer,
}

func init() {
	// Initialize flags
	apiCmd.Flags().StringVar(&bindAddr, "bind", ":8080", "The address to bind the server to")
	apiCmd.Flags().StringVar(&dbURN, "db-urn", "", "Database URN")
	apiCmd.Flags().StringVar(&redisURL, "redis-url", "localhost:6379", "Redis URL")

	// PayPal configuration
	apiCmd.Flags().StringVar(&rbapi.PayPalClientID, "paypal-client-id", "", "PayPal Client ID")
	apiCmd.Flags().StringVar(&rbapi.PayPalClientSecret, "paypal-client-secret", "", "PayPal Client Secret")
	apiCmd.Flags().StringVar(&rbapi.PayPalWebhookID, "paypal-webhook-id", "", "PayPal Webhook ID")

	// Discord configuration
	apiCmd.Flags().StringVar(&rbapi.DiscordBotToken, "discord-bot-token", "", "Discord bot token for role management")

	apiCmd.Flags().StringVar(&corsAllowedOrigins, "cors-allowed-origins", "*", "Allowed CORS origins")
	apiCmd.Flags().DurationVar(&requestTimeout, "request-timeout", 20*time.Minute, "Request timeout")
	apiCmd.Flags().DurationVar(&shutdownTimeout, "shutdown-timeout", 21*time.Minute, "Shutdown timeout")

	// Mark required flags
	if err := apiCmd.MarkFlagRequired("db-urn"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to mark flag as required: %v\n", err)
	}
}

func runAPIServer(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Initialize PayPal
	rbapi.SetupPayPal()

	// Create service
	svcOpts := rbapi.ServiceOpts{
		Logger:      logger.Named("service"),
		DBUrn:       dbURN,
		RedisConfig: rbapi.RedisConfig{Addr: redisURL},
	}

	svc, err := rbapi.NewService(ctx, svcOpts)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	// Create server
	serverOpts := rbapi.ServerOpts{
		Logger:             logger.Named("server"),
		Bind:               bindAddr,
		CORSAllowedOrigins: corsAllowedOrigins,
		RequestTimeout:     requestTimeout,
		ShutdownTimeout:    shutdownTimeout,
	}

	server, err := rbapi.NewServer(ctx, svc, svc.DB(), svc.Redis(), serverOpts)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	defer server.Close()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// Start server
	logger.Info("starting server",
		zap.String("bind", bindAddr),
		zap.String("cors_allowed_origins", corsAllowedOrigins),
	)

	if err := server.Run(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
