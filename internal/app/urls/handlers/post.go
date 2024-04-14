package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/nomardt/urlshortener-x/internal/domain/urls"
)

func (h *Handler) PostURI(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || string(body) == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var u *urls.URL
	if h.defaultRoute == "" {
		u, err = urls.NewURLWithoutID(string(body))
	} else {
		u, err = urls.NewURL(string(body), h.defaultRoute)

	}
	if err != nil {
		http.Error(w, "Please enter a valid URL", http.StatusBadRequest)
		return
	}
	h.SaveURL(u)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("http://127.0.0.1/" + u.ID()))
	if err != nil {
		fmt.Println(err)
	}
}
