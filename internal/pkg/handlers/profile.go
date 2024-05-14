package handlers

import (
	"fmt"
	"net/http"

	"github.com/bradydean/go-website/internal/pkg/components"
	"github.com/bradydean/go-website/internal/pkg/profile"
	"github.com/labstack/echo/v4"
)

type profileHandler struct{}

func NewProfileHandler() profileHandler {
	return profileHandler{}
}

func (h profileHandler) Handler(c echo.Context) error {
	profile, err := profile.MustGet(c)

	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	return components.Render(c, http.StatusOK, components.Layout("Profile", &profile, components.Profile(profile)))
}
