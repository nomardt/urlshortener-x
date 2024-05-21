package urls

import (
	"reflect"
	"testing"
)

func Test_NewURL(t *testing.T) {
	type args struct {
		longURL string
		id      string
	}
	tests := []struct {
		name    string
		args    args
		want    *URL
		wantErr bool
	}{
		{
			name:    "Valid URL",
			args:    args{longURL: "https://example.com/?abc", id: "aaa"},
			want:    &URL{correlationID: "aaa", id: "aaa", longURL: "https://example.com/?abc"},
			wantErr: false,
		},
		{
			name:    "Invalid schema",
			args:    args{longURL: "gopher://example.com/?abc", id: "aaa"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewURL(tt.args.longURL, tt.args.id, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
