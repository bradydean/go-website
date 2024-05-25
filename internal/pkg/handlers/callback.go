package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/authenticator"
	"github.com/bradydean/go-website/internal/pkg/profile"
	"github.com/labstack/echo-contrib/session"
)

type callbackHandler struct {
	Authenticator *authenticator.Authenticator
}

func NewCallbackHandler(authenticator *authenticator.Authenticator) callbackHandler {
	return callbackHandler{
		Authenticator: authenticator,
	}
}

func (h callbackHandler) Handler(c echo.Context) error {
	sess, err := session.Get("__session", c)

	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	queryState := c.QueryParam("state")

	if sessionState, ok := sess.Values["state"].(string); !ok || queryState != sessionState {
		return c.String(http.StatusBadRequest, "Invalid state parameter.")
	}

	token, err := h.Authenticator.Exchange(c.Request().Context(), c.QueryParam("code"))

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	idToken, err := h.Authenticator.VerifyIDToken(c.Request().Context(), token)

	if err != nil {
		return fmt.Errorf("failed to verify id token: %w", err)
	}

	var claims profile.Profile

	if err := idToken.Claims(&claims); err != nil {
		return fmt.Errorf("failed to parse id token claims: %w", err)
	}

	sess.Values[profile.ProfileKey{}] = claims
	sess.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(os.Getenv("APP_URL"), "https://"),
	}

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	returnTo := c.QueryParam("returnTo")

	if returnTo != "" && strings.HasPrefix(returnTo, os.Getenv("APP_URL")) {
		return c.Redirect(http.StatusTemporaryRedirect, returnTo)
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
