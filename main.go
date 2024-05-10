package main

import (
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/bradydean/go-website/components"
)

//go:generate go run github.com/a-h/templ/cmd/templ generate

type RenderOption func(*RenderOptions)

type RenderOptions struct {
	Status      int
	ContentType string
}

func WithStatus(status int) RenderOption {
	return func(opt *RenderOptions) {
		opt.Status = status
	}
}

func WithContentType(contentType string) RenderOption {
	return func(opt *RenderOptions) {
		opt.ContentType = contentType
	}
}

func Render(c echo.Context, component templ.Component, options ...RenderOption) error {
	opts := RenderOptions{Status: http.StatusOK, ContentType: echo.MIMETextHTML}

	for _, opt := range options {
		opt(&opts)
	}

	c.Response().Header().Set(echo.HeaderContentType, opts.ContentType)
	c.Response().Status = opts.Status

	return component.Render(c.Request().Context(), c.Response())
}

func main() {
	e := echo.New()

	e.Group("/static").Use(middleware.Static("static"))

	e.GET("/", func(c echo.Context) error {
		return Render(c, components.Index())
	})

	e.GET("/search", func(c echo.Context) error {
		q := c.QueryParam("q")
		return Render(c, components.SearchResults(strings.Split(q, "")))
	})

	if err := e.Start(":8000"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
