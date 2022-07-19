package defaultregistry

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/armosec/registryx/registries/dockerregistry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// this is just a wrapper around the go-container & remote catalog
type DefaultRegistry struct {
	Registry *name.Registry
	Auth     *authn.AuthConfig
	Cfg      *common.RegistryOptions
	This     interfaces.IRegistry
}

func NewRegistry(auth *authn.AuthConfig, registry *name.Registry, registryCfg *common.RegistryOptions) (interfaces.IRegistry, error) {
	if registry.Name() == "" {
		return nil, fmt.Errorf("must provide a non empty registry")
	}
	reg := &DefaultRegistry{Auth: auth, Registry: registry, Cfg: registryCfg}
	reg.This = reg
	return reg, nil

}

func (reg *DefaultRegistry) GetMaxPageSize() int {
	return 1000
}

func (reg *DefaultRegistry) GetAuth() *authn.AuthConfig {
	return reg.Auth
}
func (reg *DefaultRegistry) GetRegistry() *name.Registry {
	return reg.Registry
}

func (reg *DefaultRegistry) getURL(urlSuffix string) *url.URL {

	return &url.URL{
		Scheme: reg.Registry.Scheme(),
		Host:   reg.Registry.RegistryStr(),
		Path:   fmt.Sprintf("/v2/%s", urlSuffix),
	}
}

func (reg *DefaultRegistry) List(repoName string, pagination common.PaginationOption, options ...remote.Option) ([]string, *common.PaginationOption, error) {
	repoData, err := common.MakeRepoWithRegistry(repoName, reg.Registry)
	if err != nil {
		return nil, nil, err
	}
	tags, err := remote.List(*repoData, options...)
	//TODO handle pagination
	return tags, nil, err
}

func (reg *DefaultRegistry) Catalog(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption, authenticator authn.Authenticator) ([]string, *common.PaginationOption, error) {
	regis := reg.GetRegistry().Name()
	if regis == "index.docker.io" && authenticator == nil {
		token, err := dockerregistry.Token(reg.GetAuth(), reg.GetRegistry())
		if err != nil {
			return nil, nil, err
		}
		if reg.Auth == nil {
			reg.Auth = &authn.AuthConfig{}
		}
		reg.Auth.RegistryToken = token.Token
	}
	//auth part not working though w/o removing scope
	if err := common.ValidateAuth(reg.GetAuth()); err == nil {
		// res, err := remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authn.FromConfig(*reg.GetAuth())))
		if authenticator == nil {
			authenticator = authn.FromConfig(*reg.GetAuth())
		}
		res, err := remote.Catalog(ctx, *reg.GetRegistry(), remote.WithAuth(authenticator))
		if err != nil {
			return nil, nil, err
		}
		return res, nil, err
	}
	if regis == "index.docker.io" {
		repos, err := remote.Catalog(ctx, *reg.GetRegistry(), remote.WithAuth(authn.FromConfig(*reg.GetAuth())))
		return repos, nil, err
	}
	repos, err := remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authn.Anonymous))
	return repos, common.CalcNextV2Pagination(repos, pagination.Size), err
}

//GetLatestTags returns the latest tags for a given repository in descending order by the image creation time
//multiple tags on a single image will be sent as a comma separated string
//e.g ["latest,v3" ,"v2", "v1"]
func (reg *DefaultRegistry) GetLatestTags(repoName string, depth int, options ...remote.Option) ([]string, error) {

	tagsInfos := tagsInfo{}
	for tagsPage, nextPage, err := reg.This.List(repoName, common.MakePagination(100), options...); ; tagsPage, nextPage, err = reg.This.List(repoName, *nextPage, options...) {
		if err != nil {
			return nil, err
		}
		for _, tag := range tagsPage {
			imageName := fmt.Sprintf("%s/%s:%s", reg.Registry.Name(), repoName, tag)

			ref, err := name.ParseReference(imageName)
			if err != nil {
				return nil, err
			}

			image, err := remote.Image(ref, options...)
			if err != nil {
				return nil, err
			}
			imageDigest, err := image.Digest()
			if err != nil {
				return nil, err
			}
			digestString := imageDigest.String()
			if existingImage := tagsInfos.getByDigest(digestString); existingImage != nil {
				existingImage.tags = append(existingImage.tags, tag)
			} else {
				cf, err := image.ConfigFile()
				if err != nil {
					return nil, err
				}
				tagsInfos = append(tagsInfos, &tagInfo{tags: []string{tag}, created: cf.Created.Time, digest: digestString})

			}

		}

		sort.Slice(tagsInfos, func(i, j int) bool {
			return tagsInfos[i].created.After(tagsInfos[j].created)
		})
		if len(tagsInfos) > depth {
			tagsInfos = tagsInfos[:depth]
		}

		if nextPage == nil {
			break
		}
	}
	tags := []string{}
	for _, tagInfo := range tagsInfos {
		tags = append(tags, strings.Join(tagInfo.tags, ","))
	}
	return tags, nil

}

type tagInfo struct {
	tags    []string
	created time.Time
	digest  string
}
type tagsInfo []*tagInfo

func (ti tagsInfo) getByDigest(digest string) *tagInfo {
	for _, tagInfo := range ti {
		if tagInfo.digest == digest {
			return tagInfo
		}
	}
	return nil
}
