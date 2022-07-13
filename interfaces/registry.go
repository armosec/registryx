package interfaces

import (
	"context"

	"github.com/armosec/registryx/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type IRegistry interface {
	Catalog(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption, authenticator authn.Authenticator) (repositories []string, nextPage *common.PaginationOption, err error)
	List(repoName string, pagination common.PaginationOption, options ...remote.Option) (tags []string, nextPagination *common.PaginationOption, err error)
	GetAuth() *authn.AuthConfig
	GetRegistry() *name.Registry
	GetMaxPageSize() int
}
