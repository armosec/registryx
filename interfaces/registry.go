package interfaces

import (
	"context"

	"github.com/LiorAlafiArmo/registryx/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

type IRegistry interface {
	Catalog(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption, authenticator authn.Authenticator) ([]string, error)
	// List()
	// Tags()
	GetAuth() *authn.AuthConfig
	GetRegistry() *name.Registry
}
