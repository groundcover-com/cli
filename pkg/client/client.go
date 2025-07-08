package client

import (
	"net/http"

	"github.com/groundcover-com/groundcover-sdk-go/pkg/client"
	"github.com/groundcover-com/groundcover-sdk-go/pkg/transport"
)

const (
	// DefaultBaseURL is the default base URL for the groundcover API
	DefaultBaseURL = "https://app.groundcover.com"
)

// CustomTransport is a custom HTTP transport that adds the X-Tenant-UUID header
// to all outgoing requests
type CustomTransport struct {
	http.RoundTripper
	tenantUUID string
}

// RoundTrip implements the http.RoundTripper interface and adds the tenant UUID header
func (ct *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if ct.tenantUUID != "" {
		req.Header.Add("X-Tenant-UUID", ct.tenantUUID)
	}
	return ct.RoundTripper.RoundTrip(req)
}

// NewCustomTransport creates a new CustomTransport with the given tenant UUID
func NewCustomTransport(tenantUUID string) *CustomTransport {
	return &CustomTransport{
		RoundTripper: http.DefaultTransport,
		tenantUUID:   tenantUUID,
	}
}

// SDKClientFactory creates groundcover SDK clients with custom transport
type SDKClientFactory struct {
}

// NewClient creates a new groundcover SDK client with a custom transport
// that includes the tenant UUID in all requests
func (f *SDKClientFactory) NewClient(baseURL, accessToken, backendId, tenantUUID string) (*client.GroundcoverAPI, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// Create custom transport with tenant UUID
	customTransport := NewCustomTransport(tenantUUID)

	// Create SDK client with custom transport
	sdkClient, err := transport.NewSDKClient(accessToken, backendId, baseURL, transport.WithHTTPTransport(customTransport))
	if err != nil {
		return nil, err
	}

	return sdkClient, nil
}

// NewDefaultClient creates a new groundcover SDK client with default settings
// and the provided tenant UUID
func NewDefaultClient(accessToken, backendId, tenantUUID string) (*client.GroundcoverAPI, error) {
	factory := SDKClientFactory{}
	return factory.NewClient("", accessToken, backendId, tenantUUID)
}

// NewClientWithBaseURL creates a new groundcover SDK client with a custom base URL
// and the provided tenant UUID
func NewClientWithBaseURL(accessToken, backendId, tenantUUID, baseURL string) (*client.GroundcoverAPI, error) {
	factory := SDKClientFactory{}
	return factory.NewClient(baseURL, accessToken, backendId, tenantUUID)
}

// NotImplementedError represents an error for unimplemented functionality
type NotImplementedError struct {
	Message string
}

func (e *NotImplementedError) Error() string {
	return e.Message
}
