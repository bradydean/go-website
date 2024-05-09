package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/bradydean/go-website/components"
)

//go:generate go run github.com/a-h/templ/cmd/templ generate

func main() {
	e := echo.New()
	e.Group("/static").Use(middleware.Static("static"))

	e.GET("/", func(c echo.Context) error {
		return components.Index("World").Render(c.Request().Context(), c.Response().Writer)
	})

	e.POST("/clicked", func(c echo.Context) error {
		return components.Clicked().Render(c.Request().Context(), c.Response().Writer)	
	})

	if err := e.Start(":8000"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
