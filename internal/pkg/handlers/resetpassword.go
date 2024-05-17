package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/profile"
)

type resetPasswordHandler struct{}

func NewResetPasswordHandler() resetPasswordHandler {
	return resetPasswordHandler{}
}

func (h resetPasswordHandler) Handler(c echo.Context) error {
	profile, err := profile.MustGet(c)

	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	body := map[string]string{
		"client_id": os.Getenv("AUTH0_CLIENT_ID"), 
		"email": profile.Email, 
		"connection": os.Getenv("AUTH0_CONNECTION"),
	}

	jsonBody, err := json.Marshal(body)

	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	resp, err := http.Post(
		"https://"+os.Getenv("AUTH0_DOMAIN")+"/dbconnections/change_password",
		"application/json",
		bytes.NewReader(jsonBody),
	)

	if err != nil {
		return fmt.Errorf("failed to post request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to reset password: %d", resp.StatusCode)
	}

	return c.NoContent(http.StatusOK)
}
