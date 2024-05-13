package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/components"
	"github.com/bradydean/go-website/internal/pkg/profile"
)

type indexHandler struct{}

func NewIndexHandler() indexHandler {
	return indexHandler{}
}

func (h indexHandler) Handler(c echo.Context) error {
	session, err := session.Get("__session", c)

	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if profile, ok := session.Values[profile.ProfileKey{}].(profile.Profile); ok {
		layout := components.Layout("go-website", components.Index(&profile))
		return components.Render(c, http.StatusOK, layout)
	}

	layout := components.Layout("go-website", components.Index(nil))
	return components.Render(c, http.StatusOK, layout)
}
