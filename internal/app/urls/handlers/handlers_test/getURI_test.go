package handlers_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/nomardt/urlshortener-x/internal/app/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/auth"
	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func Test_GetURI(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("text/plain"))
	router.Use(auth.WithAuth("hardcodedSecret"))

	storageFile := "/tmp/" + uuid.New().String() + ".json"
	config := newMockConfig("127.0.0.1:8888", "tesst", storageFile)

	urlsRepo := urlsInfra.NewInMemoryRepo(config)

	urls.Setup(router, urlsRepo, config)
	ts := httptest.NewServer(router)
	defer ts.Close()

	testCases := []struct {
		name        string
		requestBody string
	}{
		{
			name:        "all OK",
			requestBody: "https://example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Adding new URL to the db
			resp, _ := testPostRequest(t, ts, http.MethodPost, "", "text/plain", tc.requestBody)
			defer resp.Body.Close()

			assert.Equal(t, 201, resp.StatusCode)

			// Testing shortened URL retrieval
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					// Return an error to prevent following the redirect
					return http.ErrUseLastResponse
				},
			}
			resp2, err := client.Get(ts.URL + "/tesst")
			require.NoError(t, err)
			defer resp2.Body.Close()

			resp2Body, err := io.ReadAll(resp2.Body)
			require.NoError(t, err)

			redirect := resp2.Header.Get("Location")
			assert.Equal(t, 307, resp2.StatusCode)
			assert.Equal(t, tc.requestBody, redirect)
			assert.Equal(t, "", string(resp2Body))
		})
	}

	err := os.Remove(storageFile)
	if err != nil {
		t.Errorf("Couldn't cleanup the db file. Please remove '%s' manually", err)
	}
}

func Test_GetUserURLs(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("text/plain"))
	router.Use(auth.WithAuth("hardcodedSecret"))

	storageFile := "/tmp/" + uuid.New().String() + ".json"
	config := newMockConfig("127.0.0.1:8888", "tesst", storageFile)

	urlsRepo := urlsInfra.NewInMemoryRepo(config)

	urls.Setup(router, urlsRepo, config)
	ts := httptest.NewServer(router)
	defer ts.Close()

	testCases := []struct {
		name         string
		requestBody  string
		responseBody string
	}{
		{
			name:        "all OK",
			requestBody: "https://example.com",
			responseBody: `[
				{
					"short_url": "http://127.0.0.1:8888/tesst",
					"original_url": "https://example.com"
				}
			]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Adding new URL to the db
			resp, _ := testPostRequest(t, ts, http.MethodPost, "", "text/plain", tc.requestBody)
			defer resp.Body.Close()

			assert.Equal(t, 201, resp.StatusCode)

			auth := resp.Header.Get("Authorization")
			if auth == "" {
				assert.Error(t, errors.New("no authorization in response"))
			}

			// Testing shortened URL retrieval
			req2, err := http.NewRequest(http.MethodGet, ts.URL+"/api/user/urls", nil)
			require.NoError(t, err)
			req2.Header.Add("Authorization", auth)

			resp2, err := ts.Client().Do(req2)
			require.NoError(t, err)
			defer resp2.Body.Close()

			resp2Body, err := io.ReadAll(resp2.Body)
			require.NoError(t, err)

			assert.Equal(t, 200, resp2.StatusCode)

			var respExpected []userURLsResponse
			err = json.Unmarshal([]byte(tc.responseBody), &respExpected)
			if err != nil {
				t.Error(err)
			}
			var respActual []userURLsResponse
			err = json.Unmarshal(resp2Body, &respActual)
			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, respExpected, respActual)
		})
	}

	err := os.Remove(storageFile)
	if err != nil {
		t.Errorf("Couldn't cleanup the db file. Please remove '%s' manually", err)
	}
}
