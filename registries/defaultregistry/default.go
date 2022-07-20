package defaultregistry

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	version "github.com/hashicorp/go-version"
)

type CatalogV2Response struct {
	Repositories []string `json:"repositories"`
}

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

func (reg *DefaultRegistry) GetURL(urlSuffix string) *url.URL {

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

// this is the default catalog implementation uses remote(for now)
func (reg *DefaultRegistry) Catalog(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption, authenticator authn.Authenticator) ([]string, *common.PaginationOption, error) {

	if err := common.ValidateAuth(reg.GetAuth()); err == nil {
		if authenticator == nil {
			authenticator = authn.FromConfig(*reg.GetAuth())
		}
		res, err := remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authn.FromConfig(*reg.GetAuth())))
		if err != nil {
			return nil, nil, err
		}
		return res, nil, err
	}
	repos, err := remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authn.Anonymous))
	return repos, common.CalcNextV2Pagination(repos, pagination.Size), err
}

func (reg *DefaultRegistry) GetV2Token(client *http.Client, url string) (*common.V2TokenResponse, error) {
	if reg.GetAuth() == nil {
		return nil, fmt.Errorf("no authorization found")
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if reg.GetAuth().Username != "" && reg.GetAuth().Password != "" {
		usrpwd := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", reg.GetAuth().Username, reg.GetAuth().Password)))
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", usrpwd))
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	token := &common.V2TokenResponse{}

	if err := json.NewDecoder(resp.Body).Decode(token); err != nil {
		return nil, err
	}

	if token.Token == "" {
		return nil, fmt.Errorf("recieved an empty token")
	}
	return token, nil
}

//GetLatestTags returns the latest tags for a given repository in descending order by the image creation time
//multiple tags on a single image will be sent as a comma separated string
//e.g ["latest,v3" ,"v2", "v1"]
func (reg *DefaultRegistry) GetLatestTags(repoName string, depth int, options ...remote.Option) ([]string, error) {

	tagsInfos := tagsInfo{}
	for tagsPage, nextPage, err := reg.This.List(repoName, common.MakePagination(reg.This.GetMaxPageSize()), options...); ; tagsPage, nextPage, err = reg.This.List(repoName, *nextPage, options...) {
		if err != nil {
			return nil, err
		}
		for _, tag := range tagsPage {
			imageName := fmt.Sprintf("%s/%s:%s", reg.Registry.Name(), repoName, tag)
			//get image descriptor for each tag
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

			//check for existing tags for this image
			if existingImage := tagsInfos.getByDigest(digestString); existingImage != nil {
				existingImage.tags = append(existingImage.tags, tag)
			} else { //not tags info for this image so create one
				cf, err := image.ConfigFile()
				if err != nil {
					return nil, err
				}
				tagsInfos = append(tagsInfos, &tagInfo{tags: []string{tag}, created: cf.Created.Time, digest: digestString})
			}
		}

		//cut off the list if we have reached the depth
		if len(tagsInfos) > depth {
			tagsInfos = tagsInfos[:depth]
		}
		//sort the list by creation time
		sort.Slice(tagsInfos, func(i, j int) bool {
			return tagsInfos[i].created.After(tagsInfos[j].created)
		})

		//sort multiple tags on a single image by version if possible otherwise sort by alphabetical order
		for _, tagInfo := range tagsInfos {
			sort.Slice(tagInfo.tags, func(i, j int) bool {
				//latest always comes first
				if tagInfo.tags[i] == "latest" {
					return true
				}
				if tagInfo.tags[j] == "latest" {
					return false
				}
				//try to parse the version
				vi, erri := version.NewVersion(tagInfo.tags[i])
				vj, errj := version.NewVersion(tagInfo.tags[j])
				if erri == nil && errj == nil {
					return vj.LessThan(vi)
				}
				if erri != nil && errj != nil {
					//no version so sort alphabetically
					return strings.ToLower(tagInfo.tags[i]) > strings.ToLower(tagInfo.tags[j])
				}
				//advance versions over non-versions
				if erri == nil {
					return true
				}
				return false
			})
		}

		//if no next page then we are done
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
