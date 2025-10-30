package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"raidbot.app/go/internal/jsonutil"
	"raidbot.app/go/pkg/rbapi"
)

var (
	useGRPC    bool
	serverAddr string
)

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "CLI for various API endpoints",
}

func init() {
	cliCmd.PersistentFlags().BoolVar(&useGRPC, "grpc", false, "Use gRPC instead of HTTP")
	cliCmd.PersistentFlags().StringVar(&serverAddr, "server", "http://localhost:8080", "Server address")
	cliCmd.AddCommand(sessionCmd)
}

var sessionCmd = &cobra.Command{
	Use:   "@me",
	Short: "Test user session endpoint",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		ctx := cmd.Context()

		// Check if we need to get a new token
		token, err := loadToken()
		if err != nil || token.isExpired() {
			token, err = getNewToken()
			if err != nil {
				return fmt.Errorf("failed to get new token: %w", err)
			}
			if err := saveToken(token); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}
		}

		if useGRPC {
			return testSessionGRPC(ctx, token)
		}
		return testSessionHTTP(ctx, token)
	},
}

func testSessionGRPC(ctx context.Context, token *Token) error {
	// Set up gRPC connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Create a new gRPC client connection
	client, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Add authorization metadata
	md := metadata.New(map[string]string{
		"authorization": fmt.Sprintf("%s %s", token.TokenType, token.AccessToken),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Create gRPC client and call service
	serviceClient := rbapi.NewServiceClient(client)
	resp, err := serviceClient.UserGetSession(ctx, &rbapi.UserGetSession_Input{})
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Print response
	fmt.Printf("User session (gRPC):\n")
	userData, err := json.MarshalIndent(resp.User, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}
	fmt.Printf("%s\n", string(userData))

	return nil
}

func testSessionHTTP(ctx context.Context, token *Token) error {
	// Create HTTP client with auth
	httpClient := &http.Client{
		Transport: &http.Transport{},
	}
	httpClient.Transport = &authTransport{
		token:     token,
		transport: httpClient.Transport,
	}

	// Create API client
	client := rbapi.NewHTTPClient(httpClient, serverAddr)

	// Get session
	resp, err := client.UserGetSession(ctx, &rbapi.UserGetSession_Input{})
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Print response
	fmt.Printf("User session (HTTP):\n")
	fmt.Println(jsonutil.PrettyJSONPB(resp))

	return nil
}
