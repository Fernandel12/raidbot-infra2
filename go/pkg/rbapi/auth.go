package rbapi

import (
	"context"
	"net/url"
	"strings"

	"google.golang.org/grpc/metadata"
	"raidbot.app/go/pkg/errcode"
	"raidbot.app/go/pkg/rbdb"
)

type ctxKey string

const (
	userInfoCtx     ctxKey = "user-info"
	adminGroup      ctxKey = "admin"
	DiscourseSecret string = "hHDjDtwn6ADb3Gv"
)

func (svc *service) AuthFuncOverride(ctx context.Context, path string) (context.Context, error) {
	// Check if this is a public endpoint that doesn't require authentication
	publicEndpoints := []string{
		"/raidbot.api.Service/ToolStatus",
		"/raidbot.api.Service/PublicBuildGet",
		"/raidbot.api.Service/PublicBuildList",
		"/raidbot.api.Service/PublicPickitGet",
		"/raidbot.api.Service/PublicPickitList",
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return ctx, nil
		}
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errcode.ERR_AUTH_MISSING_METADATA
	}

	auth, ok := md["authorization"]
	if !ok || len(auth) < 1 {
		return nil, errcode.ERR_AUTH_MISSING_TOKEN
	}

	// Clean up the authorization header
	tokenString := strings.TrimPrefix(auth[0], "Bearer ")

	// Check if this is an SSO token
	if strings.HasPrefix(tokenString, "SSO_") {
		// Extract SSO and signature from the token
		parts := strings.Split(strings.TrimPrefix(tokenString, "SSO_"), ".")
		if len(parts) != 2 {
			return nil, errcode.ERR_AUTH_INVALID_TOKEN
		}

		// URL decode the token part if needed since it might be URL-encoded
		decodedToken := parts[0]
		if unescaped, err := url.QueryUnescape(decodedToken); err == nil {
			decodedToken = unescaped
		}

		userInfo, err := VerifySSO(decodedToken, parts[1], DiscourseSecret)
		if err != nil {
			return nil, errcode.ERR_AUTH_INVALID_TOKEN.Wrap(err)
		}
		ctx = context.WithValue(ctx, userInfoCtx, userInfo)
		return ctx, nil
	}

	// Fall back to API token verification
	userInfo, err := VerifyTokenAndGetUser(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, userInfoCtx, userInfo)
	return ctx, nil
}

func discourseUserFromContext(ctx context.Context) (*rbdb.DiscourseUser, error) {
	discourseUser := ctx.Value(userInfoCtx)
	if discourseUser == nil {
		return nil, errcode.ERR_AUTH_MISSING_CONTEXT
	}
	return discourseUser.(*rbdb.DiscourseUser), nil
}

func isAdmin(ctx context.Context) bool {
	discourseUser, err := discourseUserFromContext(ctx)
	if err != nil {
		return false
	}

	return discourseUser.Admin
}
