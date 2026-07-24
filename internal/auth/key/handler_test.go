package key

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ulibaysya/caseit/internal/user"
)

// TODO move these things to pkg/
var discardLogger = slog.New(slog.DiscardHandler)

type requstData struct {
	body       string
	bodyReader io.Reader
	ctxFunc    func() context.Context
}

type responseData struct {
	status int
	body   string
}

type serviceData struct {
	numOfCalls int
	authKey    string
	err        error
}

type createTest struct {
	name     string
	requst   requstData
	response responseData
	service  serviceData
}

// TODO test logging
func (test createTest) run(t *testing.T) {
	mockService := &mockService{}
	if test.service.err != nil {
		mockService.returnError(test.service.err)
	}
	if test.service.authKey != "" {
		mockService.returnAuthKey(test.service.authKey)
	}

	rr := httptest.NewRecorder()

	var ctx context.Context
	if test.requst.ctxFunc == nil {
		ctx = t.Context()
	} else {
		ctx = test.requst.ctxFunc()
	}

	if test.requst.bodyReader == nil {
		test.requst.bodyReader = strings.NewReader(test.requst.body)
	}

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "https://api.caseit.io/auth/keys", test.requst.bodyReader)

	// New(slog.New(slog.NewTextHandler(t.Output(), nil)), mockService).Create(rr, req)
	NewHandler(discardLogger, mockService).Create(rr, req)

	if rr.Code != test.response.status {
		t.Fatalf("unexpected status code: %v; expected: %v", rr.Code, test.response.status)
	}

	if body := rr.Body.String(); body != test.response.body {
		t.Fatalf("unexpected body: %v; expected: %v", body, test.response.body)
	}

	if calls := mockService.numberOfCalls(); calls != test.service.numOfCalls {
		t.Fatalf("unexpected number of service calls: %v; expected: %v", calls, test.service.numOfCalls)
	}
}

func TestCreateHandler(t *testing.T) {
	tests := []createTest{
		{
			name: "EmptyBody",
			requst: requstData{
				body: "",
			},
			response: responseData{
				status: http.StatusBadRequest,
				body:   `"empty body"`,
			},
		},
		{
			name: "InvalidJSON",
			requst: requstData{
				body: "NoT jSoN",
			},
			response: responseData{
				status: http.StatusBadRequest,
				body:   `"invalid json"`,
			},
		},
		{
			name: "InappropriateRequestValues",
			requst: requstData{
				body: `{"user_id":"not appropriate"}`,
			},
			response: responseData{
				status: http.StatusBadRequest,
				body:   `"inappropriate request"`,
			},
		},
		{
			name: "InappropriateRequestKeys",
			requst: requstData{
				body: `{"hello":"not appropriate"}`,
			},
			response: responseData{
				status: http.StatusBadRequest,
				body:   `"inappropriate request"`,
			},
		},
		{
			name: "ErrorDecoding",
			requst: requstData{
				bodyReader: errReader{},
			},
			response: responseData{
				status: http.StatusInternalServerError,
				body:   `"Internal Server Error"`,
			},
		},
		{
			name: "InvalidUserID",
			requst: requstData{
				body: `{"user_id":-5}`,
			},
			response: responseData{
				status: http.StatusBadRequest,
				body:   `"invalid user_id: less than zero: -5"`,
			},
		},
		{
			name: "ServiceError",
			requst: requstData{
				body: `{"user_id":18435}`,
			},
			response: responseData{
				status: http.StatusInternalServerError,
				body:   `"Internal Server Error"`,
			},
			service: serviceData{
				numOfCalls: 1,
				err:        errors.New("some error"),
			},
		},
		{
			name: "Success",
			requst: requstData{
				body: `{"user_id":18435}`,
			},
			response: responseData{
				status: http.StatusCreated,
				body:   `"IYOUzS3M9eSY_WB1O0cFj406Tun-L5h8xQvLO6Ds7GM"`,
			},
			service: serviceData{
				numOfCalls: 1,
				authKey:    "IYOUzS3M9eSY_WB1O0cFj406Tun-L5h8xQvLO6Ds7GM",
			},
		},
	}

	for _, i := range tests {
		if !t.Run(i.name, i.run) {
			t.FailNow()
		}
	}
}

type mockService struct {
	calls   int
	err     error
	authKey string
	mu      sync.Mutex
}

func (m *mockService) Create(ctx context.Context, userID user.ID) (authKey string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls += 1

	return m.authKey, m.err
}

func (m *mockService) numberOfCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.calls
}

func (m *mockService) returnError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.err = err
}

func (m *mockService) returnAuthKey(authKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.authKey = authKey
}

type errReader struct{}

func (r errReader) Read([]byte) (int, error) {
	return 0, errors.New("some error")
}
