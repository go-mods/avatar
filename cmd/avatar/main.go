// Command avatar is an http server which generates svg avatars.
//
// GET / presents the project and tries the avatars live, GET /api returns the avatar described by its
// query parameters, GET /health tells that the server is running. Without argument, the command starts the server; "avatar healthcheck" asks a running
// server for its health, and exits with an error if it does not answer.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-mods/initials"

	"github.com/go-mods/avatar"
)

const (
	// maxNameLength limits the name of a request, in characters
	maxNameLength = 256

	// maxInitialsLength limits the initials of a request, in characters
	maxInitialsLength = 64

	// maxImageSize limits every dimension of a request
	maxImageSize = 4096

	// downloadTimeout is longer than the two font downloads an avatar can need
	downloadTimeout = 65 * time.Second

	// shutdownTimeout limits the time given to the running requests when the server stops
	shutdownTimeout = 10 * time.Second

	// healthcheckTimeout limits the time given to a server to tell its health
	healthcheckTimeout = 3 * time.Second

	// svgContentSecurityPolicy allows what an avatar holds and nothing else: its style and its font
	svgContentSecurityPolicy = "default-src 'none'; style-src 'unsafe-inline'; font-src data:"
)

func main() {
	fontDir := flag.String("fontdir", "", "Directory to store font files")
	addr := flag.String("addr", ":8080", "Address to listen on")
	flag.Parse()

	if flag.Arg(0) == "healthcheck" {
		if err := healthcheck(context.Background(), *addr); err != nil {
			log.Fatal(err)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, *addr, *fontDir); err != nil {
		log.Fatal(err)
	}
}

// run starts the server and returns once the context is done and the server stopped
func run(ctx context.Context, addr, fontDir string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("Server started on %s", browserURL(listener.Addr()))
	return serve(ctx, newServer(addr, fontDir), listener)
}

// browserURL returns the url to open in a browser: a terminal makes it clickable. An address which
// listens on every interface, like [::]:8080, cannot be opened as is.
func browserURL(addr net.Addr) string {
	host, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		host = "localhost"
	}
	return "http://" + net.JoinHostPort(host, port)
}

// serve answers the requests until the context is done, then lets the running requests finish
func serve(ctx context.Context, server *http.Server, listener net.Listener) error {
	failed := make(chan error, 1)
	go func() {
		failed <- server.Serve(listener)
	}()

	select {
	case err := <-failed:
		return err
	case <-ctx.Done():
	}

	log.Println("Server stopping")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

// newServer returns the server, with timeouts for the slow clients
func newServer(addr, fontDir string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           newRouter(fontDir),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      downloadTimeout,
		IdleTimeout:       60 * time.Second,
	}
}

// newRouter returns the router of the server, with its routes attached
func newRouter(fontDir string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.Default())

	addPageRoutes(router)
	router.GET("/api", avatarHandler(fontDir))
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

// healthcheck asks the server listening on an address for its health
func healthcheck(ctx context.Context, addr string) error {
	url, err := healthURL(addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, healthcheckTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("health of %s: %s", url, response.Status)
	}
	return nil
}

// healthURL returns the url of the health of a server: an address without host is the local machine
func healthURL(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	if host == "" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port) + "/health", nil
}

// params reads the query parameters of a request and keeps the first invalid one
type params struct {
	ctx *gin.Context
	err error
}

// fail keeps the first invalid parameter
func (p *params) fail(name, value string) {
	if p.err == nil {
		p.err = fmt.Errorf("invalid parameter %s: %q", name, value)
	}
}

// boolean returns a boolean parameter, false when it is not given
func (p *params) boolean(name string) bool {
	value := p.ctx.Query(name)
	if value == "" {
		return false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		p.fail(name, value)
	}
	return parsed
}

// integer returns a number between two values, and whether it is given
func (p *params) integer(name string, low, high int) (int, bool) {
	value := p.ctx.Query(name)
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < low || parsed > high {
		p.fail(name, value)
		return 0, false
	}
	return parsed, true
}

