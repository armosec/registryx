package common

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
)

type V2TokenResponse struct {
	Token string `json:"token"`
}

func ValidateAuth(auth *authn.AuthConfig) error {
	if auth == nil {
		return fmt.Errorf("no auth")
	}

	if auth.Username == "" && auth.Password == "" && auth.Auth == "" && auth.IdentityToken == "" && auth.RegistryToken == "" {
		return fmt.Errorf("empty auth")
	}
	return nil
}
