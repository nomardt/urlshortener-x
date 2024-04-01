package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mainPageHandler(t *testing.T) {
	type args struct {
		path        string
		body        string
		storage     uriStorage
		contentType string
		method      string
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
			name: "POST valid URL",
			args: args{
				path:        "/",
				body:        "https://example.com",
				contentType: "text/plain",
				method:      "POST",
				storage:     make(map[string]string),
			},
			want: want{
				code:     201,
				location: "",
			},
		},
		{
			name: "GET valid URL",
			args: args{
				path:        "/aaaaaaaa",
				body:        "",
				contentType: "",
				method:      "GET",
				storage: map[string]string{
					"aaaaaaaa": "https://example.com",
				},
			},
			want: want{
				code:     307,
				location: "https://example.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainPage := mainPageHandler(tt.args.storage)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.args.method, tt.args.path, strings.NewReader(tt.args.body))
			if tt.args.contentType != "" {
				r.Header.Add("Content-Type", tt.args.contentType)
			}
			mainPage(w, r)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}
}
