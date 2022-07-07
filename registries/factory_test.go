package registries

import (
	"context"
	"testing"

	"github.com/armosec/registryx/common"
	"github.com/google/go-containerregistry/pkg/name"
)

//NOT WORKING --- YET
func TestDocker(t *testing.T) {
	registry, err := name.NewRegistry("")
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	reg, err := Factory(nil, &registry, nil)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	ctx := context.Background()
	res, err := reg.Catalog(ctx, common.NoPagination(0), common.CatalogOption{}, nil)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	t.Errorf("%v", res)
}
