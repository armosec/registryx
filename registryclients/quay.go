package registryclients

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/registries"
	dockerregistry "github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"slices"
	"sort"
	"strings"
)

type QuayRegistryClient struct {
	Registry *armotypes.QuayImageRegistry
}

func (q *QuayRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	iRegistry, err := registries.Factory(&authn.AuthConfig{Username: q.Registry.RobotAccountName, Password: q.Registry.RobotAccountToken}, q.Registry.ContainerRegistryName, nil)
	if err != nil {
		return nil, err
	}
	var repos, pageRepos []string
	var nextPage *common.PaginationOption

	firstPage := common.MakePagination(iRegistry.GetMaxPageSize())
	catalogOpts := common.CatalogOption{}

	for pageRepos, nextPage, err = iRegistry.Catalog(ctx, firstPage, catalogOpts, authn.FromConfig(*iRegistry.GetAuth())); ; pageRepos, nextPage, err = iRegistry.Catalog(ctx, *nextPage, catalogOpts, authn.FromConfig(*iRegistry.GetAuth())) {
		if err != nil {
			return nil, err
		}
		if len(pageRepos) == 0 {
			break
		}
		repos = append(repos, pageRepos...)

		if nextPage == nil || nextPage.Cursor == "" {
			break
		}
	}
	return repos, nil
}

func (q *QuayRegistryClient) GetImagesToScan(ctx context.Context) (map[string]string, error) {
	images := make(map[string]string, len(q.Registry.Repositories))
	for _, repository := range q.Registry.Repositories {
		tag, err := q.getImageLatestTag(ctx, repository)
		if err != nil {
			return nil, err
		}
		images[fmt.Sprintf("%s/%s", q.Registry.ContainerRegistryName, repository)] = tag
	}
	return images, nil
}

func (q *QuayRegistryClient) GetDockerAuth() *dockerregistry.AuthConfig {
	return &dockerregistry.AuthConfig{
		Username: q.Registry.RobotAccountName,
		Password: q.Registry.RobotAccountToken,
	}
}

func (q *QuayRegistryClient) getImageLatestTag(_ context.Context, repo string) (string, error) {
	iRegistry, err := registries.Factory(&authn.AuthConfig{Username: q.Registry.RobotAccountName, Password: q.Registry.RobotAccountToken}, q.Registry.ContainerRegistryName, nil)
	if err != nil {
		return "", err
	}

	firstPage := common.MakePagination(1000)
	var tags []string
	options := []remote.Option{remote.WithAuth(authn.FromConfig(*iRegistry.GetAuth()))}
	if latestTags, err := iRegistry.GetLatestTags(repo, 1, options...); err == nil {
		for _, tag := range latestTags {
			if strings.HasSuffix(tag, ".sig") {
				continue
			}
			tagsForDigest := strings.Split(tag, ",")
			return tagsForDigest[0], nil
		}
	} else {
		for tagsPage, nextPage, err := iRegistry.List(repo, firstPage, options...); ; tagsPage, nextPage, err = iRegistry.List(repo, *nextPage) {
			if err != nil {
				return "", err
			}

			if slices.Contains(tagsPage, "latest") {
				return "latest", nil
			}

			tags = append(tags, tagsPage...)

			if nextPage == nil {
				break
			}
		}
		return getLatestTag(tags), nil
	}
	return "", nil
}

func getLatestTag(tags []string) string {
	var versions []*semver.Version
	for _, tag := range tags {
		version, err := semver.NewVersion(tag)
		if err == nil {
			versions = append(versions, version)
		}
	}
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	return versions[0].String()
}
