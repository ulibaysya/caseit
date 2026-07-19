package user

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

var discardLogger = slog.New(slog.DiscardHandler)

type create struct {
	testName    string
	requestBody string

	statusCode        int
	responseBody      string
	serviceNumOfCalls int
	serviceError      error
	name, imageURL    string
	location          string
}

func (c create) test(t *testing.T) {
	service := &mockUserService{
		error: c.serviceError,
	}
	rec := httptest.NewRecorder()
	// logger := slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug}))
	// NewCreateHandler(logger, service).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/users", bytes.NewBufferString(c.requestBody)))
	NewCreateHandler(discardLogger, service).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/users", bytes.NewBufferString(c.requestBody)))
	response := rec.Result()
	defer response.Body.Close() //nolint:errcheck

	if response.StatusCode != c.statusCode {
		t.Fatalf("unexpected status code: %v; expected status code: %v", response.StatusCode, c.statusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("unexpected error when reading all body: %v", err)
	}
	if string(responseBody) != c.responseBody {
		t.Fatalf("unexpected response body: %s; expected response body: %v", responseBody, c.responseBody)
	}

	if service.callsNumber != c.serviceNumOfCalls {
		t.Fatalf("service was called unexpected number of times: %v; expected: %v", service.callsNumber, c.serviceNumOfCalls)
	}

	if c.name != "" && service.name != c.name {
		t.Fatalf("unexpected name: %v; expected name: %v", service.name, c.name)
	}
	if c.imageURL != "" && service.imageURL != c.imageURL {
		t.Fatalf("unexpected image_url: %v; expected image_url: %v", service.imageURL, c.imageURL)
	}

	if responseLocation := response.Header.Get("Location"); c.location != "" && responseLocation != c.location {
		t.Fatalf("unexpected Location: %v; expected Location: %v", responseLocation, c.location)
	}
}

// TODO test logs
func TestCreateHandler(t *testing.T) {
	tests := []create{
		{
			testName:     "BadJSON",
			requestBody:  `I am not a JSON`,
			statusCode:   http.StatusBadRequest,
			responseBody: `"invalid json"`,
		},
		{
			testName:     "EmptyRequestBody",
			statusCode:   http.StatusBadRequest,
			responseBody: `"empty body"`,
		},
		{
			testName:          "OmittedName",
			requestBody:       `{"image_url":"https://s3.caseit.net/avatars/wmh7Gc--QRLD0O18Vu-U9rDDTkre8sZIwrcHcbpDYzs.jpg"}`,
			statusCode:        http.StatusBadRequest,
			responseBody:      `"empty name"`,
			serviceNumOfCalls: 0,
		},
		{
			testName:          "imageURLNotURL",
			requestBody:       `{"name":"imechkOkoko","image_url":"I AM NOT AN URL"}`,
			statusCode:        http.StatusBadRequest,
			responseBody:      `"image_url is not an url"`,
			serviceNumOfCalls: 0,
		},
		{
			testName:          "imageURLNotHTTPScheme",
			requestBody:       `{"name":"imechkOkoko","image_url":"ssh://localhost/12345"}`,
			statusCode:        http.StatusBadRequest,
			responseBody:      `"image_url doesn't have an http/https scheme"`,
			serviceNumOfCalls: 0,
		},
		{
			testName:          "ServiceError",
			requestBody:       `{"name":"UsErNaMe","image_url":"https://s3.caseit.net/avatars/wmh7Gc--QRLD0O18Vu-U9rDDTkre8sZIwrcHcbpDYzs.jpg"}`,
			statusCode:        http.StatusInternalServerError,
			responseBody:      `"Internal Server Error"`,
			serviceError:      fmt.Errorf("some service error"),
			serviceNumOfCalls: 1,
		},
		{
			testName:          "Normal",
			requestBody:       `{"name":"UsErNaMe","image_url":"https://s3.caseit.net/avatars/wmh7Gc--QRLD0O18Vu-U9rDDTkre8sZIwrcHcbpDYzs.jpg"}`,
			statusCode:        http.StatusCreated,
			responseBody:      "1",
			serviceNumOfCalls: 1,
			name:              "UsErNaMe",
			imageURL:          "https://s3.caseit.net/avatars/wmh7Gc--QRLD0O18Vu-U9rDDTkre8sZIwrcHcbpDYzs.jpg",
			location:          "/users/1",
		},
	}

	for _, i := range tests {
		t.Run(i.testName, i.test)
	}
}

type mockUserService struct {
	name, imageURL string
	mu             sync.Mutex
	error          error
	callsNumber    int
}

func (m *mockUserService) Create(ctx context.Context, name string, image string) (id int64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callsNumber += 1

	if m.error != nil {
		return 0, m.error
	}

	m.name = name
	m.imageURL = image

	return 1, nil
}
