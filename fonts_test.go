package avatar

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testFont = "font bytes"

// fontServer stands for the Google Fonts servers during a test
type fontServer struct {
	*httptest.Server
	styleRequests atomic.Int32
	style         func(fontURL string) string
	font          http.HandlerFunc
	lastQuery     atomic.Value
}

func newFontServer(t *testing.T) *fontServer {
	t.Helper()
	server := &fontServer{
		style: func(fontURL string) string {
			return "@font-face {\n  font-family: 'Inter';\n  src: url(" + fontURL + ") format('truetype');\n}\n"
		},
		font: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(testFont))
		},
	}
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/css" {
			server.styleRequests.Add(1)
			server.lastQuery.Store(r.URL.RawQuery)
			_, _ = w.Write([]byte(server.style(server.URL + "/files/font.ttf")))
			return
		}
		server.font(w, r)
	}))
	t.Cleanup(server.Close)

	cssURL, filePrefix := fontCSSURL, fontFilePrefix
	fontCSSURL, fontFilePrefix = server.URL+"/css", server.URL+"/files/"
	t.Cleanup(func() {
		fontCSSURL, fontFilePrefix = cssURL, filePrefix
	})
	return server
}

func TestGetSVG_embedsTheFont(t *testing.T) {
	server := newFontServer(t)

	svg, err := GetSVG("AB", WithFont("Roboto Mono", 700), WithFontDir(t.TempDir()))

	require.NoError(t, err)
	requireWellFormed(t, svg.String())
	assert.Contains(t, svg.String(), "url(data:font/ttf;base64,"+base64.StdEncoding.EncodeToString([]byte(testFont))+")")
	assert.NotContains(t, svg.String(), server.URL)
	assert.Equal(t, "family=Roboto+Mono:700&text=AB", server.lastQuery.Load())
}

func TestGetSVG_escapesTheInitialsOfTheFontURL(t *testing.T) {
	server := newFontServer(t)

	_, err := GetSVG(`é&x=<`, WithFont("Inter", 400), WithFontDir(t.TempDir()))

	require.NoError(t, err)
	assert.Equal(t, "family=Inter:400&text=%C3%A9%26x%3D%3C", server.lastQuery.Load())
}

func TestGetSVG_keepsTheFont(t *testing.T) {
	server := newFontServer(t)
	dir := t.TempDir()

	first, err := GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))
	require.NoError(t, err)
	second, err := GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))
	require.NoError(t, err)
	assert.Equal(t, first.String(), second.String())
	assert.EqualValues(t, 1, server.styleRequests.Load())

	_, err = GetSVG("CD", WithFont("Inter", 400), WithFontDir(dir))
	require.NoError(t, err)
	assert.EqualValues(t, 2, server.styleRequests.Load(), "other initials need other letters")
}

