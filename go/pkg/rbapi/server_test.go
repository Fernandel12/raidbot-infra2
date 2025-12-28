package rbapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"rslbot.com/go/internal/testutil"
)

func TestServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := testutil.Logger(t)

	// init server
	server, _, cleanup := TestingServer(t, ctx, ServerOpts{Logger: logger})
	defer cleanup()

	{ // http
		svc := fmt.Sprintf("http://%s", server.ListenerAddr())
		resp, err := http.Get(svc + "/status")
		assert.NoError(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		expected := "{\"everythingIsOk\":true}"
		assert.Equal(t, expected, string(body))
		// FIXME: check rest of the headers (CORS, Content-Type, etc)
	}

	{ // gRPC
		client, cleanup := TestingClient(t, server.ListenerAddr())
		defer cleanup()
		ret, err := client.ToolStatus(ctx, &ToolStatus_Input{})
		assert.NoError(t, err)
		assert.NotNil(t, ret, func() {
			assert.Equal(t, true, ret.EverythingIsOk)
		})
	}
}
