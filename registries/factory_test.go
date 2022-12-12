package registries

import (
	"context"
	"fmt"
	"testing"

	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/assert"
)

// //NOT WORKING --- YET

// func TestDocker(t *testing.T) {

// 	reg, err := Factory(nil, "", common.MakeRegistryOptions(false, false, false, "", "", "", "docker"))
// 	if err != nil {
// 		t.Errorf("%s", err.Error())
// 	}

// 	ctx := context.Background()
// 	res, _, err := reg.Catalog(ctx, common.NoPaginationOption(), common.CatalogOption{}, nil)
// 	if err != nil {
// 		t.Errorf("%s", err.Error())
// 	}
// 	t.Errorf("%v", res)
// }

//these tests needs a real harbor server
/*
var harborRegHost string = "myharbor1.org"

func TestHarborCommonUser(t *testing.T) {
	reg, err := Factory(&authn.AuthConfig{Username: "auser", Password: "Abc12345"}, harborRegHost,
		common.MakeRegistryOptions(true, false, false, "", "", "", common.Harbor))
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	testHarbor(reg, t)
}

func TestHarborAdmin(t *testing.T) {
	reg, err := Factory(&authn.AuthConfig{Username: "admin", Password: "Harbor12345"}, harborRegHost,
		common.MakeRegistryOptions(true, false, false, "", "", "", common.Harbor))
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	testHarbor(reg, t)
}

func TestHarborAdminProject(t *testing.T) {
	reg, err := Factory(&authn.AuthConfig{Username: "admin", Password: "Harbor12345"}, harborRegHost,
		common.MakeRegistryOptions(true, false, false, "", "", "user-project", common.Harbor))
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	testHarbor(reg, t)
}
*/
func testHarbor(reg interfaces.IRegistry, t *testing.T) {
	ctx := context.Background()
	for repos, repoNextPage, _, err := reg.Catalog(ctx, common.MakePagination(1), common.CatalogOption{}, nil); err == nil; repos, repoNextPage, _, err = reg.Catalog(ctx, *repoNextPage, common.CatalogOption{}, nil) {
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		fmt.Printf("repos: %v\n", repos)
		for _, repoName := range repos {
			tags, err := reg.GetLatestTags(repoName, 10, remote.WithAuth(authn.FromConfig(*reg.GetAuth())))
			if err != nil {
				t.Errorf("%s", err.Error())
			}
			fmt.Printf("    %s latest tags: %v\n", repoName, tags)

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

func TestManyTags(t *testing.T) {
	reg, err := Factory(nil, "quay.io", common.MakeRegistryOptions(false, false, false, "", "", "armosec", common.Quay))
	if err != nil {
		t.Errorf(" failed to create registry %s", err.Error())
	}
	tags, err := reg.GetLatestTags("armosec/kubescape", 50, remote.WithAuth(authn.Anonymous))
	if err != nil {
		t.Errorf("failed to get latest tags %s", err.Error())
	}
	assert.Equal(t, 50, len(tags))
}

func TestManyTagsWithOneDepth(t *testing.T) {
	reg, err := Factory(nil, "quay.io", common.MakeRegistryOptions(false, false, false, "", "", "armosec", common.Quay))
	if err != nil {
		t.Errorf(" failed to create registry %s", err.Error())
	}
	tags, err := reg.GetLatestTags("armosec/kubescape", 1, remote.WithAuth(authn.Anonymous))
	if err != nil {
		t.Errorf("failed to get latest tags %s", err.Error())
	}
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "latest", tags[0])
}
