package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type logoutHandler struct{}

func NewLogoutHandler() logoutHandler {
	return logoutHandler{}
}

func (h logoutHandler) Handler(c echo.Context) error {
	logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")

	if err != nil {
		return fmt.Errorf("failed to parse logout url: %w", err)
	}

	scheme := "http"

	if strings.HasPrefix(os.Getenv("AUTH0_CALLBACK_URL"), "https") {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + c.Request().Host)

	if err != nil {
		return fmt.Errorf("failed to parse returnTo url: %w", err)
	}

	params := logoutUrl.Query()
	params.Add("returnTo", returnTo.String())
	params.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	logoutUrl.RawQuery = params.Encode()

	sess, err := session.Get("__session", c)

	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	sess.Options.MaxAge = -1

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return c.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
}
