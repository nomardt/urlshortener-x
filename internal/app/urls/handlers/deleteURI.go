package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/auth"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	"go.uber.org/zap"
)

// Generates input channel from slice of keys
func generator(ctx context.Context, keys []string, _ context.CancelFunc) chan string {
	inputCh := make(chan string)

	go func() {
		defer close(inputCh)

		for _, key := range keys {
			select {
			case <-ctx.Done():
				return
			case inputCh <- key:
			}
		}
	}()

	return inputCh
}

// Starts multiple background workers to delete the requested URLs
func deleteFanOut(ctx context.Context, keysCh chan string, userID string, r urls.Repository) {
	numWorkers := 5
	var errorCounter int64
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					// Timeout
					return
				case key, ok := <-keysCh:
					if !ok {
						return
					}

					// Full error messages are logged at the infrastructure layer
					err := r.DeleteURL(key, userID)
					if err != nil {
						atomic.AddInt64(&errorCounter, 1)
					}
				}
			}
		}()
	}

	wg.Wait()
	logger.Log.Info("Deletion is finished", zap.Int64("Number of errors", errorCounter))
}

// Removes shortened URLs with the provided keys if the user is the owner of these URLs
func (h *Handler) JSONDeleteBatch(w http.ResponseWriter, r *http.Request) {
	var keys []string
	if err := json.NewDecoder(r.Body).Decode(&keys); err != nil {
		http.Error(w, "You provided invalid JSON! Please use the following format: [\"6qxTVvsy\", \"RTfd56hn\", \"Jlfd67ds\"]", http.StatusBadRequest)
		return
	}

	jwtCookie := r.Header.Get("Authorization")
	jwtCookie, found := strings.CutPrefix(jwtCookie, "Bearer ")
	if !found {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	userID, err := auth.GetUserID(jwtCookie, h.Secret)
	if err != nil {
		logger.Log.Info("Couldn't decrypt the cookie", zap.String("Authorization", jwtCookie), zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	// Try to delete the requested shortened URLs in the background
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	keysCh := generator(ctx, keys, cancel)
	go deleteFanOut(ctx, keysCh, userID, h.Repository)

	w.WriteHeader(http.StatusAccepted)
}
