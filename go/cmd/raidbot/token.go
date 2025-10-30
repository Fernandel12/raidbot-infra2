package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"raidbot.app/go/pkg/rbapi"
	"raidbot.app/go/pkg/rbdb"
)

const (
	tokenFile = "raidbot_token.json"
)

type Token struct {
	AccessToken string              `json:"access_token"`
	TokenType   string              `json:"token_type"`
	ExpiresAt   time.Time           `json:"expires_at"`
	UserInfo    *rbdb.DiscourseUser `json:"user_info"`
}

func (t *Token) isExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func loadToken() (*Token, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func saveToken(token *Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenFile, data, 0600)
}

func getNewToken() (*Token, error) {
	// Create a channel to receive the token
	tokenChan := make(chan *Token)
	errChan := make(chan error)

	// Start local server for callback
	server := &http.Server{Addr: ":8085"}

	// Handler for initial page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create SSO request like in discourse-test.go
		nonce := fmt.Sprintf("nonce-%d", time.Now().UnixNano())
		returnURL := "http://localhost:8085/callback"
		payload := fmt.Sprintf("return_sso_url=%s&nonce=%s",
			url.QueryEscape(returnURL),
			url.QueryEscape(nonce))

		// Base64 encode the payload
		base64Payload := base64.StdEncoding.EncodeToString([]byte(payload))

		// Generate signature using rbapi.DiscourseSecret
		mac := hmac.New(sha256.New, []byte(rbapi.DiscourseSecret))
		mac.Write([]byte(base64Payload))
		sig := hex.EncodeToString(mac.Sum(nil))

		// Construct SSO URL
		ssoURL := fmt.Sprintf("%s/session/sso_provider?sso=%s&sig=%s",
			rbapi.DiscoursePath,
			url.QueryEscape(base64Payload),
			sig)

		// Display page with auto-redirect
		fmt.Fprintf(w, `
			<html>
				<body>
					<h1>RaidBot SSO Login</h1>
					<p>Redirecting to Discourse for authentication...</p>
					<script>
						window.location.href = "%s";
					</script>
				</body>
			</html>
		`, ssoURL)
	})

	// Handler for callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Get SSO response parameters
		sso := r.URL.Query().Get("sso")
		sig := r.URL.Query().Get("sig")

		if sso == "" || sig == "" {
			err := fmt.Errorf("missing SSO parameters")
			logger.Error("Callback error", zap.Error(err))
			errChan <- err
			http.Error(w, "Missing SSO parameters", http.StatusBadRequest)
			return
		}

		userInfo, err := rbapi.VerifySSO(sso, sig, rbapi.DiscourseSecret)
		if err != nil {
			logger.Error("SSO verification failed", zap.Error(err))
			errChan <- fmt.Errorf("SSO verification failed: %w", err)
			http.Error(w, "Invalid SSO response", http.StatusBadRequest)
			return
		}

		// Create token
		token := &Token{
			AccessToken: fmt.Sprintf("SSO_%s.%s", sso, sig),
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserInfo:    userInfo,
		}

		// Display success page with token info
		fmt.Fprintf(w, `
			<html>
				<body style="font-family: sans-serif; max-width: 800px; margin: 0 auto; padding: 20px;">
					<h1>SSO Authentication Successful!</h1>
					<div style="background: #f0f0f0; padding: 20px; border-radius: 5px; margin: 20px 0;">
						<h2>User Information:</h2>
						<p><strong>Username:</strong> %s</p>
						<p><strong>Email:</strong> %s</p>
						<p><strong>External ID:</strong> %d</p>
						<p><strong>Groups:</strong> %s</p>
						<p><strong>Admin:</strong> %v</p>
					</div>
					<p>You can now close this window and return to the CLI.</p>
				</body>
			</html>
		`, userInfo.Username, userInfo.Email, userInfo.ExternalId, strings.Join(userInfo.Groups, ", "), userInfo.Admin)

		// Send token through channel
		tokenChan <- token

		// Shutdown server after short delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			if err := server.Shutdown(context.Background()); err != nil {
				logger.Error("Error shutting down server", zap.Error(err))
			}
		}()
	})

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	fmt.Printf("Please visit http://localhost:8085 to authenticate\n")

	// Wait for token or error
	select {
	case token := <-tokenChan:
		return token, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timeout")
	}
}

// authTransport adds authentication headers to requests
type authTransport struct {
	token     *Token
	transport http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", t.token.TokenType, t.token.AccessToken))
	return t.transport.RoundTrip(req)
}
