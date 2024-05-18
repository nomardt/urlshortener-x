package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	"go.uber.org/zap"
)

func OnlyJSON(h http.HandlerFunc) http.HandlerFunc {
	jsonFn := func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/x-gzip") {
			rg, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				logger.Log.Info("Couldn't decompress request body", zap.String("error", err.Error()))
				return
			}
			defer rg.Close()
			r.Body = rg
		} else if !strings.Contains(contentType, "application/json") {
			http.Error(w, "Please use only \"Content-Type: application/json\" for this endpoint!", http.StatusUnsupportedMediaType)
			return
		}

		h.ServeHTTP(w, r)
	}

	return jsonFn
}