// initialsOptions returns the options of the initials
func (p *params) initialsOptions() []initials.Option {
	var options []initials.Option
	if separator := p.ctx.Query("separator"); separator != "" {
		options = append(options, initials.WithSeparator(separator))
	}
	for _, flagOption := range []struct {
		name   string
		option func() initials.Option
	}{
		{"sensitive", initials.WithSensitive},
		{"lowercase", initials.WithLowercase},
		{"uppercase", initials.WithUppercase},
		{"camelcase", initials.WithCamelCase},
		{"wordLength", initials.WithWordLength},
	} {
		if p.boolean(flagOption.name) {
			options = append(options, flagOption.option())
		}
	}
	if length, ok := p.integer("length", 1, maxInitialsLength); ok {
		options = append(options, initials.WithLength(length))
	}
	return options
}

// svgOptions returns the options of the svg, and whether they ask for random colors
func (p *params) svgOptions() ([]avatar.GeneratorOption, bool) {
	var options []avatar.GeneratorOption

	for _, size := range []struct {
		name   string
		low    int
		option func(int) avatar.GeneratorOption
	}{
		{"width", 1, avatar.WithWidth},
		{"height", 1, avatar.WithHeight},
		{"fontSize", 1, avatar.WithFontSize},
		{"borderWidth", 0, avatar.WithBorderWidth},
		{"borderRadius", 0, avatar.WithBorderRadius},
		{"padding", 0, avatar.WithPadding},
	} {
		if value, ok := p.integer(size.name, size.low, maxImageSize); ok {
			options = append(options, size.option(value))
		}
	}

	for _, text := range []struct {
		name   string
		option func(string) avatar.GeneratorOption
	}{
		{"backgroundColor", avatar.WithBackgroundColor},
		{"fontColor", avatar.WithFontColor},
		{"shape", avatar.WithShape},
		{"borderDash", avatar.WithBorderDash},
		{"borderColor", avatar.WithBorderColor},
	} {
		if value := p.ctx.Query(text.name); value != "" {
			options = append(options, text.option(value))
		}
	}

	weight, ok := p.integer("fontWeight", 1, 1000)
	if !ok {
		weight = 400
	}
	if family := p.ctx.Query("fontFamily"); family != "" || ok {
		options = append(options, avatar.WithFont(family, weight))
	}

	random := false
	for _, randomOption := range []struct {
		name   string
		option func() avatar.GeneratorOption
	}{
		{"randomColor", avatar.WithRandomColor},
		{"randomBackgroundColor", avatar.WithRandomBackgroundColor},
		{"randomFontColor", avatar.WithRandomFontColor},
		{"randomBorderColor", avatar.WithRandomBorderColor},
	} {
		if p.boolean(randomOption.name) {
			options = append(options, randomOption.option())
			random = true
		}
	}
	return options, random
}

// avatarHandler returns the avatar described by the query parameters
func avatarHandler(fontDir string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := params{ctx: ctx}

		name := ctx.Query("name")
		if utf8.RuneCountInString(name) > maxNameLength {
			p.fail("name", "too long")
		}
		initialsOptions := p.initialsOptions()
		svgOptions, random := p.svgOptions()
		if fontDir != "" {
			svgOptions = append(svgOptions, avatar.WithFontDir(fontDir))
		}
		if p.err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": p.err.Error()})
			return
		}

		svg, err := avatar.GetSVGContext(ctx.Request.Context(), initials.GetInitials(name, initialsOptions...), svgOptions...)
		if err != nil {
			writeError(ctx, err)
			return
		}

		// Without random colors, the same request always gives the same avatar
		if random {
			ctx.Header("Cache-Control", "no-store")
		} else {
			ctx.Header("Cache-Control", "public, max-age=86400")
		}
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("Content-Security-Policy", svgContentSecurityPolicy)
		ctx.Data(http.StatusOK, "image/svg+xml", svg.Bytes())
	}
}

// writeError answers with an error: an invalid option is explained to the client, anything else
// stays in the log because it can hold a path of the server.
func writeError(ctx *gin.Context, err error) {
	if isRequestError(err) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// The client is gone: there is nobody left to answer to
	if errors.Is(err, context.Canceled) {
		ctx.Status(http.StatusRequestTimeout)
		return
	}
	log.Printf("Avatar generation failed: %v", err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

// isRequestError tells whether an error comes from a parameter of the request
func isRequestError(err error) bool {
	for _, target := range []error{
		avatar.ErrInvalidFontFamily,
		avatar.ErrInvalidFontWeight,
		avatar.ErrUnknownFont,
		avatar.ErrInvalidBorderDash,
		avatar.ErrInvalidSize,
		avatar.ErrInvalidColor,
		avatar.ErrInvalidShape,
	} {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
