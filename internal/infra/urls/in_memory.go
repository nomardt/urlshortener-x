package urls

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/google/uuid"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
)

type InMemoryRepo struct {
	urls      []urlInFile
	users     []userInFile
	file      string
	fileUsers string
	mu        sync.Mutex
}

type urlInFile struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
	IsDeleted     bool   `json:"is_deleted"`
}

type userInFile struct {
	ID     uuid.UUID `json:"id"`
	UserID string    `json:"user_id"`
	URLID  string    `json:"url_id"`
}

// Create a new Repo which consists of urls map[string]string
func NewInMemoryRepo(config conf.Configuration) *InMemoryRepo {
	inMemoryRepo := &InMemoryRepo{
		urls:  make([]urlInFile, 0),
		users: make([]userInFile, 0),
	}

	// Constructing path for the second file which will contain users
	// by appending "2" to the filename specified in config
	var pathUsers string
	pathUsersSlice := strings.Split(config.StorageFile, ".")
	if len(pathUsersSlice) > 1 {
		nameWithoutExtension := strings.Join(pathUsersSlice[:len(pathUsersSlice)-1], ".")
		extension := pathUsersSlice[len(pathUsersSlice)-1]
		pathUsers = fmt.Sprintf("%s2.%s", nameWithoutExtension, extension)
	} else {
		pathUsers = config.StorageFile + "2"
	}

	// Loading URLs previously stored on the hard drive. If it fails then create new files
	if err := inMemoryRepo.loadStoredURLs(config.StorageFile); err != nil {
		logger.Log.Info("Couldn't recover any previously shortened URLs!", zap.Error(err))

		// Creating a new file for urls
		if _, err := os.Create(config.StorageFile); err != nil {
			logger.Log.Info("No file will be created to store shortened URLs")
		}

		// Creating a new file for users
		if _, err := os.Create(pathUsers); err != nil {
			logger.Log.Info("No file will be created to store users", zap.Error(err))
		}

		inMemoryRepo.fileUsers = pathUsers
	}

	if err := inMemoryRepo.loadStoredUsers(pathUsers); err != nil {
		logger.Log.Info("Couldn't load previously stored user URL relations", zap.Error(err))
	}

	return inMemoryRepo
}

