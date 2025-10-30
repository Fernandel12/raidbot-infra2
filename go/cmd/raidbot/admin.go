package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"raidbot.app/go/internal/jsonutil"
	"raidbot.app/go/pkg/rbapi"
	"raidbot.app/go/pkg/rbdb"
)

var (
	userId          int64
	userEmail       string
	licenseDuration string
	licenseKey      string
	searchTerm      string
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin commands",
}

func init() {
	adminCmd.PersistentFlags().StringVar(&serverAddr, "server", "localhost:8080", "Server address")
	adminCmd.AddCommand(activeUsersCmd)

	// Add flags for CreateLicenseCmd
	CreateLicenseCmd.Flags().Int64Var(&userId, "user-id", 0, "User ID to generate license for")
	CreateLicenseCmd.Flags().StringVar(&userEmail, "user-email", "", "User Email to generate license for")
	CreateLicenseCmd.Flags().StringVar(&licenseDuration, "duration", "ONE_MONTH", "License duration (LIFETIME, ONE_DAY, ONE_WEEK, ONE_MONTH, THREE_MONTHS, ONE_YEAR)")

	// Add flags for RevokeLicenseCmd
	RevokeLicenseCmd.Flags().StringVar(&licenseKey, "key", "", "License key to revoke")

	// Add flags for SearchDatabaseCmd
	SearchDatabaseCmd.Flags().StringVar(&searchTerm, "term", "", "Search term to query the database")
	if err := SearchDatabaseCmd.MarkFlagRequired("term"); err != nil {
		panic(fmt.Sprintf("Failed to mark flag as required: %v", err))
	}

	// Add command to parent
	adminCmd.AddCommand(CreateLicenseCmd)
	adminCmd.AddCommand(RevokeLicenseCmd)
	adminCmd.AddCommand(SearchDatabaseCmd)
}

var activeUsersCmd = &cobra.Command{
	Use:   "active-users",
	Short: "Get number of active users",
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

		// Get active users
		resp, err := client.AdminGetActiveUsers(ctx, &rbapi.AdminGetActiveUsers_Input{})
		if err != nil {
			return fmt.Errorf("failed to get active users: %w", err)
		}

		// Print response
		fmt.Println("Active Users Statistics:")
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}
		fmt.Printf("%s\n", string(data))

		return nil
	},
}

var CreateLicenseCmd = &cobra.Command{
	Use:   "create-license",
	Short: "Create a license for a specific user",
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

		// Parse license duration
		var duration rbdb.LicenseKey_Duration
		switch licenseDuration {
		case "LIFETIME":
			duration = rbdb.LicenseKey_LIFETIME
		case "ONE_DAY":
			duration = rbdb.LicenseKey_ONE_DAY
		case "ONE_WEEK":
			duration = rbdb.LicenseKey_ONE_WEEK
		case "ONE_MONTH":
			duration = rbdb.LicenseKey_ONE_MONTH
		case "THREE_MONTHS":
			duration = rbdb.LicenseKey_THREE_MONTHS
		case "ONE_YEAR":
			duration = rbdb.LicenseKey_ONE_YEAR
		default:
			return fmt.Errorf("invalid license duration: %s", licenseDuration)
		}

		// Call AdminAddLicenseKey
		resp, err := client.AdminAddLicenseKey(ctx, &rbapi.AdminAddLicenseKey_Input{
			UserId:    userId,
			UserEmail: userEmail,
			Duration:  duration,
		})
		if err != nil {
			return fmt.Errorf("failed to create license: %w", err)
		}

		// Print response
		fmt.Println("License Created Successfully:")
		fmt.Println(jsonutil.PrettyJSONPB(resp.LicenseKey))

		return nil
	},
}

var RevokeLicenseCmd = &cobra.Command{
	Use:   "revoke-license",
	Short: "Revoke or unrevoke a license key",
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

		// Call AdminRevokeLicense
		resp, err := client.AdminRevokeLicense(ctx, &rbapi.AdminRevokeLicense_Input{
			Key: licenseKey,
		})
		if err != nil {
			return fmt.Errorf("failed to revoke license: %w", err)
		}

		fmt.Println("License successfully revoked:")
		fmt.Println(jsonutil.PrettyJSONPB(resp.LicenseKey))

		return nil
	},
}

var SearchDatabaseCmd = &cobra.Command{
	Use:   "search",
	Short: "Search database for records matching a term",
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

		// Search database
		resp, err := client.AdminSearchDatabase(ctx, &rbapi.AdminSearchDatabase_Input{
			SearchTerm: searchTerm,
		})
		if err != nil {
			return fmt.Errorf("failed to search database: %w", err)
		}

		// Print response
		fmt.Println("Search Results:")

		// Print Users
		if len(resp.Users) > 0 {
			fmt.Println("\nUsers found:", len(resp.Users))
			for _, user := range resp.Users {
				fmt.Println(jsonutil.PrettyJSONPB(user))
			}
		}

		// Print LicenseKeys
		if len(resp.LicenseKeys) > 0 {
			fmt.Println("\nLicense Keys found:", len(resp.LicenseKeys))
			for _, license := range resp.LicenseKeys {
				fmt.Println(jsonutil.PrettyJSONPB(license))
			}
		}

		// Print Payments
		if len(resp.Payments) > 0 {
			fmt.Println("\nPayments found:", len(resp.Payments))
			for _, payment := range resp.Payments {
				fmt.Println(jsonutil.PrettyJSONPB(payment))
			}
		}

		// Print Subscriptions
		if len(resp.Subscriptions) > 0 {
			fmt.Println("\nSubscriptions found:", len(resp.Subscriptions))
			for _, subscription := range resp.Subscriptions {
				fmt.Println(jsonutil.PrettyJSONPB(subscription))
			}
		}

		// Print summary if nothing found
		if len(resp.Users) == 0 &&
			len(resp.LicenseKeys) == 0 &&
			len(resp.Payments) == 0 &&
			len(resp.Subscriptions) == 0 {
			fmt.Println("No results found for search term:", searchTerm)
		}

		return nil
	},
}
