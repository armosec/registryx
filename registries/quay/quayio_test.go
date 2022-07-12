package quay

import (
	"context"
	"fmt"
	"testing"

	"github.com/armosec/registryx/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

//those are integration tests
// TODO: replace with asserts
func TestSimpleNoAuth(t *testing.T) {
	registry, err := name.NewRegistry("quay.io")
	if err != nil {
		t.Errorf("err1: %v", err.Error())
	}
	quayio, err := NewQuayIORegistry(nil, &registry, &common.RegistryOptions{})
	ctx := context.Background()
	repos, _, err := quayio.Catalog(ctx, common.NoPaginationOption(), common.CatalogOption{IsPublic: true, Namespaces: "quay"}, nil)
	if err != nil {
		t.Errorf("%v", err.Error())
	}

	if len(repos) == 0 {
		t.Error("expected more errors")
	}
	for _, repo := range repos {
		fullRepoName := quayio.GetRegistry().Name() + "/" + repo
		repo_data, err := name.NewRepository(fullRepoName)
		if err != nil {
			t.Error("basa")
		}
		imagestags, err := remote.List(repo_data, remote.WithAuth(authn.Anonymous))
		if err != nil {
			t.Error("basa2")
		}
		fmt.Printf("%s/%v\n", fullRepoName, imagestags)
	}
}

// func TestSimpleAuth(t *testing.T) {
// 	registry, err := name.NewRegistry("quay.io")
// 	if err != nil {
// 		t.Errorf("err1: %v", err.Error())
// 	}
// 	quayio, err := NewQuayIORegistry(&authn.AuthConfig{Username: "<user>", Password: "<secret>"}, &registry)
// 	ctx := context.Background()
// 	data, err := quayio.Catalog(ctx, common.NoPagination(0), common.CatalogOption{IsPublic: true, Namespaces: "armosec"})
// 	if err != nil {
// 		t.Errorf("%v", err)
// 		return
// 	}
// 	t.Errorf("%v", data)

// }
