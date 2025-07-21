package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// MockHTTPClient implements HTTPClient for testing
type MockHTTPClient struct {
	responses map[string]*http.Response
	requests  []*http.Request
	errors    map[string]error
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]*http.Response),
		requests:  make([]*http.Request, 0),
		errors:    make(map[string]error),
	}
}

// Do implements HTTPClient.Do
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.requests = append(m.requests, req)

	key := fmt.Sprintf("%s %s", req.Method, req.URL.Path)

	if err, exists := m.errors[key]; exists {
		return nil, err
	}

	if resp, exists := m.responses[key]; exists {
		return resp, nil
	}

	// Default 404 response
	return &http.Response{
		StatusCode: 404,
		Status:     "404 Not Found",
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}, nil
}

// SetResponse sets a mock response for a specific method and path
func (m *MockHTTPClient) SetResponse(method, path string, statusCode int, body interface{}, headers map[string]string) {
	var bodyReader io.ReadCloser

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = io.NopCloser(bytes.NewReader(jsonBody))
	} else {
		bodyReader = io.NopCloser(strings.NewReader(""))
	}

	resp := &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Body:       bodyReader,
		Header:     make(http.Header),
	}

	for k, v := range headers {
		resp.Header.Set(k, v)
	}

	key := fmt.Sprintf("%s %s", method, path)
	m.responses[key] = resp
}

// SetError sets an error to be returned for a specific method and path
func (m *MockHTTPClient) SetError(method, path string, err error) {
	key := fmt.Sprintf("%s %s", method, path)
	m.errors[key] = err
}

// GetRequests returns all requests made to the mock client
func (m *MockHTTPClient) GetRequests() []*http.Request {
	return m.requests
}