func TestGetSVG_doesNotPutTheInitialsInAFileName(t *testing.T) {
	newFontServer(t)
	dir := filepath.Join(t.TempDir(), "fonts")

	_, err := GetSVG(`../../x`, WithFont("Inter", 400), WithFontDir(dir))

	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dir, fontCacheFile("Inter", 400, `../../x`)))
	entries, err := os.ReadDir(filepath.Dir(dir))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestGetSVG_withoutAFontToDownload(t *testing.T) {
	tests := []struct {
		name     string
		initials string
		options  []GeneratorOption
	}{
		{name: "no text to draw", options: []GeneratorOption{WithFont("Inter", 400)}},
		{name: "empty family", initials: "AB", options: []GeneratorOption{WithFont("", 400)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newFontServer(t)
			dir := t.TempDir()

			svg, err := GetSVG(tt.initials, append(tt.options, WithFontDir(dir))...)

			require.NoError(t, err)
			assert.NotContains(t, svg.String(), "@font-face")
			assert.Zero(t, server.styleRequests.Load())
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}

func TestGetSVG_fontWeightMatchesTheFont(t *testing.T) {
	newFontServer(t)

	svg, err := GetSVG("AB", WithFont("Inter", 900), WithFontDir(t.TempDir()))

	require.NoError(t, err)
	assert.Contains(t, svg.String(), `font-weight="900">AB`)
}

func TestGetSVG_fontWithoutTheLetters(t *testing.T) {
	server := newFontServer(t)
	server.font = func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}
	dir := t.TempDir()

	for range 2 {
		svg, err := GetSVG("你好", WithFont("Inter", 400), WithFontDir(dir))

		require.NoError(t, err)
		requireWellFormed(t, svg.String())
		assert.NotContains(t, svg.String(), "@font-face")
		assert.Contains(t, svg.String(), `font-family="'Inter',sans-serif"`)
		assert.Contains(t, svg.String(), ">你好</text>")
	}
	assert.EqualValues(t, 1, server.styleRequests.Load(), "the answer is kept")
}

func TestGetSVG_rejectsAnUnknownFont(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()
	cssURL := fontCSSURL
	fontCSSURL = server.URL + "/css"
	t.Cleanup(func() { fontCSSURL = cssURL })
	dir := t.TempDir()

	svg, err := GetSVG("AB", WithFont("NoSuchFamily", 400), WithFontDir(dir))

	require.ErrorIs(t, err, ErrUnknownFont)
	assert.Nil(t, svg)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestGetSVG_rejectsAnUnexpectedFont(t *testing.T) {
	tests := []struct {
		name  string
		style func(fontURL string) string
		want  string
	}{
		{
			name:  "no font",
			style: func(string) string { return "@font-face {}" },
			want:  "0 fonts found",
		},
		{
			name: "too many fonts",
			style: func(fontURL string) string {
				return "@font-face { src: " + strings.Repeat("url("+fontURL+") ", maxFontFaces+1) + "}"
			},
			want: "fonts found",
		},
		{
			name:  "font from another server",
			style: func(string) string { return "@font-face { src: url(https://example.com/font.ttf); }" },
			want:  "unexpected font location",
		},
		{
			name: "markup",
			style: func(fontURL string) string {
				return "@font-face { src: url(" + fontURL + "); }</style><script>alert(1)</script>"
			},
			want: "markup in the font style",
		},
		{
			name:  "not a style",
			style: func(fontURL string) string { return "<html>url(" + fontURL + ")</html>" },
			want:  "not a font style",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newFontServer(t)
			server.style = tt.style
			dir := t.TempDir()

			svg, err := GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))

			require.ErrorContains(t, err, tt.want)
			assert.Nil(t, svg)
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}

func TestGetSVG_replacesAnInvalidCachedFile(t *testing.T) {
	server := newFontServer(t)
	dir := t.TempDir()
	path := filepath.Join(dir, fontCacheFile("Inter", 400, "AB"))
	require.NoError(t, os.WriteFile(path, []byte("</style><script>alert(1)</script>"), 0o600))

	svg, err := GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))

	require.NoError(t, err)
	assert.NotContains(t, svg.String(), "<script>")
	assert.EqualValues(t, 1, server.styleRequests.Load())
}

func TestGetSVG_failedDownloads(t *testing.T) {
	tests := []struct {
		name string
		font http.HandlerFunc
	}{
		{
			name: "error page",
			font: func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "not found", http.StatusNotFound) },
		},
		{
			name: "font too large",
			font: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write(make([]byte, maxFontResponseSize+1))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newFontServer(t)
			server.font = tt.font
			dir := t.TempDir()

			svg, err := GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))

			require.Error(t, err)
			assert.Nil(t, svg)
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			assert.Empty(t, entries, "neither the font nor a temporary file is kept")
		})
	}
}

