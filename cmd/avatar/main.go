package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-mods/avatar"
	"github.com/go-mods/initials"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Http server for avatar generation
func main() {
	// Create a new router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.Default())

	// Attach the routes
	router.GET("/api", avatarHandler)

	// server configuration
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// print the server address
	log.Println("Server started at http://localhost:8080")

	// Start the server
	log.Fatal(server.ListenAndServe())
}

func avatarHandler(ctx *gin.Context) {
	// Initials initialsOptions
	name := ctx.Query("name")
	separator := ctx.Query("separator")
	sensitive := ctx.Query("sensitive")
	lowercase := ctx.Query("lowercase")
	uppercase := ctx.Query("uppercase")
	camelcase := ctx.Query("camelcase")
	length := ctx.Query("length")
	wordLength := ctx.Query("wordLength")
	// SVG initialsOptions
	width := ctx.Query("width")
	height := ctx.Query("height")
	backgroundColor := ctx.Query("backgroundColor")
	fontFamily := ctx.Query("fontFamily")
	fontWeight := ctx.Query("fontWeight")
	fontSize := ctx.Query("fontSize")
	fontColor := ctx.Query("fontColor")
	shape := ctx.Query("shape")
	borderDash := ctx.Query("borderDash")
	borderWidth := ctx.Query("borderWidth")
	borderRadius := ctx.Query("borderRadius")
	borderColor := ctx.Query("borderColor")
	padding := ctx.Query("padding")
	randomColor := ctx.Query("randomColor")
	randomFontColor := ctx.Query("randomFontColor")
	randomBorderColor := ctx.Query("randomBorderColor")

	// List of initial initialsOptions
	var initialsOptions []initials.Option
	if separator != "" {
		initialsOptions = append(initialsOptions, initials.WithSeparator(separator))
	}
	if sensitive != "" {
		initialsOptions = append(initialsOptions, initials.WithSensitive())
	}
	if lowercase != "" && lowercase != "false" {
		initialsOptions = append(initialsOptions, initials.WithLowercase())
	}
	if uppercase != "" && uppercase != "false" {
		initialsOptions = append(initialsOptions, initials.WithUppercase())
	}
	if camelcase != "" && camelcase != "false" {
		initialsOptions = append(initialsOptions, initials.WithCamelCase())
	}
	if length != "" {
		// convert to int
		length, _ := strconv.Atoi(length)
		if length >= 1 {
			initialsOptions = append(initialsOptions, initials.WithLength(length))
		}
	}
	if wordLength != "" && wordLength != "false" {
		initialsOptions = append(initialsOptions, initials.WithWordLength())
	}

	// List of SVG initialsOptions
	var svgOptions []avatar.GeneratorOption
	if width != "" {
		// convert to int
		width, _ := strconv.Atoi(width)
		if width >= 1 {
			svgOptions = append(svgOptions, avatar.WithWidth(width))
		}
	}
	if height != "" {
		// convert to int
		height, _ := strconv.Atoi(height)
		if height >= 1 {
			svgOptions = append(svgOptions, avatar.WithHeight(height))
		}
	}
	if backgroundColor != "" {
		svgOptions = append(svgOptions, avatar.WithBackgroundColor(backgroundColor))
	}
	if fontFamily != "" && fontWeight != "" {
		// convert to int
		fontWeight, _ := strconv.Atoi(fontWeight)
		if fontWeight >= 1 {
			svgOptions = append(svgOptions, avatar.WithFont(fontFamily, fontWeight))
		}
	}
	if fontFamily != "" && fontWeight == "" {
		svgOptions = append(svgOptions, avatar.WithFont(fontFamily, 400))
	}
	if fontSize != "" {
		// convert to int
		fontSize, _ := strconv.Atoi(fontSize)
		if fontSize >= 1 {
			svgOptions = append(svgOptions, avatar.WithFontSize(fontSize))
		}
	}
	if fontColor != "" {
		svgOptions = append(svgOptions, avatar.WithFontColor(fontColor))
	}
	if shape != "" {
		svgOptions = append(svgOptions, avatar.WithShape(shape))
	}
	if borderDash != "" {
		svgOptions = append(svgOptions, avatar.WithBorderDash(borderDash))
	}
	if borderWidth != "" {
		// convert to int
		borderWidth, _ := strconv.Atoi(borderWidth)
		if borderWidth >= 1 {
			svgOptions = append(svgOptions, avatar.WithBorderWidth(borderWidth))
		}
	}
	if borderRadius != "" {
		// convert to int
		borderRadius, _ := strconv.Atoi(borderRadius)
		if borderRadius >= 1 {
			svgOptions = append(svgOptions, avatar.WithBorderRadius(borderRadius))
		}
	}
	if borderColor != "" {
		svgOptions = append(svgOptions, avatar.WithBorderColor(borderColor))
	}
	if padding != "" {
		// convert to int
		padding, _ := strconv.Atoi(padding)
		if padding >= 1 {
			svgOptions = append(svgOptions, avatar.WithPadding(padding))
		}
	}
	if randomColor != "" && randomColor != "false" {
		svgOptions = append(svgOptions, avatar.WithRandomColor())
	}
	if randomFontColor != "" && randomFontColor != "false" {
		svgOptions = append(svgOptions, avatar.WithRandomFontColor())
	}
	if randomBorderColor != "" && randomBorderColor != "false" {
		svgOptions = append(svgOptions, avatar.WithRandomBorderColor())
	}

	// Get initial from name
	initial := initials.GetInitials(name, initialsOptions...)

	// Generate the avatar
	svg := avatar.GetSVG(initial, svgOptions...)

	// Return the avatar
	ctx.Data(http.StatusOK, "image/svg+xml", svg.Bytes())
}
