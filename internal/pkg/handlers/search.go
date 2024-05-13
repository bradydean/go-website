package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/components"
)

type indexHandler struct{}

func NewIndexHandler() indexHandler {
	return indexHandler{}
}

func (h indexHandler) Handler(c echo.Context) error {
	search := c.QueryParam("search")
	results := strings.Split(search, "")
	layout := components.Layout("go-website", components.Search(results))
	return components.Render(c, http.StatusOK, layout)
}