func TestGetSVG_fontServersUnavailable(t *testing.T) {
	tests := []struct {
		name string
		fail func(server *fontServer)
	}{
		{
			name: "unreachable servers",
			fail: func(server *fontServer) { server.Close() },
		},
		{
			name: "server error",
			fail: func(server *fontServer) {
				server.font = func(w http.ResponseWriter, _ *http.Request) {
					http.Error(w, "unavailable", http.StatusServiceUnavailable)
				}
			},
		},
		{
			name: "truncated font",
			fail: func(server *fontServer) {
				server.font = func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Length", "1000")
					_, _ = w.Write([]byte("partial"))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newFontServer(t)
			tt.fail(server)
			dir := t.TempDir()

			svg, err := GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))

			require.NoError(t, err)
			requireWellFormed(t, svg.String())
			assert.NotContains(t, svg.String(), "@font-face")
			assert.Contains(t, svg.String(), `font-family="'Inter',sans-serif"`)
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			assert.Empty(t, entries, "the font is downloaded at the next call")
		})
	}
}

func TestGetSVG_fontComesBackAfterAnOutage(t *testing.T) {
	server := newFontServer(t)
	available := server.font
	server.font = func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}
	dir := t.TempDir()

	svg, err := GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))
	require.NoError(t, err)
	assert.NotContains(t, svg.String(), "@font-face")

	server.font = available
	svg, err = GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))
	require.NoError(t, err)
	assert.Contains(t, svg.String(), "@font-face")
}

func TestGetSVG_fontAppearsOnceComplete(t *testing.T) {
	server := newFontServer(t)
	started, release := make(chan struct{}), make(chan struct{})
	server.font = func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		_, _ = w.Write([]byte(testFont))
	}
	dir := t.TempDir()
	done := make(chan error)

	go func() {
		_, err := GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))
		done <- err
	}()

	<-started
	assert.NoFileExists(t, filepath.Join(dir, fontCacheFile("Inter", 400, "AB")))
	close(release)
	require.NoError(t, <-done)
	assert.FileExists(t, filepath.Join(dir, fontCacheFile("Inter", 400, "AB")))
}

func TestGetSVG_concurrentCalls(t *testing.T) {
	newFontServer(t)
	dir := t.TempDir()
	const calls = 20
	outputs := make([]string, calls)
	errs := make([]error, calls)

	var wg sync.WaitGroup
	for i := range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svg, err := GetSVG("AB", WithFont("Inter", 400), WithFontDir(dir))
			errs[i] = err
			if err == nil {
				outputs[i] = svg.String()
			}
		}()
	}
	wg.Wait()

	for i := range calls {
		require.NoError(t, errs[i])
		assert.Equal(t, outputs[0], outputs[i])
		assert.Contains(t, outputs[i], "@font-face")
	}
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "no temporary file is left")
}

func TestGetSVG_cacheLimit(t *testing.T) {
	server := newFontServer(t)
	limit := maxCachedFonts
	maxCachedFonts = 2
	t.Cleanup(func() { maxCachedFonts = limit })
	dir := t.TempDir()

	for _, initials := range []string{"AB", "CD", "EF"} {
		svg, err := GetSVG(initials, WithFont("Inter", 400), WithFontDir(dir))
		require.NoError(t, err)
		assert.Contains(t, svg.String(), "@font-face", "the font is embedded even when it is not kept")
	}

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	_, err = GetSVG("EF", WithFont("Inter", 400), WithFontDir(dir))
	require.NoError(t, err)
	assert.EqualValues(t, 4, server.styleRequests.Load())
}

func TestGetSVGContext_canceled(t *testing.T) {
	server := newFontServer(t)
	started, release := make(chan struct{}), make(chan struct{})
	server.font = func(_ http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
	}
	defer close(release)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)

	go func() {
		_, err := GetSVGContext(ctx, "AB", WithFont("Inter", 400), WithFontDir(t.TempDir()))
		done <- err
	}()
	<-started
	cancel()

	require.ErrorIs(t, <-done, context.Canceled)
}
