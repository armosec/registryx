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
	"sync"
	"time"

	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	version "github.com/hashicorp/go-version"
	"k8s.io/utils/strings/slices"
)

type CatalogV2Response struct {
	Repositories []string `json:"repositories"`
}

// this is just a wrapper around the go-container & remote catalog
type DefaultRegistry struct {
	Registry   *name.Registry
	Auth       *authn.AuthConfig
	Cfg        *common.RegistryOptions
	This       interfaces.IRegistry
	HTTPClient *http.Client
}

func NewRegistry(auth *authn.AuthConfig, registry *name.Registry, registryCfg *common.RegistryOptions) (interfaces.IRegistry, error) {
	if registry.Name() == "" {
		return nil, fmt.Errorf("must provide a non empty registry")
	}
	reg := &DefaultRegistry{Auth: auth, Registry: registry, Cfg: registryCfg, HTTPClient: &http.Client{Timeout: time.Duration(150) * time.Second}}
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
		res, _, err := reg.CatalogPage(ctx, pagination, options, authenticator)
		if err != nil {
			return nil, nil, err
		}
		return res, nil, err
	}
	repos, err := remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authn.Anonymous))
	return repos, common.CalcNextV2Pagination(repos, pagination.Size), err
}

// Build http req and append password / token as bearer token
// See https://cloud.google.com/container-registry/docs/advanced-authentication#token
func (reg *DefaultRegistry) gcrCatalogPage(pagination common.PaginationOption, options common.CatalogOption) ([]string, *common.PaginationOption, error) {
	uri := reg.GetURL("_catalog")
	q := uri.Query()

	if pagination.Size > 0 {
		q.Add("n", fmt.Sprintf("%d", pagination.Size))
	}
	if pagination.Cursor != "" {
		q.Add("last", pagination.Cursor)
	}
	uri.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, uri.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	if pagination.Cursor != "" {
		url := uri.String()
		if strings.HasPrefix(url, "https://") {
			url = strings.ReplaceAll(url, "https://", "")
		} else {
			url = strings.ReplaceAll(url, "http://", "")
		}

		req.Header.Add("Link", fmt.Sprintf("<%s>; rel=next", url))
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", reg.GetAuth().Password))
	resp, err := reg.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()
	repos := &CatalogV2Response{}
	if err := json.NewDecoder(resp.Body).Decode(repos); err != nil {
		return nil, nil, err
	}
	pgn := &common.PaginationOption{Size: pagination.Size}
	if len(repos.Repositories) > 0 && pagination.Size > 0 {
		pgn.Cursor = repos.Repositories[len(repos.Repositories)-1]
	}

	return repos.Repositories, pgn, nil
}

func (reg *DefaultRegistry) CatalogPage(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption, authenticator authn.Authenticator) ([]string, *common.PaginationOption, error) {
	var repos []string
	var err error
	var pgn *common.PaginationOption
	switch provider := getRegistryProvider(reg.Registry.RegistryStr()); provider {
	case "gcr":
		repos, pgn, err = reg.gcrCatalogPage(pagination, options)
	default:
		repos, err = remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authenticator))
		pgn = common.CalcNextV2Pagination(repos, pagination.Size)
	}
	return repos, pgn, err

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
	type imageInfo struct {
		created time.Time
		digest  string
		tag     string
		err     error
	}
	tagsInfos := tagsInfo{}
	wg := sync.WaitGroup{}
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	for tagsPage, nextPage, err := reg.This.List(repoName, common.MakePagination(reg.This.GetMaxPageSize()), options...); ; tagsPage, nextPage, err = reg.This.List(repoName, *nextPage, options...) {
		if err != nil {
			return nil, err
		}
		//if depth is one (default) and latest tag found no need to continue
		if depth == 1 && slices.Contains(tagsPage, "latest") {
			return []string{"latest"}, nil
		}
		ch := make(chan imageInfo, len(tagsPage))
		for _, tag := range tagsPage {
			wg.Add(1)
			go func(ch chan<- imageInfo, tag string, wg *sync.WaitGroup) {
				defer wg.Done()
				imageName := fmt.Sprintf("%s/%s:%s", reg.Registry.Name(), repoName, tag)
				digest, created, err := reg.getImageDigestAndCreationTime(imageName, options...)
				select {
				case <-ctx.Done():
					return
				case ch <- imageInfo{created: created, digest: digest, tag: tag, err: err}:
					return
				}
			}(ch, tag, &wg)
		}

		go func(ch chan imageInfo, wg *sync.WaitGroup) {
			wg.Wait()
			close(ch)
		}(ch, &wg)

		for info := range ch {
			if info.err != nil {
				return nil, info.err
			}
			//check if the image already collected with different tag
			if existingImage := tagsInfos.getByDigest(info.digest); existingImage != nil {
				existingImage.tags = append(existingImage.tags, info.tag)
			} else { //new image add it with the tag
				tagsInfos = append(tagsInfos, &tagInfo{tags: []string{info.tag}, created: info.created, digest: info.digest})
			}
		}
		//sort the list by creation time
		sort.Slice(tagsInfos, func(i, j int) bool {
			return tagsInfos[i].created.After(tagsInfos[j].created)
		})
		//cut off the list tail if we have reached the depth
		if len(tagsInfos) > depth {
			tagsInfos = tagsInfos[:depth]
		}
		//if no next page then we are done
		if nextPage == nil {
			break
		}
	}

	//sort multiple tags on a single image by version if possible otherwise sort by sematic version otherwise sort alphabetically
	for _, tagInfo := range tagsInfos {
		if len(tagInfo.tags) == 1 {
			continue
		}
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
	tags := []string{}
	for _, tagInfo := range tagsInfos {
		tags = append(tags, strings.Join(tagInfo.tags, ","))
	}
	return tags, nil

}

func (reg *DefaultRegistry) getImageDigestAndCreationTime(imageName string, options ...remote.Option) (string, time.Time, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return "", time.Time{}, err
	}
	desc, err := remote.Get(ref, options...)
	if err != nil {
		return "", time.Time{}, err
	}

	//first try to covert to image - this works only for schema v2
	if image, err := desc.Image(); err == nil {
		imageDigest, err := image.Digest()
		if err != nil {
			return "", time.Time{}, err
		}
		digestString := imageDigest.String()
		cf, err := image.ConfigFile()
		if err != nil {
			return "", time.Time{}, err
		}
		return digestString, cf.Created.Time, nil
	} else if strings.Contains(err.Error(), "unsupported MediaType") && (desc.MediaType == types.DockerManifestSchema1 || desc.MediaType == types.DockerManifestSchema1Signed) {
		//v1 schema need to parse the manifest
		rawManifest, err := desc.RawManifest()
		if err != nil {
			return "", time.Time{}, err
		}
		manifest, err := decodeV1Mafinset(rawManifest)
		if err != nil {
			return "", time.Time{}, err
		}
		if len(manifest.history) == 0 {
			return "", time.Time{}, fmt.Errorf("failed to parse v1 manifest history")
		}
		return desc.Digest.String(), manifest.history[0].v1Compatibility.created, nil
	} else {
		//unknown schema
		return "", time.Time{}, err
	}
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

func getRegistryProvider(registryName string) string {
	if strings.Contains(registryName, ".dkr.ecr") {
		return "ecr"
	}
	if strings.Contains(registryName, "gcr.io") {
		return "gcr"
	}
	return ""
}
