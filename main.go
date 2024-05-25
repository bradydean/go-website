package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
//go:generate go run ./internal/scripts/jet/main.go

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	auth, err := authenticator.New()

	if err != nil {
		e.Logger.Fatalf("failed to create oauth2 authenticator: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogMethod:   true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			var level slog.Level

			if v.Status >= 500 {
				level = slog.LevelError
			} else if v.Status >= 400 {
				level = slog.LevelWarn
			} else {
				level = slog.LevelInfo
			}

			if v.Error == nil {
				msg := fmt.Sprintf("method=%s uri=%s status=%d", v.Method, v.URI, v.Status)
				logger.LogAttrs(context.Background(), level, msg,
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				msg := fmt.Sprintf("method=%s uri=%s status=%d err=%v", v.Method, v.URI, v.Status, v.Error)
				logger.LogAttrs(context.Background(), level, msg,
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}

			return nil
		},
	}))

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			msg := fmt.Sprintf("err=%v", err)
			logger.LogAttrs(context.Background(), slog.LevelError, msg,
				slog.String("err", err.Error()),
				slog.String("stack", string(stack)),
			)
			return err
		},
	}))

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		CookieSecure:   strings.HasPrefix(os.Getenv("APP_URL"), "https://"),
		CookieHTTPOnly: true,
	}))

	session, err := session.New()

	if err != nil {
		e.Logger.Fatalf("failed to create session middleware: %w", err)
	}

	e.Use(session)

	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		e.Logger.Fatalf("unable to connect to database: %w\n", err)
	}

	defer db.Close()

	e.Group("/static").Use(middleware.Static("static"))

	e.GET("/login", handlers.NewLoginHandler(auth).Handler)
	e.GET("/logout", handlers.NewLogoutHandler().Handler)
	e.GET("/callback", handlers.NewCallbackHandler(auth).Handler)
	e.GET("/", handlers.NewIndexHandler().Handler)
	e.GET("/profile", handlers.NewProfileHandler().Handler, authentication.IsAuthenticated)
	e.POST("/profile/reset-password", handlers.NewResetPasswordHandler().Handler, authentication.IsAuthenticated)
	e.GET("/lists", handlers.NewListsHandler(db).Handler, authentication.IsAuthenticated)
	e.POST("/lists", handlers.NewNewListHandler(db).Handler, authentication.IsAuthenticated)
	e.GET("/lists/:list_id", handlers.NewItemsHandler(db).Handler, authentication.IsAuthenticated)
	e.POST("/lists/:list_id/items", handlers.NewNewItemHandler(db).Handler, authentication.IsAuthenticated)
	e.DELETE("/lists/:list_id", handlers.NewDeleteListHandler(db).Handler, authentication.IsAuthenticated)
	e.DELETE("/lists/:list_id/items/:item_id", handlers.NewDeleteItemHandler(db).Handler, authentication.IsAuthenticated)
	e.PATCH("/lists/:list_id/items/:item_id", handlers.NewPatchItemHandler(db).Handler, authentication.IsAuthenticated)

	go func() {
		port := os.Getenv("PORT")

		if port == "" {
			port = "8000"
		}

		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("shutting down the server: %w", err)
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatalf("error shutting down the server: %w", err)
	}
}
