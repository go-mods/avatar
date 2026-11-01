package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvatarRoute(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		status      int
		contentType string
		want        string
	}{
		{
			name:        "initials of the name",
			target:      "/api?name=John%20Doe",
			status:      http.StatusOK,
			contentType: "image/svg+xml",
			want:        ">JD</text>",
		},
		{
			name:        "square",
			target:      "/api?name=John%20Doe&shape=square&width=200&height=200",
			status:      http.StatusOK,
			contentType: "image/svg+xml",
			want:        `<rect x="0" y="0" width="200" height="200"`,
		},
		{
			name:        "invalid border dash",
			target:      "/api?name=John%20Doe&borderDash=x",
			status:      http.StatusBadRequest,
			contentType: "application/json; charset=utf-8",
			want:        "invalid border dash",
		},
		{
			name:        "padding too large",
			target:      "/api?name=John%20Doe&padding=100",
			status:      http.StatusBadRequest,
			contentType: "application/json; charset=utf-8",
			want:        "invalid size",
		},
		{
			name:        "invalid color",
			target:      "/api?name=John%20Doe&backgroundColor=notacolor",
			status:      http.StatusBadRequest,
			contentType: "application/json; charset=utf-8",
			want:        "invalid color",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.target, nil)

			newRouter().ServeHTTP(recorder, request)

			assert.Equal(t, tt.status, recorder.Code)
			assert.Equal(t, tt.contentType, recorder.Header().Get("Content-Type"))
			assert.Contains(t, recorder.Body.String(), tt.want)
		})
	}
}
