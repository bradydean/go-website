package session

import (
	"encoding/base64"
	"errors"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func New() (echo.MiddlewareFunc, error) {
	sessionKey := os.Getenv("SESSION_KEY")

	if sessionKey == "" {
		return nil, errors.New("SESSION_KEY is required")
	}

	decodedKey, err := base64.StdEncoding.DecodeString(sessionKey)

	if err != nil {
		return nil, err
	}

	return session.Middleware(sessions.NewCookieStore(decodedKey)), nil
}
