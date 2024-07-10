package urls

import (
	"os"
	"testing"

	"github.com/google/uuid"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
)

func newMockConfig(listenAddress string, path string, storageFile string) conf.Configuration {
	return conf.Configuration{
		ListenAddress: listenAddress,
		Path:          path,
		StorageFile:   storageFile,
		Secret:        "testsecret",
	}
}

func Test_SaveURL(t *testing.T) {
	storageFile := "/tmp/" + uuid.New().String() + ".json"
	repo := NewInMemoryRepo(newMockConfig("127.0.0.1:8080", "", storageFile))

	// Test case: Save a URL
	testURL, _ := urlsDomain.NewURL("https://example.com", "123", "anything")
	err := repo.SaveURL(*testURL, "123")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Remove the temporary url storage file
	err = os.Remove(storageFile)
	if err != nil {
		t.Errorf("Couldn't remove temporary storage file \"%s\". Please remove it manually", storageFile)
	}
}

func Test_GetURL(t *testing.T) {
	storageFile := "/tmp/" + uuid.New().String() + ".json"
	repo := NewInMemoryRepo(newMockConfig("127.0.0.1:8080", "", storageFile))

	// Test case: Get existing URL
	testURL, _ := urlsDomain.NewURL("https://example.com", "123", "random-anything")
	err := repo.SaveURL(*testURL, "123")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	tc := "123"
	foundURL, err := repo.GetURL(tc)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if foundURL != "https://example.com" {
		t.Errorf("Expected URL to be 'https://example.com', got '%s'", foundURL)
	}

	// Test case: Get non-existing URL
	tc = "456"
	foundURL, err = repo.GetURL(tc)
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
	if foundURL != "" {
		t.Errorf("Expected URL to be empty, got '%s'", foundURL)
	}

	// Remove the temporary url storage file
	err = os.Remove(storageFile)
	if err != nil {
		t.Errorf("Couldn't remove temporary storage file \"%s\". Please remove it manually", storageFile)
	}
}
