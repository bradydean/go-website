package profile

import (
	"encoding/gob"
	"errors"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type ProfileKey struct{}

type Profile struct {
	UserID string `json:"sub"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

func Get(c echo.Context) (*Profile, error) {
	session, err := session.Get("__session", c)

	if err != nil {
		return nil, err
	}

	if profile, ok := session.Values[ProfileKey{}].(Profile); ok {
		return &profile, nil
	}

	return nil, nil
}

func MustGet(c echo.Context) (Profile, error) {
	profile, err := Get(c)

	if err != nil {
		return Profile{}, err
	}

	if profile == nil {
		panic(errors.New("tried to access profile without being logged in"))
	}

	return *profile, nil
}

func init() {
	gob.Register(Profile{})
	gob.Register(ProfileKey{})
}
