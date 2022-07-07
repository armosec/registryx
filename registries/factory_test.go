package registries

import (
	"context"
	"testing"

	"github.com/LiorAlafiArmo/registryx/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

func TestDocker(t *testing.T) {
	registry, err := name.NewRegistry("")
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	reg, err := Factory(&authn.AuthConfig{Username: "lioralafi1", Password: "something!wild"}, &registry)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	ctx := context.Background()
	res, err := reg.Catalog(ctx, common.NoPagination(0), common.CatalogOption{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	t.Errorf("%v", res)
}
