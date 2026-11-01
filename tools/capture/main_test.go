package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoOptions(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{query: "name=Grace+Hopper", want: ""},
		{query: "name=A&shape=square&borderRadius=12", want: `avatar.WithShape("square"), avatar.WithBorderRadius(12)`},
		{query: "name=A&backgroundColor=%23fde68a", want: `avatar.WithBackgroundColor("#fde68a")`},
		{query: "name=A&borderDash=4+2", want: `avatar.WithBorderDash("4 2")`},
		{query: "name=A&fontFamily=Pacifico", want: `avatar.WithFont("Pacifico", 400)`},
		{query: "name=A&fontFamily=Inter&fontWeight=900", want: `avatar.WithFont("Inter", 900)`},
		{query: "name=A&length=3&separator=.", want: `initials.WithSeparator("."), initials.WithLength(3)`},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			code, err := goOptions(tt.query)
			require.NoError(t, err)
			assert.Equal(t, tt.want, code)
		})
	}
}

func TestGoOptions_unknownParameter(t *testing.T) {
	_, err := goOptions("name=A&unknown=1")
	require.ErrorContains(t, err, "unknown")
}

func TestGet_failsOnAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "invalid color", http.StatusBadRequest)
	}))
	defer server.Close()

	_, err := get(server.URL)

	require.ErrorContains(t, err, "400 Bad Request")
	require.ErrorContains(t, err, "invalid color")
}

func TestWriteTable(t *testing.T) {
	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile(readmeFile, []byte("before\n"+startMarker+"\nold\n"+endMarker+"\nafter\n"), 0o600))

	require.NoError(t, writeTable([]string{"| a | b | c | d |"}))

	content, err := os.ReadFile(readmeFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "| a | b | c | d |")
	assert.NotContains(t, string(content), "old")
	assert.Contains(t, string(content), "before\n"+startMarker)
	assert.Contains(t, string(content), endMarker+"\nafter")
}

func TestWriteTable_withoutMarkers(t *testing.T) {
	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile(readmeFile, []byte("no markers\n"), 0o600))

	require.ErrorContains(t, writeTable(nil), "markers not found")
}
