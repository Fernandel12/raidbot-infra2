package rbapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_UserGetSession(t *testing.T) {
	svc, cleanup := TestingService(t, ServiceOpts{})
	defer cleanup()

	// Let's add some debug logging
	ctx := context.Background()
	t.Log("Creating test context with token")
	ctx = TestingSetContextToken(ctx, t)

	// First call should create the user
	t.Log("Attempting to get user session")
	session, err := svc.UserGetSession(ctx, &UserGetSession_Input{})
	require.NoError(t, err)
	assert.Equal(t, int64(7), session.User.DiscourseId)
}
