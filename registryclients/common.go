package registryclients

import (
	"context"
	"slices"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	latestTag = "latest"
)

func getAllRepositories(ctx context.Context, registry interfaces.IRegistry) ([]string, error) {
	var repos, pageRepos []string
	var nextPage *common.PaginationOption
	var err error

	firstPage := common.MakePagination(registry.GetMaxPageSize())
	catalogOpts := common.CatalogOption{}

	for pageRepos, nextPage, err = registry.Catalog(ctx, firstPage, catalogOpts, authn.FromConfig(*registry.GetAuth())); ; pageRepos, nextPage, err = registry.Catalog(ctx, *nextPage, catalogOpts, authn.FromConfig(*registry.GetAuth())) {
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

func getImageLatestTag(repo string, registry interfaces.IRegistry) (string, error) {
	firstPage := common.MakePagination(1000)
	var tags []string
	withAuth := remote.WithAuth(authn.FromConfig(*registry.GetAuth()))
	if latestTags, err := registry.GetLatestTags(repo, 1, withAuth); err == nil {
		for _, tag := range latestTags {
			if strings.HasSuffix(tag, ".sig") {
				continue
			}
			tagsForDigest := strings.Split(tag, ",")
			return tagsForDigest[0], nil
		}
	} else {
		for tagsPage, nextPage, err := registry.List(repo, firstPage, withAuth); ; tagsPage, nextPage, err = registry.List(repo, *nextPage, withAuth) {
			if err != nil {
				return "", err
			}

			if slices.Contains(tagsPage, latestTag) {
				return latestTag, nil
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
	var nonSemverTags []string
	for _, tag := range tags {
		version, err := semver.NewVersion(tag)
		if err == nil {
			versions = append(versions, version)
		} else {
			nonSemverTags = append(nonSemverTags, tag)
		}
	}
	if len(versions) == 0 {
		if len(nonSemverTags) > 0 {
			// we assume freestyle tags are already sorted
			return nonSemverTags[len(nonSemverTags)-1]
		}
		return ""
	}
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	return versions[0].String()
}
