package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newUriHandler(t *testing.T) {
	type args struct {
		body    string
		storage UriStorage
	}
	type want struct {
		code int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "POST valid",
			args: args{
				body:    "https://example.com",
				storage: make(map[string]string),
			},
			want: want{
				code: 201,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newUri := newUriHandler(tt.args.storage)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.args.body))

			r.Header.Add("Content-Type", "text/plain")

			newUri(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

func Test_getUriHandler(t *testing.T) {
	type args struct {
		path    string
		storage UriStorage
	}
	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "GET valid URL",
			args: args{
				path: "/aaaaaaaa",
				storage: map[string]string{
					"aaaaaaaa": "https://example.com",
				},
			},
			want: want{
				code:     307,
				location: "https://example.com",
			},
		},
		{
			name: "GET invalid URL",
			args: args{
				path:    "/aaaaaaaa",
				storage: make(map[string]string),
			},
			want: want{
				code:     400,
				location: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getUri := getUriHandler(tt.args.storage, tt.args.path[1:])

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.args.path, nil)
			fmt.Println(r)
			getUri(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}
}
