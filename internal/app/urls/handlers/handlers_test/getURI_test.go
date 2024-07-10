package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nomardt/urlshortener-x/internal/app/urls"
	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testGetRequest(t *testing.T, ts *httptest.Server, method,
	path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	client := ts.Client()
	// Preventing the client from following the redirect, ErrUseLastResponse returns nil as err
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func Test_GetURI(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("text/plain"))

	config := newMockConfig("127.0.0.1:8080", "path")

	urlsRepo := urlsInfra.NewInMemoryRepo(config)

	urls.Setup(router, urlsRepo, config)
	ts := httptest.NewServer(router)
	defer ts.Close()

	type args struct {
		requestPath string
	}

	testCases := []struct {
		name         string
		method       string
		expectedCode int
		args         args
	}{
		{
			name:         "POST instead of GET",
			method:       http.MethodPost,
			expectedCode: http.StatusMethodNotAllowed,
			args:         args{"/path"},
		},
		{
			name:         "GET, all ok",
			method:       http.MethodGet,
			expectedCode: http.StatusTemporaryRedirect,
			args:         args{"/path"},
		},
		{
			name:         "GET, invalid path",
			method:       http.MethodGet,
			expectedCode: http.StatusBadRequest,
			args:         args{"/invalidpath"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Adding a shortened URL at /path
			req, _ := testPostRequest(t, ts, "POST", "/", "text/plain", "https://example.com")
			req.Body.Close()

			req, _ = testGetRequest(t, ts, tc.method, tc.args.requestPath)
			defer req.Body.Close()

			assert.Equal(t, tc.expectedCode, req.StatusCode)
		})
	}

}
