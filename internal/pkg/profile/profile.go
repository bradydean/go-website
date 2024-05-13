package profile

type ProfileKey struct{}

type Profile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
