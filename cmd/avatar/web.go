package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// pageContentSecurityPolicy allows the files of the server and the avatars displayed from a data url
const pageContentSecurityPolicy = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data:; " +
	"connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

// webFiles is the page of the server, built from the web folder by task web:build
//
//go:embed all:web/dist
var webFiles embed.FS

// addPageRoutes attaches the presentation page and its files
func addPageRoutes(router *gin.Engine) {
	page, err := fs.Sub(webFiles, "web/dist")
	if err != nil {
		panic(err)
	}
	assets, err := fs.Sub(page, "assets")
	if err != nil {
		panic(err)
	}
	index, err := fs.ReadFile(page, "index.html")
	if err != nil {
		panic(err)
	}

	router.GET("/", func(ctx *gin.Context) {
		// The page refers to files whose name changes with their content: it is always checked again
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("Content-Security-Policy", pageContentSecurityPolicy)
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})

	files := router.Group("/assets", func(ctx *gin.Context) {
		ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
		ctx.Header("X-Content-Type-Options", "nosniff")
	})
	files.StaticFS("/", http.FS(assets))
}
