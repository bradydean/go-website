package handlers

import (
	"fmt"
	"net/http"

	"github.com/bradydean/go-website/internal/pkg/components"
	"github.com/bradydean/go-website/internal/pkg/profile"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type profileHandler struct{}

func NewProfileHandler() profileHandler {
	return profileHandler{}
}

func (h profileHandler) Handler(c echo.Context) error {
	sess, err := session.Get("__session", c)

	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	profile := sess.Values[profile.ProfileKey{}].(profile.Profile)

	return components.Render(c, http.StatusOK, components.Layout("Profile", components.Profile(profile.Name, profile.Email)))
}
