package registries

import (
	"context"
	"fmt"
	"testing"

	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/google/go-containerregistry/pkg/name"
)

//NOT WORKING --- YET
/*
func TestDocker(t *testing.T) {
	registry, err := name.NewRegistry("", name.Insecure)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	reg, err := Factory(nil, &registry, nil)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	ctx := context.Background()
	res, _, err := reg.Catalog(ctx, common.NoPaginationOption(), common.CatalogOption{}, nil)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	t.Errorf("%v", res)
}
*/

//these tests needs a real harbor server
/*
var harborRegHost string = "myharbor1.org"

func TestHarborCommonUser(t *testing.T) {
	registry, err := name.NewRegistry(harborRegHost)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	reg, err := Factory(&authn.AuthConfig{Username: "auser", Password: "Abc12345"}, &registry,
		common.MakeRegistryOptions(true, false, false, "", "", "", common.Harbor))
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	testHarbor(reg, t, registry)
}

func TestHarborAdmin(t *testing.T) {
	registry, err := name.NewRegistry(harborRegHost)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	reg, err := Factory(&authn.AuthConfig{Username: "admin", Password: "Harbor12345"}, &registry,
		common.MakeRegistryOptions(true, false, false, "", "", "", common.Harbor))
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	testHarbor(reg, t, registry)
}

func TestHarborAdminProject(t *testing.T) {
	registry, err := name.NewRegistry(harborRegHost)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	reg, err := Factory(&authn.AuthConfig{Username: "admin", Password: "Harbor12345"}, &registry,
		common.MakeRegistryOptions(true, false, false, "", "", "user-project", common.Harbor))
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	testHarbor(reg, t, registry)
}
*/

func testHarbor(reg interfaces.IRegistry, t *testing.T, registry name.Registry) {
	ctx := context.Background()
	for repos, repoNextPage, err := reg.Catalog(ctx, common.MakePagination(1), common.CatalogOption{}, nil); err == nil; repos, repoNextPage, err = reg.Catalog(ctx, *repoNextPage, common.CatalogOption{}, nil) {
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		fmt.Printf("repos: %v\n", repos)
		for _, repoName := range repos {
			//TODO change interface to accept name and do this stuff inside the reg
			fmt.Printf("  Repo :%s\n", repoName)
			for tags, tagsNextPage, err := reg.List(repoName, common.MakePagination(1)); err == nil; tags, tagsNextPage, err = reg.List(repoName, *tagsNextPage) {
				if err != nil {
					t.Errorf("%s", err.Error())
				}
				fmt.Printf("    %s tags: %v\n", repoName, tags)
				if tagsNextPage == nil {
					break
				}
			}
		}
		if repoNextPage == nil {
			break
		}
	}
}
