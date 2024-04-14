package urls

import (
	"testing"

	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
)

func TestInMemoryRepo_SaveURL(t *testing.T) {
	repo := NewInMemoryRepo()

	// Test case: Save a URL
	testURL, _ := urlsDomain.NewURL("https://example.com", "123")
	err := repo.SaveURL(testURL)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestInMemoryRepo_GetURL(t *testing.T) {
	repo := NewInMemoryRepo()

	// Test case: Get existing URL
	testURL, _ := urlsDomain.NewURL("https://example.com", "123")
	repo.SaveURL(testURL)
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
