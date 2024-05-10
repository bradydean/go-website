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

func Component(c echo.Context, code int, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	c.Response().Status = code
	return component.Render(c.Request().Context(), c.Response())
}

func main() {
	e := echo.New()

	e.Group("/static").Use(middleware.Static("static"))

	e.GET("/", func(c echo.Context) error {
		return Component(c, http.StatusOK, components.Index())
	})

	e.GET("/search", func(c echo.Context) error {
		q := c.QueryParam("q")
		return Component(c, http.StatusOK, components.SearchResults(strings.Split(q, "")))
	})

	if err := e.Start(":8000"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
