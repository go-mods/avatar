package main

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-mods/avatar"
)

func get(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	newRouter(t.TempDir()).ServeHTTP(recorder, request)
	return recorder
}

func TestAvatarRoute(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{name: "initials of the name", target: "/api?name=John%20Doe", want: ">JD</text>"},
		{name: "empty name", target: "/api", want: "<circle "},
		{name: "square", target: "/api?name=John%20Doe&shape=square&width=200&height=200", want: `<rect x="0" y="0" width="200" height="200"`},
		{name: "diamond", target: "/api?shape=diamond", want: "<polygon "},
		{name: "octagon", target: "/api?shape=octagon", want: "<polygon "},
		{name: "heart", target: "/api?shape=heart", want: "<path "},
		{name: "cloud", target: "/api?shape=cloud", want: "<path "},
		{name: "width", target: "/api?width=100", want: `<svg width="100" height="48"`},
		{name: "height", target: "/api?height=100", want: `<svg width="48" height="100"`},
		{name: "font size", target: "/api?name=JD&fontSize=30", want: `font-size="30"`},
		{name: "font weight without family", target: "/api?name=JD&fontWeight=700", want: `font-weight="700">JD`},
		{name: "background color", target: "/api?backgroundColor=red", want: `fill="#ff0000"`},
		{name: "font color", target: "/api?name=JD&fontColor=red", want: `fill="#ff0000">JD`},
		{name: "border color", target: "/api?borderColor=red", want: `stroke="#ff0000"`},
		{name: "border width", target: "/api?borderWidth=3", want: `stroke-width="3"`},
		{name: "no border", target: "/api?borderWidth=0", want: `stroke-width="0"`},
		{name: "border radius", target: "/api?shape=square&borderRadius=5", want: `rx="5"`},
		{name: "border dash", target: "/api?borderDash=5,5", want: `stroke-dasharray="5,5"`},
		{name: "padding", target: "/api?padding=4", want: `r="20"`},
		{name: "separator", target: "/api?name=John%20Doe&separator=.", want: ">J.D</text>"},
		{name: "length", target: "/api?name=John%20Doe&length=3", want: ">JDO</text>"},
		{name: "lowercase", target: "/api?name=John%20Doe&lowercase=true", want: ">jd</text>"},
		{name: "lowercase disabled", target: "/api?name=John%20Doe&lowercase=false", want: ">JD</text>"},
		{name: "uppercase", target: "/api?name=john%20doe&uppercase=1", want: ">JD</text>"},
		{name: "sensitive", target: "/api?name=john%20doe&sensitive=true", want: ">jd</text>"},
		{name: "sensitive disabled", target: "/api?name=john%20doe&sensitive=false", want: ">JD</text>"},
		{name: "empty boolean", target: "/api?name=john%20doe&sensitive=", want: ">JD</text>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := get(t, tt.target)

			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, "image/svg+xml", response.Header().Get("Content-Type"))
			assert.Contains(t, response.Body.String(), tt.want)
		})
	}
}

func TestAvatarRoute_invalidRequests(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{name: "width not a number", target: "/api?width=abc", want: "invalid parameter width"},
		{name: "width zero", target: "/api?width=0", want: "invalid parameter width"},
		{name: "width negative", target: "/api?width=-5", want: "invalid parameter width"},
		{name: "width too large", target: "/api?width=4097", want: "invalid parameter width"},
		{name: "height too large", target: "/api?height=4097", want: "invalid parameter height"},
		{name: "font size too large", target: "/api?fontSize=4097", want: "invalid parameter fontSize"},
		{name: "padding not a number", target: "/api?padding=x", want: "invalid parameter padding"},
		{name: "border width not a number", target: "/api?borderWidth=x", want: "invalid parameter borderWidth"},
		{name: "border radius not a number", target: "/api?borderRadius=x", want: "invalid parameter borderRadius"},
		{name: "font weight not a number", target: "/api?fontFamily=Roboto&fontWeight=abc", want: "invalid parameter fontWeight"},
		{name: "font weight zero", target: "/api?fontFamily=Roboto&fontWeight=0", want: "invalid parameter fontWeight"},
		{name: "font weight too high", target: "/api?fontFamily=Roboto&fontWeight=2000", want: "invalid parameter fontWeight"},
		{name: "font family", target: "/api?name=JD&fontFamily=a/b", want: "invalid font family"},
		{name: "length not a number", target: "/api?length=x", want: "invalid parameter length"},
		{name: "length too large", target: "/api?length=65", want: "invalid parameter length"},
		{name: "boolean", target: "/api?lowercase=yes", want: "invalid parameter lowercase"},
		{name: "random boolean", target: "/api?randomColor=maybe", want: "invalid parameter randomColor"},
		{name: "name too long", target: "/api?name=" + strings.Repeat("a", maxNameLength+1), want: "invalid parameter name"},
		{name: "border dash", target: "/api?borderDash=x", want: "invalid border dash"},
		{name: "padding too large", target: "/api?padding=100", want: "invalid size"},
		{name: "shape", target: "/api?shape=star", want: "invalid shape"},
		{name: "background color", target: "/api?backgroundColor=notacolor", want: "invalid color"},
		{name: "font color", target: "/api?fontColor=notacolor", want: "invalid color"},
		{name: "border color", target: "/api?borderColor=notacolor", want: "invalid color"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := get(t, tt.target)

			assert.Equal(t, http.StatusBadRequest, response.Code)
			assert.Equal(t, "application/json; charset=utf-8", response.Header().Get("Content-Type"))
			assert.Contains(t, response.Body.String(), tt.want)
			assert.Empty(t, response.Header().Get("Cache-Control"))
		})
	}
}

