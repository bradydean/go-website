package main

import (
	"fmt"
	"net/http"
	"strings"
	"os"
	"encoding/base64"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/gorilla/sessions"
    "github.com/labstack/echo-contrib/session"
	"golang.org/x/oauth2"

	_ "github.com/joho/godotenv/autoload"

	"github.com/bradydean/go-website/internal/pkg/components"
	"github.com/bradydean/go-website/internal/pkg/middleware/auth"
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

	authenticator, err := auth.NewAuthenticator()

	if err != nil {
		e.Logger.Fatal(fmt.Errorf("failed to create oauth2 authenticator: %w", err))
	}

	{
		sessionKey := os.Getenv("SESSION_KEY")

		if sessionKey == "" {
			e.Logger.Fatal("SESSION_KEY environment variable must be set")
		}

		decodedKey, err := base64.StdEncoding.DecodeString(sessionKey)

		if err != nil {
			e.Logger.Fatal(fmt.Errorf("failed to decode session key: %w", err))
		}

		e.Use(session.Middleware(sessions.NewCookieStore(decodedKey)))
	}

	e.Group("/static").Use(middleware.Static("static"))

	e.GET("/login", func(c echo.Context) error {
		state, err := auth.GenerateRandomState()

		if err != nil {
			return fmt.Errorf("failed to generate random state: %w", err)
		}

		sess, err := session.Get("session", c)

		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}

		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
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

		return c.Redirect(http.StatusTemporaryRedirect, authenticator.AuthCodeURL(state, additionalParams...))
	})

	e.GET("/callback", func(c echo.Context) error {
		sess, err := session.Get("session", c)

		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
		
		queryState := c.QueryParam("state")

		if sessionState, ok := sess.Values["state"].(string); !ok || queryState != sessionState {
			return c.String(http.StatusBadRequest, "Invalid state parameter.")
		}

		token, err := authenticator.Exchange(c.Request().Context(), c.QueryParam("code"))

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		idToken, err := authenticator.VerifyIDToken(c.Request().Context(), token)

		if err != nil {
			return fmt.Errorf("failed to verify id token: %w", err)
		}

		var profile map[string]interface{}

		if err := idToken.Claims(&profile); err != nil {
			return fmt.Errorf("failed to parse id token claims: %w", err)
		}

		sess.Values["access_token"] = token.AccessToken
		sess.Values["name"] = profile["name"]
		sess.Values["email"] = profile["email"]

		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		return c.Redirect(http.StatusTemporaryRedirect, "/")
	})

	e.GET("/", func(c echo.Context) error {
		sess, err := session.Get("session", c)

		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}

		name := "Guest"

		if n, ok := sess.Values["name"].(string); ok {
			name = n
		}

		search := c.QueryParam("search")
		results := strings.Split(search, "")
		layout := components.Layout("go-website", components.Search(name, results))
		return Component(c, http.StatusOK, layout)
	})

	if err := e.Start(":8000"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
