package urls

import (
	"testing"

	conf "github.com/nomardt/urlshortener-x/cmd/config"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
)

func newMockConfig(listenAddress string, path string) conf.Configuration {
	return conf.Configuration{
		ListenAddress: listenAddress,
		Path:          path,
	}
}

func Test_SaveURL(t *testing.T) {
	repo := NewInMemoryRepo(newMockConfig("127.0.0.1:8080", ""))

	// Test case: Save a URL
	testURL, _ := urlsDomain.NewURL("https://example.com", "123")
	err := repo.SaveURL(testURL)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func Test_GetURL(t *testing.T) {
	repo := NewInMemoryRepo(newMockConfig("127.0.0.1:8080", ""))

	// Test case: Get existing URL
	testURL, _ := urlsDomain.NewURL("https://example.com", "123")
	err := repo.SaveURL(testURL)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	tc := "123"
	foundURL, err := repo.GetURL(&tc)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if foundURL != "https://example.com" {
		t.Errorf("Expected URL to be 'https://example.com', got '%s'", foundURL)
	}

	// Test case: Get non-existing URL
	tc = "456"
	foundURL, err = repo.GetURL(&tc)
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
	if foundURL != "" {
		t.Errorf("Expected URL to be empty, got '%s'", foundURL)
	}
}