func TestAvatarRoute_headers(t *testing.T) {
	tests := []struct {
		name  string
		query string
		cache string
	}{
		{name: "same request, same avatar", query: "name=JD", cache: "public, max-age=86400"},
		{name: "disabled random option", query: "name=JD&randomColor=false", cache: "public, max-age=86400"},
		{name: "random colors", query: "name=JD&randomColor=true", cache: "no-store"},
		{name: "random background", query: "name=JD&randomBackgroundColor=true", cache: "no-store"},
		{name: "random font", query: "name=JD&randomFontColor=true", cache: "no-store"},
		{name: "random border", query: "name=JD&randomBorderColor=true", cache: "no-store"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := get(t, "/api?"+tt.query)

			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, tt.cache, response.Header().Get("Cache-Control"))
			assert.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, svgContentSecurityPolicy, response.Header().Get("Content-Security-Policy"))
		})
	}
}

func TestAvatarRoute_randomColorsChange(t *testing.T) {
	outputs := map[string]bool{}
	for range 20 {
		outputs[get(t, "/api?name=JD&randomBackgroundColor=true").Body.String()] = true
	}
	assert.Greater(t, len(outputs), 1)
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		want   string
	}{
		{name: "unknown font", err: avatar.ErrUnknownFont, status: http.StatusBadRequest, want: "unknown font"},
		{
			name:   "internal error stays in the log",
			err:    errors.New(`open C:\secret\fonts\Inter: access denied`),
			status: http.StatusInternalServerError,
			want:   `{"error":"internal server error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			writeError(ctx, tt.err)

			assert.Equal(t, tt.status, recorder.Code)
			assert.Contains(t, recorder.Body.String(), tt.want)
			assert.NotContains(t, recorder.Body.String(), "secret")
		})
	}
}

func TestPageRoute(t *testing.T) {
	response := get(t, "/")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "text/html; charset=utf-8", response.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", response.Header().Get("Cache-Control"))
	assert.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, pageContentSecurityPolicy, response.Header().Get("Content-Security-Policy"))
	assert.Contains(t, response.Body.String(), "<title>avatar</title>")
}

func TestAssetRoute(t *testing.T) {
	page := get(t, "/").Body.String()
	script := regexp.MustCompile(`src="(/assets/[^"]+\.js)"`).FindStringSubmatch(page)
	require.NotNil(t, script, "page: %s", page)
	style := regexp.MustCompile(`href="(/assets/[^"]+\.css)"`).FindStringSubmatch(page)
	require.NotNil(t, style, "page: %s", page)

	for _, target := range []struct{ path, contentType string }{
		{script[1], "text/javascript"},
		{style[1], "text/css"},
	} {
		response := get(t, target.path)

		assert.Equal(t, http.StatusOK, response.Code, target.path)
		assert.Contains(t, response.Header().Get("Content-Type"), target.contentType)
		assert.Equal(t, "public, max-age=31536000, immutable", response.Header().Get("Cache-Control"))
		assert.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
		assert.NotEmpty(t, response.Body.String())
	}
}

func TestAssetRoute_missingFile(t *testing.T) {
	assert.Equal(t, http.StatusNotFound, get(t, "/assets/missing.js").Code)
	assert.Equal(t, http.StatusNotFound, get(t, "/assets/../web.go").Code)
	assert.Equal(t, http.StatusNotFound, get(t, "/missing").Code)
}

func TestHealthRoute(t *testing.T) {
	response := get(t, "/health")

	assert.Equal(t, http.StatusOK, response.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
}

func TestBrowserURL(t *testing.T) {
	tests := []struct {
		addr net.Addr
		want string
	}{
		{addr: &net.TCPAddr{IP: net.IPv6unspecified, Port: 8080}, want: "http://localhost:8080"},
		{addr: &net.TCPAddr{IP: net.IPv4zero, Port: 9000}, want: "http://localhost:9000"},
		{addr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9000}, want: "http://127.0.0.1:9000"},
		{addr: &net.TCPAddr{IP: net.IPv4(192, 168, 1, 10), Port: 80}, want: "http://192.168.1.10:80"},
		{addr: &net.TCPAddr{IP: net.IPv6loopback, Port: 9000}, want: "http://[::1]:9000"},
	}
	for _, tt := range tests {
		t.Run(tt.addr.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, browserURL(tt.addr))
		})
	}
}

func TestHealthURL(t *testing.T) {
	tests := []struct {
		addr    string
		want    string
		wantErr bool
	}{
		{addr: ":8080", want: "http://127.0.0.1:8080/health"},
		{addr: "localhost:9000", want: "http://localhost:9000/health"},
		{addr: "[::1]:9000", want: "http://[::1]:9000/health"},
		{addr: "8080", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			url, err := healthURL(tt.addr)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, url)
		})
	}
}

func TestServe_stopsAndKeepsAnsweringUntilThen(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopped := make(chan error, 1)

	go func() {
		stopped <- serve(ctx, newServer(addr, t.TempDir()), listener)
	}()

	require.NoError(t, healthcheck(context.Background(), addr))

	cancel()
	require.NoError(t, <-stopped)
	require.Error(t, healthcheck(context.Background(), addr), "the server does not answer anymore")
}

func TestServe_failsOnAClosedListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())

	require.Error(t, serve(context.Background(), newServer(addr, t.TempDir()), listener))
}

func TestHealthcheck_failsOnAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	require.ErrorContains(t, healthcheck(context.Background(), server.Listener.Addr().String()), "503")
}
