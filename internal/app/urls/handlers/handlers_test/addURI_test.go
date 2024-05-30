package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls"
	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockConfig(listenAddress string, path string, storageFile string) conf.Configuration {
	return conf.Configuration{
		ListenAddress: listenAddress,
		Path:          path,
		StorageFile:   storageFile,
		Secret:        "hardcodedSecret",
	}
}

// This function was created to test GET requests conveniently
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

	storageFile := "/tmp/" + uuid.New().String() + ".json"
	config := newMockConfig("127.0.0.1:8080", "", storageFile)

	urlsRepo := urlsInfra.NewInMemoryRepo(config)

	urls.Setup(router, urlsRepo, config)
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
			resp, _ := testPostRequest(t, ts, tc.method, tc.args.requestPath, tc.args.contentType, tc.args.urlToShorten)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}

	err := os.Remove(storageFile)
	if err != nil {
		t.Errorf("Couldn't cleanup the db file. Please remove '%s' manually", err)
	}
}

func Test_PostURIJSON(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("application/json"))

	storageFile := "/tmp/" + uuid.New().String()
	config := newMockConfig("127.0.0.1:8080", "", storageFile)
	urlsRepo := urlsInfra.NewInMemoryRepo(config)

	urls.Setup(router, urlsRepo, config)

	ts := httptest.NewServer(router)
	defer ts.Close()

	testCases := []struct {
		name         string
		method       string
		expectedCode int
		body         string
	}{
		{
			name:         "All OK, valid body",
			method:       http.MethodPost,
			expectedCode: 201,
			body:         `{"url":"https://example.com"}`,
		},
		{
			name:         "Invalid JSON",
			method:       http.MethodPost,
			expectedCode: 400,
			body:         `{"url:"https://examplet.com"}`,
		},
		{
			name:         "All OK, but two URLs",
			method:       http.MethodPost,
			expectedCode: 201,
			body:         `{"url":"https://examplez.com", "url":"https://exampled.com"}`,
		},
		{
			name:         "URL duplicate",
			method:       http.MethodPost,
			expectedCode: 409,
			body:         `{"url":"https://example.com"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, ts.URL+"/api/shorten", strings.NewReader(tc.body))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			_, err = io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}

	err := os.Remove(storageFile)
	if err != nil {
		t.Errorf("Couldn't cleanup the db file. Please remove '%s' manually", err)
	}
}

func Test_PostURIBatch(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("application/json"))

	storageFile := "/tmp/" + uuid.New().String()
	config := newMockConfig("127.0.0.1:8080", "", storageFile)
	urlsRepo := urlsInfra.NewInMemoryRepo(config)

	urls.Setup(router, urlsRepo, config)

	ts := httptest.NewServer(router)
	defer ts.Close()

	testCases := []struct {
		name         string
		method       string
		expectedCode int
		body         string
	}{
		{
			name:         "All OK, valid body",
			method:       http.MethodPost,
			expectedCode: 201,
			body:         `[{"correlation_id":"ffd","original_url":"https://testt.com"}]`,
		},
		{
			name:         "Invalid JSON",
			method:       http.MethodPost,
			expectedCode: 400,
			body:         `{"correlation_id":"ffd","original_url":"https://testt.com"}`,
		},
		{
			name:         "All OK, but two URLs",
			method:       http.MethodPost,
			expectedCode: 201,
			body:         `[{"correlation_id":"something1","original_url":"https://testie.com"},{"correlation_id":"something2","original_url":"https://ttest.com"}]`,
		},
		{
			name:         "Correlation ID duplicate",
			method:       http.MethodPost,
			expectedCode: 400,
			body:         `[{"correlation_id":"ffd","original_url":"https://ab.com"},{"correlation_id":"ffdd","original_url":"https://ab.ru"}]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, ts.URL+"/api/shorten/batch", strings.NewReader(tc.body))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			_, err = io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}

	err := os.Remove(storageFile)
	if err != nil {
		t.Errorf("Couldn't cleanup the db file. Please remove '%s' manually", err)
	}
}
