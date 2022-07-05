package quay

import (
	"context"
	"testing"

	"github.com/LiorAlafiArmo/registryx/common"
	"github.com/google/go-containerregistry/pkg/name"
)

func TestSimpleNoAuth(t *testing.T) {
	registry, err := name.NewRegistry("quay.io")
	if err != nil {
		t.Errorf("err1: %v", err.Error())
	}
	quayio, err := NewQuayIORegistry(nil, &registry)
	ctx := context.Background()
	quayio.Catalog(ctx, common.NoPagination(0), common.CatalogOption{IsPublic: true, Namespaces: "armosec"})
	t.Error("1")

}
