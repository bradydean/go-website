package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/joho/godotenv/autoload"

	"github.com/bradydean/go-website/internal/pkg/authenticator"
	"github.com/bradydean/go-website/internal/pkg/handlers"
	"github.com/bradydean/go-website/internal/pkg/middleware/authentication"
	"github.com/bradydean/go-website/internal/pkg/middleware/session"
)

//go:generate go run github.com/a-h/templ/cmd/templ generate
//go:generate npx --yes tailwindcss@latest -i ./global.css -o ./static/tailwind.css --minify

func main() {
	e := echo.New()

	gob.Register(map[string]interface{}{})

	auth, err := authenticator.New()

	if err != nil {
		e.Logger.Fatal(fmt.Errorf("failed to create oauth2 authenticator: %w", err))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				msg := fmt.Sprintf("uri=%s status=%d", v.URI, v.Status)
				logger.LogAttrs(context.Background(), slog.LevelInfo, msg,
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				msg := fmt.Sprintf("uri=%s status=%d error=%s", v.URI, v.Status, v.Error)
				logger.LogAttrs(context.Background(), slog.LevelError, msg,
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	session, err := session.New()

	if err != nil {
		e.Logger.Fatal(fmt.Errorf("failed to create session middleware: %w", err))
	}

	e.Use(session)

	e.Group("/static").Use(middleware.Static("static"))

	e.GET("/login", handlers.NewLoginHandler(auth).Handler)
	e.GET("/logout", handlers.NewLogoutHandler().Handler)
	e.GET("/callback", handlers.NewCallbackHandler(auth).Handler)
	e.GET("/", handlers.NewIndexHandler().Handler)
	e.GET("/profile", handlers.NewProfileHandler().Handler, authentication.IsAuthenticated)

	if err := e.Start(":8000"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
