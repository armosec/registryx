package quay

import (
	"context"
	"fmt"
	"testing"

	"github.com/armosec/registryx/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/assert"
)

//those are integration tests
func TestSimpleNoAuth(t *testing.T) {
	registry, err := name.NewRegistry("quay.io")
	assert.Nil(t, err, "failed to create registry")

	quayio, err := NewQuayIORegistry(nil, &registry, &common.RegistryOptions{})
	ctx := context.Background()
	repos, _, err := quayio.Catalog(ctx, common.NoPaginationOption(), common.CatalogOption{IsPublic: true, Namespaces: "quay"}, nil)
	assert.Nil(t, err, "failed to catalog")
	assert.NotEmpty(t, repos, "expected some returned images")

	repo := repos[0]
	fullRepoName := quayio.GetRegistry().Name() + "/" + repo
	repo_data, err := name.NewRepository(fullRepoName)
	assert.Nil(t, err, "failed to create repo data")

	imagestags, err := remote.List(repo_data, remote.WithAuth(authn.Anonymous))
	assert.Nil(t, err, "failed to list image tags")
	assert.NotEmpty(t, imagestags, fmt.Errorf("expected tags for %s", fullRepoName))

	latestTags, err := quayio.GetLatestTags(repo, 5, remote.WithAuth(authn.Anonymous))
	assert.Nil(t, err, "failed to get latest image tags")
	assert.NotEmpty(t, latestTags, fmt.Errorf("expected tags for %s", fullRepoName))
}

// func TestSimpleAuth(t *testing.T) {
// 	registry, err := name.NewRegistry("quay.io")
// 	assert.Nil(t, err, "")

// 	quayio, err := NewQuayIORegistry(&authn.AuthConfig{Username: "", Password: ""}, &registry, common.MakeRegistryOptions(false, false, false, "quay.io", "", "", common.Quay))
// 	ctx := context.Background()
// 	data, _, err := quayio.Catalog(ctx, common.MakePagination(quayio.GetMaxPageSize()), common.CatalogOption{Namespaces: "armosec"}, nil)
// 	assert.Nil(t, err, "")
// 	//TODO: fix pagination: https://docs.docker.com/registry/spec/api/
// 	// for pgn != nil {
// 	// 	data, pgn, err = quayio.Catalog(ctx, *pgn, common.CatalogOption{Namespaces: "armosec"}, nil)
// 	// 	assert.Nil(t, err, "")
// 	// 	all = append(all, data...)

// 	// }
// 	assert.NotEqual(t, len(data), 0, "received an empty results")
// 	assert.Nil(t, err, "failed to create repo data")
// 	fullRepoName := quayio.GetRegistry().Name() + "/" + data[0]
// 	tags, _, err := quayio.List(fullRepoName, common.NoPaginationOption(), remote.WithAuth(authn.FromConfig(*quayio.GetAuth())))
// 	t.Errorf("%v", tags)
// }