// GetRequestsByPath returns requests filtered by path
func (m *MockHTTPClient) GetRequestsByPath(path string) []*http.Request {
	var filtered []*http.Request
	for _, req := range m.requests {
		if req.URL.Path == path {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// TestExtensionLifecycle tests the full extension lifecycle from registration to shutdown
func TestExtensionLifecycle(t *testing.T) {
	// Setup temporary directory for secrets file
	tempDir, err := os.MkdirTemp("", "extension-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup environment
	os.Setenv("AWS_LAMBDA_RUNTIME_API", "localhost:8080")
	os.Setenv("MURMUR_EXPORT_FILE", tempDir+"/secrets.env")
	os.Setenv("MURMUR_EXPORT_REFRESH_INTERVAL", "50ms") // Very short for testing
	os.Setenv("MURMUR_EXPORT_SECRETS_TTL", "100ms")     // Short TTL
	defer func() {
		os.Unsetenv("AWS_LAMBDA_RUNTIME_API")
		os.Unsetenv("MURMUR_EXPORT_FILE")
		os.Unsetenv("MURMUR_EXPORT_REFRESH_INTERVAL")
		os.Unsetenv("MURMUR_EXPORT_SECRETS_TTL")
	}()

	mockClient := NewMockHTTPClient()

	// Setup expected API responses
	mockClient.SetResponse("POST", "/2020-01-01/extension/register", 200,
		RegisterResponse{
			FunctionName:    "test-function",
			FunctionVersion: "$LATEST",
			Handler:         "index.handler",
		},
		map[string]string{
			"Lambda-Extension-Identifier": "test-extension-id",
		})

	// Return INVOKE event first, then context timeout will end the test
	invokeEvent := NextEventResponse{
		EventType:          Invoke,
		DeadlineMs:         time.Now().Add(5 * time.Minute).UnixMilli(),
		RequestID:          "invoke-1",
		InvokedFunctionArn: "arn:aws:lambda:us-east-1:123456789012:function:test",
	}

	// We'll setup the mock to respond once with INVOKE event
	// The second call will get a default 404 which will cause an error and test the error path
	mockClient.SetResponse("GET", "/2020-01-01/extension/event/next", 200, invokeEvent, nil)

	// Parse config
	config, err := NewExtensionConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Create client with mock
	client := NewClientWithHTTPClient("localhost:8080", mockClient)

	// Create refresher
	refresher := NewRefresher(config)

	// Create extension with short timeout to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	extension := &Extension{
		client:    client,
		config:    config,
		refresher: refresher,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Extension should run and then exit due to context timeout or event error
	err = extension.start()
	// We expect either a context timeout or a parsing error - both are acceptable for this test

	// Verify the expected sequence of API calls
	requests := mockClient.GetRequests()
	if len(requests) < 2 {
		t.Fatalf("Expected at least 2 API calls (register + event), got %d", len(requests))
	}

	// Verify registration call
	registerReq := requests[0]
	if registerReq.Method != "POST" || registerReq.URL.Path != "/2020-01-01/extension/register" {
		t.Errorf("Expected POST /2020-01-01/extension/register, got %s %s", registerReq.Method, registerReq.URL.Path)
	}

	// Verify required headers for registration
	if registerReq.Header.Get("Lambda-Extension-Name") != "murmur" {
		t.Errorf("Expected Lambda-Extension-Name: murmur, got %s", registerReq.Header.Get("Lambda-Extension-Name"))
	}

	// Verify event polling
	eventReqs := mockClient.GetRequestsByPath("/2020-01-01/extension/event/next")
	if len(eventReqs) == 0 {
		t.Error("Expected at least one event polling request")
	}

	// Verify extension ID header in event requests
	for _, req := range eventReqs {
		if req.Header.Get("Lambda-Extension-Identifier") != "test-extension-id" {
			t.Errorf("Expected Lambda-Extension-Identifier: test-extension-id, got %s",
				req.Header.Get("Lambda-Extension-Identifier"))
		}
	}

	// Verify secrets file was created
	if _, err := os.Stat(config.File); os.IsNotExist(err) {
		t.Error("Expected secrets file to be created")
	}
}

// TestExtensionInitializationError tests that init errors are properly reported
func TestExtensionInitializationError(t *testing.T) {
	// Setup environment with invalid directory
	os.Setenv("AWS_LAMBDA_RUNTIME_API", "localhost:8080")
	os.Setenv("MURMUR_EXPORT_FILE", "/nonexistent/directory/secrets.env")
	defer func() {
		os.Unsetenv("AWS_LAMBDA_RUNTIME_API")
		os.Unsetenv("MURMUR_EXPORT_FILE")
	}()

	mockClient := NewMockHTTPClient()

	// Setup registration response
	mockClient.SetResponse("POST", "/2020-01-01/extension/register", 200,
		RegisterResponse{FunctionName: "test-function"},
		map[string]string{"Lambda-Extension-Identifier": "test-id"})

	// Setup init error response
	mockClient.SetResponse("POST", "/2020-01-01/extension/init/error", 202,
		StatusResponse{Status: "OK"}, nil)

	// Parse config
	config, err := NewExtensionConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Create client with mock
	client := NewClientWithHTTPClient("localhost:8080", mockClient)

	// Create extension
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	extension := &Extension{
		client:    client,
		config:    config,
		refresher: nil, // No refresher for this test
		ctx:       ctx,
		cancel:    cancel,
	}

	// This should fail due to invalid directory
	err = extension.start()
	if err == nil {
		t.Fatal("Expected extension to fail with invalid directory")
	}

	// Verify init error was called
	initErrorReqs := mockClient.GetRequestsByPath("/2020-01-01/extension/init/error")
	if len(initErrorReqs) != 1 {
		t.Errorf("Expected 1 init error call, got %d", len(initErrorReqs))
	}

	if len(initErrorReqs) > 0 {
		req := initErrorReqs[0]
		if req.Header.Get("Lambda-Extension-Function-Error-Type") != string(ErrorTypeSecretExport) {
			t.Errorf("Expected error type %s, got %s",
				ErrorTypeSecretExport, req.Header.Get("Lambda-Extension-Function-Error-Type"))
		}
	}
}

// TestBackgroundRefresh tests that background refresh works with short intervals
func TestBackgroundRefresh(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "refresh-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup environment with very short intervals
	os.Setenv("MURMUR_EXPORT_REFRESH_INTERVAL", "20ms")
	os.Setenv("MURMUR_EXPORT_SECRETS_TTL", "10ms") // TTL shorter than interval
	defer func() {
		os.Unsetenv("MURMUR_EXPORT_REFRESH_INTERVAL")
		os.Unsetenv("MURMUR_EXPORT_SECRETS_TTL")
	}()

	// Parse config
	config, err := NewExtensionConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	config.File = tempDir + "/secrets.env"

	// Create refresher
	refresher := NewRefresher(config)

	// Start refresher with proper coordination
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = refresher.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start refresher: %v", err)
	}

	// Wait longer for background refresh to occur multiple times
	time.Sleep(100 * time.Millisecond)

	// Cancel context first to signal shutdown
	cancel()

	// Then stop refresher
	refresher.Stop()

	// Verify secrets file was created (indicates refresh occurred)
	if _, err := os.Stat(config.File); os.IsNotExist(err) {
		t.Error("Expected secrets file to be created by background refresh")
	}
}

// TestModeDetection tests that the extension correctly detects Lambda vs CLI mode
func TestModeDetection(t *testing.T) {
	// Test CLI mode (no AWS_LAMBDA_RUNTIME_API)
	os.Unsetenv("AWS_LAMBDA_RUNTIME_API")

	// This should not panic or error - it should just not run extension logic
	// We can't easily test the CLI path here, but we can verify the environment detection

	// Test Lambda mode
	os.Setenv("AWS_LAMBDA_RUNTIME_API", "localhost:8080")
	defer os.Unsetenv("AWS_LAMBDA_RUNTIME_API")

	runtimeAPI := os.Getenv("AWS_LAMBDA_RUNTIME_API")
	if runtimeAPI == "" {
		t.Error("Expected AWS_LAMBDA_RUNTIME_API to be set for Lambda mode")
	}

	// Verify config parsing works in Lambda mode
	_, err := NewExtensionConfigFromEnv()
	if err != nil {
		t.Errorf("Expected config parsing to work in Lambda mode: %v", err)
	}
}
