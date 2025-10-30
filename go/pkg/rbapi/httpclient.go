package rbapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-querystring/query"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type HTTPClient struct {
	baseAPI    string
	httpClient *http.Client
}

func NewHTTPClient(httpClient *http.Client, baseAPI string) *HTTPClient {
	return &HTTPClient{
		baseAPI:    baseAPI,
		httpClient: httpClient,
	}
}

func (c *HTTPClient) AdminAddLicenseKey(ctx context.Context, input *AdminAddLicenseKey_Input) (*AdminAddLicenseKey_Output, error) {
	var result AdminAddLicenseKey_Output
	err := c.doPost(ctx, "/admin/add-license-key", input, &result)
	return &result, err
}

func (c *HTTPClient) AdminGetActiveUsers(ctx context.Context, input *AdminGetActiveUsers_Input) (*AdminGetActiveUsers_Output, error) {
	var result AdminGetActiveUsers_Output
	err := c.doGet(ctx, "/admin/active-users", input, &result)
	return &result, err
}

func (c *HTTPClient) AdminRevokeLicense(ctx context.Context, input *AdminRevokeLicense_Input) (*AdminRevokeLicense_Output, error) {
	var result AdminRevokeLicense_Output
	err := c.doPost(ctx, "/admin/revoke-license-key", input, &result)
	return &result, err
}

func (c *HTTPClient) AdminSearchDatabase(ctx context.Context, input *AdminSearchDatabase_Input) (*AdminSearchDatabase_Output, error) {
	var result AdminSearchDatabase_Output
	err := c.doPost(ctx, "/admin/search-database", input, &result)
	return &result, err
}

func (c *HTTPClient) UserGetSession(ctx context.Context, input *UserGetSession_Input) (*UserGetSession_Output, error) {
	var result UserGetSession_Output
	err := c.doGet(ctx, "/user/session", input, &result)
	return &result, err
}

func (c *HTTPClient) Raw(ctx context.Context, method string, path string, input []byte) ([]byte, error) {
	url := c.baseAPI + path
	b := bytes.NewBuffer(input)

	req, err := http.NewRequestWithContext(ctx, method, url, b)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invalid status code (%d): %q", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *HTTPClient) doPost(ctx context.Context, path string, input, output proto.Message) error {
	// Create marshaling options
	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}

	// Marshal the input message to JSON
	inputBytes, err := marshaler.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	ret, err := c.Raw(ctx, "POST", path, inputBytes)
	if err != nil {
		return fmt.Errorf("raw request failed: %w", err)
	}

	// Create unmarshaling options
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	// Unmarshal response data into the output message
	err = unmarshaler.Unmarshal(ret, output)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *HTTPClient) doGet(ctx context.Context, path string, input, output proto.Message) error {
	qs, err := query.Values(input)
	if err != nil {
		return fmt.Errorf("failed to build query string: %w", err)
	}
	path = path + "?" + qs.Encode()

	ret, err := c.Raw(ctx, "GET", path, nil)
	if err != nil {
		return fmt.Errorf("raw request failed: %w", err)
	}

	// Create unmarshaling options
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	// Unmarshal response data into the output message
	err = unmarshaler.Unmarshal(ret, output)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
