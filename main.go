package main

import (
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/bradydean/go-website/internal/pkg/components"
)

//go:generate go run github.com/a-h/templ/cmd/templ generate
//go:generate npx --yes tailwindcss@latest -i ./global.css -o ./static/tailwind.css --minify

func Component(c echo.Context, code int, component templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := component.Render(c.Request().Context(), buf); err != nil {
		return err
	}

	return c.HTML(code, buf.String())
}

func main() {
	e := echo.New()

	e.Group("/static").Use(middleware.Static("static"))

	e.GET("/", func(c echo.Context) error {
		search := c.QueryParam("search")
		results := strings.Split(search, "")
		layout := components.Layout("go-website", components.Search(results))
		return Component(c, http.StatusOK, layout)
	})

	if err := e.Start(":8000"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
