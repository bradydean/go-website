package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/components"
	"github.com/bradydean/go-website/internal/pkg/profile"
)

type indexHandler struct{}

func NewIndexHandler() indexHandler {
	return indexHandler{}
}

func (h indexHandler) Handler(c echo.Context) error {
	profile, err := profile.Get(c)

	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if c.Request().Header.Get("HX-Boosted") != "" {
		return components.Render(c, http.StatusOK, components.Boost("Todo Lists", components.Index(profile)))
	}

	layout := components.Layout("Todo Lists", c.Get("csrf").(string), components.Index(profile))
	return components.Render(c, http.StatusOK, layout)
}
