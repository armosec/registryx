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
