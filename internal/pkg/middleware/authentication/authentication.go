package authentication

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/profile"
)

func IsAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("__session", c)

		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}

		if sess.Values[profile.ProfileKey{}] == nil {
			returnTo, err := url.Parse(os.Getenv("APP_URL") + "/")

			if err != nil {
				return fmt.Errorf("failed to parse returnTo url: %w", err)
			}

			returnTo = returnTo.JoinPath(c.Request().URL.Path)

			return c.Redirect(http.StatusTemporaryRedirect, "/login?returnTo="+returnTo.String())
		}

		return next(c)
	}
}
