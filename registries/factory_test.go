package registries

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

/*
//these tests needs a real harbor server
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
func testHarbor(reg interfaces.IRegistry, t *testing.T, registry name.Registry) {
	ctx := context.Background()
	for repos, nextPage, err := reg.Catalog(ctx, common.MakePagination(1), common.CatalogOption{}, nil); nextPage != nil; repos, nextPage, err = reg.Catalog(ctx, *nextPage, common.CatalogOption{}, nil) {
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		fmt.Printf("repos: %v\n", repos)

		for _, repoName := range repos {
            //TODO change interface to accept name and do this stuff inside the reg
			repo, err := name.NewRepository(repoName)
			if err != nil {
				t.Errorf("%s\n", err.Error())
			}
			repo.Registry = registry

			fmt.Printf("  Repo :%s\n", repoName)
			for tags, nextPage, err := reg.List(repo, common.MakePagination(1)); nextPage != nil; tags, nextPage, err = reg.List(repo, *nextPage) {
				if err != nil {
					t.Errorf("%s", err.Error())
				}
				fmt.Printf("    %s tags: %v\n", repoName, tags)

			}

		}

	}
}
*/
