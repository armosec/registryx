package quay

import (
	"context"
	"testing"

	"github.com/LiorAlafiArmo/registryx/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

func TestSimpleNoAuth(t *testing.T) {
	registry, err := name.NewRegistry("quay.io")
	if err != nil {
		t.Errorf("err1: %v", err.Error())
	}
	quayio, err := NewQuayIORegistry(nil, &registry)
	ctx := context.Background()
	repos, err := quayio.Catalog(ctx, common.NoPagination(0), common.CatalogOption{IsPublic: true})
	if err != nil {
		t.Errorf("%v", err.Error())
	}

	t.Errorf("%v", repos)

}

func TestSimpleAuth(t *testing.T) {
	registry, err := name.NewRegistry("quay.io")
	if err != nil {
		t.Errorf("err1: %v", err.Error())
	}
	quayio, err := NewQuayIORegistry(&authn.AuthConfig{Username: "<user>", Password: "<secret>"}, &registry)
	ctx := context.Background()
	data, err := quayio.Catalog(ctx, common.NoPagination(0), common.CatalogOption{IsPublic: true, Namespaces: "armosec"})
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	t.Errorf("%v", data)

}
