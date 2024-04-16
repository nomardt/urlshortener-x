package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls"
	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockConfig(listenAddress string, path string) conf.Configuration {
	return conf.Configuration{
		ListenAddress: listenAddress,
		Path:          path,
	}
}

func testPostRequest(t *testing.T, ts *httptest.Server, method,
	path string, contentType string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func Test_PostURI(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("text/plain"))
	urlsRepo := urlsInfra.NewInMemoryRepo()

	urls.Setup(router, urlsRepo, newMockConfig("127.0.0.1:8080", ""))
	ts := httptest.NewServer(router)
	defer ts.Close()

	type args struct {
		requestPath  string
		contentType  string
		urlToShorten string
	}

	testCases := []struct {
		name         string
		method       string
		expectedCode int
		args         args
	}{
		{
			name:         "GET instead of POST",
			method:       http.MethodGet,
			expectedCode: http.StatusMethodNotAllowed,
			args:         args{"/", "", ""},
		},
		{
			name:         "POST, all ok",
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
			args:         args{"/", "text/plain", "https://example.com"},
		},
		{
			name:         "POST, invalid schema",
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			args:         args{"/", "text/plain", "gopher://example.com"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testPostRequest(t, ts, tc.method, tc.args.requestPath, tc.args.contentType, tc.args.urlToShorten)
			defer req.Body.Close()

			assert.Equal(t, tc.expectedCode, req.StatusCode)
		})
	}
}
