package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"

	"github.com/bradydean/go-website/internal/pkg/authenticator"
)

type loginHandler struct {
	Authenticator *authenticator.Authenticator
}

func NewLoginHandler(authenticator *authenticator.Authenticator) loginHandler {
	return loginHandler{
		Authenticator: authenticator,
	}
}

func (h loginHandler) Handler(c echo.Context) error {
	state, err := authenticator.GenerateRandomState()

	if err != nil {
		return fmt.Errorf("failed to generate random state: %w", err)
	}

	sess, err := session.Get("__session", c)

	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(os.Getenv("APP_URL"), "https://"),
	}

	sess.Values["state"] = state

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	var additionalParams []oauth2.AuthCodeOption

	organization := c.QueryParam("organization")

	if organization != "" {
		additionalParams = append(additionalParams, oauth2.SetAuthURLParam("organization", organization))
	}

	invitation := c.QueryParam("invitation")

	if invitation != "" {
		additionalParams = append(additionalParams, oauth2.SetAuthURLParam("invitation", invitation))
	}

	returnTo := c.QueryParam("returnTo")

	if returnTo != "" && strings.HasPrefix(returnTo, os.Getenv("APP_URL")) {
		callback := os.Getenv("APP_URL") + "/callback?returnTo=" + returnTo
		additionalParams = append(additionalParams, oauth2.SetAuthURLParam("redirect_uri", callback))
	}

	return c.Redirect(http.StatusTemporaryRedirect, h.Authenticator.AuthCodeURL(state, additionalParams...))
}
