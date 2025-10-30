package rbapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"raidbot.app/go/pkg/errcode"
)

// Discord configuration
var (
	DiscordBotToken string // Set via command-line flag
)

// Hardcoded Discord configuration
const (
	DiscordGuildID        = "1185938052389019769"
	DiscordLifetimeRoleID = "1185971497802682589"
	DiscordAPIBaseURL     = "https://discord.com/api/v10"
)

// GetDiscordIDFromDiscourse fetches the Discord ID from Discourse user profile
func GetDiscordIDFromDiscourse(ctx context.Context, discourseUserID int64) (string, error) {
	// Construct the API URL to get user details with associated accounts
	// Using the admin endpoint to get full user details including associated accounts
	url := fmt.Sprintf("%s/admin/users/%d.json", DiscoursePath, discourseUserID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", errcode.ERR_DISCOURSE_REQUEST_CREATE.Wrap(err)
	}

	// Set API headers
	req.Header.Set("Api-Key", DiscourseAPIKey)
	req.Header.Set("Api-Username", DiscourseAPIUsername)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errcode.ERR_DISCOURSE_API_REQUEST.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errcode.ERR_DISCOURSE_API_RESPONSE.Wrap(
			fmt.Errorf("status %d: %s", resp.StatusCode, string(body)))
	}

	// Read the full response body first for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errcode.ERR_DISCOURSE_RESPONSE_PARSE.Wrap(err)
	}

	// Parse the response
	var discourseResp struct {
		ExternalIDs        map[string]string `json:"external_ids"`
		AssociatedAccounts []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"associated_accounts"`
	}

	if err := json.Unmarshal(body, &discourseResp); err != nil {
		return "", errcode.ERR_DISCOURSE_RESPONSE_PARSE.Wrap(err)
	}

	// The Discord user ID is stored in external_ids
	if discourseResp.ExternalIDs != nil {
		if discordID, ok := discourseResp.ExternalIDs["discord"]; ok && discordID != "" {
			return discordID, nil
		}
	}

	return "", nil // No Discord account linked
}

// AssignDiscordRole assigns the lifetime role to a Discord user
func AssignDiscordRole(ctx context.Context, discordUserID string) error {
	if DiscordBotToken == "" {
		return errcode.ERR_DISCORD_CONFIG_MISSING.Wrap(
			fmt.Errorf("Discord bot token not configured"))
	}

	if DiscordGuildID == "" || DiscordLifetimeRoleID == "" {
		return errcode.ERR_DISCORD_CONFIG_MISSING.Wrap(
			fmt.Errorf("Discord guild ID or role ID not configured"))
	}

	// Discord API URL to add role to guild member
	url := fmt.Sprintf("%s/guilds/%s/members/%s/roles/%s",
		DiscordAPIBaseURL, DiscordGuildID, discordUserID, DiscordLifetimeRoleID)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, nil)
	if err != nil {
		return errcode.ERR_DISCORD_REQUEST_CREATE.Wrap(err)
	}

	// Set Discord bot authorization
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", DiscordBotToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Audit-Log-Reason", "Lifetime license verification")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errcode.ERR_DISCORD_API_REQUEST.Wrap(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		// Success - role added
		return nil
	case http.StatusNotFound:
		// User not in guild
		return errcode.ERR_DISCORD_USER_NOT_IN_GUILD
	case http.StatusForbidden:
		// Bot lacks permissions
		return errcode.ERR_DISCORD_BOT_NO_PERMISSION
	default:
		body, _ := io.ReadAll(resp.Body)
		return errcode.ERR_DISCORD_API_ERROR.Wrap(
			fmt.Errorf("status %d: %s", resp.StatusCode, string(body)))
	}
}

// RemoveDiscordRole removes the lifetime role from a Discord user (for revoked licenses)
func RemoveDiscordRole(ctx context.Context, discordUserID string) error {
	if DiscordBotToken == "" {
		return errcode.ERR_DISCORD_CONFIG_MISSING.Wrap(
			fmt.Errorf("Discord bot token not configured"))
	}

	if DiscordGuildID == "" || DiscordLifetimeRoleID == "" {
		return errcode.ERR_DISCORD_CONFIG_MISSING.Wrap(
			fmt.Errorf("Discord guild ID or role ID not configured"))
	}

	// Discord API URL to remove role from guild member
	url := fmt.Sprintf("%s/guilds/%s/members/%s/roles/%s",
		DiscordAPIBaseURL, DiscordGuildID, discordUserID, DiscordLifetimeRoleID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return errcode.ERR_DISCORD_REQUEST_CREATE.Wrap(err)
	}

	// Set Discord bot authorization
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", DiscordBotToken))
	req.Header.Set("X-Audit-Log-Reason", "License revoked")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errcode.ERR_DISCORD_API_REQUEST.Wrap(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		// Success - role removed
		return nil
	case http.StatusNotFound:
		// User not in guild or doesn't have role - not an error for removal
		return nil
	case http.StatusForbidden:
		// Bot lacks permissions
		return errcode.ERR_DISCORD_BOT_NO_PERMISSION
	default:
		body, _ := io.ReadAll(resp.Body)
		return errcode.ERR_DISCORD_API_ERROR.Wrap(
			fmt.Errorf("status %d: %s", resp.StatusCode, string(body)))
	}
}

// CheckDiscordMembership checks if a Discord user is in the guild
func CheckDiscordMembership(ctx context.Context, discordUserID string) (bool, error) {
	if DiscordBotToken == "" || DiscordGuildID == "" {
		return false, errcode.ERR_DISCORD_CONFIG_MISSING
	}

	// Discord API URL to get guild member
	url := fmt.Sprintf("%s/guilds/%s/members/%s",
		DiscordAPIBaseURL, DiscordGuildID, discordUserID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, errcode.ERR_DISCORD_REQUEST_CREATE.Wrap(err)
	}

	// Set Discord bot authorization
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", DiscordBotToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, errcode.ERR_DISCORD_API_REQUEST.Wrap(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// User is in guild
		return true, nil
	case http.StatusNotFound:
		// User not in guild
		return false, nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return false, errcode.ERR_DISCORD_API_ERROR.Wrap(
			fmt.Errorf("status %d: %s", resp.StatusCode, string(body)))
	}
}
