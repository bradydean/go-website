package profile

import (
	"encoding/gob"
)

type ProfileKey struct{}

type Profile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func init() {
	gob.Register(Profile{})
	gob.Register(ProfileKey{})
}