// Add the specified URL to the Repo
func (r *InMemoryRepo) SaveURL(url urlsDomain.URL, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Checking if the provided correlation ID is unique
	for _, savedURL := range r.urls {
		if savedURL.CorrelationID == url.CorrelationID() {
			logger.Log.Info(ErrCorIDNotUnique.Error(), zap.String("correlation_id", url.CorrelationID()))
			return ErrCorIDNotUnique
		}
	}

	// Checking if the provided full URI is unique
	for _, savedURL := range r.urls {
		if savedURL.OriginalURL == url.LongURL() && !savedURL.IsDeleted {
			logger.Log.Info("The specified full URI already exists", zap.String("full_uri", url.LongURL()))

			// Checking if the URL is already attached to the current user
			newEntryNeeded := true
			for _, userEntry := range r.users {
				if userEntry.URLID == savedURL.CorrelationID && userEntry.UserID == userID {
					newEntryNeeded = false
					break
				}
			}
			if newEntryNeeded {
				// Attaching the URL to user in file
				fileUsers, err := os.OpenFile(r.fileUsers, os.O_WRONLY|os.O_APPEND, 0666)
				if err != nil {
					logger.Log.Info("Couldn't open the users file", zap.String("file", r.fileUsers), zap.Error(err))
				}

				jsonUser := &userInFile{
					ID:     uuid.New(),
					UserID: userID,
					URLID:  savedURL.CorrelationID,
				}
				dataUser, err := json.Marshal(jsonUser)
				if err != nil {
					logger.Log.Info("Couldn't attach already shortened URL to user", zap.Error(err))
				}
				dataUser = append(dataUser, '\n')

				_, err = fileUsers.Write(dataUser)
				if err != nil {
					logger.Log.Info("Couldn't write URL relation to user", zap.Error(err))
				}
				fileUsers.Close() //nolint:all

				r.users = append(r.users, *jsonUser)
			}

			return newErrURINotUnique(savedURL.ShortURL)
		}
	}

	// Saving the new URL on the hard drive
	fileURL, err := os.OpenFile(r.file, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	jsonURL := &urlInFile{
		CorrelationID: url.CorrelationID(),
		ShortURL:      url.ID(),
		OriginalURL:   url.LongURL(),
		IsDeleted:     false,
	}
	dataURL, err := json.Marshal(jsonURL)
	if err != nil {
		logger.Log.Info("Couldn't store the shortened URL in the file", zap.Error(err))
		return err
	}
	dataURL = append(dataURL, '\n')

	_, err = fileURL.Write(dataURL)
	if err != nil {
		logger.Log.Info("Couldn't save the shortened URL to the file", zap.String("file", r.file), zap.Error(err))
		return err
	}
	fileURL.Close() //nolint:all

	r.urls = append(r.urls, *jsonURL)

	// Attaching the new URL to user on the hard drive
	fileUsers, err := os.OpenFile(r.fileUsers, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Log.Info("Couldn't open users file", zap.Error(err))
		return err
	}

	jsonUser := &userInFile{
		ID:     uuid.New(),
		UserID: userID,
		URLID:  url.CorrelationID(),
	}
	dataUser, err := json.Marshal(jsonUser)
	if err != nil {
		logger.Log.Info("Couldn't attach already shortened URL to user", zap.Error(err))
		return err
	}
	dataUser = append(dataUser, '\n')

	_, err = fileUsers.Write(dataUser)
	if err != nil {
		logger.Log.Info("Couldn't write URL relation to user", zap.Error(err))
		return err
	}
	fileUsers.Close() //nolint:all

	r.users = append(r.users, *jsonUser)

	return nil
}

// Sets the flag is_deleted to true if the shortened URL belongs to user
func (r *InMemoryRepo) DeleteURL(key string, userID string) error {
	// Checking if the shortened URL belongs to the specified user
	urlNotFound := true
UserCheck:
	for i, savedURL := range r.urls {
		if savedURL.ShortURL == key {
			for _, savedUser := range r.users {
				if savedUser.URLID == savedURL.CorrelationID && savedUser.UserID != userID {
					logger.Log.Info(ErrNotOwner.Error(), zap.String("user", userID), zap.String("ID", key))
					return ErrNotOwner
				} else {
					r.urls[i].IsDeleted = true
					urlNotFound = false
					break UserCheck
				}
			}
		}
	}

	if urlNotFound {
		logger.Log.Info(ErrNotFoundURL.Error(), zap.String("key", key))
		return ErrNotFoundURL
	}

	// Setting the flag is_deleted to true in r.file by rewriting it
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.Create(r.file)
	if err != nil {
		logger.Log.Info("Couldn't truncate the shortened urls file", zap.Error(err))
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, savedURL := range r.urls {
		jsonURL, err := json.Marshal(savedURL)
		if err != nil {
			logger.Log.Info("Error serializing URL in JSON", zap.Error(err))
			return err
		}

		writer.Write(jsonURL)
		writer.WriteString("\n")
	}
	writer.Flush()

	return nil
}

// Check if there is a URL stored in the Repo with the specified ID
func (r *InMemoryRepo) GetURL(id string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, url := range r.urls {
		if url.ShortURL == id && url.OriginalURL != "" {
			if url.IsDeleted {
				return "", ErrURLGone
			}

			return url.OriginalURL, nil
		}
	}

	return "", ErrNotFoundURL
}

// Returns all URLs connected to the user specified. Output is urls[key] = originalURL
func (r *InMemoryRepo) GetAllUserURLs(userID string) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	urls := make(map[string]string, 0)
	for _, user := range r.users {
		if user.UserID == userID {
			for _, url := range r.urls {
				if url.CorrelationID == user.URLID && !url.IsDeleted {
					urls[url.ShortURL] = url.OriginalURL
				}
			}
		}
	}

	return urls, nil
}

// Checks if files storing urls and users are present in the filesystem
func (r *InMemoryRepo) Ping(_ context.Context) error {
	if _, err := os.Stat(r.file); err != nil {
		return err
	} else if _, err = os.Stat(r.fileUsers); err != nil {
		return err
	} else {
		return nil
	}
}

// Load previously shortened URLs from the file specified
func (r *InMemoryRepo) loadStoredURLs(fileURLs string) error {
	r.file = fileURLs

	file, err := os.Open(fileURLs)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		url := &urlInFile{}
		if err := json.Unmarshal(line, url); err != nil {
			return err
		}

		r.urls = append(r.urls, *url)
	}

	return nil
}

// Load all URL to user relations stored in the file specified
func (r *InMemoryRepo) loadStoredUsers(fileUsers string) error {
	r.fileUsers = fileUsers

	file, err := os.Open(fileUsers)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		user := &userInFile{}
		if err := json.Unmarshal(line, user); err != nil {
			return err
		}

		r.users = append(r.users, *user)
	}

	return nil
}
